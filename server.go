package main

import (
	"bytes"
	"code.google.com/p/gorilla/gorilla/mux"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	*mux.Router
	datastore Datastore
	templates *template.Template
	Debug     bool
}

func NewServer(store Datastore) *Server {
	server := &Server{
		Router:    new(mux.Router),
		datastore: store,
		templates: template.New(""),
		Debug:     true,
	}
	server.templates.Funcs(template.FuncMap{
		"route":      server.routeFunc(),
		"eventRoute": server.eventRouteFunc(),
		"matchRoute": server.matchRouteFunc(),
		"cycle": func(i int, vals ...interface{}) interface{} {
			return vals[i%len(vals)]
		},
	})
	return server
}

func (server *Server) Templates() *template.Template {
	return server.templates
}

func (server *Server) Store() Datastore {
	return server.datastore
}

type ServerHandlerFunc func(*Server, http.ResponseWriter, *http.Request) error

func (server *Server) Handler(f ServerHandlerFunc) http.Handler {
	return serverHandler{server, f}
}

func (server *Server) logError(req *http.Request, err error) {
	var b bytes.Buffer
	fmt.Fprintf(&b, "ERROR %s %s\n", req.Method, req.URL.Path)
	fmt.Fprintf(&b, "\tMessage:\n\t\t%v\n", err)

	fmt.Fprint(&b, "\tHeaders:\n")
	for k, vv := range req.Header {
		for _, v := range vv {
			fmt.Fprintf(&b, "\t\t%s: %v\n", k, v)
		}
	}

	if server.Debug {
		req.ParseForm()
	}
	if req.Form != nil {
		fmt.Fprint(&b, "\tForm:\n")
		for k, vv := range req.Form {
			for _, v := range vv {
				fmt.Fprintf(&b, "\t\t%s: %v\n", k, v)
			}
		}
	}
	log.Print(&b)
}

func (server *Server) routeFunc() func(string, ...string) (template.URL, error) {
	return func(name string, pairs ...string) (template.URL, error) {
		route, ok := server.NamedRoutes[name]
		if !ok {
			return "", fmt.Errorf("Could not resolve route %q", name)
		}
		url := route.URL(pairs...)
		if url == nil {
			return "", fmt.Errorf("Bad set of pairs for route %q: %v", name, pairs)
		}
		return template.URL(url.String()), nil
	}
}

func (server *Server) eventRouteFunc() func(string, EventTag, ...string) (template.URL, error) {
	return func(name string, tag EventTag, pairs ...string) (template.URL, error) {
		route, ok := server.NamedRoutes[name]
		if !ok {
			return "", fmt.Errorf("Could not resolve route %q", name)
		}
		args := make([]string, 0, len(pairs)+4)
		args = append(args, "year", fmt.Sprint(tag.Year))
		args = append(args, "location", tag.LocationCode)
		args = append(args, pairs...)
		url := route.URL(args...)
		if url == nil {
			return "", fmt.Errorf("Bad set of pairs for event route %q: %v", name, pairs)
		}
		return template.URL(url.String()), nil
	}
}

func (server *Server) matchRouteFunc() func(string, EventTag, MatchType, int, ...string) (template.URL, error) {
	return func(name string, tag EventTag, matchType MatchType, matchNum int, pairs ...string) (template.URL, error) {
		route, ok := server.NamedRoutes[name]
		if !ok {
			return "", fmt.Errorf("Could not resolve route %q", name)
		}
		args := make([]string, 0, len(pairs)+8)
		args = append(args, "year", fmt.Sprint(tag.Year))
		args = append(args, "location", tag.LocationCode)
		args = append(args, "matchType", string(matchType))
		args = append(args, "matchNumber", fmt.Sprint(matchNum))
		args = append(args, pairs...)
		url := route.URL(args...)
		if url == nil {
			return "", fmt.Errorf("Bad set of pairs for match route %q: %v", name, pairs)
		}
		return template.URL(url.String()), nil
	}
}

// A serverHandler wraps a ServerHandlerFunc to implement the http.Handler interface.
type serverHandler struct {
	server *Server
	handle ServerHandlerFunc
}

func (handler serverHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	buf := new(ResponseBuffer)
	err := handler.handle(handler.server, buf, req)

	if err == nil {
		w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
		buf.Flush(w)
	} else {
		handler.server.logError(req, err)

		if handler.server.Debug {
			req.ParseForm()
			w.WriteHeader(http.StatusInternalServerError)
			handler.server.Templates().ExecuteTemplate(w, "error-debug.html", map[string]interface{}{
				"Server":    handler.server,
				"Error":     err,
				"Request":   req,
				"Variables": mux.Vars(req),
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			handler.server.Templates().ExecuteTemplate(w, "error.html", map[string]interface{}{
				"Server":  handler.server,
				"Error":   err,
				"Request": req,
			})
		}
	}
}

// A ResponseBuffer stores an entire request in memory.  The zero value is an empty response.
type ResponseBuffer struct {
	bytes.Buffer
	code   int
	header http.Header
	sent   http.Header
}

func (buffer *ResponseBuffer) StatusCode() int {
	return buffer.code
}

func (buffer *ResponseBuffer) HeaderSent() http.Header {
	return buffer.sent
}

func (buffer *ResponseBuffer) Flush(w http.ResponseWriter) error {
	for k, v := range buffer.sent {
		w.Header()[k] = v
	}
	w.WriteHeader(buffer.code)
	_, err := io.Copy(w, buffer)
	return err
}

func (buffer *ResponseBuffer) Header() http.Header {
	if buffer.header == nil {
		buffer.header = make(http.Header)
	}
	return buffer.header
}

func (buffer *ResponseBuffer) WriteHeader(code int) {
	if buffer.sent == nil {
		buffer.code = code
		buffer.sent = make(http.Header, len(buffer.header))
		for k, v := range buffer.header {
			v2 := make([]string, len(v))
			copy(v2, v)
			buffer.sent[k] = v2
		}
	}
}

func (buffer *ResponseBuffer) Write(p []byte) (int, error) {
	buffer.WriteHeader(http.StatusOK)
	return buffer.Buffer.Write(p)
}

// A Logger logs all HTTP requests sent to a handler.
type Logger struct {
	http.Handler
}

func (logger Logger) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	method, path := req.Method, req.URL.Path
	rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
	logger.Handler.ServeHTTP(rec, req)
	log.Printf("%s %s %d %d", method, path, rec.statusCode, rec.size)
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	size       int64
}

func (rec *responseRecorder) WriteHeader(statusCode int) {
	rec.ResponseWriter.WriteHeader(statusCode)
	rec.statusCode = statusCode
}

func (rec *responseRecorder) Write(p []byte) (n int, err error) {
	n, err = rec.ResponseWriter.Write(p)
	rec.size += int64(n)
	return
}
