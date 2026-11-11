package datasource

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
	arrutils "github.com/dictyBase/fp-go-loom/array"
	MO "github.com/dictyBase/fp-go-loom/matchopt"
	predarrays "github.com/dictyBase/fp-go-loom/predicate/array"
	"github.com/dictyBase/fp-go-loom/strutils"
	"github.com/dictybase/literature"
)

// pubMedClient defines the NCBI operations for a citation.
type pubMedClient interface {
	GetArticle(pmid string) (*literature.Article, error)
	Search(query string, opts ...literature.SearchOption) (*literature.SearchResult, error)
}

type europePMCClient interface {
	GetArticle(pmid string) (*literature.EuropePMCArticle, error)
	GetArticleByDOI(doi string) (*literature.EuropePMCArticle, error)
}

const (
	twoAuthors   = 2
	threeAuthors = 3
)

// PubInfo contains the citation and its optional links for an invoice row.
type PubInfo struct {
	AuthorStr string
	PubmedURL string
	DoiURL    string
}

// Publication resolves PMID and DOI identifiers into citations.
type Publication struct {
	client    pubMedClient
	europePMC europePMCClient
}

// PublicationParams contains the optional identity for literature requests.
type PublicationParams struct {
	APIKey string
	Tool   string
	Email  string
}

// NewPublication creates both clients as lazy effects.
func NewPublication(p *PublicationParams) IOE.IOEither[error, *Publication] {
	pubmed := createPubMed(p)
	europe := createEuropePMC(p)

	return F.Pipe3(
		IOE.Of[error](Publication{}),
		IOE.Bind(publicationClientLens.Set, F.Constant1[Publication](pubmed)),
		IOE.Bind(publicationEuropeLens.Set, F.Constant1[Publication](europe)),
		IOE.Map[error](func(source Publication) *Publication { return new(source) }),
	)
}

func createPubMed(p *PublicationParams) IOE.IOEither[error, pubMedClient] {
	opts := publicationOptions(p)

	return F.Pipe1(
		IOE.TryCatchError(func() (pubMedClient, error) { return literature.New(opts...) }),
		IOE.MapLeft[pubMedClient](func(err error) error {
			return fmt.Errorf("error creating literature pubmed client: %w", err)
		}),
	)
}

func createEuropePMC(p *PublicationParams) IOE.IOEither[error, europePMCClient] {
	opts := europePMCOptions(p)

	return F.Pipe1(
		IOE.TryCatchError(
			func() (europePMCClient, error) { return literature.NewEuropePMCClient(opts...) },
		),
		IOE.MapLeft[europePMCClient](func(err error) error {
			return fmt.Errorf("error creating literature Europe PMC client: %w", err)
		}),
	)
}

func someOption(apply func(string) literature.Option, value string) O.Option[literature.Option] {
	return F.Pipe2(value, O.FromPredicate(S.IsNonEmpty), O.Map(apply))
}

func publicationOptions(p *PublicationParams) []literature.Option {
	return F.Pipe1(
		[]O.Option[literature.Option]{
			F.Pipe2(p, paramsAPIKeyLens.Get, F.Bind1st(someOption, literature.WithAPIKey)),
			F.Pipe2(p, paramsToolLens.Get, F.Bind1st(someOption, literature.WithTool)),
			F.Pipe2(p, paramsEmailLens.Get, F.Bind1st(someOption, literature.WithEmail)),
		},
		arrutils.Compact[literature.Option],
	)
}

func europePMCOptions(p *PublicationParams) []literature.EuropePMCOption {
	return F.Pipe5(
		p,
		paramsEmailLens.Get,
		O.FromPredicate(S.IsNonEmpty),
		O.Map(literature.WithEuropePMCEmail),
		A.Of[O.Option[literature.EuropePMCOption]],
		arrutils.Compact[literature.EuropePMCOption],
	)
}

var (
	pmidPattern  = regexp.MustCompile(`^[1-9][0-9]*$`)
	validatePMID = E.FromPredicate(pmidPattern.MatchString, func(pmid string) error {
		return F.Pipe1(
			literature.NewErrorWithPMID(
				literature.ErrorTypeInvalidInput,
				pmid,
				"pmid must contain decimal digits",
			),
			func(err *literature.Error) error { return fmt.Errorf("error fetching publication %s: %w", pmid, err) },
		)
	})
	isResolverURL = F.Pipe1(
		S.HasPrefix("https://doi.org/"),
		P.Or(S.HasPrefix("http://doi.org/")),
	)
	isDOI              = S.Includes("/")
	validateIdentifier = F.Pipe1(isDOI, P.Fold(validatePMID, E.Of[error, string]))
	isNonZeroTime      = P.Not(time.Time.IsZero)
	formatYearText     = strutils.Wrap(" (", ")")
)

// ParsedInfo validates the identifier before any client request.
func (p *Publication) ParsedInfo(rawID string) IOE.IOEither[error, *PubInfo] {
	return F.Pipe3(
		IOE.Of[error](rawID),
		IOE.ChainEitherK(normalizeID),
		IOE.ChainEitherK(validateIdentifier),
		IOE.Chain(p.fetchWithFallback),
	)
}

func normalizeID(id string) E.Either[error, string] {
	return F.Pipe3(
		id,
		strings.TrimSpace,
		F.Pipe1(isResolverURL, P.Fold(E.Of[error, string], F.Flow3(
			E.Eitherize1(url.Parse),
			E.Map[error](F.Flow2(resolverPathLens.Get, F.Bind2nd(strings.TrimPrefix, "/"))),
			E.MapLeft[string](func(err error) error {
				return literature.WrapErrorWithDOI(
					literature.ErrorTypeInvalidInput,
					id,
					"invalid DOI resolver URL",
					err,
				)
			}),
		))),
		E.Map[error](F.Flow3(
			F.Bind2nd(strings.TrimPrefix, "doi:"),
			F.Bind2nd(strings.TrimPrefix, "DOI:"),
			strings.TrimSpace,
		)),
	)
}

func (p *Publication) fetchWithFallback(id string) IOE.IOEither[error, *PubInfo] {
	return F.Pipe3(
		id,
		F.Pipe1(isDOI, P.Fold(p.ncbiByPMID, p.ncbiByDOI)),
		IOE.OrElse(func(primary error) IOE.IOEither[error, *PubInfo] {
			return F.Pipe2(
				id,
				F.Pipe1(isDOI, P.Fold(p.europeByPMID, p.europeByDOI)),
				IOE.MapLeft[*PubInfo](
					func(secondary error) error { return errors.Join(primary, secondary) },
				),
			)
		}),
		IOE.MapLeft[*PubInfo](
			func(err error) error { return fmt.Errorf("error fetching publication %s: %w", id, err) },
		),
	)
}

func (p *Publication) ncbiByPMID(pmid string) IOE.IOEither[error, *PubInfo] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*literature.Article, error) {
			return publicationClientLens.Get(*p).GetArticle(pmid)
		}),
		IOE.Map[error](toPubInfo),
	)
}

func (p *Publication) ncbiByDOI(doi string) IOE.IOEither[error, *PubInfo] {
	return F.Pipe3(
		IOE.TryCatchError(func() (*literature.SearchResult, error) {
			return publicationClientLens.Get(*p).Search(doi+"[DOI]", literature.WithLimit(1))
		}),
		IOE.Map[error](searchArticlesLens.Get),
		IOE.ChainEitherK(F.Flow2(
			A.Head[*literature.Article],
			E.FromOption[*literature.Article](F.Constant[error](literature.NewErrorWithDOI(
				literature.ErrorTypeArticleNotFound, doi, "no PubMed article for DOI",
			))),
		)),
		IOE.Map[error](toPubInfo),
	)
}

func (p *Publication) europeByPMID(pmid string) IOE.IOEither[error, *PubInfo] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*literature.EuropePMCArticle, error) {
			return publicationEuropeLens.Get(*p).GetArticle(pmid)
		}),
		IOE.Map[error](toEuropePubInfo),
	)
}

func (p *Publication) europeByDOI(doi string) IOE.IOEither[error, *PubInfo] {
	return F.Pipe1(
		IOE.TryCatchError(func() (*literature.EuropePMCArticle, error) {
			return publicationEuropeLens.Get(*p).GetArticleByDOI(doi)
		}),
		IOE.Map[error](toEuropePubInfo),
	)
}

func toPubInfo(a *literature.Article) *PubInfo {
	return F.Pipe3(
		&PubInfo{},
		pubInfoAuthorLens.Set(strutils.JoinStrings(
			F.Pipe1(a, F.Flow2(articleAuthorsLens.Get, authorStr)),
			F.Pipe1(a, F.Flow2(articleDateLens.Get, pubYear)),
		)),
		pubInfoPMIDLens.Set(F.Pipe1(a, F.Flow2(articlePMIDLens.Get, pubmedURL))),
		pubInfoDOILens.Set(F.Pipe1(a, F.Flow2(articleDOILens.Get, doiURL))),
	)
}

func toEuropePubInfo(a *literature.EuropePMCArticle) *PubInfo {
	return F.Pipe3(
		&PubInfo{},
		pubInfoAuthorLens.Set(strutils.JoinStrings(
			F.Pipe3(
				a,
				europeAuthorsLens.Get,
				A.Map(func(author literature.EuropePMCAuthor) literature.Author {
					return F.Pipe3(
						literature.Author{},
						F.Pipe2(author, europeAuthorFullLens.Get, authorFullLens.Set),
						F.Pipe2(author, europeAuthorFirstLens.Get, authorFirstLens.Set),
						F.Pipe2(author, europeAuthorLastLens.Get, authorLastLens.Set),
					)
				}),
				authorStr,
			),
			F.Pipe5(
				a,
				europeDateLens.Get,
				O.FromNillable2[time.Time],
				O.Chain(O.FromPredicate(isNonZeroTime)),
				O.Map(pubYear),
				O.GetOrElse(F.Constant(F.Pipe2(
					a, europeYearLens.Get, strutils.MapNonEmpty(formatYearText),
				))),
			),
		)),
		pubInfoPMIDLens.Set(F.Pipe1(a, F.Flow2(europePMIDLens.Get, pubmedURL))),
		pubInfoDOILens.Set(F.Pipe1(a, F.Flow2(europeDOILens.Get, doiURL))),
	)
}

func authorStr(authors []literature.Author) string {
	return MO.First("unknown authors", []O.Option[string]{
		F.Pipe1(authors, MO.Case(predarrays.LenEq[literature.Author](1), singleAuthor)),
		F.Pipe1(authors, MO.Case(predarrays.LenEq[literature.Author](twoAuthors),
			F.Flow3(A.Map(displayName), A.Intersperse(" & "), A.Fold(S.Monoid)))),
		F.Pipe1(authors, MO.Case(predarrays.MinLen[literature.Author](threeAuthors),
			F.Flow2(singleAuthor, S.Append(" et al.")))),
	})
}

func singleAuthor(a []literature.Author) string { return displayName(a[0]) }

func displayName(a literature.Author) string {
	return F.Pipe3(
		a,
		authorFullLens.Get,
		O.FromPredicate(S.IsNonEmpty),
		O.GetOrElse(
			F.Constant(
				F.Pipe1(
					[]string{authorFirstLens.Get(a), authorLastLens.Get(a)},
					strutils.JoinSpace,
				),
			),
		),
	)
}

func formatYear(t time.Time) string { return fmt.Sprintf(" (%d)", t.Year()) }

func pubYear(t time.Time) string {
	return F.Pipe3(
		t,
		O.FromPredicate(isNonZeroTime),
		O.Map(formatYear),
		O.GetOrElse(F.Constant("")),
	)
}

func pubmedURL(pmid string) string {
	return F.Pipe1(pmid, strutils.MapNonEmpty(S.Prepend("https://pubmed.gov/")))
}

func doiURL(doi string) string {
	return F.Pipe3(
		doi,
		normalizeID,
		E.GetOrElse(F.Constant1[error]("")),
		strutils.MapNonEmpty(S.Prepend("https://doi.org/")),
	)
}
