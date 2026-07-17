// Package logger - Zap
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/versenilvis/log-pipeline/internal/utils"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

func InitLogger() {
	env := utils.GetEnv("APP_ENV", "development")
	isDev := env == "development"

	consoleEncoder := getConsoleEncoder(isDev)
	fileEncoder := getFileLog()

	writerFile := zapcore.AddSync(getWriterSync())
	writerConsole := zapcore.AddSync(os.Stderr)

	logLevel := zapcore.InfoLevel
	if isDev {
		logLevel = zapcore.DebugLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writerFile, zapcore.DebugLevel),
		zapcore.NewCore(consoleEncoder, writerConsole, logLevel),
	)

	options := []zap.Option{zap.AddCaller()}
	if isDev {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	Log = zap.New(core, options...)
}

// format logs a msg
func getFileLog() zapcore.Encoder {
	encodeConfig := zap.NewProductionEncoderConfig()
	encodeConfig.TimeKey = "time"
	encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encodeConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encodeConfig)
}

func getConsoleEncoder(isDev bool) zapcore.Encoder {
	encodeConfig := zapcore.EncoderConfig{
		TimeKey:          "TIME",
		LevelKey:         "LEVEL",
		NameKey:          "LOGGER",
		MessageKey:       "MSG",
		StacktraceKey:    "STACKTRACE",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalColorLevelEncoder,
		EncodeTime:       customTimeEncoder,
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		ConsoleSeparator: " | ",
	}
	if isDev {
		encodeConfig.StacktraceKey = "S"
	} else {
		encodeConfig.StacktraceKey = ""
	}
	return zapcore.NewConsoleEncoder(encodeConfig)
}

func getWriterSync() zapcore.WriteSyncer {
	// Create log directory if it doesn't exist
	path := utils.GetEnv("LOG_PATH", "logs/")
	if err := os.MkdirAll(path, 0o755); err != nil {
		fmt.Printf("WARN: Could not create log directory: %v\n", err)
	}
	logFileName := filepath.Join(path, "app.log")
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFileName,
		MaxSize:    1, // megabytes
		MaxBackups: 5,
		MaxAge:     5,    // days
		Compress:   true, // disabled by default
		LocalTime:  true,
	}
	syncfile := zapcore.AddSync(lumberjackLogger)
	return zapcore.NewMultiWriteSyncer(syncfile)
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("[%s]", t.Format("15:04:05-02/01/2006")))
}
