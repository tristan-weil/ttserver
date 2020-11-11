package utils

import (
	"regexp"
	"sync"

	securejoin "github.com/cyphar/filepath-securejoin"
)

// structs used by the global configuration block.
type (
	ConfigRoot struct {
		Log        *LogConfig        `json:"log,omitempty"`
		Prometheus *PrometheusConfig `json:"prometheus,omitempty"`
		Space      *SpaceConfig      `json:"space,omitempty"`
	}

	//
	// Logger
	//
	LogConfig struct {
		Level         *string `json:"level,omitempty"`
		Format        *string `json:"format,omitempty"`
		To            *string `json:"to,omitempty"`
		SyslogAddress *string `json:"syslog,omitempty"`
	}

	//
	// Cache
	//
	CacheConfig struct {
		Expiration *int               `json:"expiration,omitempty"`
		Memory     *CacheMemoryConfig `json:"memory,omitempty"`
	}

	CacheMemoryConfig struct {
		Cleanup *int `json:"cleanup,omitempty"`
	}

	//
	// Prometheus
	//
	PrometheusConfig struct {
		Address       string `json:"address,omitempty"`
		Endpoint      string `json:"endpoint,omitempty"`
		SummaryMaxAge *int   `json:"summaryMaxAge,omitempty"`
		ChanSize      *int   `json:"chanSize,omitempty"`

		Auth *PrometheusAuthConfig `json:"auth,omitempty"`
	}

	PrometheusAuthConfig struct {
		AuthBasic *PrometheusAuthBasicConfig `json:"basic,omitempty"`
	}

	PrometheusAuthBasicConfig struct {
		Username *string `json:"username,omitempty"`
		Password *string `json:"password,omitempty"`
	}

	//
	// Listener
	//
	ListenerConfig struct {
		Address       *string    `json:"address,omitempty"`
		Domains       []string   `json:"domains,omitempty"`
		TLSConfig     *TLSConfig `json:"tls,omitempty"`
		ProxyProtocol *string    `json:"proxyprotocol,omitempty"`
	}

	TLSConfig struct {
		ACME *TLSACMEConfig `json:"acme,omitempty"`
	}

	TLSACMEConfig struct {
		DNSProvider *string               `json:"dnsprovider,omitempty"`
		Email       *string               `json:"email,omitempty"`
		CA          *string               `json:"ca,omitempty"`
		TestCA      *string               `json:"testca,omitempty"`
		Storage     *TLSACMEStorageConfig `json:"storage,omitempty"`
	}

	TLSACMEStorageConfig struct {
		File *TLSACMEFileStorageConfig `json:"file,omitempty"`
	}

	TLSACMEFileStorageConfig struct {
		Path *string `json:"path,omitempty"`
	}

	//
	// Handler
	//
	HandlerConfig struct {
		Name       *string           `json:"name,omitempty"`
		Parameters map[string]string `json:"parameters,omitempty"`
	}

	//
	// Space
	//
	SpaceConfig struct {
		Handler *HandlerConfig `json:"handler,omitempty"`

		Cache *CacheConfig `json:"cache,omitempty"`

		Listener *ListenerConfig `json:"listener,omitempty"`

		BaseDir *string                 `json:"basedir,omitempty"`
		Routes  map[string]*RouteConfig `json:"routes,omitempty"`

		RoutesRegexp map[string]*RouteRegexpConfig
		Footer       *string
		Header       *string
		MutexRoutes  sync.RWMutex
	}

	//
	// Route
	//
	RouteConfig struct {
		File     *string `json:"file,omitempty"`
		Template *string `json:"template,omitempty"`

		Fetch map[string]*RouteFetchConfig `json:"fetch,omitempty"`
		Cache *RouteCacheConfig            `json:"cache,omitempty"`
		Cron  *string                      `json:"cron,omitempty"`

		RegexpCapturedGroups []string
	}

	RouteFetchConfig struct {
		Type *string `json:"type,omitempty"`
		URI  *string `json:"uri,omitempty"`
	}

	RouteCacheConfig struct {
		Expiration *int `json:"expiration,omitempty"`
	}

	RouteRegexpConfig struct {
		RouteConfig *RouteConfig
		Regexp      *regexp.Regexp
	}
)

func (sc *SpaceConfig) GetRoute(route string) *RouteConfig {
	sc.MutexRoutes.RLock()
	routeConfig := sc.Routes[route]
	sc.MutexRoutes.RUnlock()

	if routeConfig == nil {
		// try to find it in the Regexp routes
		for _, routeRegexpConf := range sc.RoutesRegexp {
			routeRegexp := routeRegexpConf.Regexp
			routeRegexpConf := routeRegexpConf.RouteConfig

			capturedGroups := routeRegexp.FindStringSubmatch(route)
			if capturedGroups == nil {
				continue
			}

			routeConfig = &RouteConfig{
				File:                 routeRegexpConf.File,
				Template:             routeRegexpConf.Template,
				Fetch:                routeRegexpConf.Fetch,
				Cache:                routeRegexpConf.Cache,
				Cron:                 routeRegexpConf.Cron,
				RegexpCapturedGroups: capturedGroups,
			}

			break
		}

		// no existing conf
		if routeConfig == nil {
			if tpl, err := securejoin.SecureJoin(StringValue(sc.BaseDir), route+".tpl"); err == nil {
				if CheckFileExists(tpl) {
					routeConfig = &RouteConfig{
						File:                 nil,
						Template:             String(tpl),
						Fetch:                nil,
						Cache:                &RouteCacheConfig{Expiration: sc.Cache.Expiration},
						Cron:                 nil,
						RegexpCapturedGroups: nil,
					}
				}
			} else if file, err := securejoin.SecureJoin(StringValue(sc.BaseDir), route); err == nil {
				if CheckFileExists(file) {
					routeConfig = &RouteConfig{
						File:                 String(file),
						Template:             nil,
						Fetch:                nil,
						Cache:                &RouteCacheConfig{Expiration: sc.Cache.Expiration},
						Cron:                 nil,
						RegexpCapturedGroups: nil,
					}
				}
			}
		}

		if routeConfig != nil {
			sc.MutexRoutes.Lock()
			sc.Routes[route] = routeConfig
			sc.MutexRoutes.Unlock()
		}
	}

	return routeConfig
}
