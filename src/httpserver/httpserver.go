/*
* @Author: detailyang
* @Date:   2016-02-20 18:09:48
* @Last Modified by:   detailyang
* @Last Modified time: 2016-02-26 23:05:09
 */

package httpserver

import (
	"github.com/Jeffail/gabs"
	"github.com/garyburd/redigo/redis"
	"log"
	"net"
	"net/http"
	"strings"
)

type HttpServer struct {
	redispool *redis.Pool
	addr      string
	staticdir string
	indexhtml string
	luaplugin map[string]string
}

func NewHttpServer(redispool *redis.Pool, addr, staticdir, indexhtml string, luaplugin map[string]string) *HttpServer {
	return &HttpServer{
		redispool: redispool,
		addr:      addr,
		staticdir: staticdir,
		indexhtml: indexhtml,
		luaplugin: luaplugin,
	}
}

func (self *HttpServer) ListenAndServe() {
	// static file server
	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir(self.staticdir))))

	// dont use another web router, we just want simple router:)
	http.HandleFunc("/api/luaplugins/", self.luaplugindetail)
	http.HandleFunc("/api/luaplugins", self.luapluginlist)

	http.HandleFunc("/api/ip/", self.ip)

	// index html
	http.HandleFunc("/", self.view)

	log.Printf("http server listen %s", self.addr)
	err := http.ListenAndServe(self.addr, nil)
	if err != nil {
		log.Println("http server listen error ", self.addr, err)
	}
}

func (self *HttpServer) luapluginlist(w http.ResponseWriter, r *http.Request) {
	jsonObj := gabs.New()
	jsonObj.SetP(0, "code")
	jsonObj.SetP("ok", "msg")
	jsonObj.Array("data", "request")
	jsonObj.Array("data", "upstream")
	jsonObj.Array("data", "response")
	for k, _ := range self.luaplugin {
		if strings.HasPrefix(k, "request") {
			jsonObj.ArrayAppendP(k, "data.request")
		} else if strings.HasPrefix(k, "upstream") {
			jsonObj.ArrayAppendP(k, "data.upstream")
		} else if strings.HasPrefix(k, "response") {
			jsonObj.ArrayAppendP(k, "data.response")
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonObj.Bytes())
}

func (self *HttpServer) luaplugindetail(w http.ResponseWriter, r *http.Request) {
	lastslash := strings.LastIndex(r.URL.Path, "/")
	if lastslash == -1 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 not found"))
		return
	}
	w.Write([]byte(self.luaplugin[r.URL.Path[lastslash+1:]]))
}

func (self *HttpServer) ipGet(w http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Println("split host port error ", r.RemoteAddr)
	}

	jsonObj := gabs.New()
	jsonObj.SetP(0, "code")
	jsonObj.SetP("ok", "msg")
	jsonObj.Array("data", "request")
	jsonObj.Array("data", "upstream")
	jsonObj.Array("data", "response")

	red := self.redispool.Get()
	// request
	requestcommasep, err := redis.String(red.Do("HGET", host, "request"))
	if err != nil && requestcommasep != "" {
		log.Printf("hget %s request error %s", host, err.Error())
	} else if requestcommasep != "" {
		for _, v := range strings.Split(requestcommasep, ",") {
			jsonObj.ArrayAppendP(v, "data.request")
		}
	}
	// upstream
	upstreamcommasep, err := redis.String(red.Do("HGET", host, "upstream"))
	if err != nil && upstreamcommasep != "" {
		log.Printf("hget %s upstream error %s", host, err.Error())
	} else {
		for _, v := range strings.Split(upstreamcommasep, ",") {
			jsonObj.ArrayAppendP(v, "data.upstream")
		}
	}
	// response
	responsecommasep, err := redis.String(red.Do("HGET", host, "response"))
	if err != nil && responsecommasep != "" {
		log.Printf("hget %s response error %s", host, err.Error())
	} else {
		for _, v := range strings.Split(responsecommasep, ",") {
			jsonObj.ArrayAppendP(v, "data.response")
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonObj.Bytes())
}

func (self *HttpServer) ipPost(w http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Println("split host port error ", r.RemoteAddr)
	}

	err = r.ParseForm()
	if err != nil {
		log.Println("parse form error ", err)
	}

	red := self.redispool.Get()
	for _, kind := range []string{"request", "upstream", "response"} {
		_, err = red.Do("HSET", host, kind, r.Form.Get(kind))
		if err != nil {
			log.Printf("HSET %s request error %s", host, err.Error())
		}
	}

	w.Write([]byte("abcd"))
}

func (self *HttpServer) ip(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		self.ipGet(w, r)
	case "POST":
		self.ipPost(w, r)
	default:
		self.ipGet(w, r)
	}
}

func (self *HttpServer) view(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, self.indexhtml)
}
