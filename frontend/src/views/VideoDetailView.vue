<script setup>
import { computed, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { deleteComment, getCommentList, publishComment } from '../api/comment'
import { followAccount, getFollowStatus, unfollowAccount } from '../api/social'
import { getVideoDetail, getVideoLikeStatus, likeVideo, unlikeVideo } from '../api/video'
import { getAccountId } from '../utils/auth'
import { getMediaUrl } from '../utils/media'

const COMMENT_PAGE_SIZE = 8

const route = useRoute()
const router = useRouter()
const video = ref(null)
const pageLoading = ref(true)
const pageError = ref('')
const liked = ref(false)
const likeStatusLoading = ref(false)
const likeSubmitting = ref(false)
const followingAuthor = ref(false)
const followStatusLoading = ref(false)
const followSubmitting = ref(false)
const interactionError = ref('')
const currentAccountId = ref('')
const comments = ref([])
const commentsLoading = ref(true)
const commentsLoadingMore = ref(false)
const commentsError = ref('')
const commentsNextCursor = ref('')
const commentsHasMore = ref(false)
const commentDraft = ref('')
const commentSubmitting = ref(false)
const deletingCommentId = ref('')

const videoId = computed(() => String(route.params.videoId || ''))
const videoSrc = computed(() => getMediaUrl(video.value?.play_url))
const coverSrc = computed(() => getMediaUrl(video.value?.cover_url))
const loggedIn = computed(() => Boolean(currentAccountId.value))
const authorId = computed(() => String(video.value?.author_id || ''))
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
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

function preparePlayer(event) {
  // 桌面浏览器使用 30% 初始音量；手机浏览器保留原生音量并交给系统媒体音量控制。
  if (!window.matchMedia('(hover: none), (pointer: coarse)').matches) {
    event.currentTarget.volume = 0.3
  }
  event.currentTarget.muted = false
}

function isOwnComment(comment) {
  return Boolean(currentAccountId.value && String(comment.account_id) === currentAccountId.value)
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
  goToLogin()
  return true
}

async function loadVideo() {
  pageLoading.value = true
  pageError.value = ''

  try {
    const data = await getVideoDetail(videoId.value)

    if (!data?.video) {
      throw new Error('视频详情响应中没有 video 字段')
    }

    video.value = data.video
  } catch (error) {
    pageError.value = error.message
  } finally {
    pageLoading.value = false
  }
}

async function loadLikeStatus() {
  liked.value = false

  if (!currentAccountId.value) {
    return
  }

  likeStatusLoading.value = true

  try {
    const data = await getVideoLikeStatus(videoId.value)
    liked.value = Boolean(data?.is_liked)
  } catch (error) {
    if (!handleUnauthorized(error)) {
      interactionError.value = error.message
    }
  } finally {
    likeStatusLoading.value = false
  }
}

async function loadFollowStatus() {
  followingAuthor.value = false

  if (!loggedIn.value || !authorId.value || isOwnVideo.value) {
    return
  }

  followStatusLoading.value = true

  try {
    const data = await getFollowStatus(authorId.value)
    followingAuthor.value = Boolean(data?.is_following)
  } catch (error) {
    if (!handleUnauthorized(error)) {
      interactionError.value = error.message
    }
  } finally {
    followStatusLoading.value = false
  }
}

async function loadComments({ append = false } = {}) {
  if (append) {
    if (commentsLoadingMore.value || !commentsHasMore.value) {
      return
    }
    commentsLoadingMore.value = true
  } else {
    commentsLoading.value = true
  }

  commentsError.value = ''

  try {
    const data = await getCommentList(videoId.value, {
      cursor: append ? commentsNextCursor.value : '',
      limit: COMMENT_PAGE_SIZE,
    })
    const newComments = Array.isArray(data?.comments) ? data.comments : []

    comments.value = append ? [...comments.value, ...newComments] : newComments
    commentsNextCursor.value = typeof data?.next_cursor === 'string' ? data.next_cursor : ''
    commentsHasMore.value = Boolean(data?.has_more && commentsNextCursor.value)
  } catch (error) {
    commentsError.value = error.message
  } finally {
    commentsLoading.value = false
    commentsLoadingMore.value = false
  }
}

async function loadPage() {
  video.value = null
  liked.value = false
  followingAuthor.value = false
  interactionError.value = ''
  comments.value = []
  commentsNextCursor.value = ''
  commentsHasMore.value = false
  commentsError.value = ''
  commentDraft.value = ''
  currentAccountId.value = getAccountId()

  await Promise.all([
    loadVideo(),
    loadComments(),
    loadLikeStatus(),
  ])

  if (video.value) {
    await loadFollowStatus()
  }
}

async function toggleLike() {
  if (!loggedIn.value) {
    await goToLogin()
    return
  }
  if (likeSubmitting.value || likeStatusLoading.value || !video.value) {
    return
  }

  likeSubmitting.value = true
  interactionError.value = ''

  try {
    if (liked.value) {
      await unlikeVideo(videoId.value)
      liked.value = false
      video.value.like_count = Math.max(0, Number(video.value.like_count || 0) - 1)
    } else {
      await likeVideo(videoId.value)
      liked.value = true
      video.value.like_count = Number(video.value.like_count || 0) + 1
    }
  } catch (error) {
    if (!handleUnauthorized(error)) {
      interactionError.value = error.message
    }
  } finally {
    likeSubmitting.value = false
  }
}

async function toggleAuthorFollow() {
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

async function handlePublishComment() {
  if (!loggedIn.value) {
    await goToLogin()
    return
  }

  const content = commentDraft.value.trim()
  commentsError.value = ''

  if (!content) {
    commentsError.value = '评论内容不能为空'
    return
  }

  commentSubmitting.value = true

  try {
    const data = await publishComment(videoId.value, content)

    if (!data?.comment) {
      throw new Error('评论发布响应中没有 comment 字段')
    }

    comments.value = [data.comment, ...comments.value]
    commentDraft.value = ''
  } catch (error) {
    if (!handleUnauthorized(error)) {
      commentsError.value = error.message
    }
  } finally {
    commentSubmitting.value = false
  }
}

async function handleDeleteComment(comment) {
  if (!isOwnComment(comment) || deletingCommentId.value) {
    return
  }
  if (!window.confirm('确定删除这条评论吗？')) {
    return
  }

  deletingCommentId.value = String(comment.id)
  commentsError.value = ''

  try {
    await deleteComment(comment.id)
    comments.value = comments.value.filter((item) => item.id !== comment.id)
  } catch (error) {
    if (!handleUnauthorized(error)) {
      commentsError.value = error.message
    }
  } finally {
    deletingCommentId.value = ''
  }
}

watch(videoId, loadPage, { immediate: true })
</script>

<template>
  <section class="video-detail-page">
    <RouterLink class="video-detail-back" to="/">← 返回公共视频流</RouterLink>

    <div v-if="pageLoading" class="video-detail-state" aria-busy="true">
      <span class="video-detail-state__mark">···</span>
      <h1>正在加载视频</h1>
      <p>正在获取播放地址和视频信息。</p>
    </div>

    <div v-else-if="pageError || !video" class="video-detail-state" role="alert">
      <span class="video-detail-state__mark">!</span>
      <h1>视频加载失败</h1>
      <p>{{ pageError || '没有找到这个视频。' }}</p>
      <button type="button" @click="loadPage">重新加载</button>
    </div>

    <div v-else class="video-detail-layout">
      <article class="video-detail-panel">
        <div class="video-player-shell">
          <video v-if="videoSrc" :src="videoSrc" :poster="coverSrc" controls playsinline preload="metadata" @loadedmetadata="preparePlayer">
            当前浏览器不支持视频播放。
          </video>
          <div v-else class="video-player-empty">视频播放地址不可用</div>
        </div>

        <div class="video-detail-copy">
          <div class="video-detail-meta">
            <div class="video-detail-author">
              <RouterLink :to="{ name: 'user-profile', params: { accountId: video.author_id } }">
                作者 {{ video.author_username || video.author_id }}
              </RouterLink>
              <button
                v-if="!isOwnVideo"
                type="button"
                :class="{ following: followingAuthor }"
                :disabled="followSubmitting || followStatusLoading"
                @click="toggleAuthorFollow"
              >
                {{ !loggedIn ? '关注作者' : followStatusLoading ? '查询中…' : followSubmitting ? '处理中…' : followingAuthor ? '已关注' : '关注作者' }}
              </button>
            </div>
            <time :datetime="video.created_at">{{ formatPublishedAt(video.created_at) }}</time>
          </div>

          <h1>{{ video.title }}</h1>
          <p class="video-detail-description">{{ video.description || '这个视频暂时没有描述。' }}</p>

          <div class="video-detail-actions">
            <button
              class="video-like-button"
              :class="{ 'video-like-button--active': liked }"
              type="button"
              :aria-pressed="liked"
              :disabled="likeSubmitting || likeStatusLoading"
              @click="toggleLike"
            >
              <span aria-hidden="true">♥</span>
              <strong>{{ video.like_count }}</strong>
              {{ !loggedIn ? '登录后点赞' : liked ? '取消点赞' : '点赞' }}
            </button>
            <p v-if="interactionError" class="video-interaction-error" role="alert">{{ interactionError }}</p>
          </div>
        </div>
      </article>

      <aside class="comments-panel" aria-labelledby="comments-title">
        <header class="comments-header">
          <div>
            <p class="eyebrow">COMMENTS</p>
            <h2 id="comments-title">评论</h2>
          </div>
          <span>已加载 {{ comments.length }} 条</span>
        </header>

        <form v-if="loggedIn" class="comment-composer" @submit.prevent="handlePublishComment">
          <label for="comment-content">说点什么</label>
          <textarea
            id="comment-content"
            v-model="commentDraft"
            name="content"
            rows="3"
            maxlength="500"
            placeholder="写下你的评论…"
            :disabled="commentSubmitting"
            @input="commentsError = ''"
          ></textarea>
          <div>
            <span>{{ commentDraft.length }}/500</span>
            <button type="submit" :disabled="commentSubmitting || !commentDraft.trim()">
              {{ commentSubmitting ? '发布中…' : '发布评论' }}
            </button>
          </div>
        </form>

        <div v-else class="comment-login-prompt">
          <p>登录后可以参与评论和点赞。</p>
          <button type="button" @click="goToLogin">去登录</button>
        </div>

        <p v-if="commentsError" class="comments-error" role="alert">{{ commentsError }}</p>

        <div v-if="commentsLoading" class="comments-loading" aria-busy="true">正在加载评论…</div>

        <div v-else-if="comments.length === 0 && !commentsError" class="comments-empty">
          <strong>还没有评论</strong>
          <p>成为第一个留下想法的人。</p>
        </div>

        <template v-else>
          <div class="comment-list">
            <article v-for="comment in comments" :key="comment.id" class="comment-item">
              <header>
                <RouterLink :to="{ name: 'user-profile', params: { accountId: comment.account_id } }">
                  {{ isOwnComment(comment) ? '我' : comment.username || `用户 ${comment.account_id}` }}
                </RouterLink>
                <time :datetime="comment.created_at">{{ formatPublishedAt(comment.created_at) }}</time>
              </header>
              <p>{{ comment.content }}</p>
              <button
                v-if="isOwnComment(comment)"
                class="comment-delete"
                type="button"
                :disabled="deletingCommentId === String(comment.id)"
                @click="handleDeleteComment(comment)"
              >
                {{ deletingCommentId === String(comment.id) ? '删除中…' : '删除' }}
              </button>
            </article>
          </div>

          <button
            v-if="commentsHasMore"
            class="comments-load-more"
            type="button"
            :disabled="commentsLoadingMore"
            @click="loadComments({ append: true })"
          >
            {{ commentsLoadingMore ? '正在加载…' : '加载更多评论' }}
          </button>
        </template>
      </aside>
    </div>
  </section>
</template>
