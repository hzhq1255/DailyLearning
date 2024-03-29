package log

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type LoggerInf interface {
	WithName(name string) LoggerInf
	WithKeysAndValues(keysAndValues ...any) LoggerInf
	StructLoggerInf
	FormatLoggerInf
}

type StructLoggerInf interface {
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Debugw(msg string, keysAndValues ...any)
	Fatalw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
	ErrorStackW(error error, msg string, keysAndValues ...any)
}

type FormatLoggerInf interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Debugf(format string, args ...any)
	Fatalf(format string, args ...any)
	Errorf(format string, args ...any)
	ErrorStackF(error error, format string, args ...any)
}

type LoggerType string

const (
	SlogLoggerType LoggerType = "slog"
	ZapLoggerType  LoggerType = "zap"
)

type Logger struct {
	Logger LoggerInf
}

func NewLogger(level Level) LoggerInf {
	return &Logger{Logger: newSlogLogger(level, true)}
}

func NewLoggerFromType(level Level, loggerType LoggerType) LoggerInf {
	switch loggerType {
	case SlogLoggerType:
		return &Logger{Logger: newSlogLogger(level, true)}
	case ZapLoggerType:
		return &Logger{Logger: newZapLogger(level)}
	}
	return NewLogger(level)
}

func (l *Logger) WithName(name string) LoggerInf {
	return &Logger{l.Logger.WithName(name)}
}

func (l *Logger) WithKeysAndValues(keysAndValues ...any) LoggerInf {
	return &Logger{l.Logger.WithKeysAndValues(keysAndValues...)}
}

func (l *Logger) Infow(msg string, keysAndValues ...any) {
	l.Logger.Infow(msg, keysAndValues...)
}

func (l *Logger) Warnw(msg string, keysAndValues ...any) {
	l.Logger.Warnw(msg, keysAndValues...)
}

func (l *Logger) Debugw(msg string, keysAndValues ...any) {
	l.Logger.Debugw(msg, keysAndValues...)
}

func (l *Logger) Fatalw(msg string, keysAndValues ...any) {
	l.Logger.Fatalw(msg, keysAndValues...)
}

func (l *Logger) Errorw(msg string, keysAndValues ...any) {
	l.Logger.Errorw(msg, keysAndValues...)
}

func (l *Logger) ErrorStackW(error error, msg string, keysAndValues ...any) {
	l.Logger.ErrorStackW(error, msg, keysAndValues...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.Logger.Infof(format, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.Logger.Warnf(format, args...)
}

func (l *Logger) Debugf(format string, args ...any) {
	l.Logger.Debugf(format, args...)
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.Logger.Fatalf(format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.Logger.Errorf(format, args...)
}

func (l *Logger) ErrorStackF(error error, format string, args ...any) {
	l.Logger.ErrorStackF(error, format, args...)
}

func printStack(skip int) {
	pcs := make([]uintptr, 32)
	numFrames := runtime.Callers(skip, pcs)
	for numFrames == len(pcs) {
		pcs = make([]uintptr, len(pcs)*2)
		numFrames = runtime.Callers(skip, pcs)
	}
	pcs = pcs[:numFrames]
	frames := runtime.CallersFrames(pcs)
	sb := strings.Builder{}
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		sb.WriteString(frame.Function)
		sb.WriteByte('\n')
		sb.WriteByte('\t')
		sb.WriteString(frame.File)
		sb.WriteByte(':')
		sb.WriteString(strconv.Itoa(frame.Line))
		sb.WriteByte('\n')
	}
	_, _ = fmt.Fprintln(os.Stderr, sb.String())
}
