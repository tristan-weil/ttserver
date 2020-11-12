package finger

import (
	"github.com/Masterminds/sprig/v3"
	ttconn "github.com/tristan-weil/ttserver/server/connection"
	tthandler "github.com/tristan-weil/ttserver/server/handler"
)

type (
	Handler struct{}
)

func (f *Handler) ServeConn(conn *ttconn.Connection) error {
	return tthandler.SimpleTextServeConnHandlerDefaultServeConn(f, conn)
}

func (f *Handler) ServeCrontab(conn *ttconn.Connection, route string, routeExtraData interface{}) error {
	return tthandler.SimpleTextServeConnHandlerDefaultServeCrontab(f, conn, route, routeExtraData)
}

func (f *Handler) Read(conn *ttconn.Connection) ([]byte, error) {
	return tthandler.SimpleTextServeConnHandlerDefaultRead(conn)
}

func (f *Handler) ParseData(conn *ttconn.Connection, inbuf []byte) (string, interface{}, error) {
	query, err := ParseQuery(string(inbuf))
	if err != nil {
		return "", nil, err
	}

	return query.Username, nil, nil
}

func (f *Handler) Process(conn *ttconn.Connection, route string, extraData interface{}, forceCacheUpdate bool) (output interface{}, err error) {

	errCodeMap := make(map[string][]byte)
	errCodeMap["200"] = []byte("OK (200)")
	errCodeMap["404"] = []byte("Not found (404)")
	errCodeMap["500"] = []byte("Internal Server Error (500)")

	return tthandler.SimpleTextServeConnHandlerCustomProcess(f, conn, route, extraData, forceCacheUpdate, errCodeMap)
}

func (f *Handler) PostProcess(conn *ttconn.Connection, route string, extraData interface{}, input interface{}) (output interface{}, err error) {

	return input, nil
}

func (f *Handler) Write(conn *ttconn.Connection, output interface{}) (n int64, err error) {
	return tthandler.SimpleTextServeConnHandlerDefaultWrite(conn, output)
}

func (f *Handler) GetTemplatesFuncMap(conn *ttconn.Connection) (tplFunc map[string]interface{}, err error) {
	sprigMap := sprig.TxtFuncMap()
	allMap := make(map[string]interface{})

	commonMap, err := tthandler.ServeConnHandlerCommonGetTextTemplatesFuncMap(conn)
	if err != nil {
		return nil, err
	}

	for _, m := range []map[string]interface{}{sprigMap, commonMap} {
		for k, v := range m {
			allMap[k] = v
		}
	}

	return allMap, nil
}

func (f *Handler) RegisterPrometheusMetrics() error {
	return nil
}
