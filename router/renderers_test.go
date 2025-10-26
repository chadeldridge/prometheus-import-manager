package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderersRenderJSON(t *testing.T) {
	require := require.New(t)

	t.Run("valid", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := RenderJSON(w, http.StatusOK, struct{ Message string }{Message: "you did it"})
		require.NoError(err, "encode() returned an error: %s", err)

		exp := `{"Message":"you did it"}` + "\n"
		got := w.Body.String()
		require.Equal(exp, got, "encode() returned wrong body")
	})

	t.Run("invalid", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := RenderJSON(w, http.StatusOK, make(chan struct{}))
		require.Error(err, "encode() did not return an error")
		require.Equal("encoder: json: unsupported type: chan struct {}", err.Error(), "encode() returned wrong error")
	})
}

func TestRenderersReadJSON(t *testing.T) {
	require := require.New(t)
	body := strings.NewReader(`{"Message":"you did it"}` + "\n")
	req := http.Request{Body: io.NopCloser(body)}

	t.Run("valid", func(t *testing.T) {
		data, err := ReadJSON[struct{ Message string }](&req)
		require.NoError(err, "decode() returned an error: %s", err)
		require.Equal(struct{ Message string }{Message: "you did it"}, data, "decode() returned wrong data")
	})

	t.Run("invalid", func(t *testing.T) {
		data, err := ReadJSON[struct{ Data int }](&req)
		require.Error(err, "decode() did not return an error")
		require.Equal(struct{ Data int }{}, data, "decode() returned wrong data")
	})
}

/*
func testHandleTest(logger *core.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			logger.Debugf("test: %s\n", r.Method)
			err := RenderJSON(w, http.StatusOK, struct{ Message string }{Message: "you did it"})
			if err != nil {
				logger.Printf("test: %v\n", err)
			}
		})
}
*/
