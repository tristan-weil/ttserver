package handler

import (
	"bufio"
	"io"
	"os"

	ttconn "github.com/tristan-weil/ttserver/server/connection"
)

type (
	SimpleTextServeConnHandler interface {
		// inherited from IServeConnHandler
		ServeConn(conn *ttconn.Connection) error

		ServeCrontab(conn *ttconn.Connection, route string, routeExtraData interface{}) error

		GetTemplatesFuncMap(conn *ttconn.Connection) (tplFunc map[string]interface{}, err error)
		RegisterPrometheusMetrics() error

		// specific
		Read(conn *ttconn.Connection) (input []byte, err error)
		ParseData(conn *ttconn.Connection, input []byte) (route string, routeExtraData interface{}, err error)

		Process(conn *ttconn.Connection, route string, routeExtraData interface{}, forceCacheUpdate bool) (output interface{}, err error)
		PostProcess(conn *ttconn.Connection, route string, routeExtraData interface{}, input interface{}) (output interface{}, err error)

		Write(conn *ttconn.Connection, output interface{}) (n int64, err error)
	}
)

//
// Default implementations to handler simple texts protocols
//

func SimpleTextServeConnHandlerDefaultServeConn(h SimpleTextServeConnHandler, conn *ttconn.Connection) error {
	// read the query line
	conn.Logger.Debugf("reading...")

	conn.State = ttconn.CONNECTION_STATUS_READING
	inbuf, err := h.Read(conn)
	if err != nil {
		conn.State = ttconn.CONNECTION_STATUS_READING_ERROR
		return err
	}

	conn.Logger.Debugf("reading... done!")

	// parse data and get route
	conn.Logger.Debugf("parsing...")

	conn.State = ttconn.CONNECTION_STATUS_PARSING
	route, extraData, err := h.ParseData(conn, inbuf)
	if err != nil {
		conn.State = ttconn.CONNECTION_STATUS_PARSING_ERROR
		return err
	}

	conn.Logger.Debugf("parsing... done!")

	if route == "" {
		route = "index"
	}

	conn.Logger = conn.Logger.
		WithField("route", route)

	// process
	conn.Logger.Debugf("processing...")

	conn.State = ttconn.CONNECTION_STATUS_PROCESSING
	outdata, err := h.Process(conn, route, extraData, false)
	if err != nil {
		conn.State = ttconn.CONNECTION_STATUS_PROCESSING_ERROR
		return err
	}

	conn.Logger.Debugf("processing... done!")

	// write
	conn.Logger.Debugf("writing...")

	conn.State = ttconn.CONNECTION_STATUS_WRITING
	_, err = h.Write(conn, outdata)
	if err != nil {
		conn.State = ttconn.CONNECTION_STATUS_WRITING_ERROR
		return err
	}
	// CRLF has to be added by the handler if needed

	conn.Logger.Debugf("writing... done!")

	conn.State = ttconn.CONNECTION_STATUS_FINISHED

	conn.Logger.Infof("access")

	return nil
}

func SimpleTextServeConnHandlerDefaultServeCrontab(h SimpleTextServeConnHandler, conn *ttconn.Connection, route string, routeExtraData interface{}) error {
	if _, err := h.Process(conn, route, routeExtraData, true); err != nil {
		return err
	}

	return nil
}

func SimpleTextServeConnHandlerDefaultGetTemplatesFuncMap(conn *ttconn.Connection) (tplFunc map[string]interface{}, err error) {
	return ServeConnHandlerCommonGetTextTemplatesFuncMap(conn)
}

func SimpleTextServeConnHandlerDefaultRegisterMetrics() error {
	return nil
}

func SimpleTextServeConnHandlerDefaultRead(conn *ttconn.Connection) (input []byte, err error) {
	scanner := bufio.NewScanner(conn.Reader)
	scanner.Split(bufio.ScanLines)

	// first line
	scanner.Scan()
	line := scanner.Text()

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return []byte(line), nil
}

func SimpleTextServeConnHandlerDefaultParseData(conn *ttconn.Connection, inbuf []byte) (route string, extraData interface{}, err error) {
	return string(inbuf), nil, nil
}

func SimpleTextServeConnHandlerDefaultProcess(conn *ttconn.Connection, route string, routeExtraData interface{}, forceCacheUpdate bool) (output interface{}, err error) {
	conn.ReturnCode = "200"

	conn.Logger = conn.Logger.
		WithField("route", route).
		WithField("code", conn.ReturnCode)

	return []byte("Hello\r\n"), nil
}

func SimpleTextServeConnHandlerDefaultWrite(conn *ttconn.Connection, output interface{}) (n int64, err error) {
	var nn int

	switch o := output.(type) {
	case []byte:
		nn, err = conn.Write(o)
		n = int64(nn)
	case *os.File:
		n, err = io.Copy(conn.Writer, o)
	}

	return n, err
}
