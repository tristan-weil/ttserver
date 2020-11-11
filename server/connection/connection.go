package connection

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	ttcache "github.com/tristan-weil/ttserver/svc/cache"
	ttprom "github.com/tristan-weil/ttserver/svc/prometheus"
	ttutils "github.com/tristan-weil/ttserver/utils"
)

type (
	Connection struct {
		// at init
		InitialConn net.Conn
		CurConn     net.Conn
		Reader      *bufio.Reader
		Writer      *bufio.Writer

		Start time.Time

		Config         *ttutils.ConfigRoot
		PrometheusFire func(*ttprom.PrometheusMetric) error
		Cache          func() ttcache.ICacheCache
		Logger         *logrus.Entry

		UUID string

		Domain string
		Port   string
		SNI    string

		LocalAddress  string
		RemoteAddress string

		// misc
		Error      error
		MutexFetch sync.RWMutex

		// at end
		State      int
		ReturnCode string
		End        time.Time
	}

	ConfigInput struct {
		InitialConn net.Conn
		CurConn     net.Conn
		UUID        string

		Logger *logrus.Entry

		MaxqueryBytes int64

		Config         *ttutils.ConfigRoot
		Cache          func() ttcache.ICacheCache
		PrometheusFire func(*ttprom.PrometheusMetric) error
	}
)

const (
	CONNECTION_STATE_NEW = iota + 1
	CONNECTION_STATUS_READING
	CONNECTION_STATUS_READING_ERROR
	CONNECTION_STATUS_PARSING
	CONNECTION_STATUS_PARSING_ERROR
	CONNECTION_STATUS_PROCESSING
	CONNECTION_STATUS_PROCESSING_ERROR
	CONNECTION_STATUS_WRITING
	CONNECTION_STATUS_WRITING_ERROR
	CONNECTION_STATUS_FINISHED
	CONNECTION_STATUS_CRON
)

func GetUUID(c net.Conn) string {
	var uuidStr string

	uuidBin, err := uuid.NewRandom()
	if err != nil {
		if c != nil {
			uuidStr = fmt.Sprintf("%p", c)
		} else {
			uuidStr = "00000000-0000-0000-0000-000000000000"
		}
	} else {
		uuidStr = uuidBin.String()
	}

	return uuidStr
}

func NewConnection(connConfig *ConfigInput) *Connection {
	// If we have a maximum amount to read, then setup a limit
	var r io.Reader = connConfig.CurConn
	if connConfig.MaxqueryBytes > 0 {
		r = io.LimitReader(connConfig.CurConn, connConfig.MaxqueryBytes)
	}

	conn := &Connection{
		InitialConn: connConfig.InitialConn,
		CurConn:     connConfig.CurConn,
		Logger:      connConfig.Logger.WithField("connection", connConfig.UUID),

		Config:         connConfig.Config,
		PrometheusFire: connConfig.PrometheusFire,
		Cache:          connConfig.Cache,

		Writer:        bufio.NewWriter(connConfig.CurConn),
		Reader:        bufio.NewReader(r),
		Start:         time.Now(),
		LocalAddress:  connConfig.CurConn.LocalAddr().String(),
		RemoteAddress: connConfig.CurConn.RemoteAddr().String(),
		UUID:          connConfig.UUID,
		State:         CONNECTION_STATE_NEW,
	}

	return conn
}

func (c *Connection) Close() error {
	return c.CurConn.Close()
}

func (c *Connection) Write(p []byte) (int, error) {
	return c.Writer.Write(p)
}

func (c *Connection) Flush() error {
	return c.Writer.Flush()
}

func (c *Connection) CacheGet(key string) (*ttcache.Item, bool) {
	return c.Cache().Get(key)
}

func (c *Connection) CacheReplace(key string, value *ttcache.Item, d time.Duration) error {
	return c.Cache().Replace(key, value, d)
}

func (c *Connection) CacheReplaceIfExists(key string, value *ttcache.Item, d time.Duration) error {
	return c.Cache().ReplaceIfExists(key, value, d)
}

func (c *Connection) CacheDelete(key string) {
	c.Cache().Delete(key)
}

func (c *Connection) CacheAdd(key string, value *ttcache.Item, d time.Duration) error {
	return c.Cache().Add(key, value, d)
}
