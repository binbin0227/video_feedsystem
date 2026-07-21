<script setup>
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getMediaUrl } from '../../utils/media'

const props = defineProps({
  video: {
    type: Object,
    required: true,
  },
})

const coverFailed = ref(false)
const coverSrc = computed(() => getMediaUrl(props.video.cover_url))
const authorLabel = computed(() => props.video.author_username || `作者 ${props.video.author_id}`)
const authorTitle = computed(() => {
  if (props.video.author_username) {
    return `作者：${props.video.author_username}（ID：${props.video.author_id}）`
  }

  return `作者 ID：${props.video.author_id}`
})

function formatPublishedAt(value) {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return '发布时间未知'
  }

  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}
</script>

<template>
  <RouterLink class="feed-video-card" :to="{ name: 'video-detail', params: { videoId: video.id } }">
    <div class="feed-video-card__cover">
      <img v-if="coverSrc && !coverFailed" :src="coverSrc" :alt="`${video.title}的视频封面`" loading="lazy" @error="coverFailed = true">
      <div v-else class="feed-video-card__fallback" aria-hidden="true">
        <span>FRAMEFLOW</span>
        <strong>暂无封面</strong>
      </div>
      <span class="feed-video-card__type">VIDEO</span>
      <span class="feed-video-card__likes" aria-label="点赞数">
        <span aria-hidden="true">♥</span>
        {{ video.like_count }}
      </span>
    </div>

    <div class="feed-video-card__body">
      <div class="feed-video-card__meta">
        <span class="feed-video-card__author" :title="authorTitle">{{ authorLabel }}</span>
        <time :datetime="video.created_at">{{ formatPublishedAt(video.created_at) }}</time>
      </div>
      <h2>{{ video.title }}</h2>
      <p>{{ video.description || '这个视频暂时没有描述。' }}</p>
      <div class="feed-video-card__footer">
        <span>查看视频</span>
        <span aria-hidden="true">→</span>
      </div>
    </div>
  </RouterLink>
</template>
