package cron

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	ttconn "github.com/tristan-weil/ttserver/server/connection"
	tthandler "github.com/tristan-weil/ttserver/server/handler"
	ttcache "github.com/tristan-weil/ttserver/svc/cache"
	ttprom "github.com/tristan-weil/ttserver/svc/prometheus"
	ttutils "github.com/tristan-weil/ttserver/utils"
)

type (
	CronCron struct {
		config *ttutils.ConfigRoot
		cache  func() ttcache.ICacheCache

		cron *cron.Cron

		serveConnHandler tthandler.IServeConnHandler
		prometheusFire   func(*ttprom.PrometheusMetric) error

		logger        *logrus.Entry
		context       context.Context
		contextCancel context.CancelFunc

		isInitialized bool
	}

	ConfigInput struct {
		ServeConnHandler tthandler.IServeConnHandler
		Config           *ttutils.ConfigRoot
		Cache            func() ttcache.ICacheCache
		PrometheusFire   func(metric *ttprom.PrometheusMetric) error
		Logger           *logrus.Entry
	}
)

func NewCron(cronConfig *ConfigInput) (*CronCron, error) {
	c := CronCron{
		config:           cronConfig.Config,
		prometheusFire:   cronConfig.PrometheusFire,
		cache:            cronConfig.Cache,
		cron:             cron.New(cron.WithSeconds()),
		serveConnHandler: cronConfig.ServeConnHandler,
		logger:           cronConfig.Logger,
	}

	for routeName, routeConf := range c.config.Space.Routes {
		if routeConf.Cron != nil {
			c.logger.Infof("creating cron for %s...", routeName)

			if routeConf.Cron == nil {
				c.logger.Warnf("unable to start a cron without spec")
				continue
			}

			if routeName[0] == '~' {
				c.logger.Warnf("unable to start a cron on a regexp route")
				continue
			}

			fakeConn := &ttconn.Connection{
				InitialConn: nil,
				CurConn:     nil,
				Reader:      nil,
				Writer:      nil,

				Logger:         c.logger.WithField("connection", routeName+":cron"),
				PrometheusFire: c.prometheusFire,
				Config:         c.config,
				Cache:          c.cache,

				Start:         time.Now(),
				LocalAddress:  ttutils.StringValue(c.config.Space.Listener.Address),
				RemoteAddress: ttutils.StringValue(c.config.Space.Listener.Address),
				UUID:          routeName + ":cron",
				State:         ttconn.CONNECTION_STATUS_CRON,
			}

			// copy to avoid pointer
			curRoute := routeName

			// add entry
			if _, err := c.cron.AddFunc(ttutils.StringValue(routeConf.Cron), func() {
				c.logger.
					WithField("route", curRoute).
					Infof("scheduling route")

				if ferr := c.serveConnHandler.ServeCrontab(fakeConn, curRoute, nil); ferr != nil {
					c.logger.
						WithField("route", curRoute).
						WithField("code", fakeConn.ReturnCode).
						Errorf("%s", ferr)
				} else {
					c.logger.
						WithField("route", curRoute).
						WithField("code", fakeConn.ReturnCode).
						Debugf("success")
				}
			}); err != nil {
				c.logger.Infof("creating cron for %s... %s", routeName, err)
				return nil, err
			}

			c.logger.Debugf("creating cron for %s... done!", routeName)
		}
	}

	return &c, nil
}

func (c *CronCron) Start() {
	c.logger.
		Infof("starting...")

	listenCtx, listenCancelCtx := context.WithCancel(context.Background())
	c.context = listenCtx
	c.contextCancel = listenCancelCtx

	c.isInitialized = true

	c.cron.Start()

	c.logger.
		Infof("starting... done!")
}

func (c *CronCron) Shutdown() {
	if c.cron != nil && c.isInitialized {
		c.logger.
			Infof("stopping...")

		c.contextCancel()
		c.isInitialized = false

		var entries []cron.EntryID
		for _, e := range c.cron.Entries() {
			entries = append(entries, e.ID)
		}

		for _, e := range entries {
			c.cron.Remove(e)
		}

		c.cron.Stop()

		c.logger.
			Infof("stopping... done!")
	}
}
