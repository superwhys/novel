package library

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestLoadSortsAndParsesChapters(t *testing.T) {
	contentFS := fstest.MapFS{
		"2.txt": {Data: []byte("第二章 挑战\n第二章的正文。\n又一段。")},
		"1.txt": {Data: []byte("第一章梦想\n第一章的正文。")},
	}

	lib, err := Load(contentFS)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := lib.Novel().ChapterCount; got != 2 {
		t.Fatalf("ChapterCount = %d, want 2", got)
	}
	chapters := lib.Chapters()
	if chapters[0].ID != 1 || chapters[0].ShortTitle != "梦想" {
		t.Fatalf("first chapter = %#v", chapters[0])
	}
	if chapters[1].Paragraphs != 2 {
		t.Fatalf("paragraphs = %d, want 2", chapters[1].Paragraphs)
	}
}

func TestLoadRejectsMissingChapterFiles(t *testing.T) {
	if _, err := Load(fstest.MapFS{}); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestLoadNovelContentFromDirectory(t *testing.T) {
	contentDir := filepath.Join("..", "..", "docs", "story", "content")
	lib, err := Load(os.DirFS(contentDir))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := lib.Novel().ChapterCount; got != 10 {
		t.Fatalf("ChapterCount = %d, want 10", got)
	}
	if chapter, ok := lib.Chapter(1); !ok || chapter.ShortTitle != "梦想" {
		t.Fatalf("first chapter = %#v, exists = %v", chapter, ok)
	}
}
