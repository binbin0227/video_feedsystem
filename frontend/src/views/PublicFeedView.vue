<script setup>
import { onMounted, ref } from 'vue'
import { getPublicFeed } from '../api/feed'
import VideoCard from '../components/video/VideoCard.vue'

const FEED_PAGE_SIZE = 2
const videos = ref([])
const nextCursor = ref('')
const hasMore = ref(false)
const initialLoading = ref(true)
const loadingMore = ref(false)
const errorMessage = ref('')

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
    const data = await getPublicFeed({
      cursor: append ? nextCursor.value : '',
      limit: FEED_PAGE_SIZE,
    })
    const newVideos = Array.isArray(data?.videos) ? data.videos : []

    videos.value = append ? [...videos.value, ...newVideos] : newVideos
    nextCursor.value = typeof data?.next_cursor === 'string' ? data.next_cursor : ''
    hasMore.value = Boolean(data?.has_more && nextCursor.value)
  } catch (error) {
    errorMessage.value = error.message
  } finally {
    initialLoading.value = false
    loadingMore.value = false
  }
}

onMounted(() => loadFeed())
</script>

<template>
  <section class="feed-page">
    <header class="feed-heading">
      <div>
        <p class="eyebrow">PUBLIC FEED</p>
        <h1>正在发生</h1>
        <p class="feed-heading__description">浏览社区最新发布的视频，发现每一个正在被记录的瞬间。</p>
      </div>
      <div class="feed-heading__status" aria-live="polite">
        <span>按发布时间排序</span>
        <strong v-if="videos.length">已加载 {{ videos.length }} 条</strong>
      </div>
    </header>

    <div v-if="initialLoading" class="feed-video-grid" aria-label="正在加载公共视频流" aria-busy="true">
      <div v-for="index in FEED_PAGE_SIZE" :key="index" class="feed-skeleton">
        <div class="feed-skeleton__cover"></div>
        <div class="feed-skeleton__line feed-skeleton__line--short"></div>
        <div class="feed-skeleton__line"></div>
        <div class="feed-skeleton__line"></div>
      </div>
    </div>

    <div v-else-if="errorMessage && videos.length === 0" class="feed-state" role="alert">
      <span class="feed-state__mark">!</span>
      <h2>视频流加载失败</h2>
      <p>{{ errorMessage }}</p>
      <button type="button" @click="loadFeed()">重新加载</button>
    </div>

    <div v-else-if="videos.length === 0" class="feed-state">
      <span class="feed-state__mark">0</span>
      <h2>还没有公开视频</h2>
      <p>发布第一条视频后，它会出现在这里。</p>
    </div>

    <template v-else>
      <div class="feed-video-grid">
        <VideoCard v-for="video in videos" :key="video.id" :video="video" />
      </div>

      <div class="feed-pagination" aria-live="polite">
        <p v-if="errorMessage" class="feed-pagination__error" role="alert">{{ errorMessage }}</p>
        <button v-if="hasMore" type="button" :disabled="loadingMore" @click="loadFeed({ append: true })">
          {{ loadingMore ? '正在加载…' : '加载更多' }}
        </button>
        <p v-else class="feed-pagination__end">已经看到全部视频</p>
      </div>
    </template>
  </section>
</template>
