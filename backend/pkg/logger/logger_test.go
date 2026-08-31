package logger

import (
	"log/slog"
	"testing"
)

func TestInitLogger(t *testing.T) {
	devLogger := InitLogger("development")
	if devLogger == nil {
		t.Fatal("expected non-nil dev logger")
	}

	prodLogger := InitLogger("production")
	if prodLogger == nil {
		t.Fatal("expected non-nil prod logger")
	}

	slog.Info("test message")
}
