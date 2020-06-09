package simplelogger

import (
	"runtime/debug"

	"strings"

	"encoding/json"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

// Level is type of log levels
type Level = zapcore.Level

var (
	// DebugLevel level
	DebugLevel = Level(zap.DebugLevel)
	// InfoLevel level
	InfoLevel = Level(zap.InfoLevel)
	// WarnLevel level
	WarnLevel = Level(zap.WarnLevel)
	// ErrorLevel level
	ErrorLevel = Level(zap.ErrorLevel)
	// PanicLevel level
	PanicLevel = Level(zap.PanicLevel)
	// FatalLevel level
	FatalLevel = Level(zap.FatalLevel)
)

var (
	cfg          zap.Config
	logger       *zap.Logger
	sugar        *zap.SugaredLogger
	currentLevel Level
)

func init() {
	var err error
	cfgJson := []byte(`{
		"level": "debug",
	"outputPaths": ["stderr"],
	"errorOutputPaths": ["stderr"],
	"encoding": "console",
		"encoderConfig": {
		"messageKey": "message",
			"levelKey": "level",
			"levelEncoder": "lowercase"
	}
}`)
	currentLevel = DebugLevel

	if err = json.Unmarshal(cfgJson, &cfg); err != nil {
		panic(err)
	}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	rebuildLoggerFromCfg()
}

// SetLevel sets the log level
func SetLevel(lv Level) {
	currentLevel = lv
	cfg.Level.SetLevel(lv)
}

// GetLevel get the current log level
func GetLevel() Level {
	return currentLevel
}

// TraceError prints the stack and error
func TraceError(format string, args ...interface{}) {
	Error(string(debug.Stack()))
	Errorf(format, args...)
}

// SetOutput sets the output writer
func SetOutput(outputs []string) {
	cfg.OutputPaths = outputs
	rebuildLoggerFromCfg()
}

// ParseLevel converts string to Levels
func ParseLevel(s string) Level {
	if strings.ToLower(s) == "debug" {
		return DebugLevel
	} else if strings.ToLower(s) == "info" {
		return InfoLevel
	} else if strings.ToLower(s) == "warn" || strings.ToLower(s) == "warning" {
		return WarnLevel
	} else if strings.ToLower(s) == "error" {
		return ErrorLevel
	} else if strings.ToLower(s) == "panic" {
		return PanicLevel
	} else if strings.ToLower(s) == "fatal" {
		return FatalLevel
	}
	Errorf("ParseLevel: unknown level: %s", s)
	return DebugLevel
}

func rebuildLoggerFromCfg() {
	if newLogger, err := cfg.Build(); err == nil {
		if logger != nil {
			logger.Sync()
		}
		logger = newLogger
		setSugar(logger.Sugar())
	} else {
		panic(err)
	}
}

func Debugf(format string, args ...interface{}) {
	args = append([]interface{}{time.Now().Format("2006-01-02 15:04:05.000")}, args...)
	sugar.Debugf("%s - "+format, args...)
}

func Infof(format string, args ...interface{}) {
	args = append([]interface{}{time.Now().Format("2006-01-02 15:04:05.000")}, args...)
	sugar.Infof("%s - "+format, args...)
}

func Warnf(format string, args ...interface{}) {
	args = append([]interface{}{time.Now().Format("2006-01-02 15:04:05.000")}, args...)
	sugar.Warnf("%s - "+format, args...)
}

func Errorf(format string, args ...interface{}) {
	args = append([]interface{}{time.Now().Format("2006-01-02 15:04:05.000")}, args...)
	sugar.Errorf("%s - "+format, args...)
}

func Panicf(format string, args ...interface{}) {
	args = append([]interface{}{time.Now().Format("2006-01-02 15:04:05.000")}, args...)
	sugar.Panicf("%s - "+format, args...)
}

func Fatalf(format string, args ...interface{}) {
	args = append([]interface{}{time.Now().Format("2006-01-02 15:04:05.000")}, args...)
	sugar.Fatalf("%s - "+format, args...)
}

func Error(args ...interface{}) {
	args = append([]interface{}{time.Now().Format("2006-01-02 15:04:05.000"), " - "}, args...)
	sugar.Error(args...)
}

func Panic(args ...interface{}) {
	args = append([]interface{}{time.Now().Format("2006-01-02 15:04:05.000"), " - "}, args...)
	sugar.Panic(args...)
}

func Fatal(args ...interface{}) {
	args = append([]interface{}{time.Now().Format("2006-01-02 15:04:05.000"), " - "}, args...)
	sugar.Fatal(args...)
}

func PanicIfError(err error, args ...interface{}) {
	if len(args) == 0 {
		args = []interface{}{err}
	}
	if err != nil {
		Panic(args...)
	}
}

func PanicfIfError(err error, format string, args ...interface{}) {
	if err != nil {
		Panicf(format, args...)
	}
}

func FatalIfError(err error, args ...interface{}) {
	if len(args) == 0 {
		args = []interface{}{err}
	}
	if err != nil {
		Fatal(args...)
	}
}

func FatalfIfError(err error, format string, args ...interface{}) {
	if err != nil {
		Fatalf(format, args...)
	}
}

func setSugar(sugar_ *zap.SugaredLogger) {
	sugar = sugar_
}

type assertLogger struct{}

func (t assertLogger) Errorf(format string, args ...interface{}) {
	Panicf(format, args...)
}

func AssertEqual(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	return assert.Equal(assertLogger{}, expected, actual, msgAndArgs...)
}

func AssertEqualf(expected interface{}, actual interface{}, msg string, args ...interface{}) bool {
	return assert.Equalf(assertLogger{}, expected, actual, msg, args...)
}

func AssertNil(object interface{}, msgAndArgs ...interface{}) bool {
	return assert.Nil(assertLogger{}, object, msgAndArgs...)
}

func AssertNotNil(object interface{}, msgAndArgs ...interface{}) bool {
	return assert.NotNil(assertLogger{}, object, msgAndArgs...)
}

func AssertNilF(object interface{}, msg string, args ...interface{}) bool {
	return assert.Nilf(assertLogger{}, object, msg, args...)
}

func AssertNotNilF(object interface{}, msg string, args ...interface{}) bool {
	return assert.NotNilf(assertLogger{}, object, msg, args...)
}

func AssertTrue(value bool, msgAndArgs ...interface{}) bool {
	return assert.True(assertLogger{}, value, msgAndArgs...)
}

func AssertFalse(value bool, msgAndArgs ...interface{}) bool {
	return assert.False(assertLogger{}, value, msgAndArgs...)
}

func AssertTruef(value bool, msg string, args ...interface{}) bool {
	return assert.Truef(assertLogger{}, value, msg, args...)
}

func AssertFalsef(value bool, msg string, args ...interface{}) bool {
	return assert.Falsef(assertLogger{}, value, msg, args...)
}
