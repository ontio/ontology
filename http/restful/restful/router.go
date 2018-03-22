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

//https://github.com/emostafa/garson
type Params map[string]string

type Route struct {
	Method           string
	Path             *regexp.Regexp
	RegisteredParams []string
	Handler          http.HandlerFunc
}
type Router struct {
	Routes []*Route
}

var paramsRegexp = regexp.MustCompile(`:(\w+)`)

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) Try(path string, method string) (http.HandlerFunc, Params, error) {

	for _, route := range r.Routes {
		if route.Method == method {
			match := route.Path.MatchString(path)
			if match == false {
				continue
			}
			params := Params{}
			// check if this route has registered params, and then parse them
			if len(route.RegisteredParams) > 0 {
				params = parseParams(route, path)
			}
			return route.Handler, params, nil

		}
	}
	return nil, Params{}, errors.New("Route not found")

}

func (r *Router) add(method string, path string, handler http.HandlerFunc) {

	route := &Route{}
	route.Method = method
	path = "^" + path + "$"
	route.Handler = handler

	if strings.Contains(path, ":") {
		matches := paramsRegexp.FindAllStringSubmatch(path, -1)
		if matches != nil {
			for _, v := range matches {
				route.RegisteredParams = append(route.RegisteredParams, v[1])
				// remove the :params from the url path and replace them with regex
				path = strings.Replace(path, v[0], `(\w+)`, 1)
			}
		}
	}
	compiledPath, err := regexp.Compile(path)
	if err != nil {
		panic(err)
	}
	route.Path = compiledPath
	r.Routes = append(r.Routes, route)
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

func (r *Router) Head(path string, handler http.HandlerFunc) {
	r.add("HEAD", path, handler)
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
	ctx := context.WithValue(req.Context(), "route_params", params)
	handler(w, req.WithContext(ctx))

}

func parseParams(route *Route, path string) Params {

	matches := route.Path.FindAllStringSubmatch(path, -1)
	params := Params{}
	matchedParams := matches[0][1:]

	for k, v := range matchedParams {
		params[route.RegisteredParams[k]] = v
	}
	return params
}

func getParam(r *http.Request, key string) string {
	ctx := r.Context()
	params := ctx.Value("route_params").(Params)
	val, _ := params[key]
	return val
}
