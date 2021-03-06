/*
* @Author: detailyang
* @Date:   2016-02-11 17:35:16
* @Last Modified by:   detailyang
* @Last Modified time: 2016-02-25 00:12:43
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
	redispool           *redis.Pool
	luaplguin           map[string]string
	httpsaddr, httpaddr string
}

func NewHttpProxy(redispool *redis.Pool, httpaddr, httpsaddr string, luaplugin map[string]string) *HttpProxy {
	return &HttpProxy{
		redispool: redispool,
		luaplguin: luaplugin,
		httpaddr:  httpaddr,
		httpsaddr: httpsaddr,
	}
}

func (self *HttpProxy) ListenAndServe() {
	log.Printf("http listen on %s", self.httpaddr)
	err := http.ListenAndServe(self.httpaddr, http.HandlerFunc(self.handle("http")))
	if err != nil {
		log.Println("http listen error ", self.httpaddr, err)
	}
}

func (self *HttpProxy) ListenAndServeTLS(certfile, keyfile string) {
	log.Printf("https listen on %s", self.httpsaddr)
	err := http.ListenAndServeTLS(self.httpsaddr, certfile, keyfile, http.HandlerFunc(self.handle("https")))
	if err != nil {
		log.Println("https listen error ", self.httpsaddr, err)
	}
}

func (self *HttpProxy) getLuaEnvMap(goreq *http.Request) luar.Map {
	host, port, err := net.SplitHostPort(goreq.RemoteAddr)
	if err != nil {
		log.Println("split error ", err)
	}
	r := self.redispool.Get()
	defer r.Close()

	username, err := redis.String(r.Do("hget", host, "username"))
	if err != nil && username != "" {
		log.Printf("get %s username error ", err)
	}
	if username == "" {
		username = "hijack"
	}

	env := luar.Map{
		"host":     host,
		"port":     port,
		"username": username,
	}

	return env
}

func (self *HttpProxy) loadLuaCode(luaplugintype int, ip string) []string {
	r := self.redispool.Get()
	defer r.Close()

	var err error
	var csv string

	switch luaplugintype {
	case LuaRequestType:
		csv, err = redis.String(r.Do("hget", ip, "request"))
	case LuaResponseType:
		csv, err = redis.String(r.Do("hget", ip, "response"))
	case LuaUpstreamType:
		csv, err = redis.String(r.Do("hget", ip, "upstream"))
	default:
		csv, err = redis.String(r.Do("hget", ip, "request"))
	}

	if err != nil && csv != "" {
		log.Println(err)
		return []string{}
	}

	rv := make([]string, 0)
	filenames := strings.Split(csv, ",")
	for _, filename := range filenames {
		code := self.luaplguin[filename]
		rv = append(rv, code)
	}

	return rv
}

func (self *HttpProxy) request(goreq *http.Request) *http.Request {
	goreq.Header.Add("host", goreq.Host)
	goreq.Host = goreq.Host
	goreq.URL.Host = goreq.Host
	goreq.RequestURI = ""
	luareq := utils.GoReqToLuaReq(goreq)
	luaenv := self.getLuaEnvMap(goreq)

	luavm := luar.Init()
	defer luavm.Close()
	// register context
	luar.Register(luavm, "", luar.Map{
		"req": luareq,
		"env": luaenv,
	})

	// load request
	for _, luacode := range self.loadLuaCode(LuaRequestType, luaenv["host"].(string)) {
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
	for _, luacode := range self.loadLuaCode(LuaUpstreamType, luaenv["host"].(string)) {
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
	luaenv := self.getLuaEnvMap(goreq)

	luavm := luar.Init()
	defer luavm.Close()
	// register context
	luar.Register(luavm, "", luar.Map{
		"res": luares,
		"req": luareq,
		"env": luaenv,
	})

	for _, luacode := range self.loadLuaCode(LuaResponseType, luaenv["host"].(string)) {
		if luacode == "" {
			continue
		}
		log.Println(luacode)
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

func (self *HttpProxy) isKillSelf(hostport string) bool {
	_, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return true
	}
	if port == self.httpaddr || port == self.httpsaddr {
		return true
	}

	return false
}

func (self *HttpProxy) handle(protocol string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Scheme = protocol
		goreq := self.request(r)
		if self.isKillSelf(goreq.URL.Host) == true {
			http.Error(w, "you shouldnt proxy server itself:(, please set the right upstream", 400)
			return
		}
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
		defer gores.Body.Close()

		self.response(goreq, gores)
		for k, v := range gores.Header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}
		w.WriteHeader(gores.StatusCode)
		body, err := ioutil.ReadAll(gores.Body)
		if err != nil {
			if gores.StatusCode != 301 || gores.StatusCode != 302 {
				log.Println("read body error ", err)
			}
			return
		}
		defer gores.Body.Close()
		nw, err := w.Write(body)
		if err != nil {
			log.Println("write body error ", err)
		}
		log.Printf("write %d bytes to client", nw)
	}
}
