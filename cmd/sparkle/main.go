package main

import (
	"net/http"

	"github.com/streamingfast/logging"
	"github.com/streamingfast/sparkle/cmd/sparkle/cmd"
	"go.uber.org/zap"
)

var zlog = zap.NewNop()

func init() {
	zlog, _ = logging.ApplicationLogger("sparkle", "github.com/streamingfast/sparkle/cmd/sparkle")
}

func main() {
	go func() {
		listenAddr := "localhost:6060"
		err := http.ListenAndServe(listenAddr, nil)
		if err != nil {
			zlog.Error("unable to start profiling server", zap.Error(err), zap.String("listen_addr", listenAddr))
		}
	}()

	cmd.Execute()
}
