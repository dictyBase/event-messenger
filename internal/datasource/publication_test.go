package datasource

import (
	"errors"
	"fmt"
	"testing"
	"time"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	L "github.com/IBM/fp-go/v2/optics/lens"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictybase/literature"
	"github.com/stretchr/testify/require"
)

const (
	testPMID  = "4893433"
	huberName = "Huber RJ"
)

// fakePubMed records the external requests from the publication source.
type fakePubMed struct {
	articles      map[string]*literature.Article
	errs          map[string]error
	calls         []string
	searchResult  *literature.SearchResult
	searchErr     error
	queries       []string
	searchOptions []literature.SearchOption
}

func (f *fakePubMed) GetArticle(pmid string) (*literature.Article, error) {
	f.calls = append(f.calls, pmid)
	return f.articles[pmid], f.errs[pmid]
}

func (f *fakePubMed) Search(
	query string,
	opts ...literature.SearchOption,
) (*literature.SearchResult, error) {
	f.queries = append(f.queries, query)
	f.searchOptions = opts

	return f.searchResult, f.searchErr
}

var (
	fakeArticlesLens = L.MakeLensRef(
		func(f *fakePubMed) map[string]*literature.Article { return f.articles },
		func(f *fakePubMed, v map[string]*literature.Article) *fakePubMed { f.articles = v; return f },
	)
	fakeErrorsLens = L.MakeLensRef(
		func(f *fakePubMed) map[string]error { return f.errs },
		func(f *fakePubMed, v map[string]error) *fakePubMed { f.errs = v; return f },
	)
	fakeEuropeErrorLens = L.MakeLensRef(
		func(f *fakeEuropePMC) error { return f.err },
		func(f *fakeEuropePMC, v error) *fakeEuropePMC { f.err = v; return f },
	)
)

func newPublication(c *fakePubMed) *Publication {
	return F.Pipe3(
		Publication{},
		publicationClientLens.Set(c),
		publicationEuropeLens.Set(
			F.Pipe1(
				&fakeEuropePMC{},
				fakeEuropeErrorLens.Set(errors.New("Europe PMC unavailable")),
			),
		),
		func(p Publication) *Publication { return new(p) },
	)
}

func newFakePubMed(article *literature.Article) *fakePubMed {
	return F.Pipe1(
		&fakePubMed{},
		fakeArticlesLens.Set(map[string]*literature.Article{testPMID: article}),
	)
}

func newFakePubMedWithError(err error) *fakePubMed {
	return F.Pipe1(&fakePubMed{}, fakeErrorsLens.Set(map[string]error{testPMID: err}))
}

func testArticle(authors []literature.Author) *literature.Article {
	return F.Pipe4(
		&literature.Article{},
		articlePMIDLens.Set(testPMID),
		articleDOILens.Set(indexedDOI),
		articleDateLens.Set(time.Date(2018, time.July, 23, 0, 0, 0, 0, time.UTC)),
		articleAuthorsLens.Set(authors),
	)
}

func authorByName(name string) literature.Author {
	return F.Pipe1(literature.Author{}, authorFullLens.Set(name))
}

func TestParsedInfoAuthors(t *testing.T) {
	t.Parallel()

	for _, names := range [][]string{
		nil,
		{huberName},
		{huberName, "Mathavarajah S"},
		{huberName, "Mathavarajah S", "Third A"},
	} {
		name := fmt.Sprint(names)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			source := F.Pipe4(
				names,
				A.Map(authorByName),
				testArticle,
				newFakePubMed,
				newPublication,
			)
			info := mustPubInfo(t, source, testPMID)
			want := map[int]string{
				0: "unknown authors (2018)",
				1: "Huber RJ (2018)",
				2: "Huber RJ & Mathavarajah S (2018)",
				3: "Huber RJ et al. (2018)",
			}
			require.Equal(t, want[len(names)], info.AuthorStr)
			require.Equal(t, "https://pubmed.gov/4893433", info.PubmedURL)
			require.Equal(t, "https://doi.org/10.1016/j.bbamcr.2018.07.017", info.DoiURL)
		})
	}
}

func TestParsedInfoMissingDOI(t *testing.T) {
	t.Parallel()

	source := F.Pipe4(
		[]literature.Author{},
		testArticle,
		articleDOILens.Set(""),
		newFakePubMed,
		newPublication,
	)
	info := mustPubInfo(t, source, testPMID)
	require.Empty(t, info.DoiURL)
}

func TestParsedInfoZeroPublishDate(t *testing.T) {
	t.Parallel()

	source := F.Pipe4(
		[]literature.Author{},
		testArticle,
		articleDateLens.Set(time.Time{}),
		newFakePubMed,
		newPublication,
	)
	info := mustPubInfo(t, source, testPMID)
	require.Equal(t, "unknown authors", info.AuthorStr)
}

func TestParsedInfoTypedError(t *testing.T) {
	t.Parallel()

	typed := literature.NewError(literature.ErrorTypeInvalidInput, "bad pmid")
	europe := F.Pipe1(&fakeEuropePMC{}, fakeEuropeErrorLens.Set(typed))
	source := F.Pipe3(
		error(typed),
		newFakePubMedWithError,
		newPublication,
		testEuropeLens.Set(europe),
	)
	err := F.Pipe2(
		testPMID,
		source.ParsedInfo,
		F.Flow2(ioeutils.ToEither[error, *PubInfo], E.ToError[*PubInfo]),
	)
	require.ErrorIs(t, err, typed)
}

func TestAuthorStr(t *testing.T) {
	t.Parallel()

	for _, names := range [][]string{nil, {}, {"A"}, {"A", "B"}, {"A", "B", "C"}} {
		name := fmt.Sprint(names)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := F.Pipe2(names, A.Map(authorByName), authorStr)
			want := map[int]string{0: "unknown authors", 1: "A", 2: "A & B", 3: "A et al."}
			require.Equal(t, want[len(names)], got)
		})
	}
}

func TestPublicationOptions(t *testing.T) {
	t.Parallel()

	for _, values := range [][]string{
		{"", "", ""},
		{"key", "", ""},
		{"", "event-messenger", ""},
		{"", "", "dev@example.org"},
		{"key", "event-messenger", "dev@example.org"},
	} {
		name := fmt.Sprint(values)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			params := F.Pipe3(
				&PublicationParams{},
				paramsAPIKeyLens.Set(values[0]),
				paramsToolLens.Set(values[1]),
				paramsEmailLens.Set(values[2]),
			)
			want := 0

			for _, value := range values {
				if value != "" {
					want++
				}
			}

			require.Len(t, publicationOptions(params), want)

			wantEurope := 0
			if values[2] != "" {
				wantEurope = 1
			}

			require.Len(t, europePMCOptions(params), wantEurope)
		})
	}
}

func TestNewPublicationEmptyParams(t *testing.T) {
	t.Parallel()

	pub := F.Pipe3(
		&PublicationParams{},
		NewPublication,
		ioeutils.ToEither[error, *Publication],
		E.Fold(
			func(err error) *Publication { t.Fatalf("create publication source: %v", err); return nil },
			F.Identity[*Publication],
		),
	)
	require.NotNil(t, publicationClientLens.Get(*pub))
	require.NotNil(t, publicationEuropeLens.Get(*pub))
}
