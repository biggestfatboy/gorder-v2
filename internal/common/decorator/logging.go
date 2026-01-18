package decorator

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type queryLoggingDecorator[C, R any] struct {
	logger *logrus.Entry
	base   QueryHandler[C, R]
}

func (q queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	logger := q.logger.WithFields(logrus.Fields{
		"query":      generateActionName(cmd),
		"query_body": fmt.Sprintf("%#v", cmd),
	})

	logger.Debug("Executing query")
	defer func() {
		if err == nil {
			logger.Info("Query execute successfully")
		} else {
			logger.Error("Failed to execute query", err)
		}
	}()
	res, err := q.base.Handle(ctx, cmd)
	return res, err
}

type commandLoggingDecorator[C, R any] struct {
	logger *logrus.Entry
	base   CommandHandler[C, R]
}

func (q commandLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	logger := q.logger.WithFields(logrus.Fields{
		"command":      generateActionName(cmd),
		"command_body": fmt.Sprintf("%#v", cmd),
	})

	logger.Debug("Executing command")
	defer func() {
		if err == nil {
			logger.Info("Command execute successfully")
		} else {
			logger.Error("Failed to execute command", err)
		}
	}()
	res, err := q.base.Handle(ctx, cmd)
	return res, err
}

func generateActionName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
}
