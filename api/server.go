package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/miebyte/goutils/logging"

	"github.com/superwhys/novel/internal/library"
)

type API struct {
	library *library.Library
	mux     *http.ServeMux
}

func New(lib *library.Library) http.Handler {
	api := &API{library: lib, mux: http.NewServeMux()}
	api.routes()
	return api.withMiddleware(api.mux)
}

func (a *API) routes() {
	a.mux.HandleFunc("GET /health", a.health)
	a.mux.HandleFunc("GET /novel", a.novel)
	a.mux.HandleFunc("GET /chapters", a.chapters)
	a.mux.HandleFunc("GET /chapters/{id}", a.chapter)
}

func (a *API) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) novel(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, struct {
		Novel    library.Novel            `json:"novel"`
		Chapters []library.ChapterSummary `json:"chapters"`
	}{a.library.Novel(), a.library.Chapters()})
}

func (a *API) chapters(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"chapters": a.library.Chapters()})
}

func (a *API) chapter(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "无效的章节编号")
		return
	}
	chapter, ok := a.library.Chapter(id)
	if !ok {
		writeError(w, http.StatusNotFound, "章节不存在")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"chapter": chapter})
}

func (a *API) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin(r.Header.Get("Origin")))
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Vary", "Origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
		logging.Infow(
			r.Context(),
			"http request",
			logging.String("method", r.Method),
			logging.String("path", r.URL.Path),
			logging.Duration("duration", time.Since(started)),
		)
	})
}

func allowedOrigin(origin string) string {
	if origin == "" || strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
		if origin != "" {
			return origin
		}
	}
	return "http://localhost:5173"
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		logging.Errorf("encode response error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
