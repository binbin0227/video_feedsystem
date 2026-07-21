<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import PagePlaceholder from '../components/common/PagePlaceholder.vue'

const route = useRoute()
const description = computed(() => `已进入视频 ${route.params.videoId} 的详情路由。播放、点赞和评论功能将在后续阶段接入。`)
</script>

<template>
  <PagePlaceholder eyebrow="视频详情" title="播放、点赞与评论" :description="description" />
</template>
