// Package content 提供编译进后端二进制的小说章节资源。
package content

import "embed"

// Files 包含当前目录下所有以数字命名的小说章节。
//
//go:embed *.txt
var Files embed.FS
