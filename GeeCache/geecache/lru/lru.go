package lru

import "container/list"

type Value interface {
	Len() int
}

type entry struct {
	key string
	value Value
}

type Cache struct {
	// 允许使用的最大内存
	maxBytes uint64
	// 当前已使用的内存
	usedBytes uint64
	ll *list.List
	cache map[string]*list.Element
	//某条记录被移除时的回调函数，可以为 nil
	OnEvicted func(key string,value Value)
}

func New(maxBytes uint64,onEvicted func(string,Value)) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		ll:list.New(),
		cache: make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get 从缓存中获取一个值
func (c *Cache) Get(key string) (value Value,ok bool) {
	if ele,ok:=c.cache[key];ok{
		c.ll.MoveToFront(ele)
		kv:=ele.Value.(*entry)

		return  kv.value,true
	}

	return
}

// RemoveOldest 缓存淘汰。移除最近最少访问的节点（队首）
func  (c *Cache) RemoveOldest()  {
	ele:=c.ll.Back()
	if ele!=nil{
		c.ll.Remove(ele)
		kv:=ele.Value.(*entry)
		delete(c.cache,kv.key)
		c.usedBytes-=uint64(len(kv.key))+uint64(kv.value.Len())

		if c.OnEvicted!=nil{
			c.OnEvicted(kv.key,kv.value)
		}
	}
}

// Add 向缓存中添加值
func (c *Cache) Add(key string,value Value)  {
	if ele,ok:=c.cache[key];ok{
		c.ll.MoveToFront(ele)
		kv:=ele.Value.(*entry)
		c.usedBytes+=uint64(value.Len())-uint64(kv.value.Len())
		kv.value=value
	}else{
		ele:=c.ll.PushFront(&entry{key,value})
		c.cache[key]=ele
		c.usedBytes+=uint64(len(key))+uint64(value.Len())
	}
	for c.maxBytes!=0&&c.maxBytes<c.usedBytes{
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return  c.ll.Len()
}