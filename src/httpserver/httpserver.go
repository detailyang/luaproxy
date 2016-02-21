/*
* @Author: detailyang
* @Date:   2016-02-20 18:09:48
* @Last Modified by:   detailyang
* @Last Modified time: 2016-02-21 23:47:35
 */

package httpserver

import (
	"github.com/Jeffail/gabs"
	"log"
	"net/http"
	"strings"
)

type HttpServer struct {
	addr      string
	staticdir string
	indexhtml string
	luaplugin map[string]string
}

func NewHttpServer(addr, staticdir, indexhtml string, luaplugin map[string]string) *HttpServer {
	return &HttpServer{
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

	// index html
	http.HandleFunc("/", self.view)
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
	log.Println(r.URL.Path)
	lastslash := strings.LastIndex(r.URL.Path, "/")
	if lastslash == -1 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 not found"))
		return
	}
	w.Write([]byte(self.luaplugin[r.URL.Path[lastslash+1:]]))
}

func (self *HttpServer) view(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, self.indexhtml)
}
