<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getAccountProfile } from '../api/account'
import { followAccount, getFollowStatus, unfollowAccount } from '../api/social'
import { getVideosByAuthor } from '../api/video'
import VideoCard from '../components/video/VideoCard.vue'
import { getAccountId } from '../utils/auth'

const VIDEO_PAGE_SIZE = 6

const route = useRoute()
const router = useRouter()
const profile = ref(null)
const profileLoading = ref(true)
const profileError = ref('')
const videos = ref([])
const videosLoading = ref(true)
const videosLoadingMore = ref(false)
const videosError = ref('')
const videosNextCursor = ref('')
const videosHasMore = ref(false)
const currentAccountId = ref('')
const following = ref(false)
const followStatusLoading = ref(false)
const followSubmitting = ref(false)
const followError = ref('')

const accountId = computed(() => String(route.params.accountId || ''))
const isSelf = computed(() => Boolean(currentAccountId.value && currentAccountId.value === accountId.value))
const loggedIn = computed(() => Boolean(currentAccountId.value))
const displayedVideos = computed(() => videos.value.map((video) => ({
  ...video,
  author_username: profile.value?.username || video.author_username,
})))

function formatDate(value) {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return '加入时间未知'
  }

  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  }).format(date)
}

function formatCount(value) {
  return Number(value || 0).toLocaleString('zh-CN')
}

function goToLogin() {
  return router.push({ name: 'login', query: { redirect: route.fullPath } })
}

async function loadProfile() {
  profileLoading.value = true
  profileError.value = ''

  try {
    const data = await getAccountProfile(accountId.value)

    if (!data?.profile) {
      throw new Error('用户主页响应中没有 profile 字段')
    }

    profile.value = data.profile
  } catch (error) {
    profileError.value = error.message
  } finally {
    profileLoading.value = false
  }
}

async function loadVideos({ append = false } = {}) {
  if (append) {
    if (videosLoadingMore.value || !videosHasMore.value) {
      return
    }
    videosLoadingMore.value = true
  } else {
    videosLoading.value = true
  }

  videosError.value = ''

  try {
    const data = await getVideosByAuthor(accountId.value, {
      cursor: append ? videosNextCursor.value : '',
      limit: VIDEO_PAGE_SIZE,
    })
    const newVideos = Array.isArray(data?.videos) ? data.videos : []

    videos.value = append ? [...videos.value, ...newVideos] : newVideos
    videosNextCursor.value = typeof data?.next_cursor === 'string' ? data.next_cursor : ''
    videosHasMore.value = Boolean(data?.has_more && videosNextCursor.value)
  } catch (error) {
    videosError.value = error.message
  } finally {
    videosLoading.value = false
    videosLoadingMore.value = false
  }
}

async function loadFollowStatus() {
  following.value = false
  followError.value = ''

  if (!loggedIn.value || isSelf.value) {
    return
  }

  followStatusLoading.value = true

  try {
    const data = await getFollowStatus(accountId.value)
    following.value = Boolean(data?.is_following)
  } catch (error) {
    if (error.response?.status === 401) {
      currentAccountId.value = ''
      return
    }

    followError.value = error.message
  } finally {
    followStatusLoading.value = false
  }
}

async function toggleFollow() {
  if (!loggedIn.value) {
    await goToLogin()
    return
  }
  if (isSelf.value || followSubmitting.value || followStatusLoading.value) {
    return
  }

  followSubmitting.value = true
  followError.value = ''

  try {
    if (following.value) {
      await unfollowAccount(accountId.value)
      following.value = false
      profile.value.follower_count = Math.max(0, Number(profile.value.follower_count || 0) - 1)
    } else {
      await followAccount(accountId.value)
      following.value = true
      profile.value.follower_count = Number(profile.value.follower_count || 0) + 1
    }
  } catch (error) {
    if (error.response?.status === 401) {
      currentAccountId.value = ''
      await goToLogin()
      return
    }

    followError.value = error.message
  } finally {
    followSubmitting.value = false
  }
}

async function loadPage() {
  profile.value = null
  videos.value = []
  videosNextCursor.value = ''
  videosHasMore.value = false
  following.value = false
  followError.value = ''
  currentAccountId.value = getAccountId()

  await loadProfile()

  if (profile.value) {
    await Promise.all([loadVideos(), loadFollowStatus()])
  }
}

watch(accountId, loadPage, { immediate: true })
</script>

<template>
  <section class="profile-page">
    <div v-if="profileLoading" class="profile-state" aria-busy="true">
      <span>···</span>
      <h1>正在加载用户主页</h1>
    </div>

    <div v-else-if="profileError || !profile" class="profile-state" role="alert">
      <span>!</span>
      <h1>用户主页加载失败</h1>
      <p>{{ profileError || '没有找到这个用户。' }}</p>
      <button type="button" @click="loadPage">重新加载</button>
    </div>

    <template v-else>
      <header class="profile-hero">
        <div class="profile-hero__copy">
          <p class="eyebrow">{{ isSelf ? 'MY PROFILE' : 'CREATOR PROFILE' }}</p>
          <h1>{{ profile.username }}</h1>
          <div class="profile-identity">
            <span>ID {{ profile.account_id }}</span>
            <span>{{ formatDate(profile.created_at) }} 加入</span>
          </div>
        </div>

        <div class="profile-hero__action">
          <span v-if="isSelf" class="profile-self-badge">这是你的主页</span>
          <button v-else type="button" :class="{ following }" :disabled="followSubmitting || followStatusLoading" @click="toggleFollow">
            {{ !loggedIn ? '登录后关注' : followStatusLoading ? '查询中…' : followSubmitting ? '处理中…' : following ? '已关注 · 取消' : '关注' }}
          </button>
          <p v-if="followError" role="alert">{{ followError }}</p>
        </div>

        <dl class="profile-stats">
          <div><dt>作品</dt><dd>{{ formatCount(profile.video_count) }}</dd></div>
          <div><dt>获赞</dt><dd>{{ formatCount(profile.received_like_count) }}</dd></div>
          <div><dt>关注</dt><dd>{{ formatCount(profile.following_count) }}</dd></div>
          <div><dt>粉丝</dt><dd>{{ formatCount(profile.follower_count) }}</dd></div>
        </dl>
      </header>

      <section class="profile-works" aria-labelledby="profile-works-title">
        <header>
          <div>
            <p class="eyebrow">PUBLISHED WORKS</p>
            <h2 id="profile-works-title">作者作品</h2>
          </div>
          <span>已加载 {{ videos.length }} 条</span>
        </header>

        <div v-if="videosLoading" class="profile-video-grid" aria-busy="true">
          <div v-for="index in 3" :key="index" class="feed-skeleton">
            <div class="feed-skeleton__cover"></div>
            <div class="feed-skeleton__line feed-skeleton__line--short"></div>
            <div class="feed-skeleton__line"></div>
          </div>
        </div>

        <div v-else-if="videosError && videos.length === 0" class="profile-works-state" role="alert">
          <strong>作品加载失败</strong>
          <p>{{ videosError }}</p>
          <button type="button" @click="loadVideos()">重新加载</button>
        </div>

        <div v-else-if="videos.length === 0" class="profile-works-state">
          <strong>还没有发布作品</strong>
          <p>作者发布视频后会显示在这里。</p>
        </div>

        <template v-else>
          <div class="profile-video-grid">
            <VideoCard v-for="video in displayedVideos" :key="video.id" :video="video" />
          </div>
          <div class="profile-pagination">
            <p v-if="videosError" role="alert">{{ videosError }}</p>
            <button v-if="videosHasMore" type="button" :disabled="videosLoadingMore" @click="loadVideos({ append: true })">
              {{ videosLoadingMore ? '正在加载…' : '加载更多作品' }}
            </button>
            <span v-else>已经看到全部作品</span>
          </div>
        </template>
      </section>
    </template>
  </section>
</template>
