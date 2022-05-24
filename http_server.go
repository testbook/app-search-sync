package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"
)

type httpServerCtx struct {
	httpServer  *http.Server
	indexConfig *indexClient
	shutdown    bool
	started     time.Time
}

func (ctx *httpServerCtx) buildServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/started", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		data := time.Since(ctx.started).String()
		w.Write([]byte(data))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	if ctx.indexConfig.config.Stats {
		mux.HandleFunc("/stats", func(w http.ResponseWriter, _ *http.Request) {
			stats, err := json.MarshalIndent(ctx.indexConfig.stats, "", "    ")
			if err == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
				w.Write(stats)
				fmt.Fprintln(w)
			} else {
				w.WriteHeader(500)
				fmt.Fprintf(w, "Unable to print statistics: %s", err)
			}
		})
	}

	if ctx.indexConfig.config.Pprof {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	s := &http.Server{
		Addr:     ctx.indexConfig.config.HTTPServerAddr,
		Handler:  mux,
		ErrorLog: ctx.indexConfig.config.ErrorLogger,
	}
	ctx.httpServer = s
}

func (ctx *httpServerCtx) serveHTTP() {
	s := ctx.httpServer
	if ctx.indexConfig.config.Verbose {
		ctx.indexConfig.config.InfoLogger.Printf("Starting http server at %s", s.Addr)
	}
	ctx.started = time.Now()
	err := s.ListenAndServe()
	if !ctx.shutdown {
		ctx.indexConfig.config.ErrorLogger.Fatalf("Unable to serve http at address %s: %s", s.Addr, err)
	}
}

func startHTTPServer(ctx *httpServerCtx) {
	ctx.buildServer()
	ctx.started = time.Now()
	ctx.serveHTTP()
}
