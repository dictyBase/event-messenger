package datasource

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictybase/literature"
	"github.com/stretchr/testify/require"
)

const indexedDOI = "10.1016/j.bbamcr.2018.07.017"

func TestParsedInfoDOINCBIFirst(t *testing.T) {
	t.Parallel()

	for _, id := range []string{
		indexedDOI,
		"doi:" + indexedDOI,
		"DOI:" + indexedDOI,
		"https://doi.org/" + indexedDOI,
		"http://doi.org/" + indexedDOI,
		"  doi: " + indexedDOI + "  ",
		"https://doi.org/10.1016%2Fj.bbamcr.2018.07.017",
		"http://doi.org/10.1016%2fj.bbamcr.2018.07.017?source=invoice#citation",
		"https://doi.org/" + indexedDOI + "?source=invoice#citation",
	} {
		t.Run(id, func(t *testing.T) {
			t.Parallel()

			fake := F.Pipe1([]literature.Author{}, testArticle)
			ncbi := newFakePubMed(fake)
			ncbi.searchResult = F.Pipe1(
				&literature.SearchResult{},
				searchArticlesLens.Set([]*literature.Article{fake}),
			)
			europe := new(fakeEuropePMC)
			source := F.Pipe2(ncbi, newPublication, testEuropeLens.Set(europe))
			lookup := source.ParsedInfo(id)

			require.Empty(t, ncbi.queries)
			require.Empty(t, europe.calls)

			info := F.Pipe2(
				lookup,
				ioeutils.ToEither[error, *PubInfo],
				E.Fold(func(err error) *PubInfo {
					t.Fatalf("DOI lookup failed: %v", err)
					return nil
				}, F.Identity[*PubInfo]),
			)
			require.Equal(t, "unknown authors (2018)", info.AuthorStr)
			require.Equal(t, "https://doi.org/"+indexedDOI, info.DoiURL)
			require.Equal(t, []string{indexedDOI + "[DOI]"}, ncbi.queries)
			require.Len(t, ncbi.searchOptions, 1)
			require.Empty(t, ncbi.calls)
			require.Empty(t, europe.calls)
		})
	}
}
