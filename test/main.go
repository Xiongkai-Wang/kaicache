package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"xiongkai.com/kaicache"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *kaicache.Group {
	return kaicache.NewGroup("scores", 2<<10, kaicache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key ", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exists", key)
		},
	))
}

func startCacheServer(addr string, addrs []string, kaiGroup *kaicache.Group) {
	peers := kaicache.NewHTTPPool(addr)
	peers.Set(addrs...)
	kaiGroup.RegisterPeers(peers)
	log.Println("kaicache is running at ", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, kaiGroup *kaicache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := kaiGroup.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		},
	))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {
	// kaicache.NewGroup("scores", 2<<10, kaicache.GetterFunc(
	// 	func(key string) ([]byte, error) {
	// 		log.Println("[SlowDB] search key ", key)
	// 		if v, ok := db[key]; ok {
	// 			return []byte(v), nil
	// 		}
	// 		return nil, fmt.Errorf("%s not exists", key)
	// 	},
	// ))
	// addr := "localhost:9999"
	// peers := kaicache.NewHTTPPool(addr)
	// log.Println("kaicache is running at", addr)
	// log.Fatal(http.ListenAndServe(addr, peers))
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "kaicache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://127.0.0.1:9999"
	addrMap := map[int]string{
		8001: "http://127.0.0.1:8001",
		8002: "http://127.0.0.1:8002",
		8003: "http://127.0.0.1:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	kaiGroup := createGroup()
	if api {
		go startAPIServer(apiAddr, kaiGroup)
	}
	startCacheServer(addrMap[port], addrs, kaiGroup)
}
