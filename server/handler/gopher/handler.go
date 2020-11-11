package gopher

import (
	"bufio"
	"bytes"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	ttconn "github.com/tristan-weil/ttserver/server/connection"
	tthandler "github.com/tristan-weil/ttserver/server/handler"
)

type (
	Handler struct{}
)

func (f *Handler) MaxQueryBytes() int {
	return 512
}

func (f *Handler) ServeConn(conn *ttconn.Connection) error {
	return tthandler.SimpleTextServeConnHandlerDefaultServeConn(f, conn)
}

func (f *Handler) ServeCrontab(conn *ttconn.Connection, route string, routeExtraData interface{}) error {
	return tthandler.SimpleTextServeConnHandlerDefaultServeCrontab(f, conn, route, routeExtraData)
}

func (f *Handler) Read(conn *ttconn.Connection) ([]byte, error) {
	return tthandler.SimpleTextServeConnHandlerDefaultRead(conn)
}

func (f *Handler) ParseData(conn *ttconn.Connection, inbuf []byte) (route string, extraData interface{}, err error) {
	query, err := ParseQuery(string(inbuf))
	if err != nil {
		return "", nil, err
	}

	var data string

	if query.ExtData != "" {
		data = query.ExtData
	}

	var domain string
	var port string

	addrSplit := strings.Split(conn.LocalAddress, ":")
	if resp_domain, ok := conn.Config.Space.Handler.Parameters["response_domain"]; ok {
		domain = resp_domain
	} else if conn.SNI != "" {
		domain = conn.SNI
	} else if len(conn.Config.Space.Listener.Domains) > 0 {
		domain = conn.Config.Space.Listener.Domains[0]
	} else {
		domain = addrSplit[0]
	}

	if resp_port, ok := conn.Config.Space.Handler.Parameters["response_port"]; ok {
		port = resp_port
	} else {
		addrSplit := strings.Split(conn.LocalAddress, ":")
		port = addrSplit[1]
	}

	conn.Domain = domain
	conn.Port = port

	return query.Selector, &data, nil
}

func (f *Handler) Process(conn *ttconn.Connection, route string, extraData interface{}, forceCacheUpdate bool) (output interface{}, err error) {
	var gopherExtensionData string

	if extraData != nil {
		switch extraData.(type) {
		case *string:
			gopherExtensionData = *(extraData.(*string))
		default:
			gopherExtensionData = ""
		}
	}

	urlMessage := []byte("" +
		"<HTML>" +
		"<HEAD>" +
		"<META HTTP-EQUIV=\"refresh\" content=\"2;URL=" + gopherExtensionData + "\">" +
		"</HEAD>" +
		"<BODY>" +
		"You are following an external link to a Web site. You will be automatically taken to the site shortly." +
		"<BR>If you do not get sent there, please click " +
		"<A HREF=\"" + gopherExtensionData + "\">here</A> to go to the web site." +
		"<P>" +
		"The URL linked is: " + gopherExtensionData +
		"<P>" +
		"<A HREF=\"" + gopherExtensionData + "\">" + gopherExtensionData + "</A>" +
		"<P>" +
		"Thanks for using Gopher!" +
		"</BODY>" +
		"</HTML>")

	//
	// gopher extension
	//
	if route == "URL" && gopherExtensionData != "" {
		conn.ReturnCode = "200"

		return urlMessage, nil
	}

	//
	// normal handling
	//
	errCodeMap := make(map[string][]byte)

	code200 := gopherItem{
		Type:        INFO,
		Description: "OK (200)",
		Selector:    "",
		Host:        "localhost",
		Port:        "0",
	}
	code404 := gopherItem{
		Type:        ERROR,
		Description: "Not found (404)",
		Selector:    "",
		Host:        "localhost",
		Port:        "0",
	}
	code500 := gopherItem{
		Type:        ERROR,
		Description: "Internal Server Error (500)",
		Selector:    "",
		Host:        "localhost",
		Port:        "0",
	}

	errCodeMap["200"] = []byte(code200.String())
	errCodeMap["404"] = []byte(code404.String())
	errCodeMap["500"] = []byte(code500.String())

	return tthandler.SimpleTextServeConnHandlerCustomProcess(f, conn, route, extraData, forceCacheUpdate, errCodeMap)
}

func (f *Handler) PostProcess(conn *ttconn.Connection, route string, extraData interface{}, input interface{}) (output interface{}, err error) {
	reader := bytes.NewReader(input.([]byte))
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	newOutbuf := []byte{}

	// make all lines become INFO type if they have no other type
	for scanner.Scan() {
		line := scanner.Text()
		if !TYPES_REGEXP.MatchString(line) {
			gi := gopherItem{
				Type:        INFO,
				Description: line,
				Selector:    "",
				Host:        "localhost",
				Port:        "0",
			}
			newOutbuf = append(newOutbuf, gi.Bytes()...)
		} else {
			newOutbuf = append(newOutbuf, []byte(line)...)
			newOutbuf = append(newOutbuf, []byte(tthandler.CRLF)...)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return newOutbuf, nil
}

func (f *Handler) Write(conn *ttconn.Connection, output interface{}) (n int64, err error) {
	return tthandler.SimpleTextServeConnHandlerDefaultWrite(conn, output)
}

func (f *Handler) GetTemplatesFuncMap(conn *ttconn.Connection) (tplFunc map[string]interface{}, err error) {
	var handlerMap template.FuncMap

	handlerMap = template.FuncMap{
		"gurl_for": func(selector string) string {
			if strings.HasPrefix(selector, "/") {
				return selector
			}

			return "/" + selector
		},

		"gmenu": func(selector string, description string) string {
			gi := gopherItem{
				Type:        MENU,
				Description: description,
				Selector:    selector,
				Host:        conn.Domain,
				Port:        conn.Port,
			}

			return gi.String()
		},

		"ginfo": func(text string) string {
			splitted := strings.Split(text, "\n")

			var result string

			for i, s := range splitted {
				if i == len(splitted)-1 && s == "" {
					continue
				}

				gi := gopherItem{
					Type:        INFO,
					Description: s,
					Selector:    "",
					Host:        "localhost",
					Port:        "0",
				}

				result += gi.String()
			}

			return result
		},

		"gerror": func(text string) string {
			gi := gopherItem{
				Type:        ERROR,
				Description: text,
				Selector:    "",
				Host:        "localhost",
				Port:        "0",
			}

			return gi.String()
		},

		"gtitle": func(text string) string {
			gi := gopherItem{
				Type:        INFO,
				ExtraType:   "TITLE",
				Description: text,
				Selector:    "",
				Host:        "localhost",
				Port:        "0",
			}

			return gi.String()
		},

		"gurl": func(url string, description string) string {
			gi := gopherItem{
				Type:        HTML,
				ExtraType:   "URL",
				Description: description,
				Selector:    url,
				Host:        conn.Domain,
				Port:        conn.Port,
			}

			return gi.String()
		},
	}

	sprigMap := sprig.TxtFuncMap()

	commonMap, err := tthandler.ServeConnHandlerCommonGetTextTemplatesFuncMap(conn)
	if err != nil {
		return nil, err
	}

	var allmap = make(map[string]interface{})
	for _, m := range []map[string]interface{}{sprigMap, commonMap, handlerMap} {
		for k, v := range m {
			allmap[k] = v
		}
	}

	return allmap, nil
}

func (f *Handler) RegisterPrometheusMetrics() error {
	return nil
}
