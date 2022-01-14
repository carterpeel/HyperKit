package airplayserver

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/pprof"
)

func StartProfiler() {
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/index", pprof.Index)
		log.Debugln(http.ListenAndServe(":6060", mux))
	}()
}
