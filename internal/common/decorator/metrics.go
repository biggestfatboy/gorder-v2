package decorator

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type MetricsClient interface {
	Inc(key string, value int)
}

type queryMetricsDecorator[C, R any] struct {
	base   QueryHandler[C, R]
	client MetricsClient
}

func (q queryMetricsDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	start := time.Now()
	actionNmae := strings.ToLower(generateActionName(cmd))
	defer func() {
		end := time.Since(start)
		q.client.Inc(fmt.Sprintf("querys.%x.duration", actionNmae), int(end.Seconds()))
		if err == nil {
			q.client.Inc(fmt.Sprintf("querys.%x.success", actionNmae), 1)
		} else {
			q.client.Inc(fmt.Sprintf("querys.%x.failure", actionNmae), 1)
		}
	}()
	return q.base.Handle(ctx, cmd)
}
