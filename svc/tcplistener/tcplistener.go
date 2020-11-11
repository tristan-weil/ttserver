package tcplistener

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	proxyprotocol "github.com/pires/go-proxyproto"
	"github.com/sirupsen/logrus"
	ttconn "github.com/tristan-weil/ttserver/server/connection"
	tthandler "github.com/tristan-weil/ttserver/server/handler"
	ttcache "github.com/tristan-weil/ttserver/svc/cache"
	ttprom "github.com/tristan-weil/ttserver/svc/prometheus"
	ttutils "github.com/tristan-weil/ttserver/utils"
)

type (
	TCPListener struct {
		config *ttutils.ConfigRoot
		cache  func() ttcache.ICacheCache

		address       string
		domains       []string
		tlsConfig     *ttutils.TLSConfig
		proxyProtocol string

		serveConnHandler tthandler.IServeConnHandler
		prometheusFire   func(*ttprom.PrometheusMetric) error

		// ReadTimeout is the maximum duration before timing out reads of the
		// response. This sets a deadline on the connection and isn't a handler
		// timeout.
		readTimeout time.Duration

		// WriteTimeout is the maximum duration before timing out writes of the
		// response. This sets a deadline on the connection and isn't a handler
		// timeout.
		writeTimeout time.Duration

		// MaxQueryBytes is the maximum amount of bytes that will be read from
		// the connection to determine the query.
		maxQueryBytes int

		logger *logrus.Entry

		listener   *netListenerWrapper
		activeConn map[net.Conn]*ttconn.Connection

		shuttingDown                ttutils.AtomicBool
		alreadyNewTLSGetCertificate ttutils.AtomicBool

		muActiveConn sync.RWMutex
	}

	TCPListenerConfigInput struct {
		Config           *ttutils.ConfigRoot
		Cache            func() ttcache.ICacheCache
		ServeConnHandler tthandler.IServeConnHandler
		PrometheusFire   func(*ttprom.PrometheusMetric) error
		Logger           *logrus.Entry
	}

	netListenerWrapper struct {
		netListener           net.Listener
		proxyProtocolListener *proxyprotocol.Listener

		tlsConfig *tls.Config
	}
)

var (
	shutdownActiConnPollInterval = 500 * time.Millisecond
	shutdownServerTimeout        = 5 * time.Second
)

func NewTCPListener(serverConfig *TCPListenerConfigInput) *TCPListener {
	t := TCPListener{
		config:           serverConfig.Config,
		cache:            serverConfig.Cache,
		serveConnHandler: serverConfig.ServeConnHandler,
		prometheusFire:   serverConfig.PrometheusFire,

		logger: serverConfig.Logger,
	}

	return &t
}

func (t *TCPListener) Initialize() {
	if !t.IsServing() {
		t.address = ttutils.StringValue(t.config.Space.Listener.Address)
		t.domains = t.config.Space.Listener.Domains
		t.proxyProtocol = ttutils.StringValue(t.config.Space.Listener.ProxyProtocol)
		t.tlsConfig = t.config.Space.Listener.TLSConfig

		// TODO: by config
		t.readTimeout = 1 * time.Minute
		t.writeTimeout = 1 * time.Minute
		t.maxQueryBytes = 512
	}
}

// Listen listens on the TCP network address s.Addr and then
// calls serve to handle incoming connections.
func (t *TCPListener) Listen() error {
	if !t.IsServing() {
		t.logger.
			Debugf("creating...")

		listenStr := ""

		ln, err := net.Listen("tcp", t.address)
		if err != nil {
			return err
		}

		t.listener = &netListenerWrapper{netListener: ln}

		if t.proxyProtocol == "v1" || t.proxyProtocol == "v2" || t.proxyProtocol == "enabled" {
			listenStr += fmt.Sprintf(" + PROXY protocol (%s)", t.proxyProtocol)
		}

		if t.tlsConfig != nil {
			if t.tlsConfig.ACME == nil {
				logrus.Warn("no tls configuration configured")
			} else {
				if t.tlsConfig.ACME != nil {
					if err := t.configureACME(); err != nil {
						return err
					}
				}

				listenStr += " + tls"

				if !t.alreadyNewTLSGetCertificate.IsSet() {
					oldGetCertificate := t.listener.tlsConfig.GetCertificate

					t.muActiveConn.Lock()
					defer t.muActiveConn.Unlock()

					t.alreadyNewTLSGetCertificate.SetTrue()

					t.listener.tlsConfig.GetCertificate = func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
						sni := info.ServerName

						t.muActiveConn.RLock()
						conn := t.activeConn[info.Conn]
						t.muActiveConn.RUnlock()

						if conn != nil {
							for _, d := range t.domains {
								if d == sni {
									conn.SNI = sni

									t.logger.
										WithField("connection", conn.UUID).
										Tracef("SNI found for %q", sni)

									break
								}
							}
						}

						return oldGetCertificate(info)
					}
				}
			}
		}

		if t.proxyProtocol == "v1" || t.proxyProtocol == "v2" || t.proxyProtocol == "enabled" {
			t.listener.proxyProtocolListener = &proxyprotocol.Listener{Listener: ln}
		}

		t.logger.
			Infof("listening%s...", listenStr)

		t.logger.
			Debugf("creating... done!")
	}

	return nil
}

func (t *TCPListener) Reset(newConfig *ttutils.ConfigRoot) (*TCPListener, error) {
	defer func() {
		t.config = newConfig
	}()

	if ttutils.StringValue(newConfig.Space.Listener.Address) != t.address ||
		ttutils.StringValue(newConfig.Space.Listener.ProxyProtocol) != t.proxyProtocol ||
		!cmp.Equal(newConfig.Space.Listener.Domains, t.domains) {

		t.logger.
			Infof("reloading... stopping listener...")

		if err := t.Shutdown(); err != nil {
			return t, fmt.Errorf("unable to stop listener: %s", err)
		}

		t.logger.
			Debugf("reloading... stopping listener... done!")

		return nil, nil
	}

	return t, nil
}

func (t *TCPListener) isShuttingDown() bool {
	return t.shuttingDown.IsSet()
}

func (t *TCPListener) Shutdown() error {
	t.shuttingDown.SetTrue()

	t.logger.
		Debugf("stopping...")

	t.logger.Infof("closing...")

	var err error

	if t.listener.proxyProtocolListener != nil {
		err = t.listener.proxyProtocolListener.Listener.Close()
		t.listener.proxyProtocolListener = nil
	} else {
		err = t.listener.netListener.Close()
		t.listener.netListener = nil
	}

	if err != nil {
		t.logger.Infof("unable to close, %s", err)
	}

	t.muActiveConn.RLock()
	if len(t.activeConn) != 0 {
		t.logger.Infof("waiting to end all connections")
	}
	t.muActiveConn.RUnlock()

	ctxShutdown, ctxShutdownCancel := context.WithTimeout(context.Background(), shutdownServerTimeout)
	defer ctxShutdownCancel()

	ticker := time.NewTicker(shutdownActiConnPollInterval)
	defer ticker.Stop()

	for {
		t.muActiveConn.RLock()
		if len(t.activeConn) == 0 {
			t.muActiveConn.RUnlock()
			break
		}
		t.muActiveConn.RUnlock()

		select {
		case <-ctxShutdown.Done():
			break
		case <-ticker.C:
		}
	}

	t.activeConn = nil
	t.listener = nil

	t.logger.
		Debugf("stopping... done!")

	return err
}

// serve accepts incoming connections on the netListener l, creating a new
// service goroutine for each.
//
// The netListener is closed when this function returns.
func (t *TCPListener) Serve() error {
	if !t.IsServing() {
		t.logger.Infof("serving...")

		var c net.Conn
		var err error
		var curConnection *ttconn.Connection

		t.activeConn = make(map[net.Conn]*ttconn.Connection)

		for {
			// Accept
			if t.listener.proxyProtocolListener != nil {
				c, err = t.listener.proxyProtocolListener.Accept()
			} else if t.listener.netListener != nil {
				c, err = t.listener.netListener.Accept()
			} else {
				err = fmt.Errorf("no more listener available")
			}
			if err != nil {
				if t.isShuttingDown() {
					t.logger.Tracef("asked to close nicely")
					return nil
				}

				return err
			}

			// Handle connection
			uuidStr := ttconn.GetUUID(c)
			t.logger.
				WithField("connection", uuidStr).
				Debugf("connection accepted")

			// config Connection
			var t0 = time.Now()
			if t.readTimeout > 0 {
				c.SetReadDeadline(t0.Add(t.readTimeout))
			}

			if t.writeTimeout > 0 {
				c.SetWriteDeadline(t0.Add(t.writeTimeout))
			}

			curConn := c

			if t.listener.tlsConfig != nil {
				curConn = tls.Server(c, t.listener.tlsConfig)
			}

			curConnection = ttconn.NewConnection(&ttconn.ConfigInput{
				InitialConn:   c,
				CurConn:       curConn,
				MaxqueryBytes: int64(t.maxQueryBytes),
				UUID:          uuidStr,

				Config:         t.config,
				PrometheusFire: t.prometheusFire,
				Cache:          t.cache,
				Logger:         t.logger,
			})

			t.trackConn(curConnection, true)

			// serving
			go func(conn *ttconn.Connection) {
				defer func() {
					t.logger.
						WithField("connection", conn.UUID).
						Debugf("end of connection")

					conn.Flush()
					conn.Close()
					t.trackConn(conn, false)
				}()

				if err := t.serveConnHandler.ServeConn(conn); err != nil {
					t.logger.
						WithField("connection", conn.UUID).
						Errorf("%s", err)

					conn.Error = err
				}
			}(curConnection)
		}
	}

	return nil
}

func (t *TCPListener) IsServing() bool {
	return t.activeConn != nil && t.listener.netListener != nil
}

func (t *TCPListener) trackConn(conn *ttconn.Connection, add bool) {
	t.muActiveConn.Lock()

	if add {
		t.activeConn[conn.InitialConn] = conn
		t.muActiveConn.Unlock()

		if err := t.prometheusFire(&ttprom.PrometheusMetric{
			Metric: ttprom.PrometheusConnDurationSummary,
			Labels: []string{conn.ReturnCode},
			Action: "set",
			Values: float64(len(t.activeConn)),
		}); err != nil {
			t.logger.
				WithField("connection", conn.UUID).
				Errorf("firing prometheus failed -> %s", err)
		}
	} else {
		conn.End = time.Now()
		delete(t.activeConn, conn.InitialConn)
		t.muActiveConn.Unlock()

		if err := t.prometheusFire(&ttprom.PrometheusMetric{
			Metric: ttprom.PrometheusConnDurationSummary,
			Labels: []string{conn.ReturnCode},
			Action: "observe",
			Values: float64(conn.End.Sub(conn.Start).Microseconds()),
		}); err != nil {
			t.logger.
				WithField("connection", conn.UUID).
				Errorf("firing prometheus failed -> %s", err)
		}
	}

	if err := t.prometheusFire(&ttprom.PrometheusMetric{
		Metric: ttprom.PrometheusActiveConnGauge,
		Labels: []string{},
		Action: "set",
		Values: float64(len(t.activeConn)),
	}); err != nil {
		t.logger.
			WithField("connection", conn.UUID).
			Errorf("firing prometheus failed -> %s", err)
	}
}
