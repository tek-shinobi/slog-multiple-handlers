package slogmultiplehandlers_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	slogmultiplehandlers "github.com/tek-shinobi/slog-multiple-handlers"
)

func TestMultiTargetHandler(t *testing.T) {

	t.Run("level routing", func(t *testing.T) {
		standardBuf := &bytes.Buffer{}
		errorBuf := &bytes.Buffer{}

		standardHandler := slog.NewJSONHandler(standardBuf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		errorHandler := slog.NewJSONHandler(errorBuf, &slog.HandlerOptions{
			Level: slog.LevelError,
		})

		logger := slogmultiplehandlers.New().
			WithOutputHandlers(standardHandler).
			WithErrorOutputHandlers(errorHandler).
			Logger()

		logger.Info("info message")
		logger.Error("error message")

		// Verify standard handler got info but not error
		standardOutput := standardBuf.String()
		if !strings.Contains(standardOutput, "info message") {
			t.Error("Standard handler did not receive info message")
		}
		if strings.Contains(standardOutput, "error message") {
			t.Error("Standard handler incorrectly received error message")
		}

		// Verify error handler got error but not info
		errorOutput := errorBuf.String()
		if !strings.Contains(errorOutput, "error message") {
			t.Error("Error handler did not receive error message")
		}
		if strings.Contains(errorOutput, "info message") {
			t.Error("Error handler incorrectly received info message")
		}
	})

	t.Run("with attrs and groups", func(t *testing.T) {
		standardBuf := &bytes.Buffer{}
		errorBuf := &bytes.Buffer{}

		standardHandler := slog.NewJSONHandler(standardBuf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		errorHandler := slog.NewJSONHandler(errorBuf, &slog.HandlerOptions{
			Level: slog.LevelError,
		})

		handler := slogmultiplehandlers.New().
			WithOutputHandlers(standardHandler).
			WithErrorOutputHandlers(errorHandler)

		withAttrs := handler.WithAttrs([]slog.Attr{slog.String("testattr", "testvalue")})
		// Note: due to WithAttrs returning slog.Handler, beyond this point Logger() is not available.
		// Use slog.New() instead to generate a logger
		withGroups := withAttrs.WithGroup("testgroup")

		logger := slog.New(withGroups)
		logger.Info("grouped message", slog.String("keygrp", "valuegrp"))
		logger.Error("grouped error")

		standardOutput := standardBuf.String()
		if !strings.Contains(standardOutput, "testvalue") && !strings.Contains(standardOutput, "testattr") {
			t.Error("Standard handler did not receive attribute")
		}
		if !strings.Contains(standardOutput, "testgroup") {
			t.Error("Standard handler did not receive group")
		}

		errorOutput := errorBuf.String()
		if !strings.Contains(errorOutput, "testvalue") && !strings.Contains(standardOutput, "testattr") {
			t.Error("Error handler did not receive attribute")
		}
	})
}
