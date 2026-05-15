package webhook

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dictyBase/event-messenger/internal/client"
	"github.com/dictyBase/event-messenger/internal/http/server"
	"github.com/dictyBase/event-messenger/internal/logger"
	araobo "github.com/dictyBase/go-obograph/storage/arangodb"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/urfave/cli/v3"
)

const (
	exitCode          = 2
	readHeaderTimeout = 5 * time.Second
)

func RunOntoServer(_ context.Context, c *cli.Command) error {
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
		return cli.Exit(err.Error(), exitCode)
	}

	l, err := logger.NewLogger(c)
	if err != nil {
		return cli.Exit(err.Error(), exitCode)
	}

	server := &server.OntoServer{
		DataSource: ds,
		Logger:     l,
		Client:     client.GetGithubClient(c.String("token")),
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/ontologies", server.DeploymentWebhookHandler)

	tsrv := &http.Server{
		Addr:              fmt.Sprintf(":%s", c.String("port")),
		ReadHeaderTimeout: readHeaderTimeout,
	}
	if err := tsrv.ListenAndServe(); err != nil {
		return cli.Exit(
			fmt.Sprintf("error in running webhook server %s", err),
			exitCode,
		)
	}

	return nil
}
