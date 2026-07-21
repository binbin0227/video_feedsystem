<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getLikedVideos } from '../api/video'
import VideoCard from '../components/video/VideoCard.vue'

const PAGE_SIZE = 6

const route = useRoute()
const router = useRouter()
const videos = ref([])
const nextCursor = ref('')
const hasMore = ref(false)
const initialLoading = ref(true)
const loadingMore = ref(false)
const errorMessage = ref('')

async function loadVideos({ append = false } = {}) {
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
    const data = await getLikedVideos({
      cursor: append ? nextCursor.value : '',
      limit: PAGE_SIZE,
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

onMounted(() => loadVideos())
</script>

<template>
  <section class="liked-page">
    <header class="collection-heading">
      <div>
        <p class="eyebrow">LIKED VIDEOS</p>
        <h1>我点赞的视频</h1>
        <p>按最近点赞时间查看当前账号喜欢过的作品。</p>
      </div>
      <span v-if="videos.length">已加载 {{ videos.length }} 条</span>
    </header>

    <div v-if="initialLoading" class="collection-video-grid" aria-busy="true">
      <div v-for="index in 3" :key="index" class="feed-skeleton">
        <div class="feed-skeleton__cover"></div>
        <div class="feed-skeleton__line feed-skeleton__line--short"></div>
        <div class="feed-skeleton__line"></div>
      </div>
    </div>

    <div v-else-if="errorMessage && videos.length === 0" class="collection-state" role="alert">
      <strong>点赞列表加载失败</strong>
      <p>{{ errorMessage }}</p>
      <button type="button" @click="loadVideos()">重新加载</button>
    </div>

    <div v-else-if="videos.length === 0" class="collection-state">
      <strong>还没有点赞视频</strong>
      <p>在视频详情页点赞后，作品会出现在这里。</p>
    </div>

    <template v-else>
      <div class="collection-video-grid">
        <VideoCard v-for="video in videos" :key="video.id" :video="video" />
      </div>

      <div class="collection-pagination">
        <p v-if="errorMessage" role="alert">{{ errorMessage }}</p>
        <button v-if="hasMore" type="button" :disabled="loadingMore" @click="loadVideos({ append: true })">
          {{ loadingMore ? '正在加载…' : '加载更多视频' }}
        </button>
        <span v-else>已经看到全部点赞视频</span>
      </div>
    </template>
  </section>
</template>
