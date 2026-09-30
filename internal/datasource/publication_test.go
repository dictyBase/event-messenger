package datasource

import (
	"errors"
	"fmt"
	"testing"
	"time"

	E "github.com/IBM/fp-go/v2/either"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictybase/literature"
	"github.com/stretchr/testify/require"
)

const (
	testPMID  = "4893433"
	huberName = "Huber RJ"
)

// fakePubMed is an in-memory pubMedClient stub that records every id it
// is asked for.
type fakePubMed struct {
	articles map[string]*literature.Article
	errs     map[string]error
	calls    []string
}

func (f *fakePubMed) GetArticle(pmid string) (*literature.Article, error) {
	f.calls = append(f.calls, pmid)

	if err := f.errs[pmid]; err != nil {
		return nil, err
	}

	return f.articles[pmid], nil
}

func newFakePubMed(article *literature.Article) *fakePubMed {
	return &fakePubMed{
		articles: map[string]*literature.Article{testPMID: article},
		errs:     map[string]error{},
	}
}

func newFakePubMedWithError(pmid string, err error) *fakePubMed {
	return &fakePubMed{
		articles: map[string]*literature.Article{},
		errs:     map[string]error{pmid: err},
	}
}

func testArticle(authors []literature.Author) *literature.Article {
	return &literature.Article{
		PMID:        testPMID,
		DOI:         "10.1016/j.bbamcr.2018.07.017",
		PublishDate: time.Date(2018, time.July, 23, 0, 0, 0, 0, time.UTC),
		Authors:     authors,
	}
}

func parsedInfo(t *testing.T, fake *fakePubMed) *PubInfo {
	t.Helper()

	info, err := E.UnwrapError(
		ioeutils.ToEither(newPublication(fake).ParsedInfo(testPMID)),
	)
	require.NoError(t, err, "should parse the fake article without error")

	return info
}

func TestParsedInfoSingleAuthor(t *testing.T) {
	t.Parallel()

	info := parsedInfo(t, newFakePubMed(testArticle(
		[]literature.Author{{FullName: huberName}},
	)))

	require.Equal(t, "Huber RJ (2018)", info.AuthorStr)
	require.Equal(t, "https://pubmed.gov/4893433", info.PubmedURL)
	require.Equal(
		t,
		"https://doi.org/10.1016/j.bbamcr.2018.07.017",
		info.DoiURL,
	)
}

func TestParsedInfoTwoAuthors(t *testing.T) {
	t.Parallel()

	info := parsedInfo(t, newFakePubMed(testArticle(
		[]literature.Author{
			{FullName: huberName},
			{FullName: "Mathavarajah S"},
		},
	)))

	require.Equal(t, "Huber RJ & Mathavarajah S (2018)", info.AuthorStr)
}

func TestParsedInfoManyAuthors(t *testing.T) {
	t.Parallel()

	info := parsedInfo(t, newFakePubMed(testArticle(
		[]literature.Author{
			{FullName: huberName},
			{FullName: "Mathavarajah S"},
			{FullName: "Third A"},
		},
	)))

	require.Equal(t, "Huber RJ et al. (2018)", info.AuthorStr)
}

func TestParsedInfoNoAuthors(t *testing.T) {
	t.Parallel()

	info := parsedInfo(t, newFakePubMed(testArticle(nil)))

	require.Equal(t, "unknown authors (2018)", info.AuthorStr)
}

func TestParsedInfoMissingDOI(t *testing.T) {
	t.Parallel()

	article := testArticle([]literature.Author{{FullName: huberName}})
	article.DOI = ""

	info := parsedInfo(t, newFakePubMed(article))

	require.Empty(t, info.DoiURL, "missing doi must not build a doi.org url")
}

func TestParsedInfoZeroPublishDate(t *testing.T) {
	t.Parallel()

	article := testArticle([]literature.Author{{FullName: huberName}})
	article.PublishDate = time.Time{}

	info := parsedInfo(t, newFakePubMed(article))

	require.Equal(t, huberName, info.AuthorStr)
	require.NotContains(t, info.AuthorStr, "(1)")
}

func TestParsedInfoError(t *testing.T) {
	t.Parallel()

	_, err := E.UnwrapError(ioeutils.ToEither(
		newPublication(newFakePubMedWithError(
			testPMID,
			errors.New("network down"),
		)).ParsedInfo(testPMID),
	))

	require.Error(t, err)
	require.ErrorContains(t, err, testPMID)
	require.ErrorContains(t, err, "network down")
}

func TestParsedInfoTypedError(t *testing.T) {
	t.Parallel()

	typed := literature.NewError(literature.ErrorTypeInvalidInput, "bad pmid")

	_, err := E.UnwrapError(ioeutils.ToEither(
		newPublication(newFakePubMedWithError(testPMID, typed)).
			ParsedInfo(testPMID),
	))

	require.Error(t, err)

	var litErr *literature.Error
	require.ErrorAs(t, err, &litErr)
	require.Equal(t, literature.ErrorTypeInvalidInput, litErr.Type)
}

func TestParsedInfoInvalidPMID(t *testing.T) {
	t.Parallel()

	for _, pmid := range []string{"", "   ", "abc", "123x", "0"} {
		t.Run(fmt.Sprintf("pmid %q", pmid), func(t *testing.T) {
			t.Parallel()

			// The stub answers every id, so a rejection here can only
			// come from validation in front of the client call.
			fake := newFakePubMed(testArticle(
				[]literature.Author{{FullName: huberName}},
			))
			fake.articles[pmid] = fake.articles[testPMID]

			_, err := E.UnwrapError(ioeutils.ToEither(
				newPublication(fake).ParsedInfo(pmid),
			))
			require.Error(t, err)
			require.ErrorContains(
				t,
				err,
				"pmid must contain decimal digits",
			)

			var litErr *literature.Error
			require.ErrorAs(t, err, &litErr)
			require.Equal(
				t,
				literature.ErrorTypeInvalidInput,
				litErr.Type,
			)
			require.Empty(
				t,
				fake.calls,
				"an invalid pmid must be rejected before the client call",
			)
		})
	}
}

func TestAuthorStr(t *testing.T) {
	t.Parallel()

	authors := func(names ...string) []literature.Author {
		out := make([]literature.Author, 0, len(names))
		for _, name := range names {
			out = append(out, literature.Author{FullName: name})
		}

		return out
	}

	for _, tc := range []struct {
		name    string
		authors []literature.Author
		want    string
	}{
		{name: "no authors", authors: nil, want: "unknown authors"},
		{name: "empty slice", authors: authors(), want: "unknown authors"},
		{name: "one author", authors: authors("A"), want: "A"},
		{name: "two authors", authors: authors("A", "B"), want: "A & B"},
		{name: "three authors", authors: authors("A", "B", "C"), want: "A et al."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, authorStr(tc.authors))
		})
	}
}

func TestPublicationOptions(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		args *PublicationParams
		want int
	}{
		{name: "empty params", args: &PublicationParams{}, want: 0},
		{
			name: "api key only",
			args: &PublicationParams{APIKey: "key"},
			want: 1,
		},
		{
			name: "tool only",
			args: &PublicationParams{Tool: "event-messenger"},
			want: 1,
		},
		{
			name: "email only",
			args: &PublicationParams{Email: "dev@example.org"},
			want: 1,
		},
		{
			name: "full identity",
			args: &PublicationParams{
				APIKey: "key",
				Tool:   "event-messenger",
				Email:  "dev@example.org",
			},
			want: 3,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Len(t, publicationOptions(tc.args), tc.want)
		})
	}
}

func TestNewPublicationEmptyParams(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	pub, err := E.UnwrapError(ioeutils.ToEither(
		NewPublication(&PublicationParams{}),
	))
	require.NoError(err, "expect no error, received %s", err)
	require.NotNil(pub, "expect a publication source")
}
