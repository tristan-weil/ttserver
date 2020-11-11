package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/goji/httpauth"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	ttutils "github.com/tristan-weil/ttserver/utils"
)

type (
	PrometheusMetric struct {
		Metric interface{}
		Labels []string
		Action string
		Values interface{}
	}

	PrometheusServer struct {
		config *ttutils.ConfigRoot
		logger *logrus.Entry

		address   string
		endpoint  string
		authBasic *PrometheusAuthBasic

		summaryMaxAge int

		httpServer  *http.Server
		metricsChan chan *PrometheusMetric

		context       context.Context
		contextCancel context.CancelFunc

		isInitialized bool
	}

	PrometheusAuthBasic struct {
		Username string
		Password string
	}

	PrometheusConfigInput struct {
		Config *ttutils.ConfigRoot
		Logger *logrus.Entry
	}
)

var (
	PrometheusRouteCacheStatusCounter *prometheus.CounterVec
	PrometheusActiveConnGauge         *prometheus.GaugeVec
	PrometheusProcessDurationSummary  *prometheus.SummaryVec
	PrometheusConnDurationSummary     *prometheus.SummaryVec

	shutdownChanPollInterval  = 500 * time.Millisecond
	shutdownChanTimeout       = 5 * time.Second
	shutdownHTTPServerTimeout = 5 * time.Second
)

func NewPrometheus(prometheusConfig *PrometheusConfigInput) *PrometheusServer {
	p := PrometheusServer{
		config: prometheusConfig.Config,
		logger: prometheusConfig.Logger,
	}

	return &p
}

func (p *PrometheusServer) Initialize() {
	if p.httpServer == nil && !p.isInitialized {
		p.logger.
			Debugf("creating...")

		p.address = p.config.Prometheus.Address
		p.endpoint = p.config.Prometheus.Endpoint
		p.summaryMaxAge = ttutils.IntValue(p.config.Prometheus.SummaryMaxAge)
		p.metricsChan = make(chan *PrometheusMetric, ttutils.IntValue(p.config.Prometheus.ChanSize))

		if p.config.Prometheus.Auth != nil &&
			p.config.Prometheus.Auth.AuthBasic != nil &&
			ttutils.NotStringEmpty(p.config.Prometheus.Auth.AuthBasic.Username) &&
			ttutils.NotStringEmpty(p.config.Prometheus.Auth.AuthBasic.Password) {

			p.authBasic = &PrometheusAuthBasic{
				Username: ttutils.StringValue(p.config.Prometheus.Auth.AuthBasic.Username),
				Password: ttutils.StringValue(p.config.Prometheus.Auth.AuthBasic.Password),
			}
		}

		go func() {
			for m := range p.metricsChan {
				switch pm := (m.Metric).(type) {
				case prometheus.Summary:
					p.logger.Tracef("firing %s(%s) on [%s]", m.Action, "prometheus.Summary", strings.Join(m.Labels, ", "))
					switch m.Action {
					case "observe":
						pm.Observe(m.Values.(float64))
					}
				case *prometheus.SummaryVec:
					p.logger.Tracef("firing %s(%s) on [%s]", m.Action, "prometheus.SummaryVec", strings.Join(m.Labels, ", "))

					if pm == nil {
						continue
					}

					switch m.Action {
					case "observe":
						pm.WithLabelValues(m.Labels...).Observe(m.Values.(float64))
					}
				case prometheus.Gauge:
					p.logger.Tracef("firing %s(%s) on [%s]", m.Action, "prometheus.Gauge", strings.Join(m.Labels, ", "))
					switch m.Action {
					case "set":
						pm.Set(m.Values.(float64))
					case "inc":
						pm.Inc()
					case "dec":
						pm.Dec()
					case "add":
						pm.Add(m.Values.(float64))
					case "sub":
						pm.Sub(m.Values.(float64))
					}
				case *prometheus.GaugeVec:
					p.logger.Tracef("firing %s(%s) on [%s]", m.Action, "prometheus.GaugeVec", strings.Join(m.Labels, ", "))

					if pm == nil {
						continue
					}

					switch m.Action {
					case "set":
						pm.WithLabelValues(m.Labels...).Set(m.Values.(float64))
					case "inc":
						pm.WithLabelValues(m.Labels...).Inc()
					case "dec":
						pm.WithLabelValues(m.Labels...).Dec()
					case "add":
						pm.WithLabelValues(m.Labels...).Add(m.Values.(float64))
					case "sub":
						pm.WithLabelValues(m.Labels...).Sub(m.Values.(float64))
					}
				case prometheus.Counter:
					p.logger.Tracef("firing %s(%s) on [%s]", m.Action, "prometheus.Counter", strings.Join(m.Labels, ", "))

					switch m.Action {
					case "inc":
						pm.Inc()
					case "add":
						pm.Add(m.Values.(float64))
					}
				case *prometheus.CounterVec:
					p.logger.Tracef("firing %s(%s) on [%s]", m.Action, "prometheus.CounterVec", strings.Join(m.Labels, ", "))

					if pm == nil {
						continue
					}

					switch m.Action {
					case "inc":
						pm.WithLabelValues(m.Labels...).Inc()
					case "add":
						pm.WithLabelValues(m.Labels...).Add(m.Values.(float64))
					}
				}
			}
		}()

		p.logger.
			Debugf("creating... done!")
	}
}

func (p *PrometheusServer) Reset(newConfig *ttutils.ConfigRoot) (*PrometheusServer, error) {
	defer func() {
		p.config = newConfig
	}()

	if (newConfig.Prometheus.Address != p.address || p.config.Prometheus.Endpoint != p.endpoint) ||
		((newConfig.Prometheus.Auth != nil && newConfig.Prometheus.Auth.AuthBasic != nil) && p.authBasic == nil) ||
		((newConfig.Prometheus.Auth != nil && newConfig.Prometheus.Auth.AuthBasic == nil) && p.authBasic != nil) ||
		((newConfig.Prometheus.Auth == nil) && p.authBasic != nil) ||
		((newConfig.Prometheus.Auth != nil && newConfig.Prometheus.Auth.AuthBasic != nil && p.authBasic != nil) && (ttutils.StringValue(newConfig.Prometheus.Auth.AuthBasic.Username) != p.authBasic.Username)) ||
		((newConfig.Prometheus.Auth != nil && newConfig.Prometheus.Auth.AuthBasic != nil && p.authBasic != nil) && (ttutils.StringValue(newConfig.Prometheus.Auth.AuthBasic.Password) != p.authBasic.Password)) {

		p.logger.
			Infof("reloading... stopping prometheus...")

		if err := p.Shutdown(); err != nil {
			return p, fmt.Errorf("unable to shutdown prometheus: %s", err)
		}

		p.logger.
			Debugf("reloading... stopping prometheus... done!")

		return nil, nil
	}

	return p, nil
}

func (p *PrometheusServer) Start() error {
	p.logger.
		Infof("starting...")

	var errListener error
	errChanPrometheus := make(chan error)

	if err := p.registerDefaultMetrics(); err != nil {
		return err
	}

	listenCtx, listenCancelCtx := context.WithCancel(context.Background())
	p.context = listenCtx
	p.contextCancel = listenCancelCtx

	httpMux := http.NewServeMux()

	if p.authBasic != nil {
		authOpts := httpauth.AuthOptions{
			User:     p.authBasic.Username,
			Password: p.authBasic.Password,
		}

		httpMux.Handle(p.endpoint, httpauth.BasicAuth(authOpts)(promhttp.Handler()))
	} else {
		httpMux.Handle(p.endpoint, promhttp.Handler())
	}

	p.httpServer = &http.Server{Addr: p.address, Handler: httpMux}

	go func() {
		if err := p.httpServer.ListenAndServe(); err != nil {
			select {
			case <-p.context.Done():
				errChanPrometheus <- nil
				return
			default:
				errChanPrometheus <- err
				return
			}
		}

		errChanPrometheus <- nil
	}()

	p.isInitialized = true

	p.logger.
		Infof("starting... done!")

	// blocking
	for err := range errChanPrometheus {
		if err != nil {
			errListener = err
		}
	}

	// shutting down
	p.httpServer = nil

	ctxShutdown, ctxShutdownCancel := context.WithTimeout(context.Background(), shutdownChanTimeout)
	defer ctxShutdownCancel()

	ticker := time.NewTicker(shutdownChanPollInterval)
	defer ticker.Stop()

	for {
		if len(p.metricsChan) == 0 {
			break
		}

		select {
		case <-ctxShutdown.Done():
			break
		case <-ticker.C:
		}
	}

	close(p.metricsChan)
	p.metricsChan = nil

	return errListener
}

func (p *PrometheusServer) Shutdown() error {
	p.logger.
		Infof("stopping...")

	if p.httpServer != nil && p.isInitialized {
		defer func() {
			p.contextCancel()
		}()

		p.isInitialized = false

		ctxShutdown, _ := context.WithTimeout(p.context, shutdownHTTPServerTimeout)
		if err := p.httpServer.Shutdown(ctxShutdown); err != nil {
			return fmt.Errorf("unable to shutdown prometheus: %s", err)
		}
	}

	p.logger.
		Infof("stopping... done!")

	return nil
}

func (p *PrometheusServer) registerDefaultMetrics() error {
	// update conf
	prometheusMetricRouteCacheStatusOpts := prometheus.Opts{
		Name: "ttserver_route_cache_status_total",
		Help: "The total number of request per cache status and per route",
	}
	PrometheusRouteCacheStatusCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts(prometheusMetricRouteCacheStatusOpts),
		[]string{"route", "status"},
	)

	prometheusMetricActiveConnOpts := prometheus.Opts{
		Name: "ttserver_active_conn",
		Help: "The current number of active connections",
	}
	PrometheusActiveConnGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts(prometheusMetricActiveConnOpts),
		[]string{},
	)

	prometheusMetricProcessDurationOpts := prometheus.SummaryOpts{
		Name:       "ttserver_process_duration_microseconds",
		Help:       "The process duration in microseconds",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		MaxAge:     time.Duration(p.summaryMaxAge) * time.Second,
	}
	PrometheusProcessDurationSummary = prometheus.NewSummaryVec(
		prometheusMetricProcessDurationOpts,
		[]string{"route", "code"},
	)

	prometheusMetricConnDurationOpts := prometheus.SummaryOpts{
		Name:       "ttserver_request_duration_microseconds",
		Help:       "The request duration in microseconds",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		MaxAge:     time.Duration(p.summaryMaxAge) * time.Second,
	}
	PrometheusConnDurationSummary = prometheus.NewSummaryVec(
		prometheusMetricConnDurationOpts,
		[]string{"code"},
	)

	// register
	for collector, opts := range map[prometheus.Collector]prometheus.Opts{
		PrometheusRouteCacheStatusCounter: prometheusMetricRouteCacheStatusOpts,
		PrometheusActiveConnGauge:         prometheusMetricActiveConnOpts,
	} {
		if err := prometheus.Register(collector); err != nil {
			return fmt.Errorf("unable to register %s, %s", opts.Name, err)
		}
	}

	for collector, opts := range map[prometheus.Collector]prometheus.SummaryOpts{
		PrometheusConnDurationSummary:    prometheusMetricConnDurationOpts,
		PrometheusProcessDurationSummary: prometheusMetricProcessDurationOpts,
	} {
		if err := prometheus.Register(collector); err != nil {
			return fmt.Errorf("unable to register %s, %s", opts.Name, err)
		}
	}

	return nil
}

func (p *PrometheusServer) IsInitialized() bool {
	return p.httpServer != nil && p.isInitialized
}

func (p *PrometheusServer) Fire(metric *PrometheusMetric) error {
	if p.httpServer != nil && p.isInitialized {
		select {
		case p.metricsChan <- metric:
			return nil
		default:
			return fmt.Errorf("channel is full")
		}
	}

	return nil
}
