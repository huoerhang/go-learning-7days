package main

import (
	"flag"
	"fmt"
	"geecache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func  createGroup() *geecache.Group {
	return geecache.NewGroup("scores",2<<10,geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key",key)
			if v,ok:=db[key];ok{
				return []byte(v),nil
			}
			return nil,fmt.Errorf("%s not exists",key)
		},
	))
}

func startCacheServer(addr string, adds []string,group *geecache.Group)  {
	peers:=geecache.NewHTTPPool(addr)
	peers.Set(adds...)
	group.RegisterPeers(peers)
	log.Println("geecache is running at",addr)
	log.Fatal(http.ListenAndServe(addr[7:],peers))
}

func startAPIServer(apiAddr string,group *geecache.Group)  {
	http.Handler("/api",http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key:=r.URL.Query().Get("key")
			view,err:=group.Get(key)
			if err!=nil{
				http.Error(w,err.Error(),http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type","application/octet-stream")
			w.Write(view.ByteSlice())
		},
	))
	log.Println("fonted server is running at",apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:],nil))
}

func main()  {
	var port int
	var api bool
	flag.IntVar(&port,"port",8001,"Geecache server port")
	flag.BoolVar(&api,"api",false,"Start a api server?")
	flag.Parse()

	apiAddr:="http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var adds []string
	for _,v:=range addrMap{
		adds=append(adds,v)
	}

	group:=createGroup()
	if api{
		go startAPIServer(apiAddr,group)
	}
	startCacheServer(addrMap[port],[]string(adds),group)
}