package logger

import "gofr.dev/pkg/gofr"

func Logf(format string, args ...any) {
	logger := gofr.New().Logger()
	logger.Logf("[LOG] "+format, args...)
}

func Infof(format string, args ...any) {
	logger := gofr.New().Logger()
	logger.Infof("[INFO] "+format, args...)
}

func Fatalf(format string, args ...any) {
	logger := gofr.New().Logger()
	logger.Fatalf("[FATAL] "+format, args...)
}
