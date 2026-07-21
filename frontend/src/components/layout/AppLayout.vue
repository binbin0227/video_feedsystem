<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { getAccountId, isLoggedIn, removeToken } from '../../utils/auth'

const route = useRoute()
const router = useRouter()
const loggedIn = ref(isLoggedIn())
const accountId = ref(getAccountId())

function syncLoginState() {
  loggedIn.value = isLoggedIn()
  accountId.value = getAccountId()
}

function handleLogout() {
  removeToken()
  loggedIn.value = false
  accountId.value = ''
  router.push('/')
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
onMounted(() => {
  window.addEventListener('storage', syncLoginState)
  window.addEventListener('auth-changed', syncLoginState)
})
onBeforeUnmount(() => {
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
          <RouterLink to="/following">关注流</RouterLink>
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
  </div>
</template>
