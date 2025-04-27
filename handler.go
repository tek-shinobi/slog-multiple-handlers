package slogmultiplehandlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
)

type MultiTargetHandler struct {
	outputHandlers      []slog.Handler
	errorOutputHandlers []slog.Handler
}

var _ slog.Handler = (*MultiTargetHandler)(nil)

func New(outputHandlers []slog.Handler, errorOutputHandlers []slog.Handler) *MultiTargetHandler {
	h := &MultiTargetHandler{
		outputHandlers:      outputHandlers,
		errorOutputHandlers: errorOutputHandlers,
	}
	return h
}

func (h *MultiTargetHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if level >= slog.LevelError {
		for i := range h.errorOutputHandlers {
			if h.errorOutputHandlers[i].Enabled(ctx, level) {
				return true
			}
		}
		return false
	}
	for i := range h.outputHandlers {
		if h.outputHandlers[i].Enabled(ctx, level) {
			return true
		}
	}

	return false
}

func (h *MultiTargetHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	if r.Level >= slog.LevelError {
		for i := range h.errorOutputHandlers {
			if h.errorOutputHandlers[i].Enabled(ctx, r.Level) {
				err := try(func() error {
					return h.errorOutputHandlers[i].Handle(ctx, r.Clone())
				})
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
		return errors.Join(errs...)
	}

	for i := range h.outputHandlers {
		if h.outputHandlers[i].Enabled(ctx, r.Level) {
			err := try(func() error {
				return h.outputHandlers[i].Handle(ctx, r.Clone())
			})
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

func (h *MultiTargetHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	outputHandlers := mapHandlers(h.outputHandlers, func(handler slog.Handler) slog.Handler {
		return handler.WithAttrs(slices.Clone(attrs))
	})
	errorOutputHandlers := mapHandlers(h.errorOutputHandlers, func(handler slog.Handler) slog.Handler {
		return handler.WithAttrs(slices.Clone(attrs))
	})
	return &MultiTargetHandler{
		outputHandlers:      outputHandlers,
		errorOutputHandlers: errorOutputHandlers,
	}
}

func (h *MultiTargetHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	outputHandlers := mapHandlers(h.outputHandlers, func(handler slog.Handler) slog.Handler {
		return handler.WithGroup(name)
	})
	errorOutputHandlers := mapHandlers(h.errorOutputHandlers, func(handler slog.Handler) slog.Handler {
		return handler.WithGroup(name)
	})
	return &MultiTargetHandler{
		outputHandlers:      outputHandlers,
		errorOutputHandlers: errorOutputHandlers,
	}
}

func mapHandlers(handlers []slog.Handler, mapFunc func(handler slog.Handler) slog.Handler) []slog.Handler {
	newHandlers := make([]slog.Handler, len(handlers))

	for i := range handlers {
		newHandlers[i] = mapFunc(handlers[i])
	}

	return newHandlers
}

func try(callback func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("unexpected error: %+v", r)
			}
		}
	}()

	err = callback()

	return
}
