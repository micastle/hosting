package logger

import (
	"sync"
	"testing"

	"go.uber.org/zap/zapcore"
	"goms.io/azureml/mir/mir-vmagent/pkg/host/types"
)

var initLoggerOnce sync.Once

func InitLoggerForTest() {
	initLoggerOnce.Do(func() {
		InitStdOutLogger("UnitTest", zapcore.DebugLevel)
	})
}

func TestLoggerFactory(t *testing.T) {
	InitLoggerForTest()

	logF := NewDefaultLoggerFactory()

	logger := logF.GetDefaultLogger()
	logger.Info("test log info")
	logger.Warn("test log warn")
	logger.Error("test log error")
	logger.Infow("test log info", "key", "value")
	logger.Warnw("test log warn", "key", "value")
	logger.Errorw("test log error", "key", "value")

	logger = logF.GetLogger(types.Of(t).Name())
	logger.Info("test log info")
	logger.Warn("test log warn")
	logger.Error("test log error")
	logger.Infow("test log info", "key", "value")
	logger.Warnw("test log warn", "key", "value")
	logger.Errorw("test log error", "key", "value")

	instance := &LoggingComponent{}
	logger = logF.GetLogger(types.Of(instance).Name())
	logger.Info("test log info")
	logger.Warn("test log warn")
	logger.Error("test log error")
	logger.Infow("test log info", "key", "value")
	logger.Warnw("test log warn", "key", "value")
	logger.Errorw("test log error", "key", "value")
}

func TestLoggerUtilityAPIs(t *testing.T) {
	InitLoggerForTest()

	logger := GetLogger("MyLogger")
	logger.Info("test log info")
	logger.Warn("test log warn")
	logger.Error("test log error")
	logger.Infow("test log info", "key", "value")
	logger.Warnw("test log warn", "key", "value")
	logger.Errorw("test log error", "key", "value")

	rawLogger := GetRawLogger(logger)
	rawLogger.Info("test log info")
	rawLogger.Warn("test log warn")
	rawLogger.Error("test log error")
	rawLogger.Infow("test log info", "key", "value")
	rawLogger.Warnw("test log warn", "key", "value")
	rawLogger.Errorw("test log error", "key", "value")

	logger = With(logger, "Mode", "Test")
	logger.Info("test log info")
	logger.Warn("test log warn")
	logger.Error("test log error")
	logger.Infow("test log info", "key", "value")
	logger.Warnw("test log warn", "key", "value")
	logger.Errorw("test log error", "key", "value")
}

type LoggingComponent struct {
}

// LoggerContract method
func (lc *LoggingComponent) LoggerName() string {
	return "CustomizedLoggerName"
}
