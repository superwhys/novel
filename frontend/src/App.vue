<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { fetchChapter, fetchNovel } from './api'

const novel = ref(null)
const chapters = ref([])
const activeChapter = ref(null)
const loading = ref(true)
const chapterLoading = ref(false)
const error = ref('')
const query = ref('')
const directoryOpen = ref(false)
const readProgress = ref(0)
const lastChapterID = ref(Number(localStorage.getItem('novel:lastChapter')) || 0)
const fontSize = ref(Number(localStorage.getItem('novel:fontSize')) || 19)
const lineHeight = ref(Number(localStorage.getItem('novel:lineHeight')) || 2)
const theme = ref(localStorage.getItem('novel:theme') || 'paper')

const filteredChapters = computed(() => {
  const keyword = query.value.trim().toLowerCase()
  if (!keyword) return chapters.value
  return chapters.value.filter((chapter) =>
    `${chapter.title}${chapter.excerpt}`.toLowerCase().includes(keyword),
  )
})

const contentParagraphs = computed(() =>
  activeChapter.value?.content
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean) || [],
)

const currentIndex = computed(() =>
  chapters.value.findIndex((chapter) => chapter.id === activeChapter.value?.id),
)

const previousChapter = computed(() => chapters.value[currentIndex.value - 1] || null)
const nextChapter = computed(() => chapters.value[currentIndex.value + 1] || null)
const savedChapter = computed(() => chapters.value.find((chapter) => chapter.id === lastChapterID.value))

function formatCharacters(value) {
  if (!value) return '0'
  return value >= 10000 ? `${(value / 10000).toFixed(1)} 万` : value.toLocaleString('zh-CN')
}

function reloadPage() {
  window.location.reload()
}

async function loadChapter(id, updateHistory = true) {
  if (!id || chapterLoading.value) return
  chapterLoading.value = true
  error.value = ''
  directoryOpen.value = false
  try {
    const payload = await fetchChapter(id)
    activeChapter.value = payload.chapter
    lastChapterID.value = id
    localStorage.setItem('novel:lastChapter', String(id))
    if (updateHistory && window.location.hash !== `#chapter-${id}`) {
      window.history.pushState({ chapter: id }, '', `#chapter-${id}`)
    }
    await nextTick()
    window.scrollTo({ top: 0, behavior: 'instant' })
    updateProgress()
  } catch (err) {
    error.value = err.message
  } finally {
    chapterLoading.value = false
  }
}

function closeReader(updateHistory = true) {
  activeChapter.value = null
  directoryOpen.value = false
  if (updateHistory) window.history.pushState({}, '', window.location.pathname)
  nextTick(() => window.scrollTo({ top: 0, behavior: 'instant' }))
}

function syncFromLocation() {
  const match = window.location.hash.match(/^#chapter-(\d+)$/)
  if (match) loadChapter(Number(match[1]), false)
  else closeReader(false)
}

function updateProgress() {
  if (!activeChapter.value) {
    readProgress.value = 0
    return
  }
  const max = document.documentElement.scrollHeight - window.innerHeight
  readProgress.value = max > 0 ? Math.min(100, Math.max(0, (window.scrollY / max) * 100)) : 100
}

function changeFont(delta) {
  fontSize.value = Math.min(25, Math.max(16, fontSize.value + delta))
  localStorage.setItem('novel:fontSize', String(fontSize.value))
}

function changeLineHeight() {
  lineHeight.value = lineHeight.value === 2 ? 2.25 : lineHeight.value === 2.25 ? 1.8 : 2
  localStorage.setItem('novel:lineHeight', String(lineHeight.value))
}

function toggleTheme() {
  theme.value = theme.value === 'paper' ? 'night' : 'paper'
  localStorage.setItem('novel:theme', theme.value)
}

onMounted(async () => {
  try {
    const payload = await fetchNovel()
    novel.value = payload.novel
    chapters.value = payload.chapters
    syncFromLocation()
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
  window.addEventListener('scroll', updateProgress, { passive: true })
  window.addEventListener('popstate', syncFromLocation)
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', updateProgress)
  window.removeEventListener('popstate', syncFromLocation)
})
</script>

<template>
  <div v-if="loading" class="loading-screen">
    <div class="loading-ball" aria-hidden="true"></div>
    <p>正在打开球场边的故事…</p>
  </div>

  <div v-else-if="!novel" class="error-screen">
    <span class="error-number" aria-hidden="true">暂歇</span>
    <h1>故事暂时没有加载出来</h1>
    <p>{{ error }}</p>
    <button class="button button-primary" @click="reloadPage">重新加载</button>
  </div>

  <main v-else-if="!activeChapter" class="site-shell">
    <nav class="top-nav" aria-label="主导航">
      <a href="#top" class="brand" aria-label="回到首页">
        <span class="brand-mark">距</span>
        <span>我与篮球的距离</span>
      </a>
      <div class="nav-links">
        <a href="#about">故事</a>
        <a href="#chapters">章节</a>
        <button v-if="savedChapter" class="nav-continue" @click="loadChapter(savedChapter.id)">
          继续阅读
        </button>
      </div>
    </nav>

    <section id="top" class="hero">
      <div class="court-lines" aria-hidden="true">
        <span class="court-circle"></span>
        <span class="court-arc"></span>
      </div>
      <div class="hero-copy">
        <p class="eyebrow">A CAMPUS BASKETBALL STORY · 2014</p>
        <h1>我与篮球<br /><em>的距离</em></h1>
        <p class="hero-description">{{ novel.subtitle }}</p>
        <div class="hero-actions">
          <button class="button button-light" @click="loadChapter(savedChapter?.id || chapters[0].id)">
            <span class="play-icon">▶</span>
            {{ savedChapter ? `继续第 ${savedChapter.number} 章` : '开始阅读' }}
          </button>
          <a href="#chapters" class="button button-ghost">查看目录</a>
        </div>
        <div class="hero-stats" aria-label="小说数据">
          <div><strong>{{ novel.chapterCount }}</strong><span>已发布章节</span></div>
          <div><strong>{{ formatCharacters(novel.totalCharacters) }}</strong><span>正文字符</span></div>
          <div><strong>{{ novel.totalReadingMinutes }}</strong><span>分钟读完</span></div>
        </div>
      </div>

      <div class="hero-art" aria-hidden="true">
        <div class="year-stamp">2014</div>
        <div class="backboard"><span></span></div>
        <div class="hoop"><i></i><i></i><i></i><i></i></div>
        <div class="basketball"><span></span><span></span></div>
        <p>GONGMING<br />HIGH SCHOOL</p>
      </div>
      <a href="#about" class="scroll-cue"><span></span>向下探索</a>
    </section>

    <section id="about" class="about-section">
      <div class="section-index">01</div>
      <div class="about-heading">
        <p class="section-kicker">故事简介 / THE STORY</p>
        <h2>梦想不会自己<br />落进篮筐。</h2>
      </div>
      <div class="about-copy">
        <p>{{ novel.description }}</p>
        <div class="story-tags">
          <span>{{ novel.genre }}</span>
          <span>{{ novel.era }}</span>
        </div>
        <blockquote>
          <span>“</span>
          他们不可能打倒我，除非杀了我。任何不能杀死我的，只会令我更坚强。
        </blockquote>
      </div>
    </section>

    <section id="chapters" class="chapters-section">
      <div class="section-title-row">
        <div>
          <p class="section-kicker">章节目录 / CHAPTERS</p>
          <h2>每一次上场，<br />都离梦想更近。</h2>
        </div>
        <label class="chapter-search">
          <span aria-hidden="true">⌕</span>
          <input v-model="query" type="search" placeholder="搜索章节或片段" />
        </label>
      </div>

      <div v-if="error" class="inline-error">{{ error }}</div>

      <div class="chapter-list">
        <button
          v-for="chapter in filteredChapters"
          :key="chapter.id"
          class="chapter-card"
          @click="loadChapter(chapter.id)"
        >
          <span class="chapter-number">{{ String(chapter.number).padStart(2, '0') }}</span>
          <span class="chapter-info">
            <strong>{{ chapter.shortTitle }}</strong>
            <small>{{ chapter.excerpt }}</small>
          </span>
          <span class="chapter-meta">
            {{ chapter.readingMinutes }} 分钟
            <i>↗</i>
          </span>
        </button>
      </div>
      <p v-if="!filteredChapters.length" class="empty-result">没有找到相关章节，换个关键词试试。</p>
    </section>

    <footer class="site-footer">
      <span class="footer-mark">距</span>
      <div><strong>我与篮球的距离</strong><p>献给那些在球场上挥洒过汗水的少年。</p></div>
      <a href="#top">回到顶部 ↑</a>
    </footer>
  </main>

  <main v-else :class="['reader-shell', `theme-${theme}`]">
    <div class="progress-track"><span :style="{ width: `${readProgress}%` }"></span></div>
    <header class="reader-header">
      <button class="icon-button back-button" aria-label="返回书页" @click="closeReader()">←</button>
      <button class="reader-book-title" @click="closeReader()">{{ novel.title }}</button>
      <span class="reader-chapter-label">{{ activeChapter.title }}</span>
      <div class="reader-tools">
        <button aria-label="减小字号" @click="changeFont(-1)">A−</button>
        <button aria-label="增大字号" @click="changeFont(1)">A+</button>
        <button aria-label="调整行距" @click="changeLineHeight">↕</button>
        <button :aria-label="theme === 'paper' ? '切换深色模式' : '切换浅色模式'" @click="toggleTheme">
          {{ theme === 'paper' ? '◐' : '☼' }}
        </button>
        <button aria-label="打开目录" @click="directoryOpen = true">☰</button>
      </div>
    </header>

    <div v-if="chapterLoading" class="chapter-loading">正在翻页…</div>

    <article
      class="reader-article"
      :style="{ '--reader-font-size': `${fontSize}px`, '--reader-line-height': lineHeight }"
    >
      <div class="article-meta">
        <span>CHAPTER {{ String(activeChapter.number).padStart(2, '0') }}</span>
        <span>{{ activeChapter.readingMinutes }} MIN READ</span>
      </div>
      <h1>{{ activeChapter.shortTitle }}</h1>
      <div class="title-rule"><span></span></div>
      <div class="article-content">
        <p v-for="(paragraph, index) in contentParagraphs" :key="index">{{ paragraph }}</p>
      </div>
      <div class="chapter-end">
        <span>END OF CHAPTER {{ String(activeChapter.number).padStart(2, '0') }}</span>
      </div>
      <nav class="chapter-navigation" aria-label="章节切换">
        <button :disabled="!previousChapter" @click="previousChapter && loadChapter(previousChapter.id)">
          <small>← 上一章</small>
          <strong>{{ previousChapter?.shortTitle || '已经是第一章' }}</strong>
        </button>
        <button class="next" :disabled="!nextChapter" @click="nextChapter && loadChapter(nextChapter.id)">
          <small>下一章 →</small>
          <strong>{{ nextChapter?.shortTitle || '未完待续' }}</strong>
        </button>
      </nav>
    </article>

    <div v-if="directoryOpen" class="drawer-backdrop" @click.self="directoryOpen = false">
      <aside class="directory-drawer" aria-label="章节目录">
        <div class="drawer-heading">
          <div><small>TABLE OF CONTENTS</small><h2>章节目录</h2></div>
          <button aria-label="关闭目录" @click="directoryOpen = false">×</button>
        </div>
        <div class="drawer-list">
          <button
            v-for="chapter in chapters"
            :key="chapter.id"
            :class="{ active: chapter.id === activeChapter.id }"
            @click="loadChapter(chapter.id)"
          >
            <span>{{ String(chapter.number).padStart(2, '0') }}</span>
            <strong>{{ chapter.shortTitle }}</strong>
            <small>{{ chapter.readingMinutes }} 分钟</small>
          </button>
        </div>
      </aside>
    </div>
  </main>
</template>
