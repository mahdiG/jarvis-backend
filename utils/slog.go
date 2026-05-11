package utils

import (
	"context"
	"fmt"
	"jarvis/configs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/lmittmann/tint"
)

// InitSlog configures slog with a colorful tint handler that includes:
// - Clickable source file:line
// - Stack traces on error-level records
// - Colorized output
func InitSlog() {
	slog.Info("configs.Envs", "IsStackTraceEnabled", configs.Envs.IsStackTraceEnabled)
	// Don't use tint's AddSource — we add source ourselves so path format
	// is consistent with stacktrace paths (project-relative via shortenPath).
	tintHandler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
		AddSource:  true,
		NoColor:    false,
	})

	var slogHandler slog.Handler = tintHandler
	if configs.Envs.IsStackTraceEnabled {
		slogHandler = &StacktraceHandler{
			next: tintHandler,
		}
	}

	slog.SetDefault(slog.New(slogHandler))
}

// StacktraceHandler wraps a slog.Handler and adds project-relative source
// locations and stack traces for error-level records.
type StacktraceHandler struct {
	next slog.Handler
}

func (handler *StacktraceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return handler.next.Enabled(ctx, level)
}

func (handler *StacktraceHandler) Handle(ctx context.Context, record slog.Record) error {
	// Capture and override the source location using our own shortenPath logic.
	const sourceSkipFrames = 4 // slog -> this Handle -> caller
	var sourcePC [1]uintptr
	if runtime.Callers(sourceSkipFrames, sourcePC[:]) > 0 {
		frames := runtime.CallersFrames(sourcePC[:])
		frame, _ := frames.Next()
		if frame.File != "" {
			relativePath := shortenPath(frame.File)
			source := fmt.Sprintf("%s:%d", relativePath, frame.Line)
			record.AddAttrs(slog.String("source", source))
		}
	}

	// Capture stack trace only for error level and above.
	if record.Level >= slog.LevelError {
		const depth = 8
		var pcs [depth]uintptr
		framesCaptured := runtime.Callers(3, pcs[:])
		if framesCaptured > 0 {
			frames := runtime.CallersFrames(pcs[:framesCaptured])

			var stackBuffer strings.Builder
			first := true
			for {
				frame, more := frames.Next()
				// Skip tint/slog internal frames.
				if strings.Contains(frame.File, "github.com/lmittmann/tint") ||
					strings.Contains(frame.File, "log/slog") {
					if !more {
						break
					}
					continue
				}

				relativePath := shortenPath(frame.File)
				if !first {
					stackBuffer.WriteString("\n")
				}
				fmt.Fprintf(&stackBuffer, "  %s:%d %s", relativePath, frame.Line, frame.Function)
				first = false

				if !more {
					break
				}
			}

			if stackBuffer.Len() > 0 {
				record.AddAttrs(slog.String("stacktrace", stackBuffer.String()))
			}
		}
	}

	return handler.next.Handle(ctx, record)
}

func (handler *StacktraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &StacktraceHandler{
		next: handler.next.WithAttrs(attrs),
	}
}

func (handler *StacktraceHandler) WithGroup(name string) slog.Handler {
	return &StacktraceHandler{
		next: handler.next.WithGroup(name),
	}
}

// --- Error source wrapping ---

// ErrWithSource wraps an error with the source file:line where it was created.
type ErrWithSource struct {
	Err    error
	Source string
}

func (e ErrWithSource) Error() string {
	if e.Source != "" {
		return e.Err.Error() + " (" + e.Source + ")"
	}
	return e.Err.Error()
}
func (e ErrWithSource) Unwrap() error { return e.Err }

// LogValue implements slog.LogValuer so that when this error is passed as a
// slog attribute, the source origin appears as a structured field alongside the
// error message.
func (e ErrWithSource) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("message", e.Err.Error()),
		slog.String("origin", e.Source),
	)
}

// WrapError annotates err with the caller's source file and line number.
// If err is nil, it returns nil. If err already has source info, it is kept.
func WrapError(err error) error {
	if err == nil {
		return nil
	}
	var existing ErrWithSource
	if errorsAs(err, &existing) {
		return err
	}

	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return err
	}

	relativePath := shortenPath(file)
	return ErrWithSource{Err: err, Source: fmt.Sprintf("%s:%d", relativePath, line)}
}

// errorsAs is a type-parameterized helper to unwrap into a target.
func errorsAs[T error](err error, target *T) bool {
	for err != nil {
		if matched, ok := err.(T); ok {
			*target = matched
			return true
		}
		err = unwrap(err)
	}
	return false
}

func unwrap(err error) error {
	unwrapper, ok := err.(interface{ Unwrap() error })
	if !ok {
		return nil
	}
	return unwrapper.Unwrap()
}

// shortenPath converts an absolute path to a project-relative one if possible.
func shortenPath(absPath string) string {
	moduleRoot, err := findModuleRoot(absPath)
	if err == nil && moduleRoot != "" {
		relativePath, err := filepath.Rel(moduleRoot, absPath)
		if err == nil {
			return relativePath
		}
	}
	return absPath
}

// findModuleRoot walks up from filePath looking for go.mod.
func findModuleRoot(filePath string) (string, error) {
	directory := filepath.Dir(filePath)
	for {
		if directory == "/" || directory == "." {
			return "", fmt.Errorf("no go.mod found")
		}
		info, err := os.Stat(filepath.Join(directory, "go.mod"))
		if err == nil && !info.IsDir() {
			return directory, nil
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("no go.mod found")
		}
		directory = parent
	}
}
