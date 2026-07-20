package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/superwhys/novel/internal/library"
)

func TestChapterEndpoint(t *testing.T) {
	lib, err := library.Load(fstest.MapFS{
		"1.txt": {Data: []byte("第一章 梦想\n这是正文。")},
	})
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/chapters/1", nil)
	New(lib).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "这是正文") {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestMissingChapterReturnsNotFound(t *testing.T) {
	lib, _ := library.Load(fstest.MapFS{
		"1.txt": {Data: []byte("第一章 梦想\n正文")},
	})
	recorder := httptest.NewRecorder()
	New(lib).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/chapters/99", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
}
