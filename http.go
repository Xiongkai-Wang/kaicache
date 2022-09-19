package kaicache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// 提供被其他节点访问的能力(基于http)

const defaultBasePath = "/_kaicache/"

// // HTTPPool implements PeerPicker for a pool of HTTP peers: multiple cache nodes
type HTTPPool struct {
	self     string // "https://example.net:8000"
	basePath string // 节点间通讯地址的前缀，默认是 /_kaicache/
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// implement ServeHTTP, become a handler
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// url.URL   [scheme:][//[userinfo@]host][/]path[?query][#fragment]
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "Application/octet-stream")
	w.Write(view.ByteSlice())

}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Println("[Server %s received] %s", p.self, fmt.Sprintf(format, v...))
}
