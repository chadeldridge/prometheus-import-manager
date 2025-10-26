package web

import (
	"net/http"

	"github.com/chadeldridge/prometheus-import-manager/router"
)

func handleIndex(server *router.HTTPServer) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err := router.RenderHTML(w, http.StatusOK, "pim http exporter")
			if err != nil {
				server.Logger.Printf("%s %s: %s\n", r.Method, r.RequestURI, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				// handleError(logger, w, r, http.StatusInternalServerError, "internal server error", nil)
			}
		})
}

/*
type ErrorHandler func(error) templ.Component

func handleError(
	logger *core.Logger,
	w http.ResponseWriter,
	r *http.Request,
	status int,
	pageMsg string,
	errMsg error,
) {
	logger.Printf("%s %s: %s\n", r.Method, r.RequestURI, errMsg)
	err := components.ErrorPage(status, pageMsg, errMsg).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleIndex(server *router.HTTPServer) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			claims := r.Context().Value(router.ClaimsKey).(*db.Claims)
			err := components.Page("Cuttle", components.Index(claims.Username)).Render(r.Context(), w)
			if err != nil {
				handleError(server.Logger, w, r, http.StatusInternalServerError, "internal server error", nil)
			}
		})
}
*/
