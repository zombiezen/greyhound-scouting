package main

import (
	"bytes"
	"code.google.com/p/gorilla/mux"
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
		"route": server.routeFunc(),
		"percent": func(x float64) string {
			return fmt.Sprintf("%.1f%%", x*100)
		},
		"cycle": func(i int, vals ...interface{}) interface{} {
			return vals[i%len(vals)]
		},
		"map": func(vals ...interface{}) (map[string]interface{}, error) {
			if len(vals)%2 != 0 {
				return nil, fmt.Errorf("map must be given an even number of arguments, %d given", len(vals))
			}

			m := make(map[string]interface{}, len(vals)/2)
			for i := 0; i < len(vals); i += 2 {
				if k, ok := vals[i].(string); ok {
					m[k] = vals[i+1]
				} else {
					return nil, fmt.Errorf("argument %d must be string key, got %#v", i, vals[i])
				}
			}
			return m, nil
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

func (server *Server) routeFunc() func(string, ...interface{}) (template.URL, error) {
	return func(name string, pairs ...interface{}) (template.URL, error) {
		route := server.GetRoute(name)
		if route == nil {
			return "", fmt.Errorf("Could not resolve route %q", name)
		}
		stringPairs := make([]string, len(pairs))
		for i := range pairs {
			stringPairs[i] = fmt.Sprint(pairs[i])
		}
		url, err := route.URL(stringPairs...)
		if err != nil {
			return "", fmt.Errorf("Bad set of pairs for route %q: %v (%v)", name, pairs, err)
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
