package pprof

import (
	"log"
	"net/http"
	"net/http/pprof"
)

func httpServePprof(addr string) {
	pprofServer := http.NewServeMux()
	pprofServer.HandleFunc("/debug/pprof/", pprof.Index)
	pprofServer.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofServer.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofServer.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	pprofServer.HandleFunc("/debug/pprof/trace", pprof.Trace)

	if err := http.ListenAndServe(addr, pprofServer); err != nil {
		log.Printf("http.ListenAndServe init %+v error: %+v ", addr, err)
		panic(err)
	}
}
