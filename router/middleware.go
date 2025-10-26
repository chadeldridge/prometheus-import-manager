package router

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/chadeldridge/prometheus-import-manager/core"
	"github.com/felixge/httpsnoop"
)

type Middleware func(http.Handler) http.Handler

// Middleware should call the next handler on success.
// doAuth := authMiddleware(logger *core.Logger, db *db.Users) // returns func(http.Handler) http.Handler
// mux.Handle("/v1/test", doAuth(handleTest(server.logger)))
//
//func newMiddleware(server HTTPServer) Middleware {
//	// Add middleware.
//	// server.Handler = someMiddleware(server)
//	return func(next http.Handler) http.Handler {
//		return http.HandlerFunc(
//			func(w http.ResponseWriter, r *http.Request) {
//				// Do something that fails.
//				//if !something {
//				//	// Return early.
//				//	http.NotFound(w, r)
//				//	return
//				//}
//
//				// Allow original handler to run.
//				next.ServeHTTP(w, r)
//			})
//	}
//}

type ReqMetrics struct {
	ClientIP     string        `json:"client_ip"`
	RequestTime  time.Time     `json:"request_time"`
	Method       string        `json:"method"`
	URI          string        `json:"uri"`
	ResponseCode int           `json:"response_code"`
	ResponseSize int64         `json:"response_size"`
	Referer      string        `json:"referer"`
	UserAgent    string        `json:"user_agent"`
	Duration     time.Duration `json:"duration"`
}

func NewReqMetrics(r *http.Request) ReqMetrics {
	return ReqMetrics{
		ClientIP:    ClientIP(r),
		RequestTime: time.Now(),
		Method:      r.Method,
		URI:         r.RequestURI,
		Referer:     r.Referer(),
		UserAgent:   r.UserAgent(),
	}
}

func ClientIP(r *http.Request) string {
	// Headers are not case sensitive. Initial caps for readability.
	if x := r.Header.Get("X-Real-IP"); x != "" {
		return x
	}
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		// The first IP in the list should be the client IP.
		return strings.Split(x, ", ")[0]
	}

	return remoteAddr(r)
}

// remoteAddr returns the remote address from the request without the port.
func remoteAddr(r *http.Request) string {
	addr := r.RemoteAddr
	if strings.Contains(addr, ":") {
		addr, _, _ := net.SplitHostPort(addr)
		return addr
	}

	return addr
}

func LoggerMiddleware(logger *core.Logger) Middleware {
	accessLogger := core.NewLogger(logger.Writer(), "cuttle-access: ", 0, false)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				// logger.Debugf("request: %s %s\n", r.Method, r.URL.Path)
				rm := NewReqMetrics(r)
				m := httpsnoop.CaptureMetrics(next, w, r)
				rm.ResponseCode = m.Code
				rm.ResponseSize = m.Written
				rm.Duration = m.Duration

				// Add request metrics to the global metrics.
				RecordRequest(rm.ResponseCode, rm.Duration)
				log, err := json.Marshal(rm)
				if err != nil {
					logger.Printf("LoggerMiddleware: failed to marshal request metrics: %v\n", err)
					return
				}

				accessLogger.Print(string(log))
			})
	}
}
