package web

import "runtime"

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
