/*
Copyright 2019 The Feelguuds Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logs

import (
	"context"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/opentracing/opentracing-go"
	tag "github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Inspired from https://github.com/go-logr/zapr, some functions is copy from the repo.
type ILog interface {
	logr.Logger

	// Fatal logs an error message with the given message as additional context.
	//
	// The msg argument should be used to add some constant description to
	// the log line.
	Fatal(err error, msg string)

	// For wraps the current log handle with various context centric variables (utilized in request tracing)
	For(ctx context.Context) ILog

	// Warn logs a non-error message with the given key/value pairs as context.
	//
	// The msg argument should be used to add some constant description to
	// the log line.  The key/value pairs can then be used to add additional
	// variable information.  The key/value pairs should alternate string
	// keys and arbitrary values.
	Warn(msg string, keysAndValues ...interface{})

	// Debug logs a non-error message with the given key/value pairs as context.
	//
	// The msg argument should be used to add some constant description to
	// the log line.  The key/value pairs can then be used to add additional
	// variable information.  The key/value pairs should alternate string
	// keys and arbitrary values.
	Debug(msg string, keysAndValues ...interface{})
}

var (
	// JSONLogger is global json log format logr
	JSONLogger ILog

	// timeNow stubbed out for testing
	timeNow = time.Now

	// witholds a handle on a trancing object with which the library can log traces
	Span opentracing.Span

	logLevelMap = map[zapcore.Level]string {
		zapcore.DebugLevel: "debug",
		zapcore.InfoLevel: "info",
		zapcore.WarnLevel: "warn",
		zapcore.ErrorLevel: "err",
		zapcore.FatalLevel: "fatal",
		zapcore.DPanicLevel: "panic",
	}
)

const (
	CORE_LOGGER string = "CORE_LOGGER"
)

// zapLogger is a logr.Logger that uses Zap to record log.
type zapLogger struct {
	// NB: this looks very similar to zap.SugaredLogger, but
	// deals with our desire to have multiple verbosity levels.
	l   *zap.Logger
	lvl int
	span       opentracing.Span
	spanFields []zapcore.Field
}

// implement logr.Logger
var _ ILog = &zapLogger{}

// Enabled should always return true
func (l *zapLogger) Enabled() bool {
	return true
}

// Error write log message to error level
func (l *zapLogger) Error(err error, msg string, keysAndValues ...interface{}){
	var logLevelType = zapcore.ErrorLevel
	var loglevel = logLevelMap[logLevelType]

	entry := l.configureCoreLoggingLevel(msg, logLevelType)
	l.logInternal(msg, keysAndValues, entry, loglevel, err)
}

// Info write message to info level log
func (l *zapLogger) Info(msg string, keysAndVals ...interface{}) {
	var logLevelType = zapcore.InfoLevel
	var loglevel = logLevelMap[logLevelType]

	entry := l.configureCoreLoggingLevel(msg, logLevelType)
	l.logInternal(msg, keysAndVals, entry, loglevel, nil)
}

// Warn write message to warn level log
func (l *zapLogger) Warn(msg string, keysAndVals ...interface{}) {
	var logLevelType = zapcore.WarnLevel
	var loglevel = logLevelMap[logLevelType]

	entry := l.configureCoreLoggingLevel(msg, logLevelType)
	l.logInternal(msg, keysAndVals, entry, loglevel, nil)
}

// Debug write message to warn level log
func (l *zapLogger) Debug(msg string, keysAndVals ...interface{}) {
	var logLevelType = zapcore.DebugLevel
	var loglevel = logLevelMap[logLevelType]

	entry := l.configureCoreLoggingLevel(msg, logLevelType)
	l.logInternal(msg, keysAndVals, entry, loglevel, nil)
}

// Fatal writes zap  log message to error level and calls os.exit(255)
func (l *zapLogger) Fatal(err error, msg string) {
	var logLevelType = zapcore.FatalLevel
	var loglevel = logLevelMap[logLevelType]

	entry := l.configureCoreLoggingLevel(msg, logLevelType)
	l.logInternal(msg, nil, entry, loglevel, err)
	tag.Error.Set(l.span, true)

	os.Exit(255)
}

// dPanic write message to DPanicLevel level log
// we need implement this because unit test case need stub time.Now
// otherwise the ts field always changed
func (l *zapLogger) dPanic(msg string) {
	var logLevelType = zapcore.DPanicLevel
	var loglevel = logLevelMap[logLevelType]

	entry := l.configureCoreLoggingLevel(msg, logLevelType)
	l.logInternal(msg, nil, entry, loglevel, nil)
}

func (l *zapLogger) logInternal(msg string, keysAndVals []interface{}, entry zapcore.Entry, loglevel string,  err error) {
	checkedEntry := l.l.Core().Check(entry, nil)

	if entry.Level == zapcore.DPanicLevel{
		checkedEntry.Write(zap.Int("v", l.lvl))
	} else if err == nil {
		checkedEntry.Write(l.handleFields(keysAndVals)...)
	} else {
		checkedEntry.Write(l.handleFields(keysAndVals, handleError(err))...)
	}

	l.logToSpan(loglevel, msg, l.handleFields(keysAndVals)...)
}

func (l *zapLogger) configureCoreLoggingLevel(msg string, level zapcore.Level) zapcore.Entry {
	entry := zapcore.Entry{
		Level:   level,
		Time:    timeNow(),
		Message: msg,
	}
	return entry
}

// handleFields converts a bunch of arbitrary key-value pairs into Zap fields.  It takes
// additional pre-converted Zap fields, for use with automatically attached fields, like
// `error`.
func (l *zapLogger) handleFields(args []interface{}, additional ...zap.Field) []zap.Field {
	// a slightly modified version of zap.SugaredLogger.sweetenFields
	if len(args) == 0 {
		// fast-return if we have no suggared fields.
		return append(additional, zap.Int("v", l.lvl))
	}

	// unlike Zap, we can be pretty sure users aren't passing structured
	// fields (since logr has no concept of that), so guess that we need a
	// little less space.
	fields := make([]zap.Field, 0, len(args)/2+len(additional)+1)
	fields = append(fields, zap.Int("v", l.lvl))
	for i := 0; i < len(args)-1; i += 2 {
		// check just in case for strongly-typed Zap fields, which is illegal (since
		// it breaks implementation agnosticism), so we can give a better error message.
		if _, ok := args[i].(zap.Field); ok {
			l.dPanic("strongly-typed Zap Field passed to logr")
			break
		}

		// process a key-value pair,
		// ensuring that the key is a string
		key, val := args[i], args[i+1]
		keyStr, isString := key.(string)
		if !isString {
			// if the key isn't a string, stop logging
			l.dPanic("non-string key argument passed to logging, ignoring all later arguments")
			break
		}

		fields = append(fields, zap.Any(keyStr, val))
	}

	return append(fields, additional...)
}

// V return info logr.Logger  with specified level
func (l *zapLogger) V(level int) logr.Logger {
	opentracing.SetGlobalTracer(opentracing.GlobalTracer())
	return &zapLogger{
		lvl: l.lvl + level,
		l:   l.l,
		span: opentracing.StartSpan(CORE_LOGGER),
	}
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (l *zapLogger) WithSpanFields(fields ...zapcore.Field) *zap.Logger {
	l.l = l.l.With(fields...)
	return l.l
}

// WithValues return logr.Logger with some keys And Values
func (l *zapLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	l.l = l.l.With(l.handleFields(keysAndValues)...)
	return l
}

// WithName return logger Named with specified name
func (l *zapLogger) WithName(name string) logr.Logger {
	l.l = l.l.Named(name)
	return l
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (l *zapLogger) With(fields ...zapcore.Field) ILog {
	return &zapLogger{l: l.WithSpanFields(fields...), span: l.span, spanFields: l.spanFields}
}

// For decorates a log handler with various context centric variables
func (l *zapLogger) For(ctx context.Context) ILog {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		l.span = span
		if jaegerCtx, ok := span.Context().(jaeger.SpanContext); ok {
			l.spanFields = []zapcore.Field{
				zap.String("trace_id", jaegerCtx.TraceID().String()),
				zap.String("span_id", jaegerCtx.SpanID().String()),
			}
		}

		return l
	}
	return l
}

// logToSpan logs fields to span
func (l *zapLogger) logToSpan(level string, msg string, fields ...zapcore.Field) {
	// TODO rather than always converting the fields, we could wrap them into a lazy logger
	fa := fieldAdapter(make([]log.Field, 0, 2+len(fields)))
	fa = append(fa, log.String("event", msg))
	fa = append(fa, log.String("level", level))
	for _, field := range fields {
		field.AddTo(&fa)
	}
	l.span.LogFields(fa...)
}


// encoderConfig config zap encodetime format
var encoderConfig = zapcore.EncoderConfig{
	MessageKey: "msg",

	TimeKey:    "ts",
	EncodeTime: zapcore.EpochMillisTimeEncoder,
}

// NewJSONLogger creates a new json logging object using the given Zap Logger to log.
func NewJSONLogger(w zapcore.WriteSyncer, span opentracing.Span) ILog {
	l, _ := zap.NewProduction()
	if w == nil {
		w = os.Stdout
	}
	log := l.WithOptions(zap.AddCallerSkip(1),
		zap.WrapCore(
			func(zapcore.Core) zapcore.Core {
				return zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(w), zapcore.DebugLevel)
			}))

	if span == nil {
		opentracing.SetGlobalTracer(opentracing.GlobalTracer())
		span = opentracing.StartSpan(CORE_LOGGER)
	}
	return &zapLogger{
		l: log,
		span: span,
	}
}

// handleError converts an error to a named zap error
func handleError(err error) zap.Field {
	return zap.NamedError("err", err)
}

func init() {
	JSONLogger = NewJSONLogger(nil, nil)
}
