package space

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	tthandler "github.com/tristan-weil/ttserver/server/handler"
	ttcache "github.com/tristan-weil/ttserver/svc/cache"
	ttcron "github.com/tristan-weil/ttserver/svc/cron"
	ttprom "github.com/tristan-weil/ttserver/svc/prometheus"
	tttcpl "github.com/tristan-weil/ttserver/svc/tcplistener"
	ttutils "github.com/tristan-weil/ttserver/utils"
)

type (
	Space struct {
		config           *ttutils.ConfigRoot
		serveConnHandler tthandler.IServeConnHandler
		prometheusFire   func(*ttprom.PrometheusMetric) error
		logger           *logrus.Entry

		cron        *ttcron.CronCron
		cache       ttcache.ICacheCache
		tcpListener *tttcpl.TCPListener

		context       context.Context
		contextCancel context.CancelFunc
		isServing     ttutils.AtomicBool
		mu            sync.RWMutex
	}

	SpaceConfigInput struct {
		Config           *ttutils.ConfigRoot
		ServeConnHandler tthandler.IServeConnHandler
		PrometheusFire   func(*ttprom.PrometheusMetric) error
		Logger           *logrus.Entry
		Context          context.Context
	}
)

func NewSpace(config *SpaceConfigInput) *Space {
	s := Space{
		config:           config.Config,
		serveConnHandler: config.ServeConnHandler,
		prometheusFire:   config.PrometheusFire,
		logger:           config.Logger,
		context:          config.Context,
	}

	ctx, ctxCancel := context.WithCancel(config.Context)
	s.context = ctx
	s.contextCancel = ctxCancel

	return &s
}

func (s *Space) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.
		Debugf("creating...")

	//
	// cache
	//
	if s.cache == nil {
		var cache ttcache.ICacheCache

		if s.config.Space.Cache.Memory != nil {
			cache = ttcache.NewCacheMemory(&ttcache.MemoryConfigInput{
				Expiration: ttutils.IntValue(s.config.Space.Cache.Expiration),
				Cleanup:    ttutils.IntValue(s.config.Space.Cache.Memory.Cleanup),
				Logger:     s.logger.WithField("space-svc", "cache"),
			})
		}

		if cache != nil {
			cache.Start()
			s.cache = cache
		}
	}

	//
	// cron
	//
	if s.cron == nil {
		cronInst, err := ttcron.NewCron(&ttcron.ConfigInput{
			Config:           s.config,
			Cache:            s.GetCache,
			Logger:           s.logger.WithField("space-svc", "cron"),
			ServeConnHandler: s.serveConnHandler,
			PrometheusFire:   s.prometheusFire,
		})
		if err != nil {
			return fmt.Errorf("unable to create cron: %s", err)
		}

		cronInst.Start()
		s.cron = cronInst
	}

	//
	// listener
	//
	if s.tcpListener == nil {
		s.tcpListener = tttcpl.NewTCPListener(&tttcpl.TCPListenerConfigInput{
			Config:           s.config,
			Cache:            s.GetCache,
			ServeConnHandler: s.serveConnHandler,
			PrometheusFire:   s.prometheusFire,
			Logger:           s.logger.WithField("space-svc", "listener("+ttutils.StringValue(s.config.Space.Listener.Address)+")"),
		})

		s.tcpListener.Initialize()

		if err := s.tcpListener.Listen(); err != nil {
			return err
		}
	}

	s.logger.
		Debugf("creating... done!")

	return nil
}

func (s *Space) Reset(newConfig *ttutils.ConfigRoot) (*Space, error) {
	s.mu.Lock()
	defer func() {
		s.config = newConfig
		s.mu.Unlock()
	}()

	s.logger.
		Infof("reloading...")

	//
	// listener
	//
	if s.tcpListener != nil {
		if reset, err := s.tcpListener.Reset(newConfig); err != nil {
			return s, err
		} else {
			s.tcpListener = reset
			if s.tcpListener == nil {
				s.isServing.SetFalse()
			}
		}
	}

	//
	// cache
	//
	if s.cache != nil {
		if _, ok := s.cache.(*ttcache.Memory); ok {
			if newConfig.Space.Cache.Memory == nil ||
				(s.cache.(*ttcache.Memory).Expiration != ttutils.IntValue(newConfig.Space.Cache.Expiration)) ||
				(s.cache.(*ttcache.Memory).Cleanup != ttutils.IntValue(newConfig.Space.Cache.Memory.Cleanup)) {

				s.cache.Shutdown()
				s.cache = nil
			} else {
				s.cache.Flush()
			}
		} else {
			s.cache.Flush()
		}
	}

	//
	// cron
	//
	if s.cron != nil {
		s.cron.Shutdown()
		s.cron = nil
	}

	s.logger.
		Infof("reloading... done!")

	return s, nil
}

func (s *Space) Start() error {
	s.logger.
		Infof("serving...")

	s.isServing.SetTrue()

	if err := s.tcpListener.Serve(); err != nil {
		return err
	}

	s.logger.
		Debugf("serving... done!")

	return nil
}

func (s *Space) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.
		Infof("stopping...")

	// cache
	if s.cache != nil {
		s.cache.Shutdown()
		s.cache = nil
	}

	//
	// cron
	//
	if s.cron != nil {
		s.cron.Shutdown()
		s.cron = nil
	}

	//
	// listener
	//
	if s.tcpListener != nil {
		s.tcpListener.Shutdown()
		s.tcpListener = nil
		s.isServing.SetFalse()
	}

	s.logger.
		Infof("stopping... done!")

	return nil
}

func (s *Space) GetCache() ttcache.ICacheCache {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.cache
}

func (s *Space) IsServing() bool {
	return s.isServing.IsSet()
}
