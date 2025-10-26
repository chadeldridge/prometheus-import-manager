package router

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"testing"

	"github.com/chadeldridge/prometheus-import-manager/core"
	"github.com/stretchr/testify/require"
)

func testRunServer(ctx context.Context, srv *HTTPServer, timeout int) error {
	// Capture the interrupt signal to gracefully shutdown the server.
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan error)

	go func() {
		err := srv.Start(ctx, timeout)
		ch <- err
	}()
	cancel()

	err := <-ch
	return err
}

func TestServerStart(t *testing.T) {
	require := require.New(t)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var out bytes.Buffer
	l := core.NewLogger(&out, "test_server: ", log.LstdFlags, false)

	// Setup the configuration.
	core.SetTester(core.MockTester)
	core.SetReader(core.MockReader)
	core.MockWriteFile("/tmp/test_config.yaml", core.MockTestConfigYAML, true, nil)
	core.MockWriteFile("/tmp/test_cert.cert", core.MockTestCert, true, nil)
	core.MockWriteFile("/tmp/test_key.pem", core.MockTestKey, true, nil)

	flags := core.Flags{"config_file": "/tmp/test_config.yaml", "command": "run"}
	config, err := core.NewConfig(l, flags, map[string]string{})
	require.NoError(err, "NewConfig() returned an error: %s", err)

	// Update logger with config value.
	l.DebugMode = config.Debug

	// Setup the HTTP server.
	srv := NewHTTPServer(l, config)
	// Add routes here if neccasary.
	require.NoError(err, "Build() returned an error: %s", err)

	t.Run("Start", func(t *testing.T) {
		// Test the Start function.
		err := testRunServer(ctx, &srv, 5)
		require.NoError(err, "Start() returned an error: %s", err)
	})

	fmt.Println(out.String())
}
