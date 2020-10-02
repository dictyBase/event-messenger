package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/dictyBase/event-messenger/internal/client"
	"github.com/dictyBase/event-messenger/internal/logger"
	"github.com/dictyBase/go-obograph/storage"
	araobo "github.com/dictyBase/go-obograph/storage/arangodb"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type ontoServer struct {
	client     *github.Client
	logger     *logrus.Entry
	dataSource storage.DataSource
}

type payload struct {
	Files []string `json:"files"`
}

func getPayload(data []byte) (*payload, error) {
	var s string
	p := new(payload)
	if err := json.Unmarshal(data, &s); err != nil {
		return p, fmt.Errorf("error in decoding json data to string %s", err)
	}
	if err := json.Unmarshal([]byte(s), p); err != nil {
		return p, fmt.Errorf("error in decoding string to structure %s", err)
	}
	return p, nil
}

func (s *ontoServer) webhookHandler(w http.ResponseWriter, r *http.Request) {
	d := &github.Deployment{}
	if err := json.NewDecoder(r.Body).Decode(d); err != nil {
		http.Error(
			w,
			fmt.Sprintf("error in decoding json %s", err),
			http.StatusInternalServerError,
		)
		return
	}
	p, err := getPayload(d.Payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.logger.Infof(
		"going to fetch the following files %s from github",
		strings.Join(p.Files, " "),
	)
	return
}

func RunOntoServer(c *cli.Context) error {
	arPort, _ := strconv.Atoi(c.String("arangodb-port"))
	cp := &araobo.ConnectParams{
		User:     c.String("arangodb-user"),
		Pass:     c.String("arangodb-pass"),
		Host:     c.String("arangodb-host"),
		Database: c.String("arangodb-database"),
		Istls:    c.Bool("is-secure"),
		Port:     arPort,
	}
	clp := &araobo.CollectionParams{
		Term:         c.String("term-collection"),
		Relationship: c.String("rel-collection"),
		GraphInfo:    c.String("cv-collection"),
		OboGraph:     c.String("obograph"),
	}
	ds, err := araobo.NewDataSource(cp, clp)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	l, err := logger.NewLogger(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	server := &ontoServer{
		dataSource: ds,
		logger:     l,
		client:     client.GetGithubClient(c.String("token")),
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/ontologies", server.webhookHandler)
	if err := http.ListenAndServe(":9945", r); err != nil {
		return cli.NewExitError(
			fmt.Sprintf("error in running webhook server %s", err),
			2,
		)
	}
	return nil
}
