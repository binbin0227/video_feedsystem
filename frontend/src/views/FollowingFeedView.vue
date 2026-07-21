<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getFollowingFeed } from '../api/feed'
import FeedVideoSlide from '../components/video/FeedVideoSlide.vue'

const FEED_PAGE_SIZE = 2
const route = useRoute()
const router = useRouter()
const feedScroller = ref(null)
const videos = ref([])
const nextCursor = ref('')
const hasMore = ref(false)
const initialLoading = ref(true)
const loadingMore = ref(false)
const errorMessage = ref('')
const activeIndex = ref(0)
const soundEnabled = ref(false)
const volume = ref(0.3)
let wheelGestureLocked = false
let wheelUnlockTimer

async function loadFeed({ append = false } = {}) {
  if (append) {
    if (loadingMore.value || !hasMore.value) {
      return
    }
    loadingMore.value = true
  } else {
    initialLoading.value = true
  }

  errorMessage.value = ''

  try {
    const data = await getFollowingFeed({
      cursor: append ? nextCursor.value : '',
      limit: FEED_PAGE_SIZE,
    })
    const newVideos = Array.isArray(data?.videos) ? data.videos : []

    videos.value = append ? [...videos.value, ...newVideos] : newVideos
    nextCursor.value = typeof data?.next_cursor === 'string' ? data.next_cursor : ''
    hasMore.value = Boolean(data?.has_more && nextCursor.value)
  } catch (error) {
    if (error.response?.status === 401) {
      await router.replace({ name: 'login', query: { redirect: route.fullPath } })
      return
    }

    errorMessage.value = error.message
  } finally {
    initialLoading.value = false
    loadingMore.value = false
  }
}

function handleSlideActive(index) {
  activeIndex.value = index

  if (index === videos.value.length - 1 && hasMore.value) {
    loadFeed({ append: true })
  }
}

function scrollToVideo(index) {
  const slides = feedScroller.value?.querySelectorAll('.immersive-video-slide')
  const target = slides?.[index]

  if (!target || !feedScroller.value) {
    return
  }

  feedScroller.value.scrollTo({
    top: target.offsetTop,
    behavior: 'smooth',
  })
}

function handleFeedKeydown(event) {
  if (event.key === 'ArrowDown' || event.key === 'PageDown') {
    event.preventDefault()
    scrollToVideo(Math.min(activeIndex.value + 1, videos.value.length - 1))
  } else if (event.key === 'ArrowUp' || event.key === 'PageUp') {
    event.preventDefault()
    scrollToVideo(Math.max(activeIndex.value - 1, 0))
  }
}

function handleFeedWheel(event) {
  window.clearTimeout(wheelUnlockTimer)
  wheelUnlockTimer = window.setTimeout(() => {
    wheelGestureLocked = false
  }, 420)

  if (wheelGestureLocked || Math.abs(event.deltaY) < 8 || Math.abs(event.deltaY) <= Math.abs(event.deltaX)) {
    return
  }

  const direction = event.deltaY > 0 ? 1 : -1
  const targetIndex = Math.min(Math.max(activeIndex.value + direction, 0), videos.value.length - 1)

  if (targetIndex === activeIndex.value) {
    return
  }

  wheelGestureLocked = true
  activeIndex.value = targetIndex
  scrollToVideo(targetIndex)
}

onMounted(() => loadFeed())
onBeforeUnmount(() => window.clearTimeout(wheelUnlockTimer))
</script>

<template>
  <section class="immersive-feed-page">
    <div v-if="initialLoading" class="immersive-feed-state" aria-label="正在加载关注视频流" aria-busy="true">
      <span class="immersive-feed-loader"></span>
      <strong>正在准备关注流</strong>
      <p>获取你关注的作者最近发布的视频。</p>
    </div>

    <div v-else-if="errorMessage && videos.length === 0" class="immersive-feed-state" role="alert">
      <span class="immersive-feed-state__mark">!</span>
      <strong>关注视频流加载失败</strong>
      <p>{{ errorMessage }}</p>
      <button type="button" @click="loadFeed()">重新加载</button>
    </div>

    <div v-else-if="videos.length === 0" class="immersive-feed-state">
      <span class="immersive-feed-state__mark">0</span>
      <strong>关注流还是空的</strong>
      <p>关注一些创作者后，他们发布的视频会出现在这里。</p>
      <button type="button" @click="router.push({ name: 'search' })">去搜索用户</button>
    </div>

    <div
      v-else
      ref="feedScroller"
      class="immersive-feed-scroll"
      tabindex="0"
      aria-label="关注视频流"
      @keydown="handleFeedKeydown"
      @wheel.prevent="handleFeedWheel"
    >
      <FeedVideoSlide
        v-for="(video, index) in videos"
        :key="video.id"
        :video="video"
        :position="index + 1"
        :total="videos.length"
        :sound-enabled="soundEnabled"
        :volume="volume"
        :should-preload="Math.abs(index - activeIndex) <= 1"
        @active="handleSlideActive(index)"
        @toggle-sound="soundEnabled = $event"
        @update-volume="volume = $event"
      />
    </div>

    <div v-if="videos.length && (loadingMore || errorMessage)" class="immersive-feed-notice" aria-live="polite">
      <span v-if="loadingMore">正在加载更多视频…</span>
      <template v-else>
        <span role="alert">{{ errorMessage }}</span>
        <button type="button" @click="loadFeed({ append: true })">重试</button>
      </template>
    </div>
  </section>
</template>
