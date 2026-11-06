package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	w      http.ResponseWriter
	zw     *gzip.Writer
	status int
	wrote  bool
}

func (g *gzipResponseWriter) Header() http.Header {
	return g.w.Header()
}

func (g *gzipResponseWriter) Write(p []byte) (int, error) {
	if !g.wrote {
		g.wrote = true
		if g.status == 0 {
			g.status = http.StatusOK
		}
		g.w.Header().Set("Content-Encoding", "gzip")
		g.w.WriteHeader(g.status)
	}
	if g.w.Header().Get("Content-Type") != "" &&
		!strings.Contains(g.w.Header().Get("Content-Type"), "application/json") &&
		!strings.Contains(g.w.Header().Get("Content-Type"), "text/") {
		return g.w.Write(p)
	}
	return g.zw.Write(p)
}

func (g *gzipResponseWriter) WriteHeader(statusCode int) {
	if g.wrote {
		return
	}
	g.wrote = true
	g.status = statusCode
	g.w.Header().Set("Content-Encoding", "gzip")
	g.w.WriteHeader(statusCode)
}

func Compression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Vary", "Accept-Encoding")
		zw := gzip.NewWriter(w)
		defer zw.Close()
		gw := &gzipResponseWriter{w: w, zw: zw}
		next.ServeHTTP(gw, r)
	})
}

var _ io.Writer = (*gzipResponseWriter)(nil)
