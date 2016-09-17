package web

import (
	"log"
	"net"
	"net/http"
	"os"
)

// Server represents a web.go server.
type Server struct {
	Logger                 *log.Logger
	RedirectTrailingSlash  bool
	RedirectFixedPath      bool
	HandleMethodNotAllowed bool
	HandleOPTIONS          bool
	NotFound               http.Handler
	MethodNotAllowed       http.Handler
	PanicHandler           func(http.ResponseWriter, *http.Request, interface{})

	l       net.Listener
	encKey  []byte
	signKey []byte
	trees   map[string]*node
}

// New create new server
func NewServer() *Server {
	return &Server{
		Logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}
}

func (s *Server) recv(w http.ResponseWriter, r *http.Request) {
	if rcv := recover(); rcv != nil {
		s.PanicHandler(w, r, rcv)
	}
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (s *Server) Lookup(method, path string) (Handle, Params, bool) {
	if root := s.trees[method]; root != nil {
		return root.getValue(path)
	}
	return nil, nil, false
}

func (s *Server) allowed(path, reqMethod string) (allow string) {
	if path == "*" { // server-wide
		for method := range s.trees {
			if method == "OPTIONS" {
				continue
			}

			// add request method to list of allowed methods
			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
	} else { // specific path
		for method := range s.trees {
			// Skip the requested method - we already tried this one
			if method == reqMethod || method == "OPTIONS" {
				continue
			}

			handle, _, _ := s.trees[method].getValue(path)
			if handle != nil {
				// add request method to list of allowed methods
				if len(allow) == 0 {
					allow = method
				} else {
					allow += ", " + method
				}
			}
		}
	}
	if len(allow) > 0 {
		allow += ", OPTIONS"
	}
	return
}

// ServeHTTP is the interface method for Go's http server package
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.PanicHandler != nil {
		defer s.recv(w, r)
	}

	path := r.URL.Path

	if root := s.trees[r.Method]; root != nil {
		if handle, ps, tsr := root.getValue(path); handle != nil {
			handle(&Context{
				ResponseWriter: w,
				Request:        r,
				Params:         &ps,
			})
			return
		} else if r.Method != "CONNECT" && path != "/" {
			code := 301 // Permanent redirect, request with GET method
			if r.Method != "GET" {
				// Temporary redirect, request with same method
				// As of Go 1.3, Go does not support status code 308.
				code = 307
			}

			if tsr && s.RedirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					r.URL.Path = path[:len(path)-1]
				} else {
					r.URL.Path = path + "/"
				}
				http.Redirect(w, r, r.URL.String(), code)
				return
			}

			// Try to fix the request path
			if s.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					s.RedirectTrailingSlash,
				)
				if found {
					r.URL.Path = string(fixedPath)
					http.Redirect(w, r, r.URL.String(), code)
					return
				}
			}
		}
	}

	if r.Method == "OPTIONS" {
		// Handle OPTIONS requests
		if s.HandleOPTIONS {
			if allow := s.allowed(path, r.Method); len(allow) > 0 {
				w.Header().Set("Allow", allow)
				return
			}
		}
	} else {
		// Handle 405
		if s.HandleMethodNotAllowed {
			if allow := s.allowed(path, r.Method); len(allow) > 0 {
				w.Header().Set("Allow", allow)
				if s.MethodNotAllowed != nil {
					s.MethodNotAllowed.ServeHTTP(w, r)
				} else {
					http.Error(w,
						http.StatusText(http.StatusMethodNotAllowed),
						http.StatusMethodNotAllowed,
					)
				}
				return
			}
		}
	}

	// Handle 404
	if s.NotFound != nil {
		s.NotFound.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}

// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (s *Server) Handle(method, path string, handle Handle) {
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if s.trees == nil {
		s.trees = make(map[string]*node)
	}

	root := s.trees[method]
	if root == nil {
		root = new(node)
		s.trees[method] = root
	}

	root.addRoute(path, handle)
}

// Run starts the web application and serves HTTP requests for s
func (s *Server) Run(addr string) {

	mux := http.NewServeMux()

	mux.Handle("/", s)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}

	s.Logger.Printf("web.go serving %s\n", l.Addr())

	s.l = l
	err = http.Serve(s.l, mux)
	s.l.Close()
}
