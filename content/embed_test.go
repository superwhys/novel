package content_test

import (
	"testing"

	"github.com/superwhys/novel/content"
	"github.com/superwhys/novel/internal/library"
)

func TestEmbeddedNovelContent(t *testing.T) {
	lib, err := library.Load(content.Files)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := lib.Novel().ChapterCount; got != 8 {
		t.Fatalf("ChapterCount = %d, want 8", got)
	}
	if chapter, ok := lib.Chapter(1); !ok || chapter.ShortTitle != "梦想" {
		t.Fatalf("embedded first chapter = %#v, exists = %v", chapter, ok)
	}
}
