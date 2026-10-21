<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import AppIcon from '../common/AppIcon.vue'
import { logoutAccount } from '../../api/account'
import { clearAuthTokens, getAccountId, getRefreshToken, isLoggedIn } from '../../utils/auth'

const GUEST_PROMPT_DISMISSED_KEY = 'guest-login-prompt-dismissed'
const route = useRoute()
const router = useRouter()
const loggedIn = ref(isLoggedIn())
const accountId = ref(getAccountId())
const guestPromptDismissed = ref(sessionStorage.getItem(GUEST_PROMPT_DISMISSED_KEY) === '1')
const guestPromptReady = ref(false)
const showGuestLoginPrompt = computed(() => route.name === 'home' && !loggedIn.value && !guestPromptDismissed.value && guestPromptReady.value)
const searchKeyword = ref('')
const accountMenuOpen = ref(false)
function submitSearch() {
  if (searchKeyword.value.trim()) router.push({ name: 'search', query: { q: searchKeyword.value.trim() } })
}
function closeMenu(event) {
  if (event.key === 'Escape') accountMenuOpen.value = false
}
watch(() => route.fullPath, () => { accountMenuOpen.value = false })
let guestPromptTimer

function syncLoginState() {
  loggedIn.value = isLoggedIn()
  accountId.value = getAccountId()
}

async function handleLogout() {
  const refreshToken = getRefreshToken()

  try {
    if (refreshToken) {
      await logoutAccount(refreshToken)
    }
  } catch {
    // 即使服务端暂时不可用，也要允许用户清除当前浏览器中的登录状态。
  } finally {
    clearAuthTokens()
    loggedIn.value = false
    accountId.value = ''
    sessionStorage.removeItem(GUEST_PROMPT_DISMISSED_KEY)
    guestPromptDismissed.value = false
    router.push('/')
  }
}

function dismissGuestLoginPrompt() {
  sessionStorage.setItem(GUEST_PROMPT_DISMISSED_KEY, '1')
  guestPromptDismissed.value = true
}

function scheduleGuestLoginPrompt() {
  window.clearTimeout(guestPromptTimer)
  guestPromptReady.value = false

  if (route.name === 'home' && !loggedIn.value && !guestPromptDismissed.value) {
    guestPromptTimer = window.setTimeout(() => {
      guestPromptReady.value = true
    }, 1000)
  }
}

function isImmersiveRoute() {
  return route.name === 'home' || route.name === 'following-feed'
}

function syncRouteLayout() {
  const immersive = isImmersiveRoute()
  document.documentElement.classList.toggle('immersive-route', immersive)
  document.body.classList.toggle('immersive-route', immersive)
}

watch(() => route.fullPath, syncLoginState)
watch(() => route.name, syncRouteLayout, { immediate: true })
watch([() => route.name, loggedIn, guestPromptDismissed], scheduleGuestLoginPrompt, { immediate: true })
onMounted(() => {
  window.addEventListener('keydown', closeMenu)
  window.addEventListener('storage', syncLoginState)
  window.addEventListener('auth-changed', syncLoginState)
})
onBeforeUnmount(() => {
  window.clearTimeout(guestPromptTimer)
  window.removeEventListener('keydown', closeMenu)
  window.removeEventListener('storage', syncLoginState)
  window.removeEventListener('auth-changed', syncLoginState)
  document.documentElement.classList.remove('immersive-route')
  document.body.classList.remove('immersive-route')
})
</script>

<template>
  <div class="app-shell" :class="{ 'app-shell--immersive': isImmersiveRoute() }">
    <aside class="sidebar" aria-label="主要导航">
      <RouterLink class="brand" to="/" aria-label="FrameFlow 首页"><span class="brand__mark"><AppIcon name="play" :size="19" /></span><span>FrameFlow</span></RouterLink>
      <nav class="sidebar-nav">
        <RouterLink to="/"><AppIcon name="home" /><span>推荐</span></RouterLink>
        <RouterLink to="/hot"><AppIcon name="fire" /><span>热门</span></RouterLink>
        <RouterLink to="/following"><AppIcon name="users" /><span>关注</span></RouterLink>
        <RouterLink :to="accountId ? { name: 'user-profile', params: { accountId } } : '/login'"><AppIcon name="user" /><span>我的</span></RouterLink>
      </nav>
      <div class="sidebar-bottom"><span class="sidebar-caption">每一帧，都有新发现</span><span>FrameFlow · 短视频社区</span></div>
    </aside>
    <header class="site-header">
      <div class="site-header__inner">
        <form class="header-search" role="search" @submit.prevent="submitSearch">
          <input v-model="searchKeyword" type="search" aria-label="搜索用户" placeholder="搜索你感兴趣的创作者">
          <button type="submit" aria-label="搜索"><AppIcon name="search" :size="19" /><span>搜索</span></button>
        </form>
        <nav class="account-nav" aria-label="账号导航">
          <RouterLink class="primary-link" to="/publish" aria-label="发布作品"><AppIcon name="plus" :size="18" /><span>发布作品</span></RouterLink>
          <div v-if="loggedIn" class="account-menu">
            <button class="account-avatar" type="button" aria-label="打开账号菜单" :aria-expanded="accountMenuOpen" @click="accountMenuOpen = !accountMenuOpen"><AppIcon name="user" :size="21" /></button>
            <template v-if="accountMenuOpen">
              <button class="menu-dismiss" aria-label="关闭账号菜单" @click="accountMenuOpen = false"></button>
              <div class="account-dropdown">
                <RouterLink v-if="accountId" :to="{ name: 'user-profile', params: { accountId } }">我的主页</RouterLink>
                <RouterLink to="/me/liked">我的喜欢</RouterLink>
                <RouterLink to="/me/following">我的关注</RouterLink>
                <RouterLink to="/me/followers">我的粉丝</RouterLink>
                <button type="button" @click="handleLogout"><AppIcon name="logout" :size="17" />退出登录</button>
              </div>
            </template>
          </div>
          <RouterLink v-else class="login-link" to="/login">登录 / 注册</RouterLink>
        </nav>
      </div>
    </header>

    <main class="page-shell" :class="{ 'page-shell--immersive': isImmersiveRoute() }">
      <RouterView />
    </main>

    <aside v-if="showGuestLoginPrompt" class="guest-login-notice" aria-label="登录提示">
      <button class="guest-login-notice__close" type="button" aria-label="关闭登录提示" @click="dismissGuestLoginPrompt">×</button>
      <div>
        <strong>登录后可以参与互动</strong>
        <p>不登录也能继续刷视频；登录后可点赞、评论和关注作者。</p>
      </div>
      <RouterLink :to="{ name: 'login', query: { redirect: '/' } }">去登录</RouterLink>
    </aside>
  </div>
</template>
