package web

import (
	"net/http"

	"github.com/chadeldridge/prometheus-import-manager/router"
)

func AddRoutes(server *router.HTTPServer) error {
	// Initialize middleware
	mwLogger := router.LoggerMiddleware(server.Logger)
	// mwAuth := router.WebAuthMiddleware(server.Logger, server.AuthDB, server.Config.Secret)

	// Create a new router group
	root, err := router.NewRouterGroup(server.Mux, "/", mwLogger)
	if err != nil {
		return err
	}

	server.Logger.Debug("adding targets routes")
	// Handle static assets
	server.Mux.Handle(
		"/sources/",
		http.StripPrefix("/sources/", http.FileServer(http.Dir(server.Config.Sources))),
	)
	server.Mux.Handle(
		"/targets/",
		http.StripPrefix("/targets/", http.FileServer(http.Dir(server.Config.TargetsDir))),
	)
	// root.GET("/index.html", handleIndex(server), mwAuth)
	root.GET("/index.html", handleIndex(server))

	return nil
}
