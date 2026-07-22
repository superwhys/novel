package library

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	chapterTeasersFile = "teasers.json"
	chapterStagesFile  = "stages.json"
)

var chapterFilePattern = regexp.MustCompile(`^(\d+)\.txt$`)

type chapterCandidate struct {
	number int
	name   string
}

type Novel struct {
	Title              string `json:"title"`
	Subtitle           string `json:"subtitle"`
	Description        string `json:"description"`
	Genre              string `json:"genre"`
	Era                string `json:"era"`
	ChapterCount       int    `json:"chapterCount"`
	TotalCharacters    int    `json:"totalCharacters"`
	TotalReadingMinute int    `json:"totalReadingMinutes"`
}

type ChapterSummary struct {
	ID             int    `json:"id"`
	Number         int    `json:"number"`
	Title          string `json:"title"`
	ShortTitle     string `json:"shortTitle"`
	Teaser         string `json:"teaser"`
	Characters     int    `json:"characters"`
	Paragraphs     int    `json:"paragraphs"`
	ReadingMinutes int    `json:"readingMinutes"`
}

type ChapterStage struct {
	Number       int    `json:"number"`
	Title        string `json:"title"`
	StartChapter int    `json:"startChapter"`
	EndChapter   int    `json:"endChapter"`
}

type Chapter struct {
	ChapterSummary
	Content string `json:"content"`
}

type Library struct {
	novel    Novel
	stages   []ChapterStage
	chapters []Chapter
	byID     map[int]Chapter
}

func Load(contentFS fs.FS) (*Library, error) {
	entries, err := fs.ReadDir(contentFS, ".")
	if err != nil {
		return nil, fmt.Errorf("读取小说内容目录: %w", err)
	}

	var candidates []chapterCandidate
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		match := chapterFilePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		number, _ := strconv.Atoi(match[1])
		candidates = append(candidates, chapterCandidate{number: number, name: entry.Name()})
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("小说内容目录中没有找到数字命名的章节文件")
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].number < candidates[j].number })
	teasers, err := loadChapterTeasers(contentFS, candidates)
	if err != nil {
		return nil, err
	}
	stages, err := loadChapterStages(contentFS, candidates)
	if err != nil {
		return nil, err
	}

	lib := &Library{stages: stages, byID: make(map[int]Chapter)}
	for _, item := range candidates {
		data, err := fs.ReadFile(contentFS, item.name)
		if err != nil {
			return nil, fmt.Errorf("读取章节 %s: %w", item.name, err)
		}
		if !utf8.Valid(data) {
			return nil, fmt.Errorf("章节 %s 不是有效的 UTF-8 文本", item.name)
		}
		chapter := parseChapter(item.number, string(data), teasers[item.number])
		lib.chapters = append(lib.chapters, chapter)
		lib.byID[chapter.ID] = chapter
	}

	lib.novel = Novel{
		Title:        "我与篮球的距离",
		Subtitle:     "一群少年，在球场上寻找梦想与自己的距离",
		Description:  "发生在 2014 年深圳光明新区公明中学的校园篮球故事。热血、友情、失落与成长，都从一次传球开始。",
		Genre:        "校园 · 篮球 · 青春",
		Era:          "2014 · 深圳公明",
		ChapterCount: len(lib.chapters),
	}
	for _, chapter := range lib.chapters {
		lib.novel.TotalCharacters += chapter.Characters
		lib.novel.TotalReadingMinute += chapter.ReadingMinutes
	}
	return lib, nil
}

func (l *Library) Novel() Novel {
	return l.novel
}

func (l *Library) Chapters() []ChapterSummary {
	result := make([]ChapterSummary, 0, len(l.chapters))
	for _, chapter := range l.chapters {
		result = append(result, chapter.ChapterSummary)
	}
	return result
}

func (l *Library) Stages() []ChapterStage {
	return append([]ChapterStage(nil), l.stages...)
}

func (l *Library) Chapter(id int) (Chapter, bool) {
	chapter, ok := l.byID[id]
	return chapter, ok
}

func parseChapter(number int, raw, teaser string) Chapter {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.TrimSpace(raw)
	lines := strings.Split(raw, "\n")

	title := fmt.Sprintf("第%d章", number)
	shortTitle := "未命名"
	bodyStart := 0
	for index, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		title = normalizeTitle(line, number)
		shortTitle = extractShortTitle(line)
		bodyStart = index + 1
		break
	}
	body := strings.TrimSpace(strings.Join(lines[bodyStart:], "\n"))
	characters := countReadableCharacters(body)
	readingMinutes := (characters + 499) / 500
	if readingMinutes < 1 {
		readingMinutes = 1
	}

	return Chapter{
		ChapterSummary: ChapterSummary{
			ID:             number,
			Number:         number,
			Title:          title,
			ShortTitle:     shortTitle,
			Teaser:         teaser,
			Characters:     characters,
			Paragraphs:     countParagraphs(body),
			ReadingMinutes: readingMinutes,
		},
		Content: body,
	}
}

func loadChapterTeasers(contentFS fs.FS, candidates []chapterCandidate) (map[int]string, error) {
	data, err := fs.ReadFile(contentFS, chapterTeasersFile)
	if err != nil {
		return nil, fmt.Errorf("读取章节预告 %s: %w", chapterTeasersFile, err)
	}
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("章节预告 %s 不是有效的 UTF-8 文本", chapterTeasersFile)
	}

	var stored map[string]string
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, fmt.Errorf("解析章节预告 %s: %w", chapterTeasersFile, err)
	}

	chapterNumbers := make(map[int]struct{}, len(candidates))
	for _, candidate := range candidates {
		chapterNumbers[candidate.number] = struct{}{}
	}

	teasers := make(map[int]string, len(stored))
	for rawNumber, rawTeaser := range stored {
		number, err := strconv.Atoi(rawNumber)
		if err != nil || number < 1 {
			return nil, fmt.Errorf("章节预告 %s 包含无效章节编号 %q", chapterTeasersFile, rawNumber)
		}
		if _, ok := chapterNumbers[number]; !ok {
			return nil, fmt.Errorf("章节预告 %s 中的第%d章没有对应正文", chapterTeasersFile, number)
		}
		teaser := strings.TrimSpace(rawTeaser)
		if teaser == "" {
			return nil, fmt.Errorf("章节预告 %s 中的第%d章预告为空", chapterTeasersFile, number)
		}
		if !strings.HasSuffix(teaser, "？") && !strings.HasSuffix(teaser, "?") {
			return nil, fmt.Errorf("章节预告 %s 中的第%d章预告必须是疑问句", chapterTeasersFile, number)
		}
		teasers[number] = teaser
	}

	for _, candidate := range candidates {
		if _, ok := teasers[candidate.number]; !ok {
			return nil, fmt.Errorf("章节预告 %s 缺少第%d章", chapterTeasersFile, candidate.number)
		}
	}
	return teasers, nil
}

func loadChapterStages(contentFS fs.FS, candidates []chapterCandidate) ([]ChapterStage, error) {
	data, err := fs.ReadFile(contentFS, chapterStagesFile)
	if err != nil {
		return nil, fmt.Errorf("读取章节阶段 %s: %w", chapterStagesFile, err)
	}
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("章节阶段 %s 不是有效的 UTF-8 文本", chapterStagesFile)
	}

	var stages []ChapterStage
	if err := json.Unmarshal(data, &stages); err != nil {
		return nil, fmt.Errorf("解析章节阶段 %s: %w", chapterStagesFile, err)
	}
	if len(stages) == 0 {
		return nil, fmt.Errorf("章节阶段 %s 不能为空", chapterStagesFile)
	}

	seenNumbers := make(map[int]struct{}, len(stages))
	for index := range stages {
		stage := &stages[index]
		stage.Title = strings.TrimSpace(stage.Title)
		if stage.Number < 1 || stage.Title == "" || stage.StartChapter < 1 || stage.EndChapter < stage.StartChapter {
			return nil, fmt.Errorf("章节阶段 %s 中的第%d项无效", chapterStagesFile, index+1)
		}
		if _, ok := seenNumbers[stage.Number]; ok {
			return nil, fmt.Errorf("章节阶段 %s 中的阶段编号 %d 重复", chapterStagesFile, stage.Number)
		}
		seenNumbers[stage.Number] = struct{}{}
		for previousIndex := 0; previousIndex < index; previousIndex++ {
			previous := stages[previousIndex]
			if stage.StartChapter <= previous.EndChapter && previous.StartChapter <= stage.EndChapter {
				return nil, fmt.Errorf("章节阶段 %s 中的阶段 %d 与阶段 %d 范围重叠", chapterStagesFile, previous.Number, stage.Number)
			}
		}
	}

	for _, candidate := range candidates {
		covered := false
		for _, stage := range stages {
			if candidate.number >= stage.StartChapter && candidate.number <= stage.EndChapter {
				covered = true
				break
			}
		}
		if !covered {
			return nil, fmt.Errorf("章节阶段 %s 没有覆盖第%d章", chapterStagesFile, candidate.number)
		}
	}

	sort.Slice(stages, func(i, j int) bool { return stages[i].Number < stages[j].Number })
	return stages, nil
}

func normalizeTitle(raw string, number int) string {
	trimmed := strings.TrimSpace(raw)
	prefixes := []string{
		fmt.Sprintf("第%d章", number),
		fmt.Sprintf("第%s章", chineseNumber(number)),
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(trimmed, prefix) {
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
			if rest == "" {
				return prefix
			}
			return prefix + " · " + strings.TrimLeft(rest, "：:· ")
		}
	}
	return fmt.Sprintf("第%d章 · %s", number, trimmed)
}

func extractShortTitle(raw string) string {
	if index := strings.Index(raw, "章"); index >= 0 {
		value := strings.TrimSpace(strings.TrimLeft(raw[index+len("章"):], "：:· "))
		if value != "" {
			return value
		}
	}
	return strings.TrimSpace(raw)
}

func countReadableCharacters(text string) int {
	count := 0
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			count++
		}
	}
	return count
}

func countParagraphs(text string) int {
	count := 0
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

func chineseNumber(number int) string {
	numbers := map[int]string{1: "一", 2: "二", 3: "三", 4: "四", 5: "五", 6: "六", 7: "七", 8: "八", 9: "九", 10: "十"}
	if value, ok := numbers[number]; ok {
		return value
	}
	return strconv.Itoa(number)
}
