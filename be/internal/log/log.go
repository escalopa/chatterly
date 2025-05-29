package log

import (
	"log/slog"
	"os"
	"strings"
)

var l *slog.Logger

func init() {
	var level slog.Level

	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelWarn
	}

	l = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
}

func Debug(msg string, args ...any) {
	l.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	l.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	l.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	l.Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	l.Error(msg, args...)
	os.Exit(1)
}

func UserID(userID string) slog.Attr {
	return slog.String("user_id", userID)
}

func ChatID(chatID string) slog.Attr {
	return slog.String("chat_id", chatID)
}

func Err(err error) slog.Attr {
	return slog.String("err", err.Error())
}

func String(key string, value string) slog.Attr {
	return slog.String(key, value)
}
