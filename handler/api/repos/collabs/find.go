// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

// +build !oss

package collabs

import (
	"net/http"

	"github.com/joseluisq/drone/core"
	"github.com/joseluisq/drone/handler/api/render"
	"github.com/joseluisq/drone/logger"

	"github.com/go-chi/chi"
)

// HandleFind returns an http.HandlerFunc that writes a json-encoded
// repository collaborator details to the response body.
func HandleFind(
	users core.UserStore,
	repos core.RepositoryStore,
	members core.PermStore,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			login     = chi.URLParam(r, "member")
			namespace = chi.URLParam(r, "owner")
			name      = chi.URLParam(r, "name")
		)

		repo, err := repos.FindName(r.Context(), namespace, name)
		if err != nil {
			render.NotFound(w, err)
			logger.FromRequest(r).
				WithError(err).
				WithField("namespace", namespace).
				WithField("name", name).
				Debugln("api: repository not found")
			return
		}
		user, err := users.FindLogin(r.Context(), login)
		if err != nil {
			render.NotFound(w, err)
			logger.FromRequest(r).
				WithError(err).
				WithField("namespace", namespace).
				WithField("name", name).
				WithField("member", login).
				Debugln("api: user not found")
			return
		}
		member, err := members.Find(r.Context(), repo.UID, user.ID)
		if err != nil {
			render.NotFound(w, err)
			logger.FromRequest(r).
				WithError(err).
				WithField("member", login).
				WithField("namespace", namespace).
				WithField("name", name).
				Debugln("api: membership not found")
			return
		}
		render.JSON(w, member, 200)
	}
}
