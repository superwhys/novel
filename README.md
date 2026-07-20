# 《我与篮球的距离》

这是《我与篮球的距离》的小说内容与在线阅读服务项目。

## 仓库目录

- `docs/story/content/`：小说章节正文；后端在运行时从该目录读取，不会编译进二进制文件。
- `docs/story/`：集中存放小说正文、大纲、人物档案、故事设定与续写规范。
- `api/`、`internal/`：后端服务代码。
- `frontend/`：网页阅读界面。

## 小说资料

- [小说大纲与章节架构](docs/story/OUTLINE.md)
- [故事设定档案](docs/story/STORY_BIBLE.md)
- [三位核心人物与重要辅助人物档案](docs/story/characters/README.md)
- [续写风格与时代设定](docs/story/WRITING_GUIDE.md)

## 运行配置

`config.json` 中的 `server.contentDir` 用于指定小说章节目录，默认值为 `docs/story/content`。目录中的数字命名文本文件（如 `1.txt`）会在服务启动时读取。
