package web

import (
	"log"
	"net"
	"net/http"
	"os"
)

type Server struct {
	NotFound         http.Handler
	MethodNotAllowed http.Handler
	PanicHandler     func(http.ResponseWriter, *http.Request, interface{})

	encKey                 []byte
	signKey                []byte
	trees                  map[string]*node
	logger                 *log.Logger
	redirectTrailingSlash  bool
	redirectFixedPath      bool
	handleMethodNotAllowed bool
	handleOPTIONS          bool
}

func NewServer() *Server {
	return &Server{
		logger:                 log.New(os.Stdout, "", log.Ldate|log.Ltime),
		redirectTrailingSlash:  true,
		redirectFixedPath:      true,
		handleMethodNotAllowed: true,
		handleOPTIONS:          true,
	}
}

func (s *Server) SetLogger(logger *log.Logger) {
	s.logger = logger
}

func (s *Server) SetRedirectTrailingSlash(redirectTrailingSlash bool) {
	s.redirectTrailingSlash = redirectTrailingSlash
}

func (s *Server) SetRedirectFixedPath(redirectFixedPath bool) {
	s.redirectFixedPath = redirectFixedPath
}

func (s *Server) SetHandleMethodNotAllowed(handleMethodNotAllowed bool) {
	s.handleMethodNotAllowed = handleMethodNotAllowed
}

func (s *Server) SetHandleOPTIONS(handleOPTIONS bool) {
	s.handleOPTIONS = handleOPTIONS
}

func (s *Server) recv(w http.ResponseWriter, r *http.Request) {
	if rcv := recover(); rcv != nil {
		s.PanicHandler(w, r, rcv)
	}
}

func (s *Server) Lookup(method, path string) (Handle, Params, bool) {
	if root := s.trees[method]; root != nil {
		return root.getValue(path)
	}
	return nil, nil, false
}

func (s *Server) allowed(path, reqMethod string) (allow string) {
	if path == "*" {
		for method := range s.trees {
			if method == "OPTIONS" {
				continue
			}

			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
	} else { // specific path
		for method := range s.trees {
			if method == reqMethod || method == "OPTIONS" {
				continue
			}

			handle, _, _ := s.trees[method].getValue(path)
			if handle != nil {
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
			code := 301
			if r.Method != "GET" {
				code = 307
			}

			if tsr && s.redirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					r.URL.Path = path[:len(path)-1]
				} else {
					r.URL.Path = path + "/"
				}
				http.Redirect(w, r, r.URL.String(), code)
				return
			}

			if s.redirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					s.redirectTrailingSlash,
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
		if s.handleOPTIONS {
			if allow := s.allowed(path, r.Method); len(allow) > 0 {
				w.Header().Set("Allow", allow)
				return
			}
		}
	} else {
		if s.handleMethodNotAllowed {
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

	if s.NotFound != nil {
		s.NotFound.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}

func (s *Server) GET(path string, handle Handle) {
	s.Handle("GET", path, handle)
}

func (s *Server) HEAD(path string, handle Handle) {
	s.Handle("HEAD", path, handle)
}

func (s *Server) OPTIONS(path string, handle Handle) {
	s.Handle("OPTIONS", path, handle)
}

func (s *Server) POST(path string, handle Handle) {
	s.Handle("POST", path, handle)
}

func (s *Server) PUT(path string, handle Handle) {
	s.Handle("PUT", path, handle)
}

func (s *Server) PATCH(path string, handle Handle) {
	s.Handle("PATCH", path, handle)
}

func (s *Server) DELETE(path string, handle Handle) {
	s.Handle("DELETE", path, handle)
}

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

func (s *Server) Run(addr string) {

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}

	s.logger.Printf("web.go serving %s\n", l.Addr())

	s.logger.Fatal(http.Serve(l, s))
}
