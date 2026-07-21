package memories

import (
	"embed"
	"path"
	"sort"
	"strconv"
	"strings"
)

// imageFS is compiled into the Go executable. Adding an image to images/
// therefore requires rebuilding the backend binary.
//
//go:embed images/*
var imageFS embed.FS

type imageFile struct {
	name        string
	contentType string
}

type Image struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type Asset struct {
	Data        []byte
	ContentType string
}

var imageFiles = loadImageFiles()

func loadImageFiles() []imageFile {
	entries, err := imageFS.ReadDir("images")
	if err != nil {
		panic("读取内嵌回忆图片失败: " + err.Error())
	}

	files := make([]imageFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		contentType, ok := supportedContentType(path.Ext(entry.Name()))
		if !ok {
			continue
		}
		files = append(files, imageFile{name: entry.Name(), contentType: contentType})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })
	return files
}

func supportedContentType(extension string) (string, bool) {
	switch strings.ToLower(extension) {
	case ".jpg", ".jpeg":
		return "image/jpeg", true
	case ".png":
		return "image/png", true
	case ".webp":
		return "image/webp", true
	case ".gif":
		return "image/gif", true
	case ".avif":
		return "image/avif", true
	default:
		return "", false
	}
}

func Images(apiPrefix string) []Image {
	apiPrefix = strings.TrimRight(apiPrefix, "/")
	images := make([]Image, 0, len(imageFiles))
	for index := range imageFiles {
		id := index + 1
		images = append(images, Image{ID: id, URL: apiPrefix + "/memories/" + strconv.Itoa(id)})
	}
	return images
}

func Read(id int) (Asset, bool) {
	if id < 1 || id > len(imageFiles) {
		return Asset{}, false
	}
	image := imageFiles[id-1]
	data, err := imageFS.ReadFile("images/" + image.name)
	if err != nil {
		return Asset{}, false
	}
	return Asset{Data: data, ContentType: image.contentType}, true
}
