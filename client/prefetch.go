package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"strings"

	"github.com/progrhyme/binq/internal/erron"
	"github.com/progrhyme/binq/schema"
)

var errIndexDataNotFound = errors.New("Index Data is Not Found on given URL")

// prefetch query metadata for item info to fetch
func (r *Runner) prefetch() (err error) {
	if strings.HasPrefix(r.Source, "http") {
		r.sourceURL = r.Source
		return nil
	}
	if r.ServerURL == nil {
		return fmt.Errorf("No server is configured. Can't deal with source: %s", r.Source)
	}

	// Use different timeout for index server
	hc := newHttpClient(httpTimeoutToQueryIndex)

	item, _err := r.prefetchItemByURL(hc, r.Source)
	switch _err {
	case nil:
		// OK
	case errIndexDataNotFound:
		// Retry
		if item, err = r.prefetchItemByIndex(hc); err != nil {
			return err
		}
	default:
		return _err
	}

	srcURL, err := item.GetLatestURL(schema.ItemURLParam{OS: runtime.GOOS, Arch: runtime.GOARCH})
	if err != nil {
		return err
	}
	if srcURL == "" {
		return fmt.Errorf("Can't get source URL from JSON")
	}

	r.sourceURL = srcURL

	return err
}

func (r *Runner) prefetchItemByURL(hc *http.Client, urlPath string) (item *schema.Item, err error) {
	// Copy r.ServerURL
	urlItem, _err := url.Parse(r.ServerURL.String())
	if _err != nil {
		// Unexpected case
		return item, erron.Errorwf(_err, "Failed to parse server URL: %v", r.ServerURL)
	}
	urlItem.Path = path.Join(urlItem.Path, urlPath)

	headers := make(map[string]string)
	headers["Accept"] = "application/json"
	req, err := newHttpGetRequest(urlItem.String(), headers)
	if err != nil {
		return item, err
	}
	r.Logger.Infof("GET %s", urlItem.String())

	res, _err := hc.Do(req)
	if _err != nil {
		err = erron.Errorwf(_err, "Failed to execute HTTP request")
		return item, err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 200:
		// OK
	case 404:
		r.Logger.Debugf("Index Item Data is Not Found: %s", urlItem.String())
		return item, errIndexDataNotFound
	default:
		err = fmt.Errorf("HTTP response is not OK. Code: %d, URL: %s", res.StatusCode, urlItem.String())
		return item, err
	}

	var b strings.Builder
	_, _err = io.Copy(&b, res.Body)
	if _err != nil {
		err = erron.Errorwf(_err, "Failed to read HTTP response")
		return item, err
	}

	item, err = schema.DecodeItemJSON([]byte(b.String()))
	if err != nil {
		return item, err
	}
	r.Logger.Debugf("Decoded JSON: %s", item)

	return item, nil
}

func (r *Runner) prefetchItemByIndex(hc *http.Client) (item *schema.Item, err error) {
	index, _err := r.prefetchIndex(hc, "")
	switch _err {
	case nil:
		// OK
	case errIndexDataNotFound:
		if index, _err = r.prefetchIndex(hc, "index.json"); _err != nil {
			err = erron.Errorwf(_err, "Can't get Index data from server: %s", r.ServerURL.String())
		}
		return item, err
	default:
		return item, err
	}

	urlPath := index.FindPath(r.Source)
	switch urlPath {
	case "":
		err = fmt.Errorf("Can't find item in index. Server: %s", r.ServerURL.String())
		return item, err
	case r.Source:
		err = fmt.Errorf(
			"Found path equals to specified Source. Won't retry. Source: %s, Server: %s",
			r.Source, r.ServerURL.String())
		return item, err
	default:
		// OK
	}

	if item, _err = r.prefetchItemByURL(hc, urlPath); _err != nil {
		err = erron.Errorwf(_err, "Failed to get Item Data on path: %s", urlPath)
		return item, err
	}

	return item, nil
}

func (r *Runner) prefetchIndex(hc *http.Client, url string) (index *schema.Index, err error) {
	if url == "" {
		url = r.ServerURL.String()
	}
	headers := make(map[string]string)
	headers["Accept"] = "application/json"
	req, err := newHttpGetRequest(url, headers)
	if err != nil {
		return index, err
	}
	r.Logger.Infof("GET %s", url)

	res, _err := hc.Do(req)
	if _err != nil {
		err = erron.Errorwf(_err, "Failed to execute HTTP request")
		return index, err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 200:
		// OK
	case 404:
		r.Logger.Debugf("Index Data is Not Found: %s")
		return index, errIndexDataNotFound
	default:
		err = fmt.Errorf("HTTP response is not OK. Code: %d, URL: %s", res.StatusCode, url)
		return index, err
	}

	var b strings.Builder
	_, _err = io.Copy(&b, res.Body)
	if _err != nil {
		err = erron.Errorwf(_err, "Failed to read HTTP response")
		return index, err
	}

	index, err = schema.DecodeIndexJSON([]byte(b.String()))
	if err != nil {
		return index, err
	}
	r.Logger.Debugf("Decoded JSON: %s", index)

	return index, nil
}