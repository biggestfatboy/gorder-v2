package logging

import (
	"context"
	"github.com/biggestfatboy/gorder-v2/common/tracing"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"strconv"
	"time"
)

//要么用logging.Infof, Warnf..
//或者直接加到hook里面，用logrus.Infof...

func Init() {
	SetFormatter(logrus.StandardLogger())
	logrus.SetLevel(logrus.DebugLevel)
	logrus.AddHook(&traceHook{})
}

func SetFormatter(logger *logrus.Logger) {
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyMsg:   "message",
		},
	})

	if isLocal, _ := strconv.ParseBool(os.Getenv("LOCAL_ENV")); isLocal {
		logger.SetFormatter(&prefixed.TextFormatter{
			ForceColors:     true,
			ForceFormatting: true,
			TimestampFormat: time.RFC3339,
		})
	}
}

func InfoWithCost(ctx context.Context, fields logrus.Fields, start time.Time, format string, args ...any) {
	fields["cost"] = time.Since(start).Milliseconds()
	Infof(ctx, fields, format, args)
}

func logf(ctx context.Context, level logrus.Level, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Logf(level, format, args)
}

func Infof(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Infof(format, args)
}

func Errof(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Errorf(format, args)
}

func Warnf(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Warnf(format, args)
}

func Panicf(ctx context.Context, fields logrus.Fields, format string, args ...any) {
	logrus.WithContext(ctx).WithFields(fields).Panicf(format, args)
}

type traceHook struct {
}

func (t traceHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (t traceHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		entry.Data["trace"] = tracing.TraceID(entry.Context)
		entry = entry.WithTime(time.Now())
	}
	return nil
}
