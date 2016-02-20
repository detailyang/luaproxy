/*
* @Author: detailyang
* @Date:   2016-02-11 17:34:23
* @Last Modified by:   detailyang
* @Last Modified time: 2016-02-20 18:17:19
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

	luarequest := make(map[string]string)
	requestdir := v.GetString("plugin.requestdir")
	files, err := ioutil.ReadDir(requestdir)
	if err != nil {
		log.Println("read dir error", err)
	}
	for _, file := range files {
		luacode, err := ioutil.ReadFile(filepath.Join(requestdir, file.Name()))
		if err != nil {
			log.Println("read file error ", err)
			continue
		}
		luarequest[file.Name()] = string(luacode)
	}

	luaupstream := make(map[string]string)
	upstreamdir := v.GetString("plugin.upstreamdir")
	files, err = ioutil.ReadDir(upstreamdir)
	if err != nil {
		log.Println("read dir error", err)
	}

	for _, file := range files {
		luacode, err := ioutil.ReadFile(filepath.Join(upstreamdir, file.Name()))
		if err != nil {
			log.Println("read file error ", err)
			continue
		}
		luaupstream[file.Name()] = string(luacode)
	}

	luaresponse := make(map[string]string)
	responsedir := v.GetString("plugin.responsedir")
	files, err = ioutil.ReadDir(responsedir)
	if err != nil {
		log.Println("read dir error", err)
	}
	for _, file := range files {
		luacode, err := ioutil.ReadFile(filepath.Join(responsedir, file.Name()))
		if err != nil {
			log.Println("read file error ", err)
			continue
		}
		luaresponse[file.Name()] = string(luacode)
	}

	hp := httpproxy.NewHttpProxy(redispool, luarequest, luaupstream, luaresponse)

	wg.Add(1)
	go func() {
		httpaddress := v.GetString("proxy.http.address")
		hp.ListenAndServe(httpaddress)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		httpsaddress := v.GetString("proxy.https.address")
		httpscert := v.GetString("proxy.https.cert")
		httpskey := v.GetString("proxy.https.key")
		hp.ListenAndServeTLS(httpsaddress, httpscert, httpskey)
		wg.Done()
	}()

	hsaddress := v.GetString("server.address")
	hs := httpserver.NewHttpServer(hsaddress)

	wg.Add(1)
	go func() {
		hs.ListenAndServe()
		wg.Done()
	}()
	wg.Wait()
}
