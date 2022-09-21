package kaicache

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"xiongkai.com/kaicache/consistenthash"
)

// 提供被其他节点访问的能力(基于http)

const (
	defaultBasePath = "/_kaicache/"
	defalutReplicas = 50
)

var _ PeerPicker = (*HTTPPool)(nil)

// HTTPPool implements PeerPicker for a pool of HTTP peers: multiple cache nodes
type HTTPPool struct {
	self     string // "https://example.net:8000"
	basePath string // 节点间通讯地址的前缀，默认是 /_kaicache/

	mu          sync.Mutex // guards peers and httpGetters
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter // keyed by e.g. "http://10.0.0.2:8008"
}

type httpGetter struct { // every httpGetter maps to a remote node
	baseURL string // http://example.com/_kaicache/
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// implement ServeHTTP, become a handler(as a server)
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

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())

}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s received:] %s", p.self, fmt.Sprintf(format, v...))
}

// register peers and its httpGetter
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defalutReplicas, nil) //?
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// call consistenthash
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

// get value from remote node
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u) // return *http.Response
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}
