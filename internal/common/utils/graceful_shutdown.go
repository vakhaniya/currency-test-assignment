package utils

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func WaitForShutdown(
	ctx context.Context,
	srv *http.Server,
	cancel context.CancelFunc,
	timeoutInSeconds int,
) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	slog.Info("shutting down...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Info("server shutdown error:" + err.Error())
	}

	slog.Info("exited gracefully")
}
