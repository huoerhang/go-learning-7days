package geecache

import (
	"lru"
	"sync"
)

type  cache struct {
	mu sync.Mutex
	lru *lru.Cache
	cacheBytesSize uint64
}

func (c *cache) add(key string,value ByteView)  {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru==nil{
		c.lru=lru.New(c.cacheBytesSize,nil)
	}
	c.lru.Add(key,value)
}

func (c *cache) get(key string) (value ByteView,ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru==nil{
		return
	}

	if v,ok:=c.lru.Get(key);ok{
		return v.(ByteView),ok
	}

	return
}