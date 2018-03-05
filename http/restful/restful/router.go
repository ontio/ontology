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
