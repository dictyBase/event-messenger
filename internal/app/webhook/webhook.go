package webhook

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dictyBase/event-messenger/internal/client"
	"github.com/dictyBase/event-messenger/internal/http/server"
	"github.com/dictyBase/event-messenger/internal/logger"
	araobo "github.com/dictyBase/go-obograph/storage/arangodb"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/urfave/cli"
)

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
	server := &server.OntoServer{
		DataSource: ds,
		Logger:     l,
		Client:     client.GetGithubClient(c.String("token")),
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/ontologies", server.DeploymentWebhookHandler)
	if err := http.ListenAndServe(":9945", r); err != nil {
		return cli.NewExitError(
			fmt.Sprintf("error in running webhook server %s", err),
			2,
		)
	}
	return nil
}
