package pprof

import (
	"net/http"
	"net/http/pprof"
)

// InitPprof start http pprof.
func InitPprof() {
	pprofServeMux := http.NewServeMux()
	pprofServeMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofServeMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofServeMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofServeMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	go func() {
		if err := http.ListenAndServe("0.0.0.0:6971", pprofServeMux); err != nil {
			// log.Error("http.ListenAndServe(\"%s\", pproServeMux) error(%v)", addr)
			panic(err)
		}
	}()
}
