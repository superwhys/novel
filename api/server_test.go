package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/superwhys/novel/internal/library"
)

func TestChapterEndpoint(t *testing.T) {
	lib, err := library.Load(fstest.MapFS{
		"1.txt":        {Data: []byte("第一章 梦想\n这是正文。")},
		"teasers.json": {Data: []byte(`{"1":"梦想会从这里开始吗？"}`)},
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
		"1.txt":        {Data: []byte("第一章 梦想\n正文")},
		"teasers.json": {Data: []byte(`{"1":"梦想会从这里开始吗？"}`)},
	})
	recorder := httptest.NewRecorder()
	New(lib).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/chapters/99", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
}

func TestNovelListsEmbeddedMemoryImages(t *testing.T) {
	lib, _ := library.Load(fstest.MapFS{
		"1.txt":        {Data: []byte("第一章 梦想\n正文")},
		"teasers.json": {Data: []byte(`{"1":"梦想会从这里开始吗？"}`)},
	})
	recorder := httptest.NewRecorder()
	New(lib).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/novel", nil))

	var payload struct {
		Chapters []struct {
			Teaser string `json:"teaser"`
		} `json:"chapters"`
		MemoryImages []struct {
			ID  int    `json:"id"`
			URL string `json:"url"`
		} `json:"memoryImages"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.MemoryImages) == 0 || payload.MemoryImages[0].URL != "/api/memories/1" {
		t.Fatalf("memoryImages = %#v", payload.MemoryImages)
	}
	if len(payload.Chapters) != 1 || payload.Chapters[0].Teaser != "梦想会从这里开始吗？" {
		t.Fatalf("chapters = %#v", payload.Chapters)
	}
}

func TestEmbeddedMemoryImageEndpoint(t *testing.T) {
	lib, _ := library.Load(fstest.MapFS{
		"1.txt":        {Data: []byte("第一章 梦想\n正文")},
		"teasers.json": {Data: []byte(`{"1":"梦想会从这里开始吗？"}`)},
	})
	recorder := httptest.NewRecorder()
	New(lib).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/memories/1", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Fatalf("Content-Type = %q, want image/jpeg", got)
	}
	if body := recorder.Body.Bytes(); len(body) < 3 || body[0] != 0xff || body[1] != 0xd8 || body[2] != 0xff {
		t.Fatal("response is not a JPEG image")
	}
}
