package datasource

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func pubTestData() ([]byte, error) {
	dir, err := os.Getwd()
	if err != nil {
		return []byte(""), fmt.Errorf("unable to get current dir %s", err)
	}
	path := filepath.Join(
		filepath.Dir(dir), "../testdata", "publication.json",
	)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return b, errors.New("unable to read test file")
	}
	return b, nil
}

func pubHandler(w http.ResponseWriter, r *http.Request) {
	b, err := pubTestData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(b); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func TestParsedInfo(t *testing.T) {
	assert := require.New(t)
	ts := httptest.NewServer(http.HandlerFunc(pubHandler))
	defer ts.Close()
	p := NewPublication(ts.URL)
	info, err := p.ParsedInfo("4893433")
	assert.NoError(err, "should not be error from parsing pubmed id")
	assert.Equal(
		info.DoiURL,
		"https://doi.org/10.1016/j.bbamcr.2018.07.017",
		"should match doi url",
	)
	assert.Equal(
		info.PubmedURL,
		"https://pubmed.gov/30048658",
		"should match pubmed url",
	)
	assert.Contains(info.AuthorStr, "Huber", "should contain author last name")
}
