package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dictyBase/go-obograph/graph"
	"github.com/dictyBase/go-obograph/storage"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

type OntoServer struct {
	Client     *github.Client
	Logger     *logrus.Entry
	DataSource storage.DataSource
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

func (s *OntoServer) DeploymentWebhookHandler(w http.ResponseWriter, r *http.Request) {
	d := &github.DeploymentEvent{}
	if err := json.NewDecoder(r.Body).Decode(d); err != nil {
		http.Error(
			w,
			fmt.Sprintf("error in decoding json %s", err),
			http.StatusInternalServerError,
		)
		return
	}
	p, err := getPayload(d.GetDeployment().Payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	msg, ok := s.setDeploymentStatus(d, s.fetchAndLoadFiles(p.Files, d))
	if !ok {
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *OntoServer) setDeploymentStatus(d *github.DeploymentEvent, err error) (string, bool) {
	if err != nil {
		_, _, err := s.Client.Repositories.CreateDeploymentStatus(
			context.Background(),
			d.GetRepo().GetOwner().GetLogin(),
			d.GetRepo().GetName(),
			d.GetDeployment().GetID(),
			&github.DeploymentStatusRequest{
				State:       github.String("error"),
				Description: github.String(err.Error()),
			})
		if err != nil {
			err = errors.New(err.Error())
		}
		return fmt.Sprintf("error in loading files %s", err), false
	}
	_, _, err = s.Client.Repositories.CreateDeploymentStatus(
		context.Background(),
		d.GetRepo().GetOwner().GetLogin(),
		d.GetRepo().GetName(),
		d.GetDeployment().GetID(),
		&github.DeploymentStatusRequest{
			State:       github.String("success"),
			Description: github.String("success in loading files"),
		})
	if err != nil {
		return fmt.Sprintf(
			"files loaded but error in setting deployment status %s",
			err,
		), false
	}
	return "success in loading file", true
}

func (s *OntoServer) fetchAndLoadFiles(files []string, d *github.DeploymentEvent) error {
	s.Logger.Infof(
		"going to fetch the following files %s from github",
		strings.Join(files, " "),
	)
	var ac []*github.RepositoryContent
	for _, f := range files {
		c, _, _, err := s.Client.Repositories.GetContents(
			context.Background(),
			d.GetRepo().GetOwner().GetLogin(),
			d.GetRepo().GetName(),
			f,
			&github.RepositoryContentGetOptions{
				Ref: d.GetDeployment().GetSHA(),
			})
		if err != nil {
			return fmt.Errorf(
				"error in fetching file %s from github repo %s %s",
				f,
				d.GetRepo().GetFullName(),
				err,
			)
		}
		ac = append(ac, c)
	}
	return s.loadFiles(ac)
}

func (s *OntoServer) loadFiles(contents []*github.RepositoryContent) error {
	for _, c := range contents {
		ct, err := c.GetContent()
		if err != nil {
			return fmt.Errorf("error in getting content for %s %s", c.GetPath(), err)
		}
		if err := s.loadGraph(c, ct); err != nil {
			return err
		}
	}
	return nil
}

func (s *OntoServer) loadGraph(c *github.RepositoryContent, ct string) error {
	g, err := graph.BuildGraph(strings.NewReader(ct))
	if err != nil {
		return fmt.Errorf("error in building graph from %s %s", c.GetPath(), err)
	}
	if !s.DataSource.ExistsOboGraph(g) {
		return s.loadNewGraph(g, c)
	}
	return s.loadExistingGraph(g, c)
}

func (s *OntoServer) loadNewGraph(g graph.OboGraph, c *github.RepositoryContent) error {
	logger := s.Logger
	ds := s.DataSource
	logger.Infof("obograph %s does not exist, have to be loaded", c.GetPath())
	err := ds.SaveOboGraphInfo(g)
	if err != nil {
		return fmt.Errorf("error in saving graph %s", err)
	}
	nt, err := ds.SaveTerms(g)
	if err != nil {
		return fmt.Errorf("error in saving terms %s", err)
	}
	logger.Infof("saved %d terms", nt)
	nr, err := ds.SaveRelationships(g)
	if err != nil {
		return fmt.Errorf("error in saving relationships %s", err)
	}
	logger.Infof("saved %d relationships", nr)
	return nil
}

func (s *OntoServer) loadExistingGraph(g graph.OboGraph, c *github.RepositoryContent) error {
	logger := s.Logger
	ds := s.DataSource
	logger.Infof("obograph %s exist, have to be updated", c.GetPath())
	if err := ds.UpdateOboGraphInfo(g); err != nil {
		return fmt.Errorf("error in updating graph information %s", err)
	}
	it, ut, err := ds.SaveOrUpdateTerms(g)
	if err != nil {
		return fmt.Errorf("error in updating terms %s", err)
	}
	logger.Infof("saved: %d and updated: %d terms", it, ut)
	ur, err := ds.SaveNewRelationships(g)
	if err != nil {
		return fmt.Errorf("error in saving relationships %s", err)
	}
	logger.Infof("updated %d relationships", ur)
	return nil
}
