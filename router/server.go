package router

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/chadeldridge/prometheus-import-manager/core"
)

type HTTPServer struct {
	Logger  *core.Logger
	Config  *core.Config
	Handler http.Handler
	// Mux saves the http.ServeMux instance. This provides easier access to the
	// mux without having to enforce a ref type on HTTPServer.Handler everytime.
	// We can now use HTTPServer.Mux.Handle() instead of HTTPServer.Handler.(*http.ServeMux).Handle().
	Mux *http.ServeMux
}

func NewHTTPServer(logger *core.Logger, config *core.Config) HTTPServer {
	mux := http.NewServeMux()
	return HTTPServer{Logger: logger, Config: config, Handler: mux, Mux: mux}
}

func (s *HTTPServer) Start(ctx context.Context, timeoutSec int) error {
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(s.Config.APIHost, s.Config.APIPort),
		Handler: s.Handler,
	}

	// Start the server.
	srvErr := make(chan error)
	go func() {
		var err error
		if s.Config.TLSCertFile != "" && s.Config.TLSKeyFile != "" {
			s.Logger.Printf("starting HTTPS server with TLS")
			err = httpServer.ListenAndServeTLS(s.Config.TLSCertFile, s.Config.TLSKeyFile)
		} else {
			s.Logger.Printf("starting HTTP server (no TLS)")
			err = httpServer.ListenAndServe()
		}

		if err != nil {
			if err == http.ErrServerClosed {
				s.Logger.Printf("server closed")
				close(srvErr)
			} else {
				s.Logger.Debugf("http server error: %v\n", err)
				srvErr <- err
			}
		}
	}()
	s.Logger.Printf("http server listening on %s\n", httpServer.Addr)

	// Create a wait group to handle a graceful shutdown.
	var wg sync.WaitGroup
	wg.Add(1)
	wgErr := make(chan error)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(
			shutdownCtx,
			time.Duration(timeoutSec)*time.Second,
		)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			s.Logger.Debugf("http server shutdown error: %v\n", err)
			wgErr <- fmt.Errorf("http server shutdown error: %w", err)
		}
	}()
	wg.Wait()

	select {
	case err := <-srvErr:
		if err != nil {
			return err
		}
	case err := <-wgErr:
		if err != nil {
			return err
		}
	}

	return nil
}
