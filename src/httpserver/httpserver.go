/*
* @Author: detailyang
* @Date:   2016-02-20 18:09:48
* @Last Modified by:   detailyang
* @Last Modified time: 2016-02-20 18:18:40
 */

package httpserver

import (
	"log"
	"net/http"
)

type HttpServer struct {
	addr string
}

func NewHttpServer(addr string) *HttpServer {
	return &HttpServer{
		addr: addr,
	}
}

func (self *HttpServer) ListenAndServe() {
	err := http.ListenAndServe(self.addr, http.HandlerFunc(self.handle))
	if err != nil {
		log.Println("http server listen error ", self.addr, err)
	}
}

func (self *HttpServer) handle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("abcd"))
}
