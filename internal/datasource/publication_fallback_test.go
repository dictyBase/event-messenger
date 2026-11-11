package datasource

import (
	"errors"
	"fmt"
	"testing"
	"time"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	L "github.com/IBM/fp-go/v2/optics/lens"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictybase/literature"
	"github.com/stretchr/testify/require"
)

const preprintDOI = "10.1101/582072"

var testEuropeLens = L.MakeLensRef(
	func(p *Publication) europePMCClient { return p.europePMC },
	func(p *Publication, c europePMCClient) *Publication { p.europePMC = c; return p },
)

type fakeEuropePMC struct {
	article *literature.EuropePMCArticle
	err     error
	calls   []string
}

func (f *fakeEuropePMC) GetArticle(id string) (*literature.EuropePMCArticle, error) {
	f.calls = append(f.calls, "pmid:"+id)
	return f.article, f.err
}

func (f *fakeEuropePMC) GetArticleByDOI(id string) (*literature.EuropePMCArticle, error) {
	f.calls = append(f.calls, "doi:"+id)
	return f.article, f.err
}

func mustPubInfo(t *testing.T, p *Publication, id string) *PubInfo {
	t.Helper()

	return F.Pipe3(id, p.ParsedInfo, ioeutils.ToEither[error, *PubInfo], E.Fold(
		func(err error) *PubInfo { t.Fatalf("publication lookup: %v", err); return nil },
		F.Identity[*PubInfo],
	))
}

func TestParsedInfoEuropeFallback(t *testing.T) {
	t.Parallel()

	for _, id := range []string{"doi:" + preprintDOI, testPMID} {
		for _, networkFailure := range []bool{false, true} {
			name := fmt.Sprintf("%s/network=%t", id, networkFailure)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				ncbi := newFakePubMedWithError(errors.New("NCBI unavailable"))

				ncbi.searchResult = new(literature.SearchResult)
				if networkFailure {
					ncbi.searchErr = errors.New("search unavailable")
				}

				europe := new(fakeEuropePMC)
				europe.article = new(literature.EuropePMCArticle)
				europe.article.DOI = preprintDOI
				europe.article.PubYear = "2019"
				europe.article.Authors = []literature.EuropePMCAuthor{
					F.Pipe1(literature.EuropePMCAuthor{}, europeAuthorFullLens.Set("Smith A")),
				}
				p := F.Pipe2(ncbi, newPublication, testEuropeLens.Set(europe))
				info := mustPubInfo(t, p, id)
				require.Equal(t, "Smith A (2019)", info.AuthorStr)
				require.Equal(t, "https://doi.org/10.1101/582072", info.DoiURL)
				require.Empty(t, info.PubmedURL)

				want := "doi:" + preprintDOI
				if id == testPMID {
					want = "pmid:" + testPMID
				}

				require.Equal(t, []string{want}, europe.calls)
			})
		}
	}
}

func TestParsedInfoEuropeCitation(t *testing.T) {
	t.Parallel()

	for _, names := range [][]string{nil, {"A"}, {"A", "B"}, {"A", "B", "C"}} {
		name := fmt.Sprint(names)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			europe := new(fakeEuropePMC)
			europe.article = new(literature.EuropePMCArticle)
			europe.article.PMID = testPMID
			europe.article.PubYear = "1990"

			europe.article.PublishDate = new(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))
			for _, name := range names {
				europe.article.Authors = append(
					europe.article.Authors,
					F.Pipe2(
						literature.EuropePMCAuthor{},
						europeAuthorFirstLens.Set(name),
						europeAuthorLastLens.Set("Author"),
					),
				)
			}

			p := F.Pipe3(
				errors.New("offline"),
				newFakePubMedWithError,
				newPublication,
				testEuropeLens.Set(europe),
			)
			info := mustPubInfo(t, p, testPMID)
			want := map[int]string{
				0: "unknown authors (2020)",
				1: "A Author (2020)",
				2: "A Author & B Author (2020)",
				3: "A Author et al. (2020)",
			}
			require.Equal(t, want[len(names)], info.AuthorStr)
			require.Equal(t, "https://pubmed.gov/4893433", info.PubmedURL)
			require.Empty(t, info.DoiURL)
		})
	}
}

func TestParsedInfoBothProvidersFail(t *testing.T) {
	t.Parallel()

	europe := new(fakeEuropePMC)
	europe.err = literature.NewError(literature.ErrorTypeNetworkError, "Europe PMC unavailable")
	primary := errors.New("NCBI unavailable")
	p := F.Pipe3(
		primary,
		newFakePubMedWithError,
		newPublication,
		testEuropeLens.Set(europe),
	)
	err := F.Pipe3(testPMID, p.ParsedInfo, ioeutils.ToEither[error, *PubInfo], E.ToError[*PubInfo])
	require.ErrorIs(t, err, europe.err)
	require.ErrorIs(t, err, primary)
	require.ErrorContains(t, err, testPMID)
}

func TestParsedInfoPMIDNCBIFirst(t *testing.T) {
	t.Parallel()

	ncbi := F.Pipe2([]literature.Author{}, testArticle, newFakePubMed)
	europe := new(fakeEuropePMC)
	source := F.Pipe2(ncbi, newPublication, testEuropeLens.Set(europe))
	lookup := source.ParsedInfo(" 4893433 ")

	require.Empty(t, ncbi.calls)
	require.Empty(t, europe.calls)

	info := F.Pipe2(lookup, ioeutils.ToEither[error, *PubInfo], E.Fold(
		func(err error) *PubInfo { t.Fatalf("PMID lookup: %v", err); return nil },
		F.Identity[*PubInfo],
	))
	require.Equal(t, "unknown authors (2018)", info.AuthorStr)
	require.Equal(t, []string{testPMID}, ncbi.calls)
	require.Empty(t, ncbi.queries)
	require.Empty(t, europe.calls)
}

func TestEuropePubInfoMissingYear(t *testing.T) {
	t.Parallel()

	for _, date := range []*time.Time{nil, new(time.Time)} {
		info := F.Pipe2(&literature.EuropePMCArticle{}, europeDateLens.Set(date), toEuropePubInfo)
		require.Equal(t, "unknown authors", info.AuthorStr)
		require.Empty(t, info.PubmedURL)
		require.Empty(t, info.DoiURL)
	}
}

func TestParsedInfoInvalidSkipsBothProviders(t *testing.T) {
	t.Parallel()

	for _, id := range []string{"", "   ", "abc", "0", "0123", "123x", "https://doi.org/10.1101/%GG"} {
		t.Run(id, func(t *testing.T) {
			t.Parallel()

			ncbi := newFakePubMed(nil)
			ncbi.searchErr = errors.New("invalid input reached NCBI")
			europe := new(fakeEuropePMC)
			europe.err = literature.NewError(
				literature.ErrorTypeNetworkError,
				"invalid input reached Europe PMC",
			)
			p := F.Pipe2(ncbi, newPublication, testEuropeLens.Set(europe))
			err := F.Pipe3(
				id,
				p.ParsedInfo,
				ioeutils.ToEither[error, *PubInfo],
				E.ToError[*PubInfo],
			)

			var typed *literature.Error
			require.ErrorAs(t, err, &typed)
			require.Equal(t, literature.ErrorTypeInvalidInput, typed.Type)
			require.Empty(t, ncbi.calls)
			require.Empty(t, ncbi.queries)
			require.Empty(t, europe.calls)
		})
	}
}
