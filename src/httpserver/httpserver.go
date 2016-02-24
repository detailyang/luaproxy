/*
* @Author: detailyang
* @Date:   2016-02-20 18:09:48
* @Last Modified by:   detailyang
* @Last Modified time: 2016-02-25 00:47:13
 */

package httpserver

import (
	"github.com/Jeffail/gabs"
	"github.com/garyburd/redigo/redis"
	"log"
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

	//
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

func (self *HttpServer) ip(w http.ResponseWriter, r *http.Request) {
	lastslash := strings.LastIndex(r.URL.Path, "/")
	if lastslash == -1 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 not found"))
		return
	}
	ip := r.URL.Path[lastslash+1:]

	jsonObj := gabs.New()
	jsonObj.SetP(0, "code")
	jsonObj.SetP("ok", "msg")
	jsonObj.Array("data", "request")
	jsonObj.Array("data", "upstream")
	jsonObj.Array("data", "response")

	red := self.redispool.Get()
	// request
	requestcommasep, err := redis.String(red.Do("HGET", ip, "request"))
	if err != nil {
		log.Printf("hget %s request error %s", ip, err.Error())
	} else if requestcommasep != "" {
		for _, v := range strings.Split(requestcommasep, ",") {
			jsonObj.ArrayAppendP(v, "data.request")
		}
	}
	// upstream
	upstreamcommasep, err := redis.String(red.Do("HGET", ip, "upstream"))
	if err != nil {
		log.Printf("hget %s upstream error %s", ip, err.Error())
	} else if upstreamcommasep != "" {
		for _, v := range strings.Split(upstreamcommasep, ",") {
			jsonObj.ArrayAppendP(v, "data.upstream")
		}
	}
	// response
	responsecommasep, err := redis.String(red.Do("HGET", ip, "response"))
	if err != nil {
		log.Printf("hget %s response error %s", ip, err.Error())
	} else if responsecommasep != "" {
		for _, v := range strings.Split(responsecommasep, ",") {
			jsonObj.ArrayAppendP(v, "data.response")
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonObj.Bytes())
}

func (self *HttpServer) view(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, self.indexhtml)
}
