package handler

import (
	ttconn "github.com/tristan-weil/ttserver/server/connection"
)

type (
	IServeConnHandler interface {
		// what to do with a new connection
		ServeConn(conn *ttconn.Connection) error

		// what to do with a new connection from the cron server
		ServeCrontab(conn *ttconn.Connection, route string, routeExtraData interface{}) error

		// retrieve a specifiv template functions map
		GetTemplatesFuncMap(conn *ttconn.Connection) (tplFunc map[string]interface{}, err error)

		// register new specific metrics
		RegisterPrometheusMetrics() error
	}
)

const (
	CRLF = "\r\n"
	TAB  = byte('\t')
)
