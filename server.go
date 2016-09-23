package web

import (
	"bytes"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

var server *Server
var once sync.Once

type Server struct {
	NotFound         http.Handler
	MethodNotAllowed http.Handler
	PanicHandler     func(http.ResponseWriter, *http.Request, interface{})

	viewDir                string
	cookieSecret           string
	encKey                 []byte
	signKey                []byte
	trees                  map[string]*node
	logger                 *log.Logger
	redirectTrailingSlash  bool
	redirectFixedPath      bool
	handleMethodNotAllowed bool
	handleOptions          bool
}

func CreateServer() *Server {
	once.Do(func() {
		server = createServer()
	})
	return server
}

func createServer() *Server {
	viewDir := env("AFXCN_WEB_VIEW_DIR")

	if len(viewDir) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			viewDir = path.Join(path.Dir(os.Args[0]), "views")
		} else {
			viewDir = path.Join(wd, "views")
		}
	}

	cookieSecret := envOrRandom("AFXCN_WEB_COOKIE_SECRET", 64)

	return &Server{
		viewDir:                viewDir,
		cookieSecret:           cookieSecret,
		encKey:                 genKey(cookieSecret, envOrRandom("AFXCN_WEB_COOKIE_ENC_SALT", 16)),
		signKey:                genKey(cookieSecret, envOrRandom("AFXCN_WEB_COOKIE_SIGN_SALT", 16)),
		logger:                 log.New(os.Stdout, "", log.Ldate|log.Ltime),
		redirectTrailingSlash:  true,
		redirectFixedPath:      true,
		handleMethodNotAllowed: true,
		handleOptions:          true,
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

func (s *Server) SetHandleOptions(handleOptions bool) {
	s.handleOptions = handleOptions
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
	startTime := time.Now()

	if s.PanicHandler != nil {
		defer s.recv(w, r)
	}

	path := r.URL.Path

	if root := s.trees[r.Method]; root != nil {
		if handler, params, tsr := root.getValue(path); handler != nil {

			runTime := time.Now()

			ctx := &Context{
				Server:         s,
				ResponseWriter: w,
				Request:        r,
				Params:         &params,
			}

			handler(ctx)

			s.logRequest(ctx, startTime, runTime)
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
		if s.handleOptions {
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

	if s.logger != nil {
		s.logger.Printf("web.go serving %s\n", l.Addr())
	}

	log.Fatal(http.Serve(l, s))
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

func (s *Server) logRequest(ctx *Context, startTime time.Time, runTime time.Time) {

	if s.logger == nil {
		return
	}

	r := ctx.Request
	path := r.URL.Path

	now := time.Now()
	duration := now.Sub(startTime)
	runDuration := now.Sub(runTime)

	var client string

	pos := strings.LastIndex(r.RemoteAddr, ":")
	if pos > 0 {
		client = r.RemoteAddr[0:pos]
	} else {
		client = r.RemoteAddr
	}

	var log bytes.Buffer
	log.WriteString(client)
	log.WriteString(" - " + r.Method + " " + path)
	log.WriteString(" - " + duration.String())
	log.WriteString(" - " + runDuration.String() + "\n")

	s.logger.Print(log.String())
}
