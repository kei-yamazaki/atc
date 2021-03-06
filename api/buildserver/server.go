package buildserver

import (
	"net/http"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc/auth"
	"github.com/concourse/atc/dbng"
	"github.com/concourse/atc/engine"
	"github.com/concourse/atc/worker"
)

type EventHandlerFactory func(lager.Logger, dbng.Build) http.Handler

type Server struct {
	logger lager.Logger

	externalURL string

	engine              engine.Engine
	workerClient        worker.Client
	teamFactory         dbng.TeamFactory
	buildFactory        dbng.BuildFactory
	eventHandlerFactory EventHandlerFactory
	drain               <-chan struct{}
	rejector            auth.Rejector

	httpClient *http.Client
}

func NewServer(
	logger lager.Logger,
	externalURL string,
	engine engine.Engine,
	workerClient worker.Client,
	teamFactory dbng.TeamFactory,
	buildFactory dbng.BuildFactory,
	eventHandlerFactory EventHandlerFactory,
	drain <-chan struct{},
) *Server {
	return &Server{
		logger: logger,

		externalURL: externalURL,

		engine:              engine,
		workerClient:        workerClient,
		teamFactory:         teamFactory,
		buildFactory:        buildFactory,
		eventHandlerFactory: eventHandlerFactory,
		drain:               drain,

		rejector: auth.UnauthorizedRejector{},

		httpClient: &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: 5 * time.Minute,
			},
		},
	}
}
