package handler

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/Jeffail/tunny"
	ttconn "github.com/tristan-weil/ttserver/server/connection"
	ttcache "github.com/tristan-weil/ttserver/svc/cache"
	ttprom "github.com/tristan-weil/ttserver/svc/prometheus"
	ttutils "github.com/tristan-weil/ttserver/utils"
)

type (
	simpleTextHandlerTemplateDataItem struct {
		Route        string
		Domain       string
		Port         string
		LocalAddress string
		LocalPort    string

		URI  string
		Type string
		Data interface{}
	}

	simpleTextHandlerFetcher struct {
		fetcher   func(string) (interface{}, error)
		name      string
		fetchType string
		uri       string
	}
)

var getFeedTimeout = 30 * time.Second

func SimpleTextServeConnHandlerCustomProcess(
	h SimpleTextServeConnHandler,
	conn *ttconn.Connection,
	route string,
	extraData interface{},
	forceCacheUpdate bool,
	errCodeMap map[string][]byte) (output interface{}, err error) {

	//
	// error map
	//
	if errCodeMap == nil {
		errCodeMap := make(map[string][]byte)
		errCodeMap["200"] = []byte("OK (200)")
		errCodeMap["404"] = []byte("Not found (404)")
		errCodeMap["500"] = []byte("Internal Server Error (500)")
	}

	outbuf, returnCode, cacheStatus := doSimpleTextServeConnHandlerCustomProcess(h, conn, route, extraData, forceCacheUpdate, errCodeMap)
	conn.ReturnCode = returnCode

	conn.Logger = conn.Logger.
		WithField("code", conn.ReturnCode).
		WithField("cache", cacheStatus)

	return outbuf, nil
}

func doSimpleTextServeConnHandlerCustomProcess(
	h SimpleTextServeConnHandler,
	conn *ttconn.Connection,
	route string,
	routeExtraData interface{},
	forceCacheUpdate bool,
	errCodeMap map[string][]byte) (output interface{}, code string, cacheStatus string) {

	//
	// vars
	//
	var (
		spaceConfig       = conn.Config.Space
		routeConfig       = conn.Config.Space.GetRoute(route)
		startTime         = time.Now()
		returnData        interface{}
		returnCode        = "200"
		returnCacheStatus = "NOCACHE"
		err               error
		tpl               *template.Template
		templateData      = make(map[string]interface{})
		buf               = new(bytes.Buffer)
		funcMap           map[string]interface{}
		filePath          string
		templateFilePath  string
		isATemplateFile   = false
		isAFile           = false
		templateName      = route + ".tpl"
		buftmpl           []byte
		localAddrPort     = strings.Split(conn.LocalAddress, ":")
		localAddr         = localAddrPort[0]
		localPort         = localAddrPort[1]
	)

	/*
	 ********************************************************************************
	 *
	 * Something wrong with this route
	 *
	 ********************************************************************************
	 */

	if routeConfig == nil {
		conn.Logger.Errorf("unable to find a configuration for this route")

		if route == "404" || route == "500" {
			returnData = errCodeMap[route]
			returnData = append(returnData.([]byte), []byte(CRLF)...)
			returnCode = route
		} else {
			returnData, returnCode, returnCacheStatus = doSimpleTextServeConnHandlerCustomProcess(h, conn, "404", routeExtraData, false, errCodeMap)
			returnCode = "404"
		}

		goto GOTO_NO_CACHE
	}

	/*
	 ********************************************************************************
	 *
	 * Check for all kind of files in the cache
	 *
	 ********************************************************************************
	 */

	//
	// check cache
	//
	if !forceCacheUpdate && conn.Cache().IsEnabled() {
		conn.Logger.Tracef("checking cache...")

		cachedData, ok := conn.CacheGet(route)
		if ok {
			returnData = cachedData.Item.([]byte)
			returnCode = cachedData.Code
			returnCacheStatus = "HIT"

			conn.Logger.Tracef("checking cache... found!")

			goto GOTO_NO_CACHE
		}

		conn.Logger.Tracef("checking cache... not found!")
	}

	//
	// building path
	//
	filePath = ttutils.StringValue(routeConfig.File)
	templateFilePath = ttutils.StringValue(routeConfig.Template)

	if ttutils.NotStringEmpty(routeConfig.Template) {
		templateFilePath = *routeConfig.Template

		if ttutils.NotStringSliceEmpty(routeConfig.RegexpCapturedGroups) {
			for i, capturedGroup := range routeConfig.RegexpCapturedGroups {
				if i == 0 {
					continue
				}
				templateFilePath = strings.Replace(templateFilePath, "$"+strconv.Itoa(i), capturedGroup, -1)
			}
		}
	}

	//
	// not found in cache, check fs
	// if not found in fs, return err
	//
	isATemplateFile = ttutils.CheckFileExists(templateFilePath)
	isAFile = ttutils.CheckFileExists(filePath) && !(ttutils.NotStringEmpty(routeConfig.Template) && isATemplateFile)

	if !isATemplateFile && !isAFile {
		conn.Logger.Errorf("not found on FS: %s%s", filePath, templateFilePath)

		if route == "404" || route == "500" {
			returnData = errCodeMap[route]
			returnData = append(returnData.([]byte), []byte(CRLF)...)
			returnCode = route
		} else {
			returnData, returnCode, returnCacheStatus = doSimpleTextServeConnHandlerCustomProcess(h, conn, "404", routeExtraData, false, errCodeMap)
			returnCode = "404"
		}

		goto GOTO_ADD_TO_CACHE
	} else {
		conn.Logger.Tracef("found on FS: %s%s", filePath, templateFilePath)
	}

	/*
	 ********************************************************************************
	 *
	 * Serving file
	 *
	 ********************************************************************************
	 */
	if isAFile {
		file, err := os.Open(filePath)
		if err != nil {
			conn.Logger.Errorf("unable to open %s", filePath)

			returnData, returnCode, returnCacheStatus = doSimpleTextServeConnHandlerCustomProcess(h, conn, "500", routeExtraData, false, errCodeMap)
			returnCode = "500"
		}

		if conn.Cache().IsDisabled() {
			returnData = file

			goto GOTO_NO_CACHE
		} else {
			if _, err = io.Copy(buf, file); err != nil {
				conn.Logger.Errorf("unable to open %s", file)

				returnData, returnCode, returnCacheStatus = doSimpleTextServeConnHandlerCustomProcess(h, conn, "500", routeExtraData, false, errCodeMap)
				returnCode = "500"
			}

			returnData = buf.Bytes()

			goto GOTO_ADD_TO_CACHE
		}
	}

	/*
	 ********************************************************************************
	 *
	 * Templating
	 *
	 ********************************************************************************
	 */

	//
	// not found in cache, parsing templates
	//
	conn.Logger.Tracef("parsing template...")

	// get funcMap
	funcMap, err = h.GetTemplatesFuncMap(conn)

	//
	// loading templates
	//

	// main template
	buftmpl, err = ioutil.ReadFile(templateFilePath)
	if err == nil {
		tpl, err = template.New(templateName).
			Funcs(funcMap).
			Parse(string(buftmpl))

		if err == nil {
			// header
			if ttutils.NotStringEmpty(spaceConfig.Header) {
				conn.Logger.Tracef("parsing header file %s", ttutils.StringValue(spaceConfig.Header))
				buftmpl, err = ioutil.ReadFile(ttutils.StringValue(spaceConfig.Header))
				if err == nil {
					tpl, err = tpl.New(filepath.Base(ttutils.StringValue(spaceConfig.Header))).
						Funcs(funcMap).
						Parse(string(buftmpl))
				}
			}

			// footer
			if err == nil && ttutils.NotStringEmpty(spaceConfig.Footer) {
				conn.Logger.Tracef("parsing footer file %s", ttutils.StringValue(spaceConfig.Footer))
				buftmpl, err = ioutil.ReadFile(ttutils.StringValue(spaceConfig.Footer))
				if err == nil {
					tpl, err = tpl.New(filepath.Base(ttutils.StringValue(spaceConfig.Footer))).
						Funcs(funcMap).
						Parse(string(buftmpl))
				}
			}
		}
	}

	if err != nil {
		conn.Logger.Errorf("template parsing error -> %s", err)

		if route == "404" || route == "500" {
			returnData = errCodeMap[route]
			returnData = append(returnData.([]byte), []byte(CRLF)...)
			returnCode = route
		} else {
			returnData, returnCode, returnCacheStatus = doSimpleTextServeConnHandlerCustomProcess(h, conn, "500", routeExtraData, false, errCodeMap)
			returnCode = "500"
		}

		goto GOTO_ADD_TO_CACHE
	}

	//
	// get templateData
	//
	conn.Logger.Tracef("retrieving templateData...")

	// default
	templateData["default"] = &simpleTextHandlerTemplateDataItem{
		Route:        route,
		Domain:       conn.Domain,
		Port:         conn.Port,
		LocalAddress: localAddr,
		LocalPort:    localPort,
	}

	if routeConfig.Fetch != nil {
		wgFetch := sync.WaitGroup{}
		var poolItems []*simpleTextHandlerFetcher

		// create a pool
		poolFetch := tunny.NewFunc(5, func(payload interface{}) interface{} {
			p := payload.(*simpleTextHandlerFetcher)

			conn.Logger.Tracef("fetching %s for '%s'...", p.uri, p.name)

			fetched, err := p.fetcher(p.uri)
			if err != nil {
				conn.Logger.Errorf("fetching %s for '%s' failed -> %s", p.uri, p.name, err)
				return nil
			}

			conn.Logger.Tracef("fetching done %s for '%s'...", p.uri, p.name)

			return &simpleTextHandlerTemplateDataItem{
				Route:  route,
				Domain: conn.Domain,
				Port:   conn.Port,
				URI:    p.uri,
				Type:   p.fetchType,
				Data:   fetched,
			}
		})
		defer poolFetch.Close()

		// prepare the jobs
		for fetchName, fetchItem := range routeConfig.Fetch {
			fetchType := ttutils.StringValue(fetchItem.Type)
			fetchURL := ttutils.StringValue(fetchItem.URI)

			if !ttutils.IsStringSliceEmpty(routeConfig.RegexpCapturedGroups) {
				for i, capturedGroup := range routeConfig.RegexpCapturedGroups {
					if i == 0 {
						continue
					}
					fetchURL = strings.Replace(fetchURL, "$"+strconv.Itoa(i), capturedGroup, -1)
				}
			}

			switch fetchType {
			case "html":
				poolItems = append(poolItems, &simpleTextHandlerFetcher{
					fetcher:   func(url string) (interface{}, error) { return ttutils.GetURL(getFeedTimeout, url) },
					name:      fetchName,
					fetchType: fetchType,
					uri:       fetchURL,
				})
			case "feed":
				poolItems = append(poolItems, &simpleTextHandlerFetcher{
					fetcher:   func(url string) (interface{}, error) { return ttutils.GetFeed(getFeedTimeout, url) },
					name:      fetchName,
					fetchType: fetchType,
					uri:       fetchURL,
				})
			case "json":
				poolItems = append(poolItems, &simpleTextHandlerFetcher{
					fetcher:   func(url string) (interface{}, error) { return ttutils.GetJSON(getFeedTimeout, url) },
					name:      fetchName,
					fetchType: fetchType,
					uri:       fetchURL,
				})
			case "prometheus":
				poolItems = append(poolItems, &simpleTextHandlerFetcher{
					fetcher:   func(url string) (interface{}, error) { return ttutils.GetPrometheus(getFeedTimeout, url) },
					name:      fetchName,
					fetchType: fetchType,
					uri:       fetchURL,
				})
			}
		}

		// run
		wgFetch.Add(len(poolItems))

		for i := range poolItems {
			go func(item *simpleTextHandlerFetcher) {
				defer func() {
					conn.MutexFetch.Unlock()
					wgFetch.Done()
				}()

				if result := poolFetch.Process(item); result != nil {
					conn.MutexFetch.Lock()
					templateData[item.name] = result.(*simpleTextHandlerTemplateDataItem)
				} else {
					conn.MutexFetch.Lock()
					templateData[item.name] = nil
				}

			}(poolItems[i])
		}

		// wait
		wgFetch.Wait()
	}

	//
	// execute
	//
	conn.Logger.Tracef("executing template...")

	err = tpl.ExecuteTemplate(buf, templateName, templateData)
	if err != nil {
		conn.Logger.Errorf("template execution error -> %s", err)

		if route == "404" || route == "500" {
			returnData = errCodeMap[route]
			returnData = append(returnData.([]byte), []byte(CRLF)...)
			returnCode = route
		} else {
			returnData, returnCode, returnCacheStatus = doSimpleTextServeConnHandlerCustomProcess(h, conn, "500", routeExtraData, false, errCodeMap)
			returnCode = "500"
		}

		goto GOTO_ADD_TO_CACHE
	}

	returnData = buf.Bytes()
	returnData = append(returnData.([]byte), []byte(CRLF)...)

	//
	// post processing
	//
	conn.Logger.Tracef("template post-processing...")

	returnData, err = h.PostProcess(conn, route, routeExtraData, returnData)
	if err != nil {
		conn.Logger.Errorf("template post-processing error -> %s", err)

		if route == "404" || route == "500" {
			returnData = errCodeMap[route]
			returnData = append(returnData.([]byte), []byte(CRLF)...)
			returnCode = route
		} else {
			returnData, returnCode, returnCacheStatus = doSimpleTextServeConnHandlerCustomProcess(h, conn, "500", routeExtraData, false, errCodeMap)
			returnCode = "500"
		}

		goto GOTO_ADD_TO_CACHE
	}

	// no cache if unable to fetch remote data
	for _, v := range templateData {
		if v == nil {
			goto GOTO_NO_CACHE
		}
	}

GOTO_ADD_TO_CACHE:
	if conn.Cache().IsEnabled() {
		conn.Logger.Tracef("adding to cache")

		if forceCacheUpdate {
			err = conn.CacheReplaceIfExists(
				route,
				&ttcache.Item{
					Item: returnData,
					Code: returnCode,
				},
				time.Duration(ttutils.IntValue(routeConfig.Cache.Expiration))*time.Second,
			)
		} else {
			err = conn.CacheAdd(
				route,
				&ttcache.Item{
					Item: returnData,
					Code: returnCode,
				},
				time.Duration(ttutils.IntValue(routeConfig.Cache.Expiration))*time.Second,
			)
		}

		if err != nil {
			conn.Logger.Errorf("adding to cache failed -> %s", err)
		}

		if forceCacheUpdate {
			returnCacheStatus = "REPLACE"
		} else {
			returnCacheStatus = "MISS"
		}
	} else {
		returnCacheStatus = "NOCACHE"

		conn.Logger.Tracef("not adding to cache")
	}

GOTO_NO_CACHE:
	//
	// the end
	//

	// prometheus
	var endTime = time.Now()

	if err := conn.PrometheusFire(&ttprom.PrometheusMetric{
		Metric: ttprom.PrometheusProcessDurationSummary,
		Labels: []string{route, returnCode},
		Action: "observe",
		Values: float64(endTime.Sub(startTime).Microseconds()),
	}); err != nil {
		conn.Logger.Errorf("firing prometheus failed -> %s", err)
	}

	if err := conn.PrometheusFire(&ttprom.PrometheusMetric{
		Metric: ttprom.PrometheusRouteCacheStatusCounter,
		Labels: []string{route, returnCacheStatus},
		Action: "inc",
	}); err != nil {
		conn.Logger.Errorf("firing prometheus failed -> %s", err)
	}

	// return
	return returnData, returnCode, returnCacheStatus
}
