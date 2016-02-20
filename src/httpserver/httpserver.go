/*
* @Author: detailyang
* @Date:   2016-02-20 18:09:48
* @Last Modified by:   detailyang
* @Last Modified time: 2016-02-20 23:57:47
 */

package httpserver

import (
	"fmt"
	"log"
	"net/http"
)

type HttpServer struct {
	addr      string
	staticdir string
	indexhtml string
}

func NewHttpServer(addr, staticdir, indexhtml string) *HttpServer {
	return &HttpServer{
		addr:      addr,
		staticdir: staticdir,
		indexhtml: indexhtml,
	}
}

func (self *HttpServer) ListenAndServe() {
	// static file server
	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir(self.staticdir))))

	// dont use another web router, we just want simple router:)
	http.HandleFunc("/api/", self.api)

	// index html
	http.HandleFunc("/", self.view)
	err := http.ListenAndServe(self.addr, nil)
	if err != nil {
		log.Println("http server listen error ", self.addr, err)
	}
}

func (self *HttpServer) api(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "api")
}

func (self *HttpServer) view(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, self.indexhtml)
}
