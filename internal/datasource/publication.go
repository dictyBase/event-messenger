package datasource

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
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

type Publication struct {
	apiBase string
}

func NewPublication(base string) *Publication {
	return &Publication{apiBase: base}
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
