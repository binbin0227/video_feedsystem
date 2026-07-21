<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { getFollowerAccounts, getFollowingAccounts } from '../../api/social'

const PAGE_SIZE = 12

const props = defineProps({
  mode: {
    type: String,
    required: true,
    validator: (value) => ['following', 'followers'].includes(value),
  },
})

const route = useRoute()
const router = useRouter()
const accounts = ref([])
const nextCursor = ref('')
const hasMore = ref(false)
const initialLoading = ref(true)
const loadingMore = ref(false)
const errorMessage = ref('')

const pageCopy = computed(() => props.mode === 'following'
  ? {
      eyebrow: 'FOLLOWING',
      title: '我的关注',
      description: '查看当前账号正在关注的创作者。',
      emptyTitle: '还没有关注任何人',
      emptyText: '可以从搜索或视频详情进入用户主页并关注。',
      datePrefix: '关注于',
    }
  : {
      eyebrow: 'FOLLOWERS',
      title: '我的粉丝',
      description: '查看当前账号收到的关注关系。',
      emptyTitle: '还没有粉丝',
      emptyText: '发布作品后，更多用户可能会关注你。',
      datePrefix: '成为粉丝于',
    })

function formatDate(value) {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return '时间未知'
  }

  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(date)
}

async function loadAccounts({ append = false } = {}) {
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
    const fetcher = props.mode === 'following' ? getFollowingAccounts : getFollowerAccounts
    const data = await fetcher({
      cursor: append ? nextCursor.value : '',
      limit: PAGE_SIZE,
    })
    const newAccounts = Array.isArray(data?.accounts) ? data.accounts : []

    accounts.value = append ? [...accounts.value, ...newAccounts] : newAccounts
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

onMounted(() => loadAccounts())
</script>

<template>
  <section class="account-list-page">
    <header class="collection-heading">
      <div>
        <p class="eyebrow">{{ pageCopy.eyebrow }}</p>
        <h1>{{ pageCopy.title }}</h1>
        <p>{{ pageCopy.description }}</p>
      </div>
      <span v-if="accounts.length">已加载 {{ accounts.length }} 人</span>
    </header>

    <div v-if="initialLoading" class="account-list-loading" aria-busy="true">正在加载用户列表…</div>

    <div v-else-if="errorMessage && accounts.length === 0" class="collection-state" role="alert">
      <strong>列表加载失败</strong>
      <p>{{ errorMessage }}</p>
      <button type="button" @click="loadAccounts()">重新加载</button>
    </div>

    <div v-else-if="accounts.length === 0" class="collection-state">
      <strong>{{ pageCopy.emptyTitle }}</strong>
      <p>{{ pageCopy.emptyText }}</p>
    </div>

    <template v-else>
      <div class="account-list">
        <RouterLink
          v-for="account in accounts"
          :key="account.account_id"
          class="account-list-item"
          :to="{ name: 'user-profile', params: { accountId: account.account_id } }"
        >
          <div>
            <strong>{{ account.username }}</strong>
            <span>ID {{ account.account_id }}</span>
          </div>
          <div class="account-list-item__meta">
            <time :datetime="account.followed_at">{{ pageCopy.datePrefix }} {{ formatDate(account.followed_at) }}</time>
            <span aria-hidden="true">→</span>
          </div>
        </RouterLink>
      </div>

      <div class="collection-pagination">
        <p v-if="errorMessage" role="alert">{{ errorMessage }}</p>
        <button v-if="hasMore" type="button" :disabled="loadingMore" @click="loadAccounts({ append: true })">
          {{ loadingMore ? '正在加载…' : '加载更多' }}
        </button>
        <span v-else>已经看到全部用户</span>
      </div>
    </template>
  </section>
</template>
