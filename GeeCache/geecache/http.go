package geecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath="/_geecache"

type HTTPPool struct {
	self string
	basePath string
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