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

func (s *Server) Get(path string, handler Handler) {
	s.addRoute("GET", path, handler)
}

func (s *Server) Head(path string, handler Handler) {
	s.addRoute("HEAD", path, handler)
}

func (s *Server) Options(path string, handler Handler) {
	s.addRoute("OPTIONS", path, handler)
}

func (s *Server) Post(path string, handler Handler) {
	s.addRoute("POST", path, handler)
}

func (s *Server) Put(path string, handler Handler) {
	s.addRoute("PUT", path, handler)
}

func (s *Server) Patch(path string, handler Handler) {
	s.addRoute("PATCH", path, handler)
}

func (s *Server) Delete(path string, handler Handler) {
	s.addRoute("DELETE", path, handler)
}

func (s *Server) Resource(path string, controller Controller) {
	s.Get(path, controller.Index)
	s.Post(path, controller.Create)
	s.Put(path, controller.Update)
	s.Delete(path, controller.Delete)
}

func (s *Server) addRoute(method, path string, handler Handler) {
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
	root.addRoute(path, handler)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.PanicHandler != nil {
		defer s.recv(w, r)
	}

	path := r.URL.Path

	if root := s.trees[r.Method]; root != nil {
		if handler, ps, tsr := root.getValue(path); handler != nil {
			handler(&Context{
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
					cleanPath(path),
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

func (s *Server) Run(addr string) {

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}

	s.logger.Printf("web.go serving %s\n", l.Addr())

	s.logger.Fatal(http.Serve(l, s))
}

func (s *Server) lookup(method, path string) (Handler, Params, bool) {
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

			handler, _, _ := s.trees[method].getValue(path)
			if handler != nil {
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
