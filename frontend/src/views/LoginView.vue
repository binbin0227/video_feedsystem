<script setup>
import { computed, reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { loginAccount } from '../api/account'
import { saveToken } from '../utils/auth'

const route = useRoute()
const router = useRouter()
const form = reactive({
  username: typeof route.query.username === 'string' ? route.query.username : '',
  password: '',
})
const errorMessage = ref('')
const submitting = ref(false)

const registrationMessage = computed(() => (
  route.query.registered === '1' ? '注册成功，请使用新账号登录。' : ''
))

const registerRoute = computed(() => {
  const redirect = getSafeRedirect()
  return redirect ? { name: 'register', query: { redirect } } : { name: 'register' }
})

function getSafeRedirect() {
  const redirect = route.query.redirect

  if (typeof redirect === 'string' && redirect.startsWith('/') && !redirect.startsWith('//')) {
    return redirect
  }

  return ''
}

async function handleSubmit() {
  errorMessage.value = ''
  const username = form.username.trim()

  if (!username || !form.password) {
    errorMessage.value = '请输入用户名和密码'
    return
  }

  submitting.value = true

  try {
    const data = await loginAccount(username, form.password)

    if (!data?.token) {
      throw new Error('登录响应中没有 Token')
    }

    saveToken(data.token)
    await router.replace(getSafeRedirect() || '/')
  } catch (error) {
    errorMessage.value = error.message
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section class="auth-page">
    <div class="auth-card">
      <div class="auth-card__heading">
        <p class="eyebrow">欢迎回来</p>
        <h1>登录 FrameFlow</h1>
        <p>继续浏览关注内容，管理点赞与发布作品。</p>
      </div>

      <p v-if="registrationMessage" class="form-message form-message--success" role="status">
        {{ registrationMessage }}
      </p>
      <p v-if="errorMessage" class="form-message form-message--error" role="alert">
        {{ errorMessage }}
      </p>

      <form class="auth-form" @submit.prevent="handleSubmit">
        <label class="form-field">
          <span>用户名</span>
          <input v-model="form.username" name="username" type="text" autocomplete="username" placeholder="请输入用户名" :disabled="submitting" @input="errorMessage = ''">
        </label>

        <label class="form-field">
          <span>密码</span>
          <input v-model="form.password" name="password" type="password" autocomplete="current-password" placeholder="请输入密码" :disabled="submitting" @input="errorMessage = ''">
        </label>

        <button class="auth-submit" type="submit" :disabled="submitting">
          {{ submitting ? '正在登录…' : '登录' }}
        </button>
      </form>

      <p class="auth-switch">
        还没有账号？
        <RouterLink :to="registerRoute">立即注册</RouterLink>
      </p>
    </div>

    <aside class="auth-showcase" aria-label="登录功能说明">
      <span class="auth-showcase__badge">AUTH / 01</span>
      <div>
        <p class="auth-showcase__kicker">JWT SESSION</p>
        <h2>一次登录，解锁完整互动。</h2>
        <p>登录凭证保存在当前浏览器中，访问受保护页面时会自动携带。</p>
      </div>
      <ul class="auth-feature-list">
        <li><span>01</span>关注视频流</li>
        <li><span>02</span>点赞与评论</li>
        <li><span>03</span>上传发布作品</li>
      </ul>
    </aside>
  </section>
</template>
