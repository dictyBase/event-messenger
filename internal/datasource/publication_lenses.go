package datasource

import (
	"net/url"
	"time"

	L "github.com/IBM/fp-go/v2/optics/lens"
	"github.com/dictybase/literature"
)

var (
	resolverPathLens = L.MakeLensRef(
		func(u *url.URL) string { return u.Path },
		func(u *url.URL, path string) *url.URL { u.Path = path; return u },
	)
	publicationClientLens = L.MakeLens(
		func(p Publication) pubMedClient { return p.client },
		func(p Publication, v pubMedClient) Publication { p.client = v; return p },
	)
	publicationEuropeLens = L.MakeLens(
		func(p Publication) europePMCClient { return p.europePMC },
		func(p Publication, v europePMCClient) Publication { p.europePMC = v; return p },
	)
	paramsAPIKeyLens = L.MakeLensRef(
		func(p *PublicationParams) string { return p.APIKey },
		func(p *PublicationParams, v string) *PublicationParams { p.APIKey = v; return p },
	)
	paramsToolLens = L.MakeLensRef(
		func(p *PublicationParams) string { return p.Tool },
		func(p *PublicationParams, v string) *PublicationParams { p.Tool = v; return p },
	)
	paramsEmailLens = L.MakeLensRef(
		func(p *PublicationParams) string { return p.Email },
		func(p *PublicationParams, v string) *PublicationParams { p.Email = v; return p },
	)
	pubInfoAuthorLens = L.MakeLensRef(
		func(p *PubInfo) string { return p.AuthorStr },
		func(p *PubInfo, v string) *PubInfo { p.AuthorStr = v; return p },
	)
	pubInfoPMIDLens = L.MakeLensRef(
		func(p *PubInfo) string { return p.PubmedURL },
		func(p *PubInfo, v string) *PubInfo { p.PubmedURL = v; return p },
	)
	pubInfoDOILens = L.MakeLensRef(
		func(p *PubInfo) string { return p.DoiURL },
		func(p *PubInfo, v string) *PubInfo { p.DoiURL = v; return p },
	)
	articleAuthorsLens = L.MakeLensRef(
		func(a *literature.Article) []literature.Author { return a.Authors },
		func(a *literature.Article, v []literature.Author) *literature.Article { a.Authors = v; return a },
	)
	articleDateLens = L.MakeLensRef(
		func(a *literature.Article) time.Time { return a.PublishDate },
		func(a *literature.Article, v time.Time) *literature.Article { a.PublishDate = v; return a },
	)
	articlePMIDLens = L.MakeLensRef(
		func(a *literature.Article) string { return a.PMID },
		func(a *literature.Article, v string) *literature.Article { a.PMID = v; return a },
	)
	articleDOILens = L.MakeLensRef(
		func(a *literature.Article) string { return a.DOI },
		func(a *literature.Article, v string) *literature.Article { a.DOI = v; return a },
	)
	authorFullLens = L.MakeLens(
		func(a literature.Author) string { return a.FullName },
		func(a literature.Author, v string) literature.Author { a.FullName = v; return a },
	)
	authorFirstLens = L.MakeLens(
		func(a literature.Author) string { return a.FirstName },
		func(a literature.Author, v string) literature.Author { a.FirstName = v; return a },
	)
	authorLastLens = L.MakeLens(
		func(a literature.Author) string { return a.LastName },
		func(a literature.Author, v string) literature.Author { a.LastName = v; return a },
	)
	searchArticlesLens = L.MakeLensRef(
		func(r *literature.SearchResult) []*literature.Article { return r.Articles },
		func(r *literature.SearchResult, v []*literature.Article) *literature.SearchResult {
			r.Articles = v
			return r
		},
	)
	europeAuthorsLens = L.MakeLensRef(
		func(a *literature.EuropePMCArticle) []literature.EuropePMCAuthor { return a.Authors },
		func(a *literature.EuropePMCArticle, v []literature.EuropePMCAuthor) *literature.EuropePMCArticle {
			a.Authors = v
			return a
		},
	)
	europeDateLens = L.MakeLensRef(
		func(a *literature.EuropePMCArticle) *time.Time { return a.PublishDate },
		func(a *literature.EuropePMCArticle, v *time.Time) *literature.EuropePMCArticle {
			a.PublishDate = v
			return a
		},
	)
	europeYearLens = L.MakeLensRef(
		func(a *literature.EuropePMCArticle) string { return a.PubYear },
		func(a *literature.EuropePMCArticle, v string) *literature.EuropePMCArticle { a.PubYear = v; return a },
	)
	europePMIDLens = L.MakeLensRef(
		func(a *literature.EuropePMCArticle) string { return a.PMID },
		func(a *literature.EuropePMCArticle, v string) *literature.EuropePMCArticle { a.PMID = v; return a },
	)
	europeDOILens = L.MakeLensRef(
		func(a *literature.EuropePMCArticle) string { return a.DOI },
		func(a *literature.EuropePMCArticle, v string) *literature.EuropePMCArticle { a.DOI = v; return a },
	)
	europeAuthorFullLens = L.MakeLens(
		func(a literature.EuropePMCAuthor) string { return a.FullName },
		func(a literature.EuropePMCAuthor, v string) literature.EuropePMCAuthor { a.FullName = v; return a },
	)
	europeAuthorFirstLens = L.MakeLens(
		func(a literature.EuropePMCAuthor) string { return a.FirstName },
		func(a literature.EuropePMCAuthor, v string) literature.EuropePMCAuthor { a.FirstName = v; return a },
	)
	europeAuthorLastLens = L.MakeLens(
		func(a literature.EuropePMCAuthor) string { return a.LastName },
		func(a literature.EuropePMCAuthor, v string) literature.EuropePMCAuthor { a.LastName = v; return a },
	)
)
