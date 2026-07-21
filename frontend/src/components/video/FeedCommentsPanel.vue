<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { deleteComment, getCommentList, publishComment } from '../../api/comment'

const COMMENT_PAGE_SIZE = 8

const props = defineProps({
  videoId: {
    type: [String, Number],
    required: true,
  },
  currentAccountId: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['close', 'unauthorized'])

const route = useRoute()
const router = useRouter()
const comments = ref([])
const nextCursor = ref('')
const hasMore = ref(false)
const loading = ref(true)
const loadingMore = ref(false)
const errorMessage = ref('')
const commentDraft = ref('')
const publishing = ref(false)
const deletingCommentId = ref('')

const loggedIn = computed(() => Boolean(props.currentAccountId))

function formatPublishedAt(value) {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return '时间未知'
  }

  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

function isOwnComment(comment) {
  return Boolean(props.currentAccountId && String(comment.account_id) === props.currentAccountId)
}

function goToLogin() {
  emit('close')
  return router.push({
    name: 'login',
    query: { redirect: route.fullPath },
  })
}

function handleUnauthorized(error) {
  if (error.response?.status !== 401) {
    return false
  }

  emit('unauthorized')
  goToLogin()
  return true
}

async function loadComments({ append = false } = {}) {
  if (append) {
    if (loadingMore.value || !hasMore.value) {
      return
    }
    loadingMore.value = true
  } else {
    loading.value = true
  }

  errorMessage.value = ''

  try {
    const data = await getCommentList(props.videoId, {
      cursor: append ? nextCursor.value : '',
      limit: COMMENT_PAGE_SIZE,
    })
    const newComments = Array.isArray(data?.comments) ? data.comments : []

    comments.value = append ? [...comments.value, ...newComments] : newComments
    nextCursor.value = typeof data?.next_cursor === 'string' ? data.next_cursor : ''
    hasMore.value = Boolean(data?.has_more && nextCursor.value)
  } catch (error) {
    errorMessage.value = error.message
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

async function handlePublishComment() {
  if (!loggedIn.value) {
    await goToLogin()
    return
  }

  const content = commentDraft.value.trim()
  errorMessage.value = ''

  if (!content) {
    errorMessage.value = '评论内容不能为空'
    return
  }

  publishing.value = true

  try {
    const data = await publishComment(props.videoId, content)

    if (!data?.comment) {
      throw new Error('评论发布响应中没有 comment 字段')
    }

    comments.value = [data.comment, ...comments.value]
    commentDraft.value = ''
  } catch (error) {
    if (!handleUnauthorized(error)) {
      errorMessage.value = error.message
    }
  } finally {
    publishing.value = false
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
  errorMessage.value = ''

  try {
    await deleteComment(comment.id)
    comments.value = comments.value.filter((item) => item.id !== comment.id)
  } catch (error) {
    if (!handleUnauthorized(error)) {
      errorMessage.value = error.message
    }
  } finally {
    deletingCommentId.value = ''
  }
}

function handleKeydown(event) {
  if (event.key === 'Escape') {
    emit('close')
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
  loadComments()
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <aside class="feed-comments-panel" role="dialog" aria-modal="true" :aria-labelledby="`feed-comments-title-${videoId}`" @wheel.stop>
    <header class="feed-comments-panel__header">
      <div>
        <span>COMMENTS</span>
        <h2 :id="`feed-comments-title-${videoId}`">评论</h2>
      </div>
      <div>
        <span>已加载 {{ comments.length }} 条</span>
        <button type="button" aria-label="关闭评论" @click="emit('close')">×</button>
      </div>
    </header>

    <div class="feed-comments-panel__body">
      <div v-if="loading" class="feed-comments-state" aria-busy="true">正在加载评论…</div>

      <div v-else-if="comments.length === 0 && !errorMessage" class="feed-comments-state">
        <strong>还没有评论</strong>
        <p>成为第一个留下想法的人。</p>
      </div>

      <template v-else>
        <article v-for="comment in comments" :key="comment.id" class="feed-comment-item">
          <header>
            <RouterLink :to="{ name: 'user-profile', params: { accountId: comment.account_id } }">
              {{ isOwnComment(comment) ? '我' : comment.username || `用户 ${comment.account_id}` }}
            </RouterLink>
            <time :datetime="comment.created_at">{{ formatPublishedAt(comment.created_at) }}</time>
          </header>
          <p>{{ comment.content }}</p>
          <button
            v-if="isOwnComment(comment)"
            type="button"
            :disabled="deletingCommentId === String(comment.id)"
            @click="handleDeleteComment(comment)"
          >
            {{ deletingCommentId === String(comment.id) ? '删除中…' : '删除' }}
          </button>
        </article>

        <button
          v-if="hasMore"
          class="feed-comments-load-more"
          type="button"
          :disabled="loadingMore"
          @click="loadComments({ append: true })"
        >
          {{ loadingMore ? '正在加载…' : '加载更多评论' }}
        </button>
      </template>
    </div>

    <p v-if="errorMessage" class="feed-comments-panel__error" role="alert">{{ errorMessage }}</p>

    <form v-if="loggedIn" class="feed-comments-composer" @submit.prevent="handlePublishComment">
      <label :for="`feed-comment-content-${videoId}`">发布评论</label>
      <textarea
        :id="`feed-comment-content-${videoId}`"
        v-model="commentDraft"
        rows="2"
        maxlength="500"
        placeholder="留下你的评论…"
        :disabled="publishing"
        @input="errorMessage = ''"
      ></textarea>
      <div>
        <span>{{ commentDraft.length }}/500</span>
        <button type="submit" :disabled="publishing || !commentDraft.trim()">
          {{ publishing ? '发布中…' : '发布' }}
        </button>
      </div>
    </form>

    <div v-else class="feed-comments-login">
      <span>登录后可以发布评论</span>
      <button type="button" @click="goToLogin">去登录</button>
    </div>
  </aside>
</template>
