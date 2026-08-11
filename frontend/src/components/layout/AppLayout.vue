<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { getAccountId, isLoggedIn, removeToken } from '../../utils/auth'

const GUEST_PROMPT_DISMISSED_KEY = 'guest-login-prompt-dismissed'
const route = useRoute()
const router = useRouter()
const loggedIn = ref(isLoggedIn())
const accountId = ref(getAccountId())
const guestPromptDismissed = ref(sessionStorage.getItem(GUEST_PROMPT_DISMISSED_KEY) === '1')
const guestPromptReady = ref(false)
const showGuestLoginPrompt = computed(() => route.name === 'home' && !loggedIn.value && !guestPromptDismissed.value && guestPromptReady.value)
let guestPromptTimer

function syncLoginState() {
  loggedIn.value = isLoggedIn()
  accountId.value = getAccountId()
}

function handleLogout() {
  removeToken()
  loggedIn.value = false
  accountId.value = ''
  sessionStorage.removeItem(GUEST_PROMPT_DISMISSED_KEY)
  guestPromptDismissed.value = false
  router.push('/')
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
  window.addEventListener('storage', syncLoginState)
  window.addEventListener('auth-changed', syncLoginState)
})
onBeforeUnmount(() => {
  window.clearTimeout(guestPromptTimer)
  window.removeEventListener('storage', syncLoginState)
  window.removeEventListener('auth-changed', syncLoginState)
  document.documentElement.classList.remove('immersive-route')
  document.body.classList.remove('immersive-route')
})
</script>

<template>
  <div class="app-shell" :class="{ 'app-shell--immersive': isImmersiveRoute() }">
    <header class="site-header">
      <div class="site-header__inner">
        <RouterLink class="brand" to="/" aria-label="FrameFlow 首页">
          <span class="brand__mark">FF</span>
          <span>FrameFlow</span>
        </RouterLink>

        <nav class="main-nav" aria-label="主要导航">
          <RouterLink to="/">首页</RouterLink>
          <RouterLink to="/hot">热门</RouterLink>
          <RouterLink class="main-nav__following" to="/following">关注流</RouterLink>
          <RouterLink to="/search">搜索用户</RouterLink>
        </nav>

        <nav class="account-nav" aria-label="账号导航">
          <template v-if="loggedIn">
            <RouterLink v-if="accountId" class="secondary-account-link" :to="{ name: 'user-profile', params: { accountId } }">我的主页</RouterLink>
            <RouterLink class="secondary-account-link" to="/me/liked">我的点赞</RouterLink>
            <RouterLink class="secondary-account-link" to="/me/following">我的关注</RouterLink>
            <RouterLink class="secondary-account-link" to="/me/followers">我的粉丝</RouterLink>
            <RouterLink class="primary-link" to="/publish">发布</RouterLink>
            <button class="logout-button" type="button" @click="handleLogout">退出</button>
          </template>
          <template v-else>
            <RouterLink to="/login">登录</RouterLink>
            <RouterLink class="primary-link" to="/register">注册</RouterLink>
          </template>
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
