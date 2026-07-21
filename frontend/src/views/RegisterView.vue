<script setup>
import { computed, reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { registerAccount } from '../api/account'

const route = useRoute()
const router = useRouter()
const form = reactive({
  username: '',
  password: '',
  confirmPassword: '',
})
const errorMessage = ref('')
const submitting = ref(false)

const loginRoute = computed(() => {
  const redirect = getSafeRedirect()
  return redirect ? { name: 'login', query: { redirect } } : { name: 'login' }
})

function getSafeRedirect() {
  const redirect = route.query.redirect

  if (typeof redirect === 'string' && redirect.startsWith('/') && !redirect.startsWith('//')) {
    return redirect
  }

  return ''
}

function validateForm(username) {
  if (!username || !form.password) {
    return '请输入用户名和密码'
  }
  if ([...username].length > 32) {
    return '用户名不能超过 32 个字符'
  }

  const passwordBytes = new TextEncoder().encode(form.password).length
  if (passwordBytes < 8) {
    return '密码不能少于 8 个字节'
  }
  if (passwordBytes > 72) {
    return '密码不能超过 72 个字节'
  }
  if (form.password !== form.confirmPassword) {
    return '两次输入的密码不一致'
  }

  return ''
}

async function handleSubmit() {
  errorMessage.value = ''
  const username = form.username.trim()
  const validationMessage = validateForm(username)

  if (validationMessage) {
    errorMessage.value = validationMessage
    return
  }

  submitting.value = true

  try {
    await registerAccount(username, form.password)
    const query = {
      registered: '1',
      username,
    }
    const redirect = getSafeRedirect()

    if (redirect) {
      query.redirect = redirect
    }

    await router.replace({ name: 'login', query })
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
        <p class="eyebrow">创建账号</p>
        <h1>加入 FrameFlow</h1>
        <p>用一个用户名开启你的短视频社区体验。</p>
      </div>

      <p v-if="errorMessage" class="form-message form-message--error" role="alert">
        {{ errorMessage }}
      </p>

      <form class="auth-form" @submit.prevent="handleSubmit">
        <label class="form-field">
          <span>用户名</span>
          <input v-model="form.username" name="username" type="text" autocomplete="username" placeholder="最多 32 个字符" :disabled="submitting" @input="errorMessage = ''">
        </label>

        <label class="form-field">
          <span>密码</span>
          <input v-model="form.password" name="password" type="password" autocomplete="new-password" placeholder="8～72 个字节" :disabled="submitting" @input="errorMessage = ''">
        </label>

        <label class="form-field">
          <span>确认密码</span>
          <input v-model="form.confirmPassword" name="confirm-password" type="password" autocomplete="new-password" placeholder="再次输入密码" :disabled="submitting" @input="errorMessage = ''">
        </label>

        <button class="auth-submit" type="submit" :disabled="submitting">
          {{ submitting ? '正在注册…' : '注册账号' }}
        </button>
      </form>

      <p class="auth-switch">
        已经有账号？
        <RouterLink :to="loginRoute">返回登录</RouterLink>
      </p>
    </div>

    <aside class="auth-showcase auth-showcase--register" aria-label="注册规则说明">
      <span class="auth-showcase__badge">ACCOUNT / NEW</span>
      <div>
        <p class="auth-showcase__kicker">START CREATING</p>
        <h2>从一个账号开始，记录每次表达。</h2>
        <p>注册信息会提交到 Go Hertz 后端，密码经过加密后保存。</p>
      </div>
      <ul class="auth-feature-list">
        <li><span>01</span>唯一用户名</li>
        <li><span>02</span>密码加密存储</li>
        <li><span>03</span>注册后登录</li>
      </ul>
    </aside>
  </section>
</template>
