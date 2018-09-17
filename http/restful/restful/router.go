/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package restful

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
)

type paramsMap map[string]string

//http router
type Route struct {
	Method  string
	Path    *regexp.Regexp
	Params  []string
	Handler http.HandlerFunc
}
type Router struct {
	routes []*Route
}

func NewRouter() *Router {
	return &Router{}
}

func (this *Router) Try(path string, method string) (http.HandlerFunc, paramsMap, error) {

	for _, route := range this.routes {
		if route.Method == method {
			match := route.Path.MatchString(path)
			if match == false {
				continue
			}
			params := paramsMap{}
			if len(route.Params) > 0 {
				params = parseParams(route, path)
			}
			return route.Handler, params, nil

		}
	}
	return nil, paramsMap{}, errors.New("Route not found")

}

func (this *Router) add(method string, path string, handler http.HandlerFunc) {
	route := &Route{}
	route.Method = method
	path = "^" + path + "$"
	route.Handler = handler

	if strings.Contains(path, ":") {
		matches := regexp.MustCompile(`:(\w+)`).FindAllStringSubmatch(path, -1)
		if matches != nil {
			for _, v := range matches {
				route.Params = append(route.Params, v[1])
				path = strings.Replace(path, v[0], `(\w+)`, 1)
			}
		}
	}
	compiledPath, err := regexp.Compile(path)
	if err != nil {
		panic(err)
	}
	route.Path = compiledPath
	this.routes = append(this.routes, route)
}

func (r *Router) Head(path string, handler http.HandlerFunc) {
	r.add("HEAD", path, handler)
}

func (r *Router) Connect(path string, handler http.HandlerFunc) {
	r.add("CONNECT", path, handler)
}

func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.add("GET", path, handler)
}

func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.add("POST", path, handler)
}

func (r *Router) Put(path string, handler http.HandlerFunc) {
	r.add("PUT", path, handler)
}

func (r *Router) Delete(path string, handler http.HandlerFunc) {
	r.add("DELETE", path, handler)
}

func (r *Router) Options(path string, handler http.HandlerFunc) {
	r.add("OPTIONS", path, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler, params, err := r.Try(req.URL.Path, req.Method)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	ctx := context.WithValue(req.Context(), "params", params)
	handler(w, req.WithContext(ctx))
}

func parseParams(route *Route, path string) paramsMap {
	matches := route.Path.FindAllStringSubmatch(path, -1)
	params := paramsMap{}
	matchedParams := matches[0][1:]

	for k, v := range matchedParams {
		params[route.Params[k]] = v
	}
	return params
}

func getParam(r *http.Request, key string) string {
	ctx := r.Context()
	params := ctx.Value("params").(paramsMap)
	val, _ := params[key]
	return val
}
