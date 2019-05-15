package pprof

// StartHTTPPprofServer starts http server for pprof.
func StartHTTPPprofServer(addr string, addrs ...string) {
	go httpServePprof(addr)
	for _, a := range addrs {
		go httpServePprof(a)
	}
}
