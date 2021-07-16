package geecache

import (
	"fmt"
	"geecache/singleflight"
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
	peers PeerPicker
	loader *singleflight.Group
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
		loader: &singleflight.Group{},
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

func (g *Group) load(key string) (value ByteView,err error) {
	item,err:=g.loader.Do(key, func() (interface{}, error) {
		if g.peers!=nil{
			if peer,ok:=g.peers.PickPeer(key);ok{
				if value,err=g.getFromPeer(peer,key);err==nil{
					return  value,nil
				}
				log.Println("[GeeCache] Failed to get from peer",err)
			}
		}
		return g.getLocallly(key),nil
	})

	if err==nil{
		return item.(ByteView),nil
	}

	return
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

func (g *Group) RegisterPeers(peers PeerPicker)  {
	if g.peers!=nil{
		panic("RegisterPeerPick called more than once")
	}
	g.peers=peers
}

func (g *Group) getFromPeer(peer PeerGetter,key string) (ByteView,error) {
	bytes,err:=peer.Get(g.name,key)
	if err!=nil{
		return ByteView{},err
	}
	return ByteView{b:bytes},nil
}

func (g *Group) load(key string)(value ByteView,err error)  {
	if g.peers!=nil{
		if peer,ok:=g.peers.PickPeer(key);ok{
			if value,err=g.getFromPeer(peer,key);err==nil{
				return  value,nil
			}
			log.Println("[GeeCache] Failed to get from peer",err)
		}
	}
	return  g.getLocallly(key)
}