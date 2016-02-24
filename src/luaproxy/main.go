/*
* @Author: detailyang
* @Date:   2016-02-11 17:34:23
* @Last Modified by:   detailyang
* @Last Modified time: 2016-02-25 00:07:07
 */

package main

import (
	"flag"
	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
	"httpproxy"
	"httpserver"
	"io/ioutil"
	"log"
	"path/filepath"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var config string

	flag.StringVar(&config, "config", "", "config file")
	flag.Parse()
	if config == "" {
		log.Fatalln("config file cannot be null")
	}
	v := viper.New()
	v.SetConfigFile(config)
	v.SetConfigType("json")
	err := v.ReadInConfig()
	if err != nil {
		log.Fatalln("Fatal error config file ", err)
	}
	raddress := v.GetString("redis.address")
	rmaxidle := v.GetInt("redis.maxidle")
	rmaxactive := v.GetInt("redis.maxactive")
	redispool := &redis.Pool{
		MaxIdle:   rmaxidle,
		MaxActive: rmaxactive, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", raddress)
			if err != nil {
				log.Println(err)
			}
			return c, err
		},
	}

	requestdir := v.GetString("plugin.requestdir")
	upstreamdir := v.GetString("plugin.upstreamdir")
	responsedir := v.GetString("plugin.responsedir")
	luaplguin := loadLuaCodestoMem(requestdir, upstreamdir, responsedir)

	httpsaddress := v.GetString("proxy.https.address")
	httpaddress := v.GetString("proxy.http.address")
	hp := httpproxy.NewHttpProxy(redispool, httpaddress, httpsaddress, luaplguin)

	wg.Add(1)
	go func() {
		hp.ListenAndServe()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		httpscert := v.GetString("proxy.https.cert")
		httpskey := v.GetString("proxy.https.key")
		hp.ListenAndServeTLS(httpscert, httpskey)
		wg.Done()
	}()

	hsaddress := v.GetString("server.address")
	hsstaticdir := v.GetString("server.staticdir")
	hsindexhtml := v.GetString("server.indexhtml")
	hs := httpserver.NewHttpServer(hsaddress, hsstaticdir, hsindexhtml, luaplguin)

	wg.Add(1)
	go func() {
		hs.ListenAndServe()
		wg.Done()
	}()
	wg.Wait()
}

func loadLuaCodestoMem(requestdir, upstreamdir, responsedir string) map[string]string {
	luaplugin := make(map[string]string)
	readfiles := func(dir string) {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Println("read dir error", err)
		}
		for _, file := range files {
			luacode, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				log.Println("read file error ", err)
				continue
			}
			luaplugin[file.Name()] = string(luacode)
		}
	}
	readfiles(requestdir)
	readfiles(upstreamdir)
	readfiles(responsedir)

	return luaplugin
}
