package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash Hash //哈希函数
	replicas int //虚拟节点倍数
	keys []int //哈希环
	hashMap map[int]string //虚拟节点与真实节点映射表，键是虚拟节点的哈希值，值是真实节点的名称
}

func New(replicas int,fn Hash)	*Map  {
	m:=&Map{
		replicas: replicas,
		hash: fn,
		hashMap: make(map[int]string),
	}

	if m.hash==nil{
		m.hash=crc32.ChecksumIEEE
	}

	return m
}

// Add 添加节点
// peers 真实节点
func (m *Map) Add(peers ...string)  {
	for _,p:=range peers{
		for i:=0;i<m.replicas;i++ {
			hash:=int(m.hash([]byte(strconv.Itoa(i)+p))) //虚拟节点hash
			m.keys=append(m.keys,hash)
			m.hashMap[hash]=p //映射虚拟节点与真实节点
		}
	}
}

// Get 获取真实节点
func (m *Map) Get(key string) string {
	if len(m.keys)==0{
		return ""
	}

	//计算Key的哈希值
	hash:=int(m.hash([]byte(key)))

	index:=sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i]>=hash
	})

	//如果 idx == len(m.keys)，说明应选择 m.keys[0]，因为 m.keys 是一个环状结构，所以用取余数的方式来处理这种情况。
	mid :=m.keys[index%len(m.keys)]

	//通过 hashMap 映射得到真实的节点
	return m.hashMap[mid]
}
