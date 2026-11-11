package template

import (
	"fmt"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/dictyBase/event-messenger/internal/datasource"
	"github.com/stretchr/testify/require"
)

const emailTemplateFile = "email.tmpl"

func checkPublicationLinks(t *testing.T, citation *goquery.Selection, urls []string) {
	t.Helper()
	require.Zero(t, citation.Find(`a[href=""]`).Length())

	want := 0

	for _, url := range urls {
		if url != "" {
			want++

			require.Equal(t, 1, citation.Find(`a[href="`+url+`"]`).Length())
		}
	}

	require.Equal(t, want, citation.Find("a").Length())
}

func TestEmailOptionalPublicationLinks(t *testing.T) {
	t.Parallel()

	for _, stockType := range []string{"strain", "plasmid"} {
		for _, urls := range [][]string{
			{"", "https://doi.org/10.1101/582072"},
			{"https://pubmed.gov/4893433", ""},
			{"https://pubmed.gov/4893433", "https://doi.org/10.1101/582072"},
			{"", ""},
		} {
			name := fmt.Sprint(stockType, urls)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				pub := new(datasource.PubInfo)
				pub.AuthorStr = "Smith A (2019)"
				pub.PubmedURL = urls[0]
				pub.DoiURL = urls[1]

				content := fakeStrainOnlyEmailContent()
				if stockType == "plasmid" {
					content = fakePlasmidOnlyEmailContent()
				}

				for _, row := range content.StrainData {
					row.PubInfo = []*datasource.PubInfo{pub}
				}

				for _, row := range content.PlasmidData {
					row.PubInfo = []*datasource.PubInfo{pub}
				}

				params := new(OutputParams)
				params.Path = "/"
				params.File = emailTemplateFile
				params.Content = content
				body, err := OutputHTML(params)
				require.NoError(t, err)
				doc, err := goquery.NewDocumentFromReader(body)
				require.NoError(t, err)

				citation := doc.Find("div#" + stockType + " tbody tr").First().Find("td").Last()
				require.Contains(t, citation.Text(), "Smith A (2019)")
				checkPublicationLinks(t, citation, urls)
			})
		}
	}
}
