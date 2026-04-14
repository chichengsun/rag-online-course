<template>
  <main class="auth-layout">
    <section class="brand-pane">
      <h1>RAG Online Course</h1>
      <p>教师课程管理台</p>
      <p class="subtle">支持课程、章节、资源流程一体化管理。</p>
    </section>
    <section class="form-pane card">
      <div class="mode-switch">
        <button :class="mode === 'login' ? 'active' : ''" type="button" @click="mode = 'login'">登录</button>
        <button :class="mode === 'register' ? 'active' : ''" type="button" @click="mode = 'register'">注册</button>
      </div>
      <h2>{{ mode === 'login' ? '欢迎回来' : '创建教师账号' }}</h2>
      <div v-if="error" class="alert error">{{ error }}</div>

      <form v-if="mode === 'login'" class="stack" @submit.prevent="onSubmitLogin">
        <label>
          账号（用户名或邮箱）
          <input v-model="loginForm.account" required />
        </label>
        <label>
          密码
          <input v-model="loginForm.password" required type="password" />
        </label>
        <button :disabled="loading" type="submit" class="primary">{{ loading ? '登录中...' : '登录' }}</button>
      </form>

      <form v-else class="stack" @submit.prevent="onSubmitRegister">
        <label>
          邮箱
          <input v-model="registerForm.email" required type="email" />
        </label>
        <label>
          用户名
          <input v-model="registerForm.username" required />
        </label>
        <label>
          姓名
          <input v-model="registerForm.name" required />
        </label>
        <label>
          密码
          <input v-model="registerForm.password" required type="password" minlength="6" />
        </label>
        <button :disabled="loading" type="submit" class="primary">{{ loading ? '注册中...' : '注册' }}</button>
      </form>
    </section>
  </main>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { login, register, type RegisterPayload } from '@/services/auth'
import { useAuthStore } from '@/stores/auth'

const mode = ref<'login' | 'register'>('login')
const loading = ref(false)
const error = ref('')
const router = useRouter()
const auth = useAuthStore()

const loginForm = reactive({ account: '', password: '' })
const registerForm = reactive<RegisterPayload>({
  email: '',
  username: '',
  name: '',
  password: '',
  role: 'teacher',
})

async function onSubmitLogin() {
  loading.value = true
  error.value = ''
  try {
    const data = await login(loginForm)
    auth.setAuth(data.access_token, data.user)
    router.replace('/teacher/courses')
  } catch (err) {
    error.value = err instanceof Error ? err.message : '登录失败，请重试'
  } finally {
    loading.value = false
  }
}

async function onSubmitRegister() {
  loading.value = true
  error.value = ''
  try {
    await register(registerForm)
    mode.value = 'login'
    loginForm.account = registerForm.username
    loginForm.password = registerForm.password
  } catch (err) {
    error.value = err instanceof Error ? err.message : '注册失败，请检查输入'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.brand-pane h1 {
  max-width: 12ch;
}

.form-pane {
  padding: 22px;
}

.form-pane form button.primary {
  margin-top: 4px;
}
</style>
