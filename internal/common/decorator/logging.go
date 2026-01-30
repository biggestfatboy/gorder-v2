package decorator

import (
	"context"
	"fmt"
	"github.com/biggestfatboy/gorder-v2/common/logging"
	"github.com/bytedance/gopkg/util/logger"
	"strings"

	"github.com/sirupsen/logrus"
)

type queryLoggingDecorator[C, R any] struct {
	logger *logrus.Logger
	base   QueryHandler[C, R]
}

func (q queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	fields := logrus.Fields{
		"query":      generateActionName(cmd),
		"query_body": fmt.Sprintf("%#v", cmd),
	}

	logger.Debug("Executing query")
	defer func() {
		if err == nil {
			logging.Infof(ctx, fields, "%s", "Query execute successfully")
		} else {
			logging.Errof(ctx, fields, "Failed to execute query, err=%v", err)
		}
	}()
	res, err := q.base.Handle(ctx, cmd)
	return res, err
}

type commandLoggingDecorator[C, R any] struct {
	logger *logrus.Logger
	base   CommandHandler[C, R]
}

func (q commandLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	fields := logrus.Fields{
		"command":      generateActionName(cmd),
		"command_body": fmt.Sprintf("%#v", cmd),
	}

	logger.Debug("Executing command")
	defer func() {
		if err == nil {
			logging.Infof(ctx, fields, "%s", "Query Command successfully")
		} else {
			logging.Errof(ctx, fields, "Failed to execute Command, err=%v", err)
		}
	}()
	res, err := q.base.Handle(ctx, cmd)
	return res, err
}

func generateActionName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
}
