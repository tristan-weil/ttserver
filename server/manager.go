package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"sync"
	"syscall"

	legolog "github.com/go-acme/lego/v3/log"
	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
	tthandler "github.com/tristan-weil/ttserver/server/handler"
	ttspace "github.com/tristan-weil/ttserver/server/space"
	ttprom "github.com/tristan-weil/ttserver/svc/prometheus"
	ttutils "github.com/tristan-weil/ttserver/utils"
)

type (
	Manager struct {
		configFile string

		config            *ttutils.ConfigRoot
		serveConnHandlers map[string]tthandler.IServeConnHandler
		serveConnHandler  tthandler.IServeConnHandler

		space          *ttspace.Space
		spaceServeChan chan bool
		spaceErrChan   chan error

		prometheusServer        *ttprom.PrometheusServer
		prometheusServerErrChan chan error

		signalChan chan os.Signal
		logger     *logrus.Logger
		context    context.Context

		mu sync.RWMutex
	}
)

func NewManager(ctx context.Context, serveConnHandlers map[string]tthandler.IServeConnHandler, configFile string) *Manager {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGUSR1,
	)

	return &Manager{
		configFile: configFile,

		logger:            logrus.New(),
		context:           ctx,
		serveConnHandlers: serveConnHandlers,

		prometheusServerErrChan: make(chan error),
		spaceServeChan:          make(chan bool),
		spaceErrChan:            make(chan error),
		signalChan:              signalChan,
	}
}

func (m *Manager) Start() error {
	var spaceError error

	wg := sync.WaitGroup{}
	wg.Add(4)

	//
	// err chan
	//
	go func() {
		defer wg.Done()

		for err := range m.prometheusServerErrChan {
			if err == nil {
				m.logger.
					WithField("svc", "prometheus").
					Error(err)
			}
		}
	}()

	go func() {
		defer wg.Done()

		for err := range m.spaceErrChan {
			if err == nil {
				m.logger.
					WithField("svc", "manager").
					Error(err)

				spaceError = err
			}
		}
	}()

	//
	// Initialization
	//
	m.logger.
		WithField("svc", "manager").
		Infof("reading config file: %s", m.configFile)

	if err := m.readConfig(); err != nil {
		return err
	}

	m.serveConnHandler = m.serveConnHandlers[ttutils.StringValue(m.config.Space.Handler.Name)]
	if m.serveConnHandler == nil {
		m.Shutdown()
		return fmt.Errorf("unable to find handler: %s", ttutils.StringValue(m.config.Space.Handler.Name))
	}

	if err := m.initialize(); err != nil {
		m.Shutdown()
		return err
	}

	m.logger.
		WithField("svc", "manager").
		Infof("registering prometheus metrics...")

	if err := m.serveConnHandler.RegisterPrometheusMetrics(); err != nil {
		m.Shutdown()
		return err
	}

	m.logger.
		WithField("svc", "manager").
		Infof("registering prometheus metrics... done!")

	//
	// Serving loop
	//
	go func() {
		defer wg.Done()

		m.logger.
			WithField("svc", "manager").
			Debugf("serving...")

		for {
			startServing, open := <-m.spaceServeChan
			if !open {
				break
			}

			if !startServing {
				continue
			}

			if err := m.space.Start(); err != nil {
				m.spaceErrChan <- err
				m.Shutdown()
				return
			}
		}

		m.logger.
			WithField("svc", "manager").
			Debugf("serving... done!")
	}()

	m.spaceServeChan <- true

	//
	// Signal handling loop
	//
	go func() {
		defer wg.Done()

		for s := range m.signalChan {
			m.logger.
				WithField("svc", "manager").
				Infof("signal received: %s", s.String())

			switch s {
			case syscall.SIGUSR1:
				if m.space != nil {
					m.logger.
						WithField("svc", "manager").
						Infof("flushing space's' cache...")

					m.space.GetCache().Flush()

					m.logger.
						WithField("svc", "manager").
						Debugf("flushing space's cache... done!")
				}
			case syscall.SIGHUP:
				var errinit error

				errinit = m.reset()
				if errinit == nil {
					errinit = m.initialize()
				}

				if errinit != nil {
					m.spaceErrChan <- errinit
					m.Shutdown()
					return
				}

				if !m.space.IsServing() {
					m.spaceServeChan <- true
				}
			case syscall.SIGTERM:
				fallthrough
			case syscall.SIGINT:
				m.Shutdown()
				return
			}
		}
	}()

	wg.Wait()

	return spaceError
}

func (m *Manager) initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	//
	// logger
	//
	log.SetFlags(log.Lmsgprefix)
	log.SetPrefix("certmagic = ")
	log.SetOutput(m.logger.Writer())
	legolog.Logger = log.New(m.logger.Writer(), "lego = ", log.Lmsgprefix)

	switch ttutils.StringValue(m.config.Log.Format) {
	case "text":
		m.logger.SetReportCaller(true)

		m.logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := path.Base(f.File)
				// return fmt.Sprintf("[%s()]", f.Function), fmt.Sprintf("[%s:%d]", filename, f.Line)
				return "", fmt.Sprintf("[%s:%d]", filename, f.Line)
			},
		})
	case "json":
		m.logger.SetReportCaller(true)
		
		m.logger.SetFormatter(&logrus.JSONFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				s := strings.Split(f.Function, ".")
				funcName := s[len(s)-1]
				return funcName, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
			},
		})
	}

	switch ttutils.StringValue(m.config.Log.Level) {
	case "panic":
		m.logger.SetLevel(logrus.PanicLevel)
	case "fatal":
		m.logger.SetLevel(logrus.FatalLevel)
	case "warn":
		m.logger.SetLevel(logrus.WarnLevel)
	case "debug":
		m.logger.SetLevel(logrus.DebugLevel)
	case "trace":
		m.logger.SetLevel(logrus.TraceLevel)
	case "info":
		m.logger.SetLevel(logrus.InfoLevel)
	case "error":
		m.logger.SetLevel(logrus.ErrorLevel)
	default:
		m.logger.SetLevel(logrus.InfoLevel)
	}

	switch ttutils.StringValue(m.config.Log.To) {
	case "discard":
		m.logger.SetOutput(ioutil.Discard)
	case "stderr":
		m.logger.SetOutput(os.Stderr)
	case "stdout":
		m.logger.SetOutput(os.Stdout)
	case "syslog":
		m.logger.SetOutput(ioutil.Discard)

		var syslogTp string
		var syslogAddr string

		if ttutils.StringValue(m.config.Log.SyslogAddress) != "" {
			syslogAddrSplit := strings.Split(ttutils.StringValue(m.config.Log.SyslogAddress), "://")
			syslogTp = syslogAddrSplit[0]
			syslogAddr = syslogAddrSplit[1]
		}

		hook, err := lSyslog.NewSyslogHook(syslogTp, syslogAddr, syslog.LOG_DAEMON|syslog.LOG_INFO, "")
		if err != nil {
			m.logger.
				WithField("svc", "manager").
				Errorf("unable to connect to local syslog: %s", err)
		} else {
			m.logger.AddHook(hook)
		}
	default:
		m.logger.SetOutput(os.Stderr)
	}

	// display conf
	m.logger.
		WithField("svc", "manager").
		Trace(ttutils.PrettyPrint(m.config))

	//
	// prometheus
	//
	if m.prometheusServer == nil || !m.prometheusServer.IsInitialized() {
		promConfig := &ttprom.PrometheusConfigInput{
			Config: m.config,
			Logger: m.logger.WithField("svc", "prometheus"),
		}

		m.prometheusServer = ttprom.NewPrometheus(promConfig)
		m.prometheusServer.Initialize()

		go func() {
			err := m.prometheusServer.Start()
			if err != nil {
				m.prometheusServerErrChan <- err
				return
			}

			m.prometheusServerErrChan <- nil
		}()
	}

	//
	// space
	//
	if m.space == nil {
		m.space = ttspace.NewSpace(&ttspace.SpaceConfigInput{
			Config:           m.config,
			ServeConnHandler: m.serveConnHandler,
			PrometheusFire:   m.prometheusFire,
			Logger:           m.logger.WithField("svc", "space"),
			Context:          m.context,
		})
	}

	if err := m.space.Initialize(); err != nil {
		return err
	}

	return nil
}

func (m *Manager) reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.
		WithField("svc", "manager").
		Infof("reloading...")

	//
	// config
	//
	m.logger.
		WithField("svc", "manager").
		Infof("reloading... reading config file...")

	if err := m.readConfig(); err != nil {
		return fmt.Errorf("unable to reload config file: %w", err)
	}

	m.logger.
		WithField("svc", "manager").
		Debugf("reloading... reading config file... done!")

	//
	// space
	//
	reset, err := m.space.Reset(m.config)
	if err != nil {
		return err
	}
	m.space = reset

	//
	// prometheus
	//
	if m.prometheusServer != nil && m.prometheusServer.IsInitialized() {
		reset, err := m.prometheusServer.Reset(m.config)
		if err != nil {
			return err
		}
		m.prometheusServer = reset
	}

	m.logger.
		WithField("svc", "manager").
		Infof("reloading... done!")

	return nil
}

func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.
		WithField("svc", "manager").
		Infof("stopping...")

	// space
	if m.space != nil {
		if err := m.space.Shutdown(); err != nil {
			m.logger.
				WithField("svc", "manager").
				Errorf("unable to shutdown space: %s", err)
		}
		m.space = nil
	}

	// prometheus
	if m.prometheusServer != nil {
		if err := m.prometheusServer.Shutdown(); err != nil {
			m.logger.Error(err)
		}
		m.prometheusServer = nil
	}

	close(m.spaceErrChan)
	close(m.prometheusServerErrChan)
	close(m.signalChan)
	close(m.spaceServeChan)

	m.logger.
		WithField("svc", "manager").
		Infof("stopping... done!")
}

func (m *Manager) Logger() *logrus.Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.logger.WithField("svc", "manager")
}

func (m *Manager) prometheusFire(p *ttprom.PrometheusMetric) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.prometheusServer.Fire(p)
}
