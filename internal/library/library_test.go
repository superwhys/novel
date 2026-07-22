package library

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

const testStagesJSON = `[{"number":1,"title":"起步","startChapter":1,"endChapter":2}]`

func TestLoadSortsAndParsesChapters(t *testing.T) {
	contentFS := fstest.MapFS{
		"2.txt":        {Data: []byte("第二章 挑战\n第二章的正文。\n又一段。")},
		"1.txt":        {Data: []byte("第一章梦想\n第一章的正文。")},
		"teasers.json": {Data: []byte(`{"1":"梦想会从这里开始吗？","2":"他们能接下挑战吗？"}`)},
		"stages.json":  {Data: []byte(testStagesJSON)},
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
	if chapters[0].Teaser != "梦想会从这里开始吗？" {
		t.Fatalf("first chapter teaser = %q", chapters[0].Teaser)
	}
	if chapters[1].Paragraphs != 2 {
		t.Fatalf("paragraphs = %d, want 2", chapters[1].Paragraphs)
	}
	stages := lib.Stages()
	if len(stages) != 1 || stages[0].Title != "起步" {
		t.Fatalf("stages = %#v", stages)
	}
}

func TestLoadRejectsChapterOutsideStages(t *testing.T) {
	_, err := Load(fstest.MapFS{
		"1.txt":        {Data: []byte("第一章 梦想\n正文")},
		"teasers.json": {Data: []byte(`{"1":"梦想会从这里开始吗？"}`)},
		"stages.json":  {Data: []byte(`[{"number":2,"title":"后来","startChapter":2,"endChapter":3}]`)},
	})
	if err == nil || !strings.Contains(err.Error(), "没有覆盖第1章") {
		t.Fatalf("Load() error = %v, want uncovered chapter error", err)
	}
}

func TestLoadRejectsMissingChapterTeaser(t *testing.T) {
	_, err := Load(fstest.MapFS{
		"1.txt":        {Data: []byte("第一章 梦想\n正文")},
		"teasers.json": {Data: []byte(`{}`)},
	})
	if err == nil || !strings.Contains(err.Error(), "缺少第1章") {
		t.Fatalf("Load() error = %v, want missing chapter teaser error", err)
	}
}

func TestLoadRejectsNonQuestionTeaser(t *testing.T) {
	_, err := Load(fstest.MapFS{
		"1.txt":        {Data: []byte("第一章 梦想\n正文")},
		"teasers.json": {Data: []byte(`{"1":"梦想从这里开始。"}`)},
	})
	if err == nil || !strings.Contains(err.Error(), "必须是疑问句") {
		t.Fatalf("Load() error = %v, want question teaser error", err)
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
	if got := lib.Novel().ChapterCount; got != 17 {
		t.Fatalf("ChapterCount = %d, want 17", got)
	}
	if chapter, ok := lib.Chapter(1); !ok || chapter.ShortTitle != "梦想" {
		t.Fatalf("first chapter = %#v, exists = %v", chapter, ok)
	}
}
