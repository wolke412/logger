package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"time"
	"github.com/wolke412/paint"
)

type Current struct {
	file *os.File
}

var current Current = Current{
	file: nil,
}

// Regex to match ANSI escape codes
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripAnsi(input string) string {
	return ansiEscape.ReplaceAllString(input, "")
}

type CustomHandler struct {
	w     io.Writer
	color bool
}

func NewCustomHandler(w io.Writer, color bool) *CustomHandler {
	return &CustomHandler{w: w, color: color}
}

func (h *CustomHandler) Enabled(_ context.Context, level slog.Level) bool {
	return true
}

func (h *CustomHandler) Handle(_ context.Context, r slog.Record) error {
	ts := r.Time.Format("2006-01-02 15:04:05.000") // Custom timestamp format
	level := r.Level.String()
	msg := r.Message

	// Strip incoming color if any
	if !h.color {
		msg = stripAnsi(msg)
	}
	attrs := ""
	r.Attrs(func(a slog.Attr) bool {

		key := (a.Key)
		val := (fmt.Sprint(a.Value))

		if !h.color {
			key = stripAnsi(key)
			val = stripAnsi(val)
		}

		attrs += fmt.Sprintf(" %s=%v", key, val)
		return true
	})

	levelStr := level
	if h.color {
		FG := paint.WHITE
		BG := paint.NOCOLOR
		if level == "ERROR" {
			FG = paint.RED
		}
		if level == "WARN" {
			FG = paint.YELLOW
		}
		if level == "INFO" {
			FG = paint.CYAN
		}
		levelStr = paint.Both(FG, BG, level)
	}

	_, err := fmt.Fprintf(h.w, "[%s] %s %s%s\n", ts, levelStr, msg, attrs)
	return err
}

// For simplicity, ignore structured nesting
func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

// For simplicity, ignore grouping
func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return h
}

type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(hs ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: hs}
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	var finalErr error
	for _, h := range m.handlers {
		if err := h.Handle(ctx, r); err != nil {
			finalErr = err
		}
	}
	return finalErr
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}

var LOGS_FOLDER string = "logs"

func SetPath(path string)  {
	LOGS_FOLDER = path
}

func Init() {

	// creates today
	createLogPath()

	createLogPathDaily()

	defer current.file.Close()

	// locks forever
	select {}
}

func setLoggerCallbacks() {
	// Console with colors
	consoleHandler := NewCustomHandler(os.Stdout, true)

	// File without colors
	fileHandler := NewCustomHandler(current.file, false)

	multihandler := NewMultiHandler(consoleHandler, fileHandler)

	// Combined writer: call both handlers
	handler := slog.New(multihandler)

	log.SetOutput(io.Discard) // avoid double printing
	slog.SetDefault(handler)
}

func createLogPathDaily() {

	for {
		now := time.Now()

		// Calculate the duration until the next midnight
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		durationUntilMidnight := time.Until(nextMidnight)

		// Wait until midnight
		time.Sleep(durationUntilMidnight)

		// Create the logging path
		err := createLogPath()

		if err != nil {
			log.Printf("Error creating log path: %v\n", err)
		} else {
			log.Println("Log path created successfully.")
		}
	}
}

func setLoggerPath(path string) {

	if current.file != nil {

		if current.file.Close() != nil {
			log.Println("Error closing log file.")
		}
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	// grabs file ref to gracefully handle it
	current.file = f

	setLoggerCallbacks()

	log.Println("Daily Logging started")
}

func createLogPath() error {

	// Get the current date
	now := time.Now()
	year := now.Format("2006") // Year in YYYY format
	month := now.Format("01")  // Month in MM format
	day := now.Format("02")    // Day in DD format

	// Build the path Year/Month/Day.txt
	dirPath := filepath.Join(LOGS_FOLDER, year, month)
	filePath := filepath.Join(dirPath, day+".txt")

	// Create the directory if it doesn't exist
	err := os.MkdirAll(dirPath, os.ModePerm)

	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	setLoggerPath(filePath)

	return nil
}
