package gee

import (
	"log"
	"net/http"
)

type router struct {
	handlers map[string]HandlerFunc
}

func newRouter() *router  {
	return  &router{
		handlers: make(map[string]HandlerFunc),
	}
}

func (r *router) getKey(method string,pattern string) string {
	return method+"-"+pattern
}

func (r *router) addRoute(method string,pattern string,handler HandlerFunc)  {
	log.Printf("Route %4s - %s",method,pattern)
	key:=r.getKey(method,pattern)
	r.handlers[key]=handler
}

func (r *router) handler(c *Context)  {
	key:=r.getKey(c.Method, c.Path)
	if handler,ok:=r.handlers[key];ok{
		handler(c)
		return
	}

	c.String(http.StatusNotFound,"404 NOT FOUND %s\n",c.Path)
}
