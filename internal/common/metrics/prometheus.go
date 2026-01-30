package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
)

type PrometheusMetricsClient struct {
	Registry *prometheus.Registry
}

type PrometheusMetricsClientConfig struct {
	Addr        string
	ServiceName string
}

var dynamicCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "dynamic_counter",
		Help: "Count custom keys",
	},
	[]string{"key"},
)

func NewPrometheusMetricsClient(config *PrometheusMetricsClientConfig) *PrometheusMetricsClient {
	client := &PrometheusMetricsClient{}
	client.initPrometheus(config)
	return &PrometheusMetricsClient{}
}

func (p *PrometheusMetricsClient) initPrometheus(conf *PrometheusMetricsClientConfig) {
	p.Registry = prometheus.NewRegistry()
	p.Registry.MustRegister(collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	// custom collectors
	p.Registry.Register(dynamicCounter)

	//metadata wrap
	prometheus.WrapRegistererWith(prometheus.Labels{
		"serviceName": conf.ServiceName},
		p.Registry)

	//export
	http.Handle("/metrics", promhttp.HandlerFor(p.Registry, promhttp.HandlerOpts{}))

	go func() {
		logrus.Fatalf("failed to start prometheus metrics endpoint, errr=%v", http.ListenAndServe(conf.Addr, nil))
	}()
}

func (p PrometheusMetricsClient) Inc(key string, value int) {
	dynamicCounter.WithLabelValues(key).Add(float64(value))
}
