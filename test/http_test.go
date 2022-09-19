package main

import (
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	kaicache.NewGroup("scores", 2<<10, kaicache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key ", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exists", key)
		},
	))
	addr := "localhost:9999"
	peers := kaicache.NewHTTPPool(addr)
	log.Println("kaicache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))

}
