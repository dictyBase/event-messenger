package template

import (
	"fmt"
	"testing"

	"github.com/dictyBase/event-messenger/internal/datasource"
	"github.com/stretchr/testify/require"
)

const (
	plasmidStockType = "plasmid"
	testDoiURL       = "https://doi.org/10.1101/582072"
	testPubmedURL    = "https://pubmed.gov/4893433"
)

func checkPublicationLinks(
	t *testing.T,
	entries []textEntry,
	urls []string,
	rowCount int,
) {
	t.Helper()

	var pubLinks []textEntry

	for _, entry := range entries {
		if entry.text == "Pubmed" || entry.text == "Full text" {
			require.NotEmpty(t, entry.href)
			pubLinks = append(pubLinks, entry)
		}
	}

	var want []string

	for _, url := range urls {
		if url != "" {
			want = append(want, url)
		}
	}

	require.Len(t, pubLinks, len(want)*rowCount)

	for _, url := range want {
		found := false

		for _, entry := range pubLinks {
			if entry.href == url {
				found = true
				break
			}
		}

		require.True(t, found, "expected link %s", url)
	}
}

func TestEmailOptionalPublicationLinks(t *testing.T) {
	t.Parallel()

	for _, stockType := range []string{"strain", plasmidStockType} {
		for _, urls := range [][]string{
			{"", testDoiURL},
			{testPubmedURL, ""},
			{testPubmedURL, testDoiURL},
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
				if stockType == plasmidStockType {
					content = fakePlasmidOnlyEmailContent()
				}

				for _, row := range content.StrainData {
					row.PubInfo = []*datasource.PubInfo{pub}
				}

				for _, row := range content.PlasmidData {
					row.PubInfo = []*datasource.PubInfo{pub}
				}

				rowCount := len(content.StrainData)
				if stockType == plasmidStockType {
					rowCount = len(content.PlasmidData)
				}

				entries := collectTextEntries(BuildOrderMaroto(content).GetStructure())
				texts := entryTexts(entries)
				require.Contains(t, texts, "Smith A (2019)")
				checkPublicationLinks(t, entries, urls, rowCount)
			})
		}
	}
}
