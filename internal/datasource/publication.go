package datasource

import (
	"fmt"
	"regexp"
	"time"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
	MO "github.com/dictyBase/fp-go-loom/matchopt"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	"github.com/dictybase/literature"
)

// pubMedClient is the narrow seam over the NCBI eUtils client.
type pubMedClient interface {
	GetArticle(pmid string) (*literature.Article, error)
}

const (
	// twoAuthors is the citation arity that joins both names.
	twoAuthors = 2

	// threeAuthors is the citation arity that abbreviates to "et al.".
	threeAuthors = 3
)

// PubInfo is the citation snippet rendered into an invoice row.
type PubInfo struct {
	AuthorStr string
	PubmedURL string
	DoiURL    string
}

// Publication resolves PubMed identifiers into citation snippets.
type Publication struct {
	client pubMedClient
}

// NewPublication builds the publication source over the NCBI eUtils client.
func NewPublication() IOE.IOEither[error, *Publication] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*literature.Client, error) {
			return literature.New()
		}),
		IOE.MapLeft[*literature.Client](func(err error) error {
			return fmt.Errorf("error creating literature pubmed client: %w", err)
		}),
		IOE.Map[error](func(c *literature.Client) *Publication {
			return &Publication{client: c}
		}),
	)
}

// newPublication builds a publication source over an injected client.
func newPublication(c pubMedClient) *Publication {
	return &Publication{client: c}
}

// pmidPattern matches PubMed identifiers: decimal digits with no
// leading zero.
var pmidPattern = regexp.MustCompile(`^[1-9][0-9]*$`)

// validPMID reports whether pmid is a PubMed identifier.
func validPMID(pmid string) bool {
	return pmidPattern.MatchString(pmid)
}

// invalidPMIDError presents a non-identifier pmid as a typed literature
// error.
func invalidPMIDError(pmid string) error {
	return fmt.Errorf(
		"error fetching publication %s: %w",
		pmid,
		literature.NewErrorWithPMID(
			literature.ErrorTypeInvalidInput,
			pmid,
			"pmid must contain decimal digits",
		),
	)
}

// validatePMID rejects anything that is not a PubMed identifier before
// the client is called.
var validatePMID = E.FromPredicate(validPMID, invalidPMIDError)

// ParsedInfo fetches one article and formats its citation snippet.
func (p *Publication) ParsedInfo(pmid string) IOE.IOEither[error, *PubInfo] {
	return F.Pipe3(
		pmid,
		validatePMID,
		IOE.FromEither[error, string],
		IOE.Chain(p.fetchInfo),
	)
}

// fetchInfo fetches one article through the client and formats it.
func (p *Publication) fetchInfo(pmid string) IOE.IOEither[error, *PubInfo] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*literature.Article, error) {
			return p.client.GetArticle(pmid)
		}),
		IOE.MapLeft[*literature.Article](func(err error) error {
			return fmt.Errorf("error fetching publication %s: %w", pmid, err)
		}),
		IOE.Map[error](toPubInfo),
	)
}

// toPubInfo formats one article into a citation snippet. The publish
// year comes from Article.PublishDate, which the literature client fills
// from the PubMed journal issue; a zero date renders no year at all.
func toPubInfo(a *literature.Article) *PubInfo {
	authors := authorStr(a.Authors)
	year := pubYear(a.PublishDate)

	return &PubInfo{
		AuthorStr: S.Monoid.Concat(authors, year),
		PubmedURL: fmt.Sprintf("https://pubmed.gov/%s", a.PMID),
		DoiURL:    doiURL(a.DOI),
	}
}

// matchAuthors formats the author list by arity, with a total fallback.
func matchAuthors(authors []literature.Author) string {
	isSingle := predarrays.LenEq[literature.Author](1)
	isPair := predarrays.LenEq[literature.Author](twoAuthors)
	isEtAl := predarrays.MinLen[literature.Author](threeAuthors)

	cases := []O.Option[string]{
		F.Pipe1(authors, MO.Case(isSingle, singleAuthor)),
		F.Pipe1(authors, MO.Case(isPair, pairAuthors)),
		F.Pipe1(authors, MO.Case(isEtAl, etAlAuthors)),
	}

	return MO.First("unknown authors", cases)
}

// authorStr formats the author list for the citation.
func authorStr(authors []literature.Author) string {
	matcher := matchAuthors

	return F.Pipe1(authors, matcher)
}

// singleAuthor formats a one-author citation.
func singleAuthor(a []literature.Author) string {
	return displayName(a[0])
}

// pairAuthors formats a two-author citation.
func pairAuthors(a []literature.Author) string {
	first := displayName(a[0])
	second := displayName(a[1])

	return S.IntersperseMonoid(" & ").Concat(first, second)
}

// etAlAuthors formats a three-or-more-author citation.
func etAlAuthors(a []literature.Author) string {
	name := displayName(a[0])

	return S.Monoid.Concat(name, " et al.")
}

// fullNameFallback joins the author's given and family names.
func fullNameFallback(a literature.Author) string {
	return S.IntersperseMonoid(" ").Concat(a.FirstName, a.LastName)
}

// displayName prefers the abbreviated full name over given and family names.
func displayName(a literature.Author) string {
	fallback := fullNameFallback(a)

	return F.Pipe2(
		a.FullName,
		O.FromPredicate(S.IsNonEmpty),
		O.GetOrElse(F.Constant(fallback)),
	)
}

// isNonZeroTime reports whether the timestamp is set.
var isNonZeroTime = P.Not(time.Time.IsZero)

// formatYear renders a publish year as a citation suffix.
func formatYear(t time.Time) string {
	return fmt.Sprintf(" (%d)", t.Year())
}

// pubYear renders the publish year, or nothing when the date is unset.
func pubYear(t time.Time) string {
	return F.Pipe3(
		t,
		O.FromPredicate(isNonZeroTime),
		O.Map(formatYear),
		O.GetOrElse(F.Constant("")),
	)
}

// formatDOI renders a DOI as a resolvable url.
func formatDOI(doi string) string {
	return fmt.Sprintf("https://doi.org/%s", doi)
}

// doiURL renders the DOI url, or nothing when the DOI is missing.
func doiURL(doi string) string {
	return F.Pipe3(
		doi,
		O.FromPredicate(S.IsNonEmpty),
		O.Map(formatDOI),
		O.GetOrElse(F.Constant("")),
	)
}
