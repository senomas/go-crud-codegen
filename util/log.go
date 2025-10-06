package util

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
)

type StackTraceHandler struct {
	slog.Handler
}

func (h *StackTraceHandler) Handle(ctx context.Context, r slog.Record) error {
	// Extract error field if any
	r.Attrs(func(a slog.Attr) bool {
		if _, ok := a.Value.Any().(error); ok {
			stack := string(debug.Stack())
			r.Add("stack", slog.StringValue(stack))
		}
		return true
	})
	return h.Handler.Handle(ctx, r)
}

// PrettyHandler is a custom slog.Handler that outputs colorful, human-readable logs
type PrettyHandler struct {
	opts slog.HandlerOptions
}

// NewPrettyHandler returns a console-friendly slog handler
func NewPrettyHandler(opts *slog.HandlerOptions) *PrettyHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &PrettyHandler{opts: *opts}
}

func (h *PrettyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	min := slog.LevelInfo
	if h.opts.Level != nil {
		min = h.opts.Level.Level()
	}
	return level >= min
}

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	var b strings.Builder

	// Timestamp
	b.WriteString(r.Time.Format("15:04:05"))
	b.WriteString(" ")

	// Level color
	levelColor := map[slog.Level]string{
		slog.LevelDebug: "\033[36mDEBUG\033[0m",
		slog.LevelInfo:  "\033[32mINFO\033[0m",
		slog.LevelWarn:  "\033[33mWARN\033[0m",
		slog.LevelError: "\033[31mERROR\033[0m",
	}
	levelStr := levelColor[r.Level]
	if levelStr == "" {
		levelStr = r.Level.String()
	}
	b.WriteString(levelStr)
	b.WriteString(" ")

	// Message
	b.WriteString(r.Message)

	// Attributes
	r.Attrs(func(a slog.Attr) bool {
		val := a.Value.String()

		// If the value is an error, add stack trace
		if err, ok := a.Value.Any().(error); ok {
			stack := strings.TrimSpace(string(debug.Stack()))
			val = fmt.Sprintf("%v\n%s", err, indent(stack, "    "))
		}

		b.WriteString(fmt.Sprintf("\n  • %s: %s", a.Key, val))
		return true
	})

	fmt.Fprintln(os.Stdout, b.String())
	return nil
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *PrettyHandler) WithGroup(name string) slog.Handler       { return h }

func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}
