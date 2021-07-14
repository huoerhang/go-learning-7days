package geecache

import (
	"fmt"
	"geecache/consistenthash"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath="/_geecache"
	defaultRelicas=50
)

type HTTPPool struct {
	self string
	basePath string
	mu sync.Mutex
	peers *consistenthash.Map
	httpGetter map[string]*httpGetter
}

type httpGetter struct {
	baseURL string
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self: self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string,v ...interface{})  {
	log.Printf("[Server %s] %s",p.self,fmt.Sprintf(format,v...))
}

func (p *HTTPPool) Set(peers ...string)  {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers=consistenthash.New(defaultRelicas,nil)
	p.peers.Add(peers...)
	p.httpGetter=make(map[string]*httpGetter,len(peers))
	for _,peer:=range peers{
		p.httpGetter[peer]=&httpGetter{baseURL: peer+p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string)(PeerGetter,bool)  {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer:=p.peers.Get(key);peer!="" && peer!=p.self{
		p.Log("Pick peer %s",peer)
		return p.httpGetter[peer],true
	}

	return nil,false
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter,r *http.Request)  {
	if !strings.HasPrefix(r.URL.Path,p.basePath){
		panic("HTTPPool serving unexpected  path:"+r.URL.Path)
	}

	p.Log("%s %s",r.Method,r.URL.Path)
	parts:=strings.SplitN(r.URL.Path[len(p.basePath):],"/",2)
	if len(parts)!=2{
		http.Error(w,"Bad Request",http.StatusBadRequest)
		return
	}
	groupName:=parts[0]
	key:=parts[1]
	group:=GetGroup(groupName)

	if group==nil{
		http.Error(w,"No Such Group:"+groupName,http.StatusNotFound)
		return
	}

	view,error:=group.Get(key)
	if error!=nil{
		http.Error(w,error.Error(),http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type","application/octet-stream")
	w.Write(view.ByteSlice())
}

func (h *httpGetter) Get(group string,key string)([]byte,error)  {
	u:=fmt.Sprintf("%v%v/%v",h.baseURL,url.QueryEscape(group),url.QueryEscape(key))
	res,err:=http.Get(u)
	if err!=nil{
		return  nil,err
	}
	defer res.Body.Close()
	if res.StatusCode!=http.StatusOK{
		return  nil,fmt.Errorf("server returned:%v",res.Status)
	}
	bytes,err:=ioutil.ReadAll(res.Body)
	if err!=nil{
		return  nil,fmt.Errorf("reading response body:%v",err)
	}

	return  bytes,nil
}

var _PeerGetter=(*httpGetter)(nil)