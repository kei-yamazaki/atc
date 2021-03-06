package auth

import (
	"context"
	"net/http"
	"strconv"

	"github.com/concourse/atc/dbng"
)

type CheckBuildWriteAccessHandlerFactory interface {
	HandlerFor(delegateHandler http.Handler, rejector Rejector) http.Handler
}

type checkBuildWriteAccessHandlerFactory struct {
	buildFactory dbng.BuildFactory
}

func NewCheckBuildWriteAccessHandlerFactory(
	buildFactory dbng.BuildFactory,
) *checkBuildWriteAccessHandlerFactory {
	return &checkBuildWriteAccessHandlerFactory{
		buildFactory: buildFactory,
	}
}

func (f *checkBuildWriteAccessHandlerFactory) HandlerFor(
	delegateHandler http.Handler,
	rejector Rejector,
) http.Handler {
	return checkBuildWriteAccessHandler{
		rejector:        rejector,
		buildFactory:    f.buildFactory,
		delegateHandler: delegateHandler,
	}
}

type checkBuildWriteAccessHandler struct {
	rejector        Rejector
	buildFactory    dbng.BuildFactory
	delegateHandler http.Handler
}

func (h checkBuildWriteAccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !IsAuthenticated(r) {
		h.rejector.Unauthorized(w, r)
		return
	}

	buildIDStr := r.FormValue(":build_id")
	buildID, err := strconv.Atoi(buildIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	build, found, err := h.buildFactory.Build(buildID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	authTeam, authTeamFound := GetTeam(r)
	if authTeamFound && !authTeam.IsAuthorized(build.TeamName()) {
		h.rejector.Forbidden(w, r)
		return
	}

	ctx := context.WithValue(r.Context(), BuildContextKey, build)
	h.delegateHandler.ServeHTTP(w, r.WithContext(ctx))
}
