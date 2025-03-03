package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

const (
	DEBUG = "debug"
	INFO = "info"
	WARN = "warn"
	ERROR = "error"
	FATAL = "fatal"
)

type Logger struct {
	zapLogger *zap.Logger
}

func new(level string, logFilePath string) (*Logger, error) {
	logLevel := parseLogLevel(level)

	var core zapcore.Core

	if logFilePath != "" {
		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		writer := zapcore.AddSync(file)
		encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		core = zapcore.NewCore(encoder, writer, logLevel)
	} else {
		writer := zapcore.Lock(os.Stdout)
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder := zapcore.NewConsoleEncoder(encoderConfig)
		core = zapcore.NewCore(encoder, writer, logLevel)
	}

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &Logger{zapLogger: zapLogger}, nil
}

// parseLogLevel - конвертирует строковый уровень логирования в zapcore.Level.
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case DEBUG:
		return zapcore.DebugLevel
	case INFO:
		return zapcore.InfoLevel
	case WARN:
		return zapcore.WarnLevel
	case ERROR:
		return zapcore.ErrorLevel
	case FATAL:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func (l *Logger) info(msg string, fields ...zap.Field) {
	l.zapLogger.Info(msg, fields...)
}

func (l *Logger) debug(msg string, fields ...zap.Field) {
	l.zapLogger.Debug(msg, fields...)
}

func (l *Logger) warn(msg string, fields ...zap.Field) {
	l.zapLogger.Warn(msg, fields...)
}

func (l *Logger) error(msg string, fields ...zap.Field) {
	l.zapLogger.Error(msg, fields...)
}

func (l *Logger) fatal(msg string, fields ...zap.Field) {
	l.zapLogger.Fatal(msg, fields...)
}

func (l *Logger) sync() {
	_ = l.zapLogger.Sync()
}
