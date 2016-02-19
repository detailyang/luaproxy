/*
* @Author: detailyang
* @Date:   2016-02-12 21:35:29
* @Last Modified by:   detailyang
* @Last Modified time: 2016-02-19 20:42:35
 */

package utils

import (
	"bytes"
	"github.com/stevedonovan/luar"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func GoReqMergeLuaReq(greq *http.Request, lreq luar.Map) {
	//copy header
	if lreq["header"] != nil {
		header := make(http.Header)
		for k, v := range lreq["header"].(luar.Map) {
			header.Add(k, v.(string))
		}
		greq.Header = header
	}
	//copy method
	if lreq["method"] != nil {
		method := lreq["method"].(string)
		greq.Method = method
	}
	//copy url
	if lreq["url"] != nil {
		lurl := lreq["url"].(luar.Map)
		u := &url.URL{}

		u.Scheme = lurl["scheme"].(string)
		u.Host = lurl["host"].(string)
		u.Path = lurl["path"].(string)
		u.RawQuery = lurl["query"].(string)
		u.Fragment = lurl["fragment"].(string)
		greq.URL = u
	}

	//copy body
	if lreq["body_changed"] != nil {
		//lua change the http body
		if lreq["body_changed"].(bool) != false {
			body := lreq["body"].(string)
			greq.Body = StringReaderCloser{bytes.NewBufferString(body)}
			greq.ContentLength = int64(len(body))
		}
	}

	//change upstream
	if lreq["upstream"] != nil {
		greq.URL.Host = lreq["upstream"].(string)
	}
}

func GoReqToLuaReq(r *http.Request) luar.Map {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("read request body error ", err)
		body = nil
	}
	defer r.Body.Close()

	req := luar.Map{
		"url":          luar.Map{},
		"query":        luar.Map{},
		"upstream":     r.URL.Host,
		"method":       r.Method,
		"address":      r.RemoteAddr,
		"header":       luar.Map{},
		"body":         string(body),
		"body_changed": false,
	}

	//copy url
	if r.URL.Scheme == "" {
		req["url"].(luar.Map)["scheme"] = "http"
	} else {
		req["url"].(luar.Map)["scheme"] = r.URL.Scheme
	}
	if r.URL.Host == "" {
		req["url"].(luar.Map)["host"] = r.Host
	} else {
		req["url"].(luar.Map)["host"] = r.URL.Host
	}
	req["url"].(luar.Map)["path"] = r.URL.Path
	req["url"].(luar.Map)["query"] = r.URL.RawQuery
	req["url"].(luar.Map)["fragment"] = r.URL.Fragment

	//copy header
	for k, v := range r.Header {
		for _, vv := range v {
			req["header"].(luar.Map)[strings.ToLower(k)] = vv
		}
	}

	//copy query
	for k, v := range r.URL.Query() {
		for _, vv := range v {
			req["query"].(luar.Map)[k] = vv
		}
	}

	return req
}

func GoResToLuaRes(r *http.Response) luar.Map {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("read response body error ", err)
	}
	defer r.Body.Close()

	res := luar.Map{
		"status":     r.Status,
		"statuscode": r.StatusCode,
		"header":     luar.Map{},
		"body":       string(body),
	}

	//copy header
	for k, v := range r.Header {
		for _, vv := range v {
			res["header"].(luar.Map)[strings.ToLower(k)] = vv
		}
	}

	return res
}

func GoResMergeLuaRes(gres *http.Response, lres luar.Map) {
	//copy header
	if lres["header"] != nil {
		header := make(http.Header)
		for k, v := range lres["header"].(luar.Map) {
			header.Add(k, v.(string))
		}
		gres.Header = header
	}

	//copy body
	if lres["body"] != nil {
		//lua change the http body
		body := lres["body"].(string)
		gres.Body = StringReaderCloser{bytes.NewBufferString(body)}
		gres.ContentLength = int64(len(body))
	}

	//copy status
	if lres["status"] != nil {
		status := lres["status"].(string)
		gres.Status = status
	}

	//copy status code
	if lres["statuscode"] != nil {
		statuscode := lres["statuscode"].(int)
		gres.StatusCode = statuscode
	}
}
