package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	cli "gopkg.in/urfave/cli.v1"
)

// GetTokenFromWeb uses config to request a Token.
// It returns the retrieved token.
func GetTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// TokenCacheFile generates a credential file path/filename.
// It returns the generated credential path/filename.
func TokenCacheFile(c *cli.Context) (string, error) {
	if c.IsSet("cache-file") {
		return c.String("cache-file"), nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("gmail.json")), err
}

// TokenFromFile retrieves a token from a given file path.
// It returns the retrieved token and any read errors encountered.
func TokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// SaveToken uses a file path to create a file and store the
// token in it.
func SaveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func GetGmailClient(c *cli.Context) (*gmail.Service, error) {
	var srv *gmail.Service
	cacheFile, err := TokenCacheFile(c)
	if err != nil {
		return srv, fmt.Errorf("error unable to set the token file path %s\n", err)
	}
	tok, err := TokenFromFile(cacheFile)
	if err != nil {
		return srv, fmt.Errorf("error unable to get token from cache file: possibly run the authorize-gmail command")
	}
	cont, err := ioutil.ReadFile(c.String("gmail-secret"))
	if err != nil {
		return srv, fmt.Errorf("error unable to read the secret json file %s\n", err)
	}
	config, err := google.ConfigFromJSON(
		cont,
		gmail.GmailSendScope,
		gmail.GmailComposeScope,
		gmail.GmailLabelsScope,
		gmail.GmailModifyScope,
		gmail.MailGoogleComScope,
	)
	if err != nil {
		return srv, fmt.Errorf("error unable to create oauth config from secret file %s\n", err)
	}
	client := config.Client(context.Background(), tok)
	srv, err = gmail.New(client)
	if err != nil {
		return srv, fmt.Errorf("error unable to set gmail client %s\n", err)
	}
	return srv, nil
}
