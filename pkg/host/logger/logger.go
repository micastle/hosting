package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	//With(args ...interface{}) Logger

	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})

	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
}

// default logging config and initializer
type DefaultLoggingConfig struct {
	Name  string
	Level zapcore.Level
}

func GetDefaultLoggingConfig(debug bool) *DefaultLoggingConfig {
	logLevel := zap.InfoLevel
	if debug {
		logLevel = zap.DebugLevel
	}
	return &DefaultLoggingConfig{
		Name:  "Default",
		Level: logLevel,
	}
}

func InitializeDefaultLogging(config *DefaultLoggingConfig) {
	InitStdOutLogger(config.Name, config.Level)
}

// Utility APIs to support existing code before refactoring complete to interface based hosting framework.
func GetLogger(name string) Logger {
	logger, _ := GetNamedLogger(name)
	if logger == nil {
		panic(fmt.Errorf("nil logger returned, log is not initialized?"))
	}

	//return newLogger(logger) // create wrapper around the raw logger
	return logger // no wrapper around the raw logger
}
func GetRawLogger(logger Logger) *zap.SugaredLogger {
	return logger.(*zap.SugaredLogger)
}
func With(logger Logger, args ...interface{}) Logger {
	return GetRawLogger(logger).With(args...)
}

// adapter APIs from pkg log

var singleton *zap.SugaredLogger
var once sync.Once
const defaultSyslogTimeFormat string = "Jan  2 15:04:05"

// InitStdOutLogger initialize a stdout logger
func InitStdOutLogger(logName string, logLevel zapcore.Level) {
	once.Do(func() {
		singleton = newStdOutLogger(logLevel, defaultSyslogTimeFormat).Named(logName)
	})
}

// GetLogger get a named logger, if name is empty, use the default logger
func GetNamedLogger(name string) (*zap.SugaredLogger, error) {
	if singleton == nil {
		return nil, fmt.Errorf("empty logger: InitLogger method should be called before calling GetLogger")
	}
	if name == "" {
		return singleton, nil
	}
	return singleton.Named(name), nil
}

// newStdOutLogger create a stdout logger with provided log level
func newStdOutLogger(level zapcore.Level, timeformat string) *zap.SugaredLogger {
	stdoutSyncer := zapcore.AddSync(os.Stdout)
	zapConf := zap.NewProductionEncoderConfig()
	zapConf.EncodeTime = getLogTimeEncoder(timeformat)
	zapConf.EncodeLevel = CustomLevelEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapConf),
		stdoutSyncer,
		level,
	)

	return zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	).Sugar()
}

func CustomLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}
func getLogTimeEncoder(format string) func(time.Time, zapcore.PrimitiveArrayEncoder) {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(format))
	}
}