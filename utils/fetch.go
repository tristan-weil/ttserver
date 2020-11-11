package utils

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/prom2json"
)

func GetURL(timeout time.Duration, url string) ([]byte, error) {
	ctxFeed, ctxFeedCancel := context.WithTimeout(context.Background(), timeout)
	defer ctxFeedCancel()

	req, err := http.NewRequestWithContext(ctxFeed, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []byte
	if _, err := resp.Body.Read(result); err != nil {
		return nil, err
	}

	return result, nil
}

func GetJSON(timeout time.Duration, url string) (map[string]interface{}, error) {
	myClient := &http.Client{Timeout: timeout}
	data := make(map[string]interface{})

	r, err := myClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func GetFeed(timeout time.Duration, url string) (*gofeed.Feed, error) {
	ctxFeed, ctxFeedCancel := context.WithTimeout(context.Background(), timeout)
	defer ctxFeedCancel()

	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(url, ctxFeed)

	if err != nil {
		return nil, err
	}

	return feed, nil
}

func GetPrometheus(timeout time.Duration, url string) (map[string]*prom2json.Family, error) {
	mfChan := make(chan *dto.MetricFamily, 1024)
	errChan := make(chan error)
	defer close(errChan)

	// Initialize with the DefaultTransport for sane defaults.
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// Conservatively disable HTTP keep-alives as this program will only
	// ever need a single HTTP request.
	transport.DisableKeepAlives = true
	// Timeout early if the server doesn't even return the headers.
	transport.ResponseHeaderTimeout = timeout

	go func() {
		if err := prom2json.FetchMetricFamilies(url, mfChan, transport); err != nil {
			errChan <- err
			return
		}

		errChan <- nil
	}()

	for err := range errChan {
		if err != nil {
			return nil, err
		}

		break
	}

	result := map[string]*prom2json.Family{}
	for mf := range mfChan {
		result[*mf.Name] = prom2json.NewFamily(mf)
	}

	return result, nil
}
