<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { fetchChapter, fetchNovel } from './api'

const novel = ref(null)
const chapters = ref([])
const memoryImages = ref([])
const activeChapter = ref(null)
const memoryMoments = ref([])
const previewImage = ref(null)
const previewCloseButton = ref(null)
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
const siteShell = ref(null)
const homeEntry = ref(null)
const activeMemoryPage = ref(0)
let previewTrigger = null
let bodyOverflowBeforePreview = ''

const shownMemoryImagesKey = 'novel:shownMemoryImages'

const memorySlides = [
  {
    kicker: 'PROLOGUE 01 · 落球声',
    year: '2014',
    title: ['回忆总是', '先有声音。'],
    body: '一声篮球落地，穿过午后的蝉鸣。我们循着声音跑向球场，以为那只是一个普通的夏天。',
    quote: '后来才明白，有些故事从第一次传球时，就已经写下了开头。',
    caption: '篮球落地的声音，最先叫醒那个夏天',
    variant: 'opening',
  },
  {
    kicker: 'PROLOGUE 02 · 放学以后',
    year: '17:42',
    title: ['那时的黄昏，', '好像永远不会结束。'],
    body: '书包被丢在篮架下，校服换成球衣。风从树梢吹过，夕阳把每个人的影子拉得很长。',
    quote: '我们记不清那天的比分，却一直记得谁在最后一球时喊了自己的名字。',
    caption: '放学后的风，把黄昏吹得很慢',
    variant: 'after-school',
  },
  {
    kicker: 'PROLOGUE 03 · 少年并肩',
    year: '我们',
    title: ['输赢很大，', '明天很远。'],
    body: '那时相信，只要跑得够快、跳得够高，就能追上所有想去的地方。受过的伤，睡一觉就会好。',
    quote: '青春不是赢下了多少场，而是输掉以后，身边还有人说：明天继续。',
    caption: '一记传球，从一个少年飞向另一个少年',
    variant: 'together',
  },
  {
    kicker: 'PROLOGUE 04 · 后来',
    year: 'NOW',
    title: ['我们长大了，', '篮球还在落地。'],
    body: '球场换了颜色，少年走进各自的人生。可每当熟悉的声音响起，时间还是会折返回那一年。',
    quote: '故事没有停在合照那天。它被写进往后的每一次重逢，也写进这一页。',
    caption: '许多年后，篮筐仍在原地等风',
    variant: 'return',
  },
]

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

const memoryMomentsByPosition = computed(() =>
  Object.fromEntries(memoryMoments.value.map((moment) => [moment.insertAfter, moment])),
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

function updateMemoryPage() {
  if (!siteShell.value) return
  activeMemoryPage.value = Math.min(
    memorySlides.length,
    Math.max(0, Math.round(siteShell.value.scrollTop / siteShell.value.clientHeight)),
  )
}

function scrollToHome() {
  const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  homeEntry.value?.scrollIntoView({ behavior: reduceMotion ? 'auto' : 'smooth', block: 'start' })
}

function syncPrologueAnchor() {
  const anchorID = window.location.hash.slice(1)
  if (!anchorID || anchorID.startsWith('chapter-') || !siteShell.value) return
  const anchor = document.getElementById(anchorID)
  if (!anchor) return
  const shellTop = siteShell.value.getBoundingClientRect().top
  siteShell.value.scrollTop += anchor.getBoundingClientRect().top - shellTop
  updateMemoryPage()
}

function randomInteger(maxExclusive) {
  if (maxExclusive <= 1) return 0
  if (globalThis.crypto?.getRandomValues) {
    const value = new Uint32Array(1)
    globalThis.crypto.getRandomValues(value)
    return value[0] % maxExclusive
  }
  return Math.floor(Math.random() * maxExclusive)
}

function readShownMemoryImageIDs() {
  try {
    const storedIDs = JSON.parse(localStorage.getItem(shownMemoryImagesKey) || '[]')
    return new Set(Array.isArray(storedIDs) ? storedIDs.filter(Number.isInteger) : [])
  } catch {
    return new Set()
  }
}

function writeShownMemoryImageIDs(shownIDs) {
  try {
    if (shownIDs.size) localStorage.setItem(shownMemoryImagesKey, JSON.stringify([...shownIDs]))
    else localStorage.removeItem(shownMemoryImagesKey)
  } catch {
    // localStorage 被浏览器禁用时，阅读功能仍可正常使用。
  }
}

function resetShownMemoryImageIDs() {
  try {
    localStorage.removeItem(shownMemoryImagesKey)
  } catch {
    // localStorage 被浏览器禁用时无需处理。
  }
}

function chooseMemoryMoments(chapter) {
  const paragraphCount = chapter.content
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean).length
  if (!memoryImages.value.length || paragraphCount === 0) return []

  const characterCount = chapter.content.replace(/\s/g, '').length
  const desiredCount = characterCount >= 4000 ? 3 : characterCount >= 2500 ? 2 : 1
  const firstPosition = paragraphCount < 3 ? 0 : Math.max(1, Math.floor(paragraphCount * 0.15))
  const lastPosition = Math.max(firstPosition, Math.min(paragraphCount - 2, Math.ceil(paragraphCount * 0.85)))
  const positionCount = lastPosition - firstPosition + 1
  const imageCount = Math.min(desiredCount, memoryImages.value.length, positionCount)
  const validImageIDs = new Set(memoryImages.value.map((image) => image.id))
  const shownImageIDs = new Set(
    [...readShownMemoryImageIDs()].filter((imageID) => validImageIDs.has(imageID)),
  )
  const selectedImages = []

  for (let index = 0; index < imageCount; index += 1) {
    if (shownImageIDs.size >= memoryImages.value.length) shownImageIDs.clear()

    let candidates = memoryImages.value.filter(
      (image) => !shownImageIDs.has(image.id) && !selectedImages.some((selected) => selected.id === image.id),
    )
    if (!candidates.length) {
      shownImageIDs.clear()
      candidates = memoryImages.value.filter(
        (image) => !selectedImages.some((selected) => selected.id === image.id),
      )
    }

    const image = candidates[randomInteger(candidates.length)]
    selectedImages.push(image)
    shownImageIDs.add(image.id)
  }

  if (shownImageIDs.size >= memoryImages.value.length) shownImageIDs.clear()
  writeShownMemoryImageIDs(shownImageIDs)

  return selectedImages.map((image, slot) => {
    const slotStart = firstPosition + Math.floor((positionCount * slot) / imageCount)
    const slotEnd = firstPosition + Math.floor((positionCount * (slot + 1)) / imageCount) - 1
    return {
      image,
      insertAfter: slotStart + randomInteger(slotEnd - slotStart + 1),
    }
  })
}

async function openMemoryPreview(image, event) {
  previewTrigger = event.currentTarget
  bodyOverflowBeforePreview = document.body.style.overflow
  document.body.style.overflow = 'hidden'
  previewImage.value = image
  await nextTick()
  previewCloseButton.value?.focus()
}

function closeMemoryPreview() {
  if (!previewImage.value) return
  previewImage.value = null
  document.body.style.overflow = bodyOverflowBeforePreview
  nextTick(() => {
    previewTrigger?.focus()
    previewTrigger = null
  })
}

function handlePreviewKeydown(event) {
  if (event.key === 'Escape' && previewImage.value) closeMemoryPreview()
}

async function loadChapter(id, updateHistory = true) {
  if (!id || chapterLoading.value) return
  chapterLoading.value = true
  error.value = ''
  directoryOpen.value = false
  try {
    const payload = await fetchChapter(id)
    activeChapter.value = payload.chapter
    memoryMoments.value = chooseMemoryMoments(payload.chapter)
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
  memoryMoments.value = []
  directoryOpen.value = false
  if (updateHistory) window.history.pushState({}, '', window.location.pathname)
  nextTick(() => {
    window.scrollTo({ top: 0, behavior: 'instant' })
    if (!siteShell.value || !homeEntry.value) return
    siteShell.value.scrollTop = homeEntry.value.offsetTop
    activeMemoryPage.value = memorySlides.length
  })
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
  resetShownMemoryImageIDs()
  try {
    const payload = await fetchNovel()
    novel.value = payload.novel
    chapters.value = payload.chapters
    memoryImages.value = payload.memoryImages || []
    syncFromLocation()
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
  await nextTick()
  syncPrologueAnchor()
  window.addEventListener('scroll', updateProgress, { passive: true })
  window.addEventListener('popstate', syncFromLocation)
  window.addEventListener('keydown', handlePreviewKeydown)
})

onBeforeUnmount(() => {
  if (previewImage.value) document.body.style.overflow = bodyOverflowBeforePreview
  window.removeEventListener('scroll', updateProgress)
  window.removeEventListener('popstate', syncFromLocation)
  window.removeEventListener('keydown', handlePreviewKeydown)
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

  <main
    v-else-if="!activeChapter"
    ref="siteShell"
    class="site-shell"
    @scroll.passive="updateMemoryPage"
  >
    <section
      v-for="(slide, index) in memorySlides"
      :key="slide.kicker"
      :id="`memory-${index + 1}`"
      :class="[
        'memory-page',
        `memory-page--${slide.variant}`,
        { 'is-active': activeMemoryPage === index },
      ]"
      :aria-label="`回忆序章第 ${index + 1} 页，共 ${memorySlides.length} 页`"
    >
      <div class="memory-grain" aria-hidden="true"></div>
      <div class="memory-court" aria-hidden="true"><span></span><i></i></div>

      <button class="memory-skip" type="button" @click="scrollToHome">
        跳过序章 <span aria-hidden="true">↘</span>
      </button>

      <div class="memory-copy">
        <p class="memory-kicker">{{ slide.kicker }}</p>
        <p class="memory-year" aria-hidden="true">{{ slide.year }}</p>
        <h1>{{ slide.title[0] }}<br /><em>{{ slide.title[1] }}</em></h1>
        <p class="memory-body">{{ slide.body }}</p>
        <blockquote class="memory-quote">{{ slide.quote }}</blockquote>
      </div>

      <div :class="['memory-illustration', `memory-illustration--${slide.variant}`]" aria-hidden="true">
        <div class="scene-sky">
          <span class="scene-sun"></span>
          <span class="scene-cloud scene-cloud--one"><i></i></span>
          <span class="scene-cloud scene-cloud--two"><i></i></span>
        </div>
        <div class="scene-school">
          <i></i><i></i><i></i><i></i><i></i><i></i>
        </div>
        <div class="scene-ground"></div>
        <div class="scene-lamp"><i></i></div>
        <div class="scene-hoop">
          <span class="scene-pole"></span>
          <span class="scene-board"></span>
          <span class="scene-rim"></span>
          <span class="scene-net"></span>
        </div>
        <div class="scene-player scene-player--a">
          <span class="scene-head"></span><span class="scene-body"></span>
          <span class="scene-arm scene-arm--left"></span><span class="scene-arm scene-arm--right"></span>
          <span class="scene-leg scene-leg--left"></span><span class="scene-leg scene-leg--right"></span>
        </div>
        <div class="scene-player scene-player--b">
          <span class="scene-head"></span><span class="scene-body"></span>
          <span class="scene-arm scene-arm--left"></span><span class="scene-arm scene-arm--right"></span>
          <span class="scene-leg scene-leg--left"></span><span class="scene-leg scene-leg--right"></span>
        </div>
        <div class="scene-ball"><span></span><i></i></div>
        <span class="scene-sound scene-sound--one"></span>
        <span class="scene-sound scene-sound--two"></span>
        <span class="scene-sound scene-sound--three"></span>
        <p><span>{{ String(index + 1).padStart(2, '0') }} / {{ String(memorySlides.length).padStart(2, '0') }}</span>{{ slide.caption }}</p>
      </div>

        <div class="memory-page-footer">
          <div class="memory-dots" aria-hidden="true">
            <span
              v-for="(_, dotIndex) in memorySlides"
              :key="dotIndex"
              :class="{ active: dotIndex === index }"
            ></span>
          </div>
          <p>{{ index === memorySlides.length - 1 ? '再向上滑，翻开故事' : '向上滑，继续回忆' }} <i>⌄</i></p>
        </div>
    </section>

    <div ref="homeEntry" class="home-entry">
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
    </div>
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
        <template v-for="(paragraph, index) in contentParagraphs" :key="index">
          <p>{{ paragraph }}</p>
          <figure v-if="memoryMomentsByPosition[index]" class="memory-photo">
            <button
              type="button"
              class="memory-photo-trigger"
              aria-label="放大查看这张篮球回忆照片"
              @click="openMemoryPreview(memoryMomentsByPosition[index].image, $event)"
            >
              <img
                :src="memoryMomentsByPosition[index].image.url"
                alt="少年时期的篮球回忆照片"
                loading="lazy"
                decoding="async"
                @load="updateProgress"
              />
            </button>
            <figcaption><span>MEMORY</span> 那年球场边的我们</figcaption>
          </figure>
        </template>
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

  <Teleport to="body">
    <div
      v-if="previewImage"
      class="image-lightbox"
      role="dialog"
      aria-modal="true"
      aria-label="篮球回忆照片预览"
      @click.self="closeMemoryPreview"
    >
      <button
        ref="previewCloseButton"
        type="button"
        class="image-lightbox-close"
        aria-label="关闭图片预览"
        @click="closeMemoryPreview"
      >
        ×
      </button>
      <img :src="previewImage.url" alt="放大的少年时期篮球回忆照片" />
      <p>按 Esc 或点击空白处关闭</p>
    </div>
  </Teleport>
</template>
