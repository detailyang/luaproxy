/*
* @Author: detailyang
* @Date:   2016-02-11 17:35:16
* @Last Modified by:   detailyang
* @Last Modified time: 2016-02-19 17:13:23
 */

package httpproxy

import (
	"github.com/garyburd/redigo/redis"
	"github.com/stevedonovan/luar"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"utils"
)

const (
	LuaRequestType  = 0
	LuaResponseType = 1
	LuaUpstreamType = 2
)

type HttpProxy struct {
	redispool   *redis.Pool
	luarequest  map[string]string
	luaresponse map[string]string
	luaupstream map[string]string
}

func NewHttpProxy(redispool *redis.Pool, luarequest, luaupstream, luaresponse map[string]string) *HttpProxy {
	return &HttpProxy{
		redispool:   redispool,
		luarequest:  luarequest,
		luaresponse: luaresponse,
		luaupstream: luaupstream,
	}
}

func (self *HttpProxy) ListenAndServe(addr string) {
	err := http.ListenAndServe(addr, http.HandlerFunc(self.handle("http")))
	if err != nil {
		log.Println("http listen error ", addr, err)
	}
}

func (self *HttpProxy) ListenAndServeTLS(addr, certfile, keyfile string) {
	err := http.ListenAndServeTLS(addr, certfile, keyfile, http.HandlerFunc(self.handle("https")))
	if err != nil {
		log.Println("https listen error ", addr, err)
	}
}

func (self *HttpProxy) loadLuaCode(luaplugintype int, ip string) []string {
	r := self.redispool.Get()
	defer r.Close()

	var err error
	var csv string
	var m *map[string]string

	switch luaplugintype {
	case LuaRequestType:
		csv, err = redis.String(r.Do("hget", ip, "request"))
		m = &self.luarequest
	case LuaResponseType:
		csv, err = redis.String(r.Do("hget", ip, "response"))
		m = &self.luaresponse
	case LuaUpstreamType:
		csv, err = redis.String(r.Do("hget", ip, "upstream"))
		m = &self.luaupstream
	default:
		csv, err = redis.String(r.Do("hget", ip, "request"))
		m = &self.luarequest
	}

	if err != nil && csv != "" {
		log.Println(err)
		return []string{}
	}

	rv := make([]string, 0)
	ids := strings.Split(csv, ",")
	for _, id := range ids {
		rv = append(rv, (*m)[id])
	}

	return rv
}

func (self *HttpProxy) request(goreq *http.Request) *http.Request {
	goreq.Header.Add("host", goreq.Host)
	goreq.Host = goreq.Host
	goreq.URL.Host = goreq.Host
	goreq.RequestURI = ""
	luareq := utils.GoReqToLuaReq(goreq)

	luavm := luar.Init()
	defer luavm.Close()
	// register context
	luar.Register(luavm, "", luar.Map{
		"req": luareq,
	})
	host, _, err := net.SplitHostPort(goreq.RemoteAddr)
	if err != nil {
		log.Println("split error ", err)
	}

	// load request
	for _, luacode := range self.loadLuaCode(LuaRequestType, host) {
		if luacode == "" {
			continue
		}
		err := luavm.DoString(luacode)
		if err != nil {
			log.Println("lua get error ", err)
			continue
		}
		luareqfun := luar.NewLuaObjectFromName(luavm, "request")
		_, err = luareqfun.Call()
		if err != nil {
			log.Println("lua request call get error ", err)
		}
	}

	// load upstream
	for _, luacode := range self.loadLuaCode(LuaUpstreamType, host) {
		if luacode == "" {
			continue
		}
		err := luavm.DoString(luacode)
		if err != nil {
			log.Println("lua get error ", err)
			continue
		}
		luaupstreamfun := luar.NewLuaObjectFromName(luavm, "upstream")
		_, err = luaupstreamfun.Call()
		if err != nil {
			log.Println("lua upstream call get error ", err)
		}
	}
	utils.GoReqMergeLuaReq(goreq, luareq)
	return goreq
}

func (self *HttpProxy) response(goreq *http.Request, gores *http.Response) {
	luareq := utils.GoReqToLuaReq(goreq)
	luares := utils.GoResToLuaRes(gores)

	luavm := luar.Init()
	defer luavm.Close()
	// register context
	luar.Register(luavm, "", luar.Map{
		"res": luares,
		"req": luareq,
	})

	host, _, err := net.SplitHostPort(goreq.RemoteAddr)
	if err != nil {
		log.Println("split error ", err)
	}
	for _, luacode := range self.loadLuaCode(LuaResponseType, host) {
		if luacode == "" {
			continue
		}
		err := luavm.DoString(luacode)
		if err != nil {
			log.Println("lua get error ", err)
		}
		luaresfun := luar.NewLuaObjectFromName(luavm, "response")
		_, err = luaresfun.Call()
		if err != nil {
			log.Println("lua response call get error ", err)
		}
	}
	utils.GoResMergeLuaRes(gores, luares)
}

func (self *HttpProxy) handle(protocol string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Scheme = protocol
		goreq := self.request(r)
		client := &http.Client{}
		client.CheckRedirect = func(goreq *http.Request, via []*http.Request) error {
			return utils.ErrHttpRedirect
		}
		gores, err := client.Do(goreq)
		if err != nil && gores == nil {
			http.Error(w, "Error contacting backend server.", 500)
			log.Println("request upstream error", err)
			return
		}
		self.response(goreq, gores)

		for k, v := range gores.Header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}
		w.WriteHeader(gores.StatusCode)
		body, err := ioutil.ReadAll(gores.Body)
		if err != nil && gores.StatusCode < 300 && gores.StatusCode >= 400 {
			log.Println("read body error ", err)
			return
		}
		nw, err := w.Write(body)
		if err != nil {
			log.Println("write body error ", err)
		}
		log.Printf("read %d bytes to client", nw)
	}
}
