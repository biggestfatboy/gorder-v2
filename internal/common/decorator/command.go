package decorator

import (
	"context"

	"github.com/sirupsen/logrus"
)

type CommandHandler[C any, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

func ApplyCommandDecorators[C, R any](handler CommandHandler[C, R], logger *logrus.Entry, metricsClient MetricsClient) QueryHandler[C, R] {
	return commandLoggingDecorator[C, R]{
		logger: logger,
		base: commandMetricsDecorator[C, R]{
			base:   handler,
			client: metricsClient,
		},
	}
}
