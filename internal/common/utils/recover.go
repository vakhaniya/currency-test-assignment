package utils

import (
	"log/slog"
	"runtime/debug"
)

func HandleRecover() {
	if err := recover(); err != nil {
		slog.Error(
			"Recovery",
			slog.Any("error", err),
			slog.String("stack", string(debug.Stack())),
		)
	}
}
