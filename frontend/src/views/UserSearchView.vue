<script setup>
import { ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { searchAccounts } from '../api/account'

const route = useRoute()
const router = useRouter()
const keyword = ref('')
const accounts = ref([])
const loading = ref(false)
const searched = ref(false)
const errorMessage = ref('')

function formatCount(value) {
  return Number(value || 0).toLocaleString('zh-CN')
}

async function runSearch(value) {
  const normalized = value.trim()

  if (!normalized) {
    accounts.value = []
    searched.value = false
    errorMessage.value = ''
    return
  }

  loading.value = true
  searched.value = true
  errorMessage.value = ''

  try {
    const data = await searchAccounts(normalized)
    accounts.value = Array.isArray(data?.accounts) ? data.accounts : []
  } catch (error) {
    accounts.value = []
    errorMessage.value = error.message
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  const normalized = keyword.value.trim()

  if (!normalized) {
    errorMessage.value = '请输入用户名关键词'
    return
  }

  if (route.query.q === normalized) {
    await runSearch(normalized)
    return
  }

  await router.replace({ name: 'search', query: { q: normalized } })
}

watch(
  () => route.query.q,
  (value) => {
    keyword.value = typeof value === 'string' ? value : ''
    runSearch(keyword.value)
  },
  { immediate: true },
)
</script>

<template>
  <section class="search-page">
    <header class="search-heading">
      <p class="eyebrow">DISCOVER CREATORS</p>
      <h1>搜索用户</h1>
      <p>按用户名关键词查找创作者，后端每次最多返回 20 条结果。</p>
    </header>

    <form class="search-form" role="search" @submit.prevent="handleSubmit">
      <label for="account-keyword">用户名关键词</label>
      <div>
        <input id="account-keyword" v-model="keyword" name="keyword" type="search" autocomplete="off" placeholder="例如：test" :disabled="loading" @input="errorMessage = ''">
        <button type="submit" :disabled="loading">{{ loading ? '搜索中…' : '搜索' }}</button>
      </div>
    </form>

    <p v-if="errorMessage" class="search-error" role="alert">{{ errorMessage }}</p>

    <div v-if="loading" class="account-list-loading" aria-busy="true">正在搜索用户…</div>

    <div v-else-if="searched && accounts.length === 0 && !errorMessage" class="collection-state">
      <strong>没有找到匹配用户</strong>
      <p>试试更短或不同的用户名关键词。</p>
    </div>

    <div v-else-if="accounts.length" class="search-results">
      <header>
        <strong>搜索结果</strong>
        <span>{{ accounts.length }} 人</span>
      </header>

      <div class="search-account-grid">
        <RouterLink
          v-for="account in accounts"
          :key="account.account_id"
          class="search-account-card"
          :to="{ name: 'user-profile', params: { accountId: account.account_id } }"
        >
          <span class="search-account-card__label">CREATOR</span>
          <h2>{{ account.username }}</h2>
          <p>ID {{ account.account_id }}</p>
          <dl>
            <div><dt>获赞</dt><dd>{{ formatCount(account.received_like_count) }}</dd></div>
            <div><dt>粉丝</dt><dd>{{ formatCount(account.follower_count) }}</dd></div>
          </dl>
          <span class="search-account-card__link">查看主页 →</span>
        </RouterLink>
      </div>
    </div>
  </section>
</template>
