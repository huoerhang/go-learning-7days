package gee

import (
	"gee"
	"strings"
)

type router struct {
	roots map[string]*node
	handlers map[string]gee.HandlerFunc
}

func newRouter()*router  {
	return &router{
		roots: make(map[string]*node),
		handlers: make(map[string]gee.HandlerFunc),
	}
}

func parsePattern(pattern string) []string {
	vs:=strings.Split(pattern,"/")

	parts:=make([]string,0)
	for _,item:=range vs{
		if item!=""{
			parts=append(parts,item)
			if item[0]=='*'{
				break
			}
		}
	}

	return  parts
}

func (r *router) addRoute(methond string,pattern string,handler HandlerFunc)  {
	parts:=parsePattern(pattern)

	key:=methond+"-"+pattern
	_,ok:=r.roots[methond]
	if !ok{
		r.roots[methond]=&node{}
	}
	r.roots[methond].insert(pattern,parts,0)
	r.handlers[key]=handler
}
