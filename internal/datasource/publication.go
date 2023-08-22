package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/hasura/go-graphql-client"
)

type dictyPub struct {
	Data  *pubData `json:"data"`
	Links *links   `json:"links"`
}

type links struct {
	Self string `json:"self"`
}

type pubData struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Attributes *pub   `json:"attributes"`
}

type author struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
	Initials  string `json:"initials"`
}

type pub struct {
	Abstract      string    `json:"abstract"`
	Doi           string    `json:"doi,omitempty"`
	FullTextURL   string    `json:"full_text_url,omitempty"`
	PubmedURL     string    `json:"pubmed_url,omitempty"`
	Journal       string    `json:"journal,omitempty"`
	Issn          string    `json:"issn,omitempty"`
	Page          string    `json:"page,omitempty"`
	Pubmed        string    `json:"pubmed,omitempty"`
	Title         string    `json:"title,omitempty"`
	Source        string    `json:"source,omitempty"`
	Status        string    `json:"status,omitempty"`
	PubType       string    `json:"pub_type,omitempty"`
	Issue         string    `json:"issue,omitempty"`
	Volume        string    `json:"volume,omitempty"`
	PublishedDate *pubDate  `json:"publication_date,omitempty"`
	Authors       []*author `json:"authors,omitempty"`
}

type pubDate struct {
	time.Time
}

func (pd *pubDate) UnmarshalJSON(in []byte) error {
	t, err := time.Parse("2006-01-02", strings.ReplaceAll(string(in), `"`, ""))
	if err != nil {
		return fmt.Errorf("error in parsing time %s", err)
	}
	pd.Time = t
	return nil
}

type PubInfo struct {
	AuthorStr string
	PubmedURL string
	DoiURL    string
}

type GraphqlAuthor struct {
	FirstName string `graphql:"first_name"`
	LastName  string `graphql:"last_name"`
	Initials  string
	Rank      string
}

func (agl *GraphqlAuthor) FullName() string {
	if len(agl.Initials) > 0 {
		return fmt.Sprintf("%s %s", agl.Initials, agl.LastName)
	}
	return fmt.Sprintf("%s %s", agl.FirstName, agl.LastName)
}

type PubQuery struct {
	Publication struct {
		Id      string
		PubDate time.Time `graphql:"pub_date"`
		Doi     string
		Authors []*GraphqlAuthor
	} `graphql:"publication(id: $id)"`
}

type Publication struct {
	apiBase string
	client  *graphql.Client
}

func NewPublication(base string) *Publication {
	return &Publication{apiBase: base, client: graphql.NewClient(base, nil)}
}

func (p *Publication) ParsedInfoFromGraphql(id string) (*PubInfo, error) {
	pinfo := new(PubInfo)
	query := new(PubQuery)
	err := p.client.Query(context.Background(), query, map[string]interface{}{
		"id": graphql.ID(id),
	})
	if err != nil {
		return pinfo, fmt.Errorf("error in running graphql query %s", err)
	}
	sort.Slice(query.Publication.Authors, func(i, j int) bool {
		return len(
			query.Publication.Authors[i].Rank,
		) > len(
			query.Publication.Authors[j].Rank,
		)
	})
	pinfo.AuthorStr = fmt.Sprintf(
		"%s (%d)",
		authorStrFromGrqphql(query.Publication.Authors),
		query.Publication.PubDate.Year(),
	)
	pinfo.PubmedURL = fmt.Sprintf("https://pubmed.gov/%s", query.Publication.Id)
	pinfo.DoiURL = fmt.Sprintf("https://doi.org/%s", query.Publication.Doi)
	return pinfo, nil
}

func (p *Publication) ParsedInfo(id string) (*PubInfo, error) {
	pinfo := new(PubInfo)
	res, err := pubResp(fmt.Sprintf("%s/%s", p.apiBase, id))
	if err != nil {
		return pinfo, err
	}
	defer res.Body.Close()
	pub := new(dictyPub)
	if err := json.NewDecoder(res.Body).Decode(pub); err != nil {
		return pinfo, fmt.Errorf("error in decoding json %s", err)
	}
	pinfo.AuthorStr = fmt.Sprintf(
		"%s (%d)",
		authorStr(pub.Data.Attributes.Authors),
		pub.Data.Attributes.PublishedDate.Year(),
	)
	pinfo.PubmedURL = pub.Data.Attributes.PubmedURL
	pinfo.DoiURL = pub.Data.Attributes.FullTextURL
	return pinfo, nil
}

func pubResp(pubURL string) (*http.Response, error) {
	var r *http.Response
	parsedURL, err := url.Parse(pubURL)
	if err != nil {
		return r, fmt.Errorf("error in parsing url %s %s", pubURL, err)
	}
	res, err := http.Get(parsedURL.String())
	if err != nil {
		return res, fmt.Errorf("error in http get request with %s", err)
	}
	if res.StatusCode != 200 {
		return res,
			fmt.Errorf(
				"error fetching publication %s status code %d",
				parsedURL.String(), res.StatusCode,
			)
	}
	return res, nil
}

func authorStrFromGrqphql(author []*GraphqlAuthor) string {
	var str string
	switch len(author) {
	case 1:
		str = author[0].FullName()
	case 2:
		str = fmt.Sprintf("%s & %s", author[0].FullName(), author[1].FullName())
	default:
		str = fmt.Sprintf("%s et al.", author[0].FullName())
	}
	return str
}

func authorStr(a []*author) string {
	var str string
	switch len(a) {
	case 1:
		str = a[0].FullName
	case 2:
		str = fmt.Sprintf("%s & %s", a[0].FullName, a[1].FullName)
	default:
		str = fmt.Sprintf("%s et al.", a[0].FullName)
	}
	return str
}
