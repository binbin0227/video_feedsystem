<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { followAccount, getFollowStatus, unfollowAccount } from '../../api/social'
import { getVideoLikeStatus, likeVideo, unlikeVideo } from '../../api/video'
import { getAccountId } from '../../utils/auth'
import { getMediaUrl } from '../../utils/media'
import FeedCommentsPanel from './FeedCommentsPanel.vue'

const props = defineProps({
  video: {
    type: Object,
    required: true,
  },
  soundEnabled: {
    type: Boolean,
    default: false,
  },
  volume: {
    type: Number,
    default: 0.3,
  },
  shouldPreload: {
    type: Boolean,
    default: false,
  },
  position: {
    type: Number,
    required: true,
  },
  total: {
    type: Number,
    required: true,
  },
})

const emit = defineEmits(['active', 'toggle-sound', 'update-volume'])

const route = useRoute()
const router = useRouter()
const slideElement = ref(null)
const videoElement = ref(null)
const isActive = ref(false)
const isPaused = ref(true)
const firstFrameReady = ref(false)
const mediaError = ref('')
const progress = ref(0)
const currentAccountId = ref(getAccountId())
const liked = ref(false)
const likeCount = ref(Number(props.video.like_count || 0))
const likeStatusLoading = ref(false)
const likeSubmitting = ref(false)
const followingAuthor = ref(false)
const followStatusLoading = ref(false)
const followSubmitting = ref(false)
const interactionStateLoaded = ref(false)
const interactionError = ref('')
const commentsOpen = ref(false)
const volumeSliderOpen = ref(false)
let observer
let frameRevealAnimationId
let videoFrameCallbackId
let volumeSliderTimer

const videoSrc = computed(() => getMediaUrl(props.video.play_url))
const normalizedVolume = computed(() => Math.min(1, Math.max(0, Number(props.volume) || 0)))
const volumePercent = computed(() => Math.round(normalizedVolume.value * 100))
const authorName = computed(() => props.video.author_username || `作者 ${props.video.author_id}`)
const authorInitial = computed(() => authorName.value.trim().charAt(0).toUpperCase() || '创')
const loggedIn = computed(() => Boolean(currentAccountId.value))
const authorId = computed(() => String(props.video.author_id || ''))
const isOwnVideo = computed(() => Boolean(currentAccountId.value && currentAccountId.value === authorId.value))

function formatPublishedAt(value) {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return '发布时间未知'
  }

  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(date)
}

function getPlayerVolume(volume) {
  // 手机浏览器交给系统媒体音量控制；桌面浏览器才应用页面内的音量百分比。
  return window.matchMedia('(hover: none), (pointer: coarse)').matches ? 1 : volume
}

async function playVideo(volume = normalizedVolume.value) {
  if (!videoElement.value || !videoSrc.value) {
    return
  }

  videoElement.value.volume = getPlayerVolume(volume)

  try {
    await videoElement.value.play()
    isPaused.value = false
    mediaError.value = ''
  } catch (error) {
    // 浏览器可能阻止自动有声播放；回退为静音后仍然可以自动播放。
    if (!videoElement.value.muted) {
      videoElement.value.muted = true
      emit('toggle-sound', false)

      try {
        await videoElement.value.play()
        isPaused.value = false
        return
      } catch {
        // 继续使用下面的统一提示。
      }
    }

    isPaused.value = true
    mediaError.value = '自动播放被浏览器拦截，请点击视频开始播放'
  }
}

function pauseVideo() {
  videoElement.value?.pause()
  isPaused.value = true
}

function prepareVideo() {
  if (!videoElement.value) {
    return
  }

  videoElement.value.volume = getPlayerVolume(normalizedVolume.value)
  videoElement.value.muted = !props.soundEnabled || normalizedVolume.value === 0

  if (isActive.value) {
    playVideo()
  }
}

function handleLoadStart() {
  cancelFrameReveal()
  firstFrameReady.value = false
}

function cancelFrameReveal() {
  if (videoFrameCallbackId !== undefined && videoElement.value?.cancelVideoFrameCallback) {
    videoElement.value.cancelVideoFrameCallback(videoFrameCallbackId)
  }

  videoFrameCallbackId = undefined
  window.cancelAnimationFrame(frameRevealAnimationId)
}

function handlePlaying() {
  isPaused.value = false
  cancelFrameReveal()

  if (videoElement.value?.requestVideoFrameCallback) {
    // 只有视频帧已经提交给浏览器合成器后才撤下封面，playing 事件本身还不能保证画面已经绘制。
    videoFrameCallbackId = videoElement.value.requestVideoFrameCallback(() => {
      videoFrameCallbackId = undefined
      firstFrameReady.value = true
    })
    return
  }

  // 兼容没有 requestVideoFrameCallback 的旧浏览器，至少跨过两次绘制再切换画面。
  frameRevealAnimationId = window.requestAnimationFrame(() => {
    frameRevealAnimationId = window.requestAnimationFrame(() => {
      firstFrameReady.value = true
    })
  })
}

function togglePlayback() {
  if (!videoElement.value) {
    return
  }

  if (videoElement.value.paused) {
    playVideo()
  } else {
    pauseVideo()
  }
}

function toggleSound() {
  const enableSound = !props.soundEnabled
  const nextVolume = enableSound && normalizedVolume.value === 0 ? 0.3 : normalizedVolume.value

  if (nextVolume !== normalizedVolume.value) {
    emit('update-volume', nextVolume)
  }

  emit('toggle-sound', enableSound)

  if (enableSound) {
    openVolumeSlider()
  } else {
    hideVolumeSlider()
  }

  if (!videoElement.value) {
    return
  }

  videoElement.value.volume = getPlayerVolume(nextVolume)
  videoElement.value.muted = !enableSound

  if (isActive.value) {
    playVideo(nextVolume)
  }
}

function hideVolumeSlider() {
  window.clearTimeout(volumeSliderTimer)
  volumeSliderOpen.value = false
}

function openVolumeSlider() {
  window.clearTimeout(volumeSliderTimer)
  volumeSliderOpen.value = true
}

function scheduleVolumeSliderClose() {
  window.clearTimeout(volumeSliderTimer)
  volumeSliderTimer = window.setTimeout(() => {
    volumeSliderOpen.value = false
  }, 1100)
}

function updateVolume(event) {
  const nextVolume = Math.min(1, Math.max(0, Number(event.target.value) / 100))
  const enableSound = nextVolume > 0

  emit('update-volume', nextVolume)
  emit('toggle-sound', enableSound)
  openVolumeSlider()

  if (!videoElement.value) {
    return
  }

  videoElement.value.volume = getPlayerVolume(nextVolume)
  videoElement.value.muted = !enableSound

  // 拖动滑块属于用户手势；如果当前视频暂停，可以借此机会直接恢复播放。
  if (isActive.value && videoElement.value.paused && enableSound) {
    playVideo(nextVolume)
  }
}

function updateProgress() {
  const player = videoElement.value

  if (!player || !Number.isFinite(player.duration) || player.duration <= 0) {
    progress.value = 0
    return
  }

  progress.value = Math.min(100, (player.currentTime / player.duration) * 100)
}

function handleMediaError() {
  mediaError.value = '视频文件加载失败，可以进入详情页重试'
  isPaused.value = true
}

function goToLogin() {
  return router.push({
    name: 'login',
    query: { redirect: route.fullPath },
  })
}

function handleUnauthorized(error) {
  if (error.response?.status !== 401) {
    return false
  }

  currentAccountId.value = ''
  liked.value = false
  followingAuthor.value = false
  commentsOpen.value = false
  goToLogin()
  return true
}

async function loadInteractionState() {
  if (!loggedIn.value || interactionStateLoaded.value || likeStatusLoading.value || followStatusLoading.value) {
    return
  }

  likeStatusLoading.value = true
  followStatusLoading.value = !isOwnVideo.value
  interactionError.value = ''

  try {
    const requests = [getVideoLikeStatus(props.video.id)]

    if (!isOwnVideo.value) {
      requests.push(getFollowStatus(authorId.value))
    }

    const [likeData, followData] = await Promise.all(requests)
    liked.value = Boolean(likeData?.is_liked)
    followingAuthor.value = Boolean(followData?.is_following)
    interactionStateLoaded.value = true
  } catch (error) {
    if (!handleUnauthorized(error)) {
      interactionError.value = error.message
    }
  } finally {
    likeStatusLoading.value = false
    followStatusLoading.value = false
  }
}

async function toggleLike() {
  if (!loggedIn.value) {
    await goToLogin()
    return
  }
  if (likeSubmitting.value || likeStatusLoading.value) {
    return
  }

  likeSubmitting.value = true
  interactionError.value = ''

  try {
    if (liked.value) {
      await unlikeVideo(props.video.id)
      liked.value = false
      likeCount.value = Math.max(0, likeCount.value - 1)
    } else {
      await likeVideo(props.video.id)
      liked.value = true
      likeCount.value += 1
    }
  } catch (error) {
    if (!handleUnauthorized(error)) {
      interactionError.value = error.message
    }
  } finally {
    likeSubmitting.value = false
  }
}

async function toggleFollow() {
  if (!loggedIn.value) {
    await goToLogin()
    return
  }
  if (isOwnVideo.value || followSubmitting.value || followStatusLoading.value) {
    return
  }

  followSubmitting.value = true
  interactionError.value = ''

  try {
    if (followingAuthor.value) {
      await unfollowAccount(authorId.value)
      followingAuthor.value = false
    } else {
      await followAccount(authorId.value)
      followingAuthor.value = true
    }
  } catch (error) {
    if (!handleUnauthorized(error)) {
      interactionError.value = error.message
    }
  } finally {
    followSubmitting.value = false
  }
}

function handleCommentsUnauthorized() {
  currentAccountId.value = ''
  liked.value = false
  followingAuthor.value = false
}

watch([() => props.soundEnabled, () => props.volume], async ([enabled]) => {
  await nextTick()

  if (!videoElement.value) {
    return
  }

  videoElement.value.volume = getPlayerVolume(normalizedVolume.value)
  videoElement.value.muted = !enabled || normalizedVolume.value === 0
})

watch(() => props.video.like_count, (value) => {
  likeCount.value = Number(value || 0)
})

onMounted(() => {
  observer = new IntersectionObserver(([entry]) => {
    const active = entry.isIntersecting && entry.intersectionRatio >= 0.65

    if (active === isActive.value) {
      return
    }

    isActive.value = active

    if (active) {
      emit('active')
      loadInteractionState()
      playVideo()
    } else {
      commentsOpen.value = false
      pauseVideo()
    }
  }, {
    threshold: [0, 0.65, 1],
  })

  if (slideElement.value) {
    observer.observe(slideElement.value)
  }
})

onBeforeUnmount(() => {
  observer?.disconnect()
  cancelFrameReveal()
  window.clearTimeout(volumeSliderTimer)
  videoElement.value?.pause()
})
</script>

<template>
  <article ref="slideElement" class="immersive-video-slide" :aria-label="`第 ${position} 条视频：${video.title}`">
    <div class="immersive-video-stage">
      <video
        v-if="videoSrc"
        ref="videoElement"
        class="immersive-video-player"
        :class="{ 'is-ready': firstFrameReady }"
        :src="videoSrc"
        :muted="!soundEnabled || normalizedVolume === 0"
        loop
        playsinline
        :preload="shouldPreload ? 'auto' : 'metadata'"
        tabindex="0"
        @click="togglePlayback"
        @keydown.space.prevent="togglePlayback"
        @loadstart="handleLoadStart"
        @loadedmetadata="prepareVideo"
        @play="isPaused = false"
        @playing="handlePlaying"
        @pause="isPaused = true"
        @timeupdate="updateProgress"
        @error="handleMediaError"
      >
        当前浏览器不支持视频播放。
      </video>

      <div v-else class="immersive-video-unavailable">
        <strong>播放地址不可用</strong>
        <span>可以进入详情页查看视频信息</span>
      </div>

      <button
        v-if="videoSrc && isPaused"
        class="immersive-video-play"
        type="button"
        aria-label="播放视频"
        @click="togglePlayback"
      ></button>

      <div class="immersive-video-scrim" aria-hidden="true"></div>

      <div class="immersive-video-copy">
        <div class="immersive-video-author-line">
          <RouterLink :to="{ name: 'user-profile', params: { accountId: video.author_id } }">
            @{{ authorName }}
          </RouterLink>
          <time :datetime="video.created_at">{{ formatPublishedAt(video.created_at) }}</time>
        </div>
        <h1>{{ video.title }}</h1>
        <p>{{ video.description || '这个视频暂时没有描述。' }}</p>
        <span v-if="mediaError" class="immersive-video-error" role="alert">{{ mediaError }}</span>
        <span v-if="interactionError" class="immersive-video-error" role="alert">{{ interactionError }}</span>
      </div>

      <aside class="immersive-video-actions" aria-label="视频操作">
        <button
          v-if="!isOwnVideo"
          class="immersive-video-follow"
          :class="{ 'is-following': followingAuthor }"
          type="button"
          :aria-pressed="followingAuthor"
          :disabled="followSubmitting || followStatusLoading"
          @click="toggleFollow"
        >
          <strong>{{ authorInitial }}</strong>
          <span>{{ !loggedIn ? '关注' : followStatusLoading ? '查询中' : followSubmitting ? '处理中' : followingAuthor ? '已关注' : '关注' }}</span>
        </button>

        <div
          class="immersive-volume-control"
          :class="{ 'is-slider-open': volumeSliderOpen }"
          @pointerenter="openVolumeSlider"
          @pointerleave="scheduleVolumeSliderClose"
        >
          <button
            type="button"
            :aria-label="soundEnabled ? '关闭声音' : '开启声音'"
            @click="toggleSound"
            @focus="openVolumeSlider"
            @blur="scheduleVolumeSliderClose"
          >
            <strong>
              <span class="immersive-volume-percent">{{ soundEnabled ? `${volumePercent}%` : '静音' }}</span>
              <span class="immersive-volume-system">{{ soundEnabled ? '开声' : '静音' }}</span>
            </strong>
            <span class="immersive-volume-caption">{{ soundEnabled ? '音量' : '开声音' }}</span>
            <span class="immersive-volume-system">系统音量</span>
          </button>
          <label class="immersive-volume-slider" :style="{ '--volume-level': `${volumePercent}%` }">
            <span class="immersive-volume-slider__heading">
              <span>播放音量</span>
              <output>{{ volumePercent }}%</output>
            </span>
            <input
              type="range"
              min="0"
              max="100"
              step="1"
              :value="volumePercent"
              :aria-valuetext="`${volumePercent}%`"
              @focus="openVolumeSlider"
              @blur="scheduleVolumeSliderClose"
              @input="updateVolume"
            >
          </label>
        </div>

        <button
          class="immersive-video-action immersive-video-action--likes"
          :class="{ 'is-liked': liked }"
          type="button"
          :aria-label="liked ? '取消点赞' : '点赞'"
          :aria-pressed="liked"
          :disabled="likeSubmitting || likeStatusLoading"
          @click="toggleLike"
        >
          <strong aria-hidden="true">♥</strong>
          <span>{{ likeCount }}</span>
        </button>

        <button class="immersive-video-action immersive-video-action--comments" type="button" aria-label="查看评论" @click="commentsOpen = true">
          <strong aria-hidden="true">评</strong>
          <span>评论</span>
        </button>

        <RouterLink class="immersive-video-action" :to="{ name: 'video-detail', params: { videoId: video.id } }">
          <strong aria-hidden="true">↗</strong>
          <span>详情</span>
        </RouterLink>
      </aside>

      <button v-if="commentsOpen" class="feed-comments-backdrop" type="button" aria-label="关闭评论" @click="commentsOpen = false"></button>
      <FeedCommentsPanel
        v-if="commentsOpen"
        :video-id="video.id"
        :current-account-id="currentAccountId"
        @close="commentsOpen = false"
        @unauthorized="handleCommentsUnauthorized"
      />

      <div class="immersive-video-progress" aria-hidden="true">
        <i :style="{ width: `${progress}%` }"></i>
      </div>

      <span class="immersive-video-position">{{ position }} / {{ total }}</span>
    </div>
  </article>
</template>
