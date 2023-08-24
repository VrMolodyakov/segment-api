package logging

import (
	"io"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerCfg struct {
	Development bool
	Level       string
}

type Logger interface {
	InitLogger(console io.Writer, files ...io.Writer)
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	DPanic(args ...interface{})
	DPanicf(template string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
	Printf(template string, args ...interface{})
	Infow(msg string, keysAndValues ...interface{})
}

type apiLogger struct {
	cfg         LoggerCfg
	sugarLogger *zap.SugaredLogger
}

func NewLogerConfig(
	development bool,
	level string,
) LoggerCfg {
	return LoggerCfg{
		Development: development,
		Level:       level,
	}
}

func NewLogger(cfg LoggerCfg) *apiLogger {
	return &apiLogger{cfg: cfg}
}

var loggerLevelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func (l *apiLogger) getLoggerLevel(cfg LoggerCfg) zapcore.Level {
	level, exist := loggerLevelMap[cfg.Level]
	if !exist {
		return zapcore.DebugLevel
	}

	return level
}

func (l *apiLogger) InitLogger(console io.Writer, files ...io.Writer) {
	logLevel := l.getLoggerLevel(l.cfg)
	var pe zapcore.EncoderConfig
	if l.cfg.Development {
		pe = zap.NewDevelopmentEncoderConfig()
	} else {
		pe = zap.NewProductionEncoderConfig()
	}
	pe.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC1123)
	fileEncoder := zapcore.NewJSONEncoder(pe)

	pe.EncodeCaller = func(caller zapcore.EntryCaller, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(caller.TrimmedPath())
		encoder.AppendString("|")
	}

	pe.EncodeTime = zapcore.TimeEncoderOfLayout("02/01 15:04:05 -0700") // "02/01/2006 15:04:05"
	pe.ConsoleSeparator = " "
	pe.EncodeName = func(n string, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(n)
		enc.AppendString("|")
	}

	pe.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("|")
		enc.AppendString(l.CapitalString())
		enc.AppendString("|")
	}

	consoleEncoder := zapcore.NewConsoleEncoder(pe)
	cores := make([]zapcore.Core, len(files)+1)

	cores[0] = zapcore.NewCore(consoleEncoder,
		zapcore.AddSync(console),
		logLevel,
	)

	for i := range files {
		cores[i+1] = zapcore.NewCore(fileEncoder,
			zapcore.AddSync(files[i]),
			logLevel,
		)
	}

	l.sugarLogger = zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
	).Sugar()
}

func (l *apiLogger) Debug(args ...interface{}) {
	l.sugarLogger.Debug(args...)
}

func (l *apiLogger) Debugf(template string, args ...interface{}) {
	l.sugarLogger.Debugf(template, args...)
}

func (l *apiLogger) Info(args ...interface{}) {
	l.sugarLogger.Info(args...)
}

func (l *apiLogger) Infof(template string, args ...interface{}) {
	l.sugarLogger.Infof(template, args...)
}

func (l *apiLogger) Printf(template string, args ...interface{}) {
	l.sugarLogger.Infof(template, args...)
}

func (l *apiLogger) Warn(args ...interface{}) {
	l.sugarLogger.Warn(args...)
}

func (l *apiLogger) Warnf(template string, args ...interface{}) {
	l.sugarLogger.Warnf(template, args...)
}

func (l *apiLogger) Error(args ...interface{}) {
	l.sugarLogger.Error(args...)
}

func (l *apiLogger) Errorf(template string, args ...interface{}) {
	l.sugarLogger.Errorf(template, args...)
}

func (l *apiLogger) DPanic(args ...interface{}) {
	l.sugarLogger.DPanic(args...)
}

func (l *apiLogger) DPanicf(template string, args ...interface{}) {
	l.sugarLogger.DPanicf(template, args...)
}

func (l *apiLogger) Panic(args ...interface{}) {
	l.sugarLogger.Panic(args...)
}

func (l *apiLogger) Panicf(template string, args ...interface{}) {
	l.sugarLogger.Panicf(template, args...)
}

func (l *apiLogger) Fatal(args ...interface{}) {
	l.sugarLogger.Fatal(args...)
}

func (l *apiLogger) Fatalf(template string, args ...interface{}) {
	l.sugarLogger.Fatalf(template, args...)
}

func (l *apiLogger) Infow(msg string, keysAndValues ...interface{}) {
	l.sugarLogger.Infow(msg, keysAndValues...)
}
