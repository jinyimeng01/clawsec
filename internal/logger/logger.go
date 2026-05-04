package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DBG"
	case InfoLevel:
		return "INF"
	case WarnLevel:
		return "WRN"
	case ErrorLevel:
		return "ERR"
	case FatalLevel:
		return "FTL"
	default:
		return "???"
	}
}

type Logger struct {
	level       Level
	color       bool
	json        bool
	timestamp   bool
	output      *os.File
	mu          sync.Mutex
}

var (
	defaultLogger *Logger
	once          sync.Once
)

func Init(opts ...Option) {
	once.Do(func() {
		defaultLogger = New(opts...)
	})
}

func New(opts ...Option) *Logger {
	l := &Logger{
		level:     InfoLevel,
		color:     true,
		json:      false,
		timestamp: true,
		output:    os.Stderr,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

type Option func(*Logger)

func WithLevel(level Level) Option {
	return func(l *Logger) { l.level = level }
}

func WithColor(enabled bool) Option {
	return func(l *Logger) { l.color = enabled }
}

func WithJSON(enabled bool) Option {
	return func(l *Logger) { l.json = enabled }
}

func WithOutput(w *os.File) Option {
	return func(l *Logger) { l.output = w }
}

func WithFile(path string) Option {
	return func(l *Logger) {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err == nil {
			if f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err == nil {
				l.output = f
			}
		}
	}
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf(format, args...)

	if l.json {
		fmt.Fprintf(l.output, `{"time":"%s","level":"%s","msg":"%s"}`+"\n",
			time.Now().Format(time.RFC3339),
			level.String(),
			strings.ReplaceAll(msg, `"`, `\"`),
		)
	} else {
		var prefix string
		if l.timestamp {
			prefix = time.Now().Format("2006-01-02 15:04:05") + " "
		}
		levelStr := level.String()
		if l.color {
			levelStr = colorize(level, levelStr)
		}
		fmt.Fprintf(l.output, "%s[%s] %s\n", prefix, levelStr, msg)
	}
}

func colorize(level Level, s string) string {
	switch level {
	case DebugLevel:
		return "\033[36m" + s + "\033[0m"
	case InfoLevel:
		return "\033[32m" + s + "\033[0m"
	case WarnLevel:
		return "\033[33m" + s + "\033[0m"
	case ErrorLevel:
		return "\033[31m" + s + "\033[0m"
	case FatalLevel:
		return "\033[35m" + s + "\033[0m"
	default:
		return s
	}
}

func (l *Logger) Debugf(format string, args ...interface{}) { l.log(DebugLevel, format, args...) }
func (l *Logger) Infof(format string, args ...interface{})  { l.log(InfoLevel, format, args...) }
func (l *Logger) Warnf(format string, args ...interface{})  { l.log(WarnLevel, format, args...) }
func (l *Logger) Errorf(format string, args ...interface{}) { l.log(ErrorLevel, format, args...) }
func (l *Logger) Fatalf(format string, args ...interface{}) { l.log(FatalLevel, format, args...); os.Exit(1) }

func (l *Logger) SetLevel(level Level) { l.level = level }

// Global shortcuts
func Debugf(format string, args ...interface{}) { defaultLogger.Debugf(format, args...) }
func Infof(format string, args ...interface{})  { defaultLogger.Infof(format, args...) }
func Warnf(format string, args ...interface{})  { defaultLogger.Warnf(format, args...) }
func Errorf(format string, args ...interface{}) { defaultLogger.Errorf(format, args...) }
func Fatalf(format string, args ...interface{}) { defaultLogger.Fatalf(format, args...) }
func SetLevel(level Level)                      { defaultLogger.SetLevel(level) }
