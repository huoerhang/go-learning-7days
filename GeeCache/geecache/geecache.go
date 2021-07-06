package geecache

import (
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	Get(key string)([]byte,error)
}

type GetterFunc func(key string)([]byte,error)

func (f GetterFunc) Get(key string)([]byte,error)  {
	return  f(key)
}

type  Group struct {
	name string
	mainCache cache
	getter Getter
}

var (
	mu sync.RWMutex
	groups=make(map[string]*Group)
)

func NewGroup(name string,cacheBytesSize uint64,getter Getter) *Group {
	if getter==nil{
		panic("nil Getter")
	}

	mu.Lock()
	defer  mu.Unlock()
	g:=&Group{
		name: name,
		getter: getter,
		mainCache: cache{cacheBytesSize: cacheBytesSize},
	}
	groups[name]=g
	return g
}

func GetGroup(name string)*Group  {
	mu.Lock()
	defer mu.RUnlock()
	g:=groups[name]
	return  g
}

func (g *Group) Get(key string) (ByteView,error) {
	if key==""{
		return  ByteView{},fmt.Errorf("key is required")
	}

	if v,ok:=g.mainCache.get(key);ok{
		log.Println("[GeeCache] hit")
		return v,nil
	}

	return  g.load(key)
}

func (g *Group) load(key string) (ByteView,error) {
	return g.getLocallly(key)
}

func (g *Group) getLocallly(key string)(ByteView,error)  {
	bytes,err:=g.getter.Get(key)
	if err!=nil{
		return  ByteView{},err
	}

	value:=ByteView{b: closeBytes(bytes)}
	g.populateCache(key,value)
	return  value,nil
}

func (g *Group) populateCache(key string,value ByteView)  {
	g.mainCache.add(key,value)
}