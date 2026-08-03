<script setup>
import { onMounted, ref } from 'vue'
import { getHotFeed } from '../api/feed'
import VideoCard from '../components/video/VideoCard.vue'

const videos = ref([])
const loading = ref(true)
const errorMessage = ref('')

async function loadHotVideos() {
  loading.value = true
  errorMessage.value = ''

  try {
    const data = await getHotFeed()
    videos.value = Array.isArray(data?.videos) ? data.videos.slice(0, 10) : []
  } catch (error) {
    videos.value = []
    errorMessage.value = error.message
  } finally {
    loading.value = false
  }
}

onMounted(loadHotVideos)
</script>

<template>
  <section class="hot-page">
    <header class="collection-heading">
      <div>
        <p class="eyebrow">REDIS TOP 10</p>
        <h1>全站热门</h1>
        <p>根据点赞和评论产生的热度排序，展示当前全站最热门的十条视频。</p>
      </div>
      <span v-if="videos.length">当前 {{ videos.length }} 条</span>
    </header>

    <div v-if="loading" class="collection-video-grid" aria-label="正在加载热门视频" aria-busy="true">
      <div v-for="index in 3" :key="index" class="feed-skeleton">
        <div class="feed-skeleton__cover"></div>
        <div class="feed-skeleton__line feed-skeleton__line--short"></div>
        <div class="feed-skeleton__line"></div>
      </div>
    </div>

    <div v-else-if="errorMessage" class="collection-state" role="alert">
      <strong>热门榜暂时不可用</strong>
      <p>{{ errorMessage }}</p>
      <button type="button" @click="loadHotVideos">重新加载</button>
    </div>

    <div v-else-if="videos.length === 0" class="collection-state">
      <strong>暂时没有热门视频</strong>
      <p>视频获得点赞或评论后，会按照 Redis 中的热度顺序出现在这里。</p>
    </div>

    <div v-else class="collection-video-grid">
      <VideoCard
        v-for="(video, index) in videos"
        :key="video.id"
        :video="video"
        :rank="index + 1"
      />
    </div>
  </section>
</template>
