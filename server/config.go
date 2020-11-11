package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/prometheus/client_golang/prometheus"
	ttutils "github.com/tristan-weil/ttserver/utils"
)

func (m *Manager) readConfig() error {
	jsonFile, err := os.Open(m.configFile)
	if err != nil {
		return fmt.Errorf("unable to open config file %s: %s", m.configFile, err)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return fmt.Errorf("unable to read config file %s: %s", m.configFile, err)
	}

	//
	// default values
	//
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to get current working directory, %s", err)
	}

	jsonConfig := ttutils.ConfigRoot{
		Log: &ttutils.LogConfig{
			Level:  ttutils.String("info"),
			Format: ttutils.String("text"),
			To:     ttutils.String("stderr"),
		},
		Prometheus: &ttutils.PrometheusConfig{
			Address:       "127.0.0.1:9777",
			Endpoint:      "/metrics",
			SummaryMaxAge: ttutils.Int(int(prometheus.DefMaxAge.Seconds())),
			ChanSize:      ttutils.Int(1024),
		},
		Space: &ttutils.SpaceConfig{
			Cache: &ttutils.CacheConfig{
				Expiration: ttutils.Int(300),
				Memory: &ttutils.CacheMemoryConfig{
					Cleanup: ttutils.Int(350),
				},
			},

			Listener: &ttutils.ListenerConfig{
				Address: ttutils.String("127.0.0.1:7575"),
				Domains: []string{"localhost"},
			},

			BaseDir: ttutils.String(currentWorkingDirectory),
			Routes: map[string]*ttutils.RouteConfig{
				"index": {
					File:     nil,
					Template: nil,
				},
			},
		},
	}

	// json loading
	if err := json.Unmarshal(byteValue, &jsonConfig); err != nil {
		return fmt.Errorf("unable to unmarshall config file, %s", err)
	}

	//
	// SPACE
	//

	// LISTENER
	if envListenerAddress := os.Getenv("TTSERVER_LISTENER_ADDRESS"); envListenerAddress != "" {
		jsonConfig.Space.Listener.Address = ttutils.String(envListenerAddress)
	}

	// TLS           *TLSConfig `json:"tls,omitempty"`
	if jsonConfig.Space.Listener.TLSConfig != nil {
		if jsonConfig.Space.Listener.TLSConfig.ACME != nil {
			// CA          *string               `json:"ca,omitempty"`
			if ttutils.IsStringEmpty(jsonConfig.Space.Listener.TLSConfig.ACME.CA) {
				jsonConfig.Space.Listener.TLSConfig.ACME.CA = ttutils.String("https://acme-v02.api.letsencrypt.org/directory")
			}

			// TestCA      *string               `json:"testca,omitempty"`
			if ttutils.IsStringEmpty(jsonConfig.Space.Listener.TLSConfig.ACME.TestCA) {
				jsonConfig.Space.Listener.TLSConfig.ACME.TestCA = ttutils.String("https://acme-staging-v02.api.letsencrypt.org/directory")
			}

			// Storage     *TLSACMEStorageConfig `json:"storage,omitempty"`
			if jsonConfig.Space.Listener.TLSConfig.ACME.Storage == nil {
				jsonConfig.Space.Listener.TLSConfig.ACME.Storage = &ttutils.TLSACMEStorageConfig{
					File: &ttutils.TLSACMEFileStorageConfig{
						Path: ttutils.String(".certmagic"),
					},
				}
			}

			if jsonConfig.Space.Listener.TLSConfig.ACME.Storage.File != nil &&
				ttutils.IsStringEmpty(jsonConfig.Space.Listener.TLSConfig.ACME.Storage.File.Path) {

				jsonConfig.Space.Listener.TLSConfig.ACME.Storage.File.Path = ttutils.String(".certmagic")
			}

			// DNSProvider *string                      `json:"dnsprovider,omitempty"`
			if ttutils.IsStringEmpty(jsonConfig.Space.Listener.TLSConfig.ACME.DNSProvider) {
				return fmt.Errorf("LetsEncrypt has no dns-provider configured")
			}

			// Email       *string                      `json:"email,omitempty"`
			if ttutils.IsStringEmpty(jsonConfig.Space.Listener.TLSConfig.ACME.Email) {
				return fmt.Errorf("LetsEncrypt has no email configured")
			}
		}
	}

	// Handler *HandlerConfig `json:"handler,omitempty"`
	if jsonConfig.Space.Handler == nil || ttutils.IsStringEmpty(jsonConfig.Space.Handler.Name) {
		return fmt.Errorf("no handler configured")
	}

	// BaseDir  *string `json:"basedir,omitempty"`
	jsonConfig.Space.BaseDir = ttutils.FilePathClean(jsonConfig.Space.BaseDir)

	// Footer       *string
	// Header       *string
	if page, err := securejoin.SecureJoin(ttutils.StringValue(jsonConfig.Space.BaseDir), "index.tpl"); err != nil {
		return fmt.Errorf("unable to construct index file path, %s: %s", page, err)
	} else {
		if !ttutils.CheckFileExists(page) {
			return fmt.Errorf("unable to find index path, %s: %s", page, err)
		}
	}

	if page, err := securejoin.SecureJoin(ttutils.StringValue(jsonConfig.Space.BaseDir), "header.tpl"); err != nil {
		return fmt.Errorf("unable to construct header file path, %s: %s", page, err)
	} else {
		if ttutils.CheckFileExists(page) {
			jsonConfig.Space.Header = ttutils.String(page)
		}
	}

	if page, err := securejoin.SecureJoin(ttutils.StringValue(jsonConfig.Space.BaseDir), "footer.tpl"); err != nil {
		return fmt.Errorf("unable to construct footer file path, %s: %s", page, err)
	} else {
		if ttutils.CheckFileExists(page) {
			jsonConfig.Space.Footer = ttutils.String(page)
		}
	}

	// Routes map[string]*RouteConfig `json:"routes,omitempty"`
	for routeName, routeConf := range jsonConfig.Space.Routes {
		// File              *string `json:"file,omitempty"`
		// Template              *string `json:"template,omitempty"`
		if ttutils.NotStringEmpty(routeConf.Template) {
			tpl, err := securejoin.SecureJoin(ttutils.StringValue(jsonConfig.Space.BaseDir), ttutils.StringValue(routeConf.Template))
			if err != nil {
				return fmt.Errorf("unable to construct template file path %s.tpl for route %s: %s", ttutils.StringValue(routeConf.Template), routeName, err)
			}

			routeConf.Template = ttutils.String(tpl)
			routeConf.File = nil
		} else if ttutils.NotStringEmpty(routeConf.File) {
			tpl, err := securejoin.SecureJoin(ttutils.StringValue(jsonConfig.Space.BaseDir), ttutils.StringValue(routeConf.File))
			if err != nil {
				return fmt.Errorf("unable to construct file path %s for route %s: %s", ttutils.StringValue(routeConf.File), routeName, err)
			}

			routeConf.Template = nil
			routeConf.File = ttutils.String(tpl)
		} else {
			tpl, err := securejoin.SecureJoin(ttutils.StringValue(jsonConfig.Space.BaseDir), routeName+".tpl")
			if err != nil {
				return fmt.Errorf("unable to construct template file path %s.tpl for route %s: %s", routeName, routeName, err)
			}

			routeConf.Template = ttutils.String(tpl)
			routeConf.File = nil
		}

		// Fetch map[string]*RouteFetchConfig `json:"fetch,omitempty"`
		if routeConf.Fetch != nil {
			for fetchName, fetchConf := range routeConf.Fetch {
				if ttutils.IsStringEmpty(fetchConf.Type) || ttutils.IsStringEmpty(fetchConf.URI) {
					return fmt.Errorf("unable to find valid configuration for fetch %s: %s", fetchName, err)
				}
			}
		}

		// CacheCache *RouteCacheConfig            `json:"cache,omitempty"`
		if routeConf.Cache == nil {
			routeConf.Cache = &ttutils.RouteCacheConfig{
				Expiration: jsonConfig.Space.Cache.Expiration,
			}
		}

		// Cron  *RouteCronConfig             `json:"cron,omitempty"`
		if routeConf.Cron != nil {
			routeConf.Cache = &ttutils.RouteCacheConfig{
				Expiration: ttutils.Int(0),
			}
		}

		// populating RoutesRegexp map[string]*RouteRegexpConfig
		if routeName[0] == '~' {
			if jsonConfig.Space.RoutesRegexp == nil {
				jsonConfig.Space.RoutesRegexp = make(map[string]*ttutils.RouteRegexpConfig)
			}

			strRegexp := routeName[1:]
			jsonConfig.Space.RoutesRegexp[routeName] = &ttutils.RouteRegexpConfig{
				RouteConfig: routeConf,
				Regexp:      regexp.MustCompile(strRegexp),
			}
		}
	}

	// removing duplicates
	for routeRegexpName := range jsonConfig.Space.RoutesRegexp {
		delete(jsonConfig.Space.Routes, routeRegexpName)
	}

	// add err routes
	for _, s := range []string{"404", "500"} {
		notFound := true

		for routeName := range jsonConfig.Space.Routes {
			if s == routeName {
				notFound = false
				break
			}
		}

		if notFound {
			routeCode := &ttutils.RouteConfig{}
			routeCode.Cache = &ttutils.RouteCacheConfig{
				Expiration: ttutils.Int(0),
			}

			if tpl, err := securejoin.SecureJoin(ttutils.StringValue(jsonConfig.Space.BaseDir), s+".tpl"); err == nil {
				routeCode.Template = ttutils.String(tpl)
			}

			if jsonConfig.Space.Routes == nil {
				jsonConfig.Space.Routes = make(map[string]*ttutils.RouteConfig)
			}

			jsonConfig.Space.Routes[s] = routeCode
		}
	}

	m.config = &jsonConfig

	return nil
}
