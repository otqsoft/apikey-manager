<template>
  <div class="login-page">
    <!-- 粒子网络背景（尊重 reduced-motion：不做动画） -->
    <canvas ref="canvasRef" class="bg-canvas" />
    <div class="glow glow-1" />
    <div class="glow glow-2" />

    <div class="login-card">
      <div class="login-header">
        <div class="logo-mark">
          <img src="/logo.png" alt="logo" />
        </div>
        <h1>API Key 控制台</h1>
        <p>统一纳管多平台大模型密钥、额度与对话测试</p>
      </div>

      <el-tabs v-model="activeTab" class="auth-tabs" stretch>
        <!-- 登录 -->
        <el-tab-pane label="登录" name="login">
          <el-form ref="loginRef" :model="loginForm" :rules="loginRules" label-position="top" @keyup.enter="onLogin">
            <el-form-item label="账号" prop="username">
              <el-input v-model="loginForm.username" placeholder="请输入账号" size="large" :prefix-icon="User" />
            </el-form-item>
            <el-form-item label="密码" prop="password">
              <el-input
                v-model="loginForm.password"
                type="password"
                show-password
                placeholder="请输入密码"
                size="large"
                :prefix-icon="Lock"
                @keyup.enter="onLogin"
              />
            </el-form-item>
            <button type="button" class="auth-btn" :disabled="loginLoading" @click="onLogin">
              <span v-if="!loginLoading">登 录</span>
              <span v-else class="btn-loading" />
            </button>
          </el-form>
        </el-tab-pane>

        <!-- 注册（契约 8.1：公开接口，成功不签发 token） -->
        <el-tab-pane label="注册" name="register">
          <el-form ref="regRef" :model="regForm" :rules="regRules" label-position="top" @keyup.enter="onRegister">
            <el-form-item label="用户名" prop="username">
              <el-input v-model="regForm.username" placeholder="3~32 位字母、数字或下划线" size="large" :prefix-icon="User" />
            </el-form-item>
            <el-form-item label="密码" prop="password">
              <el-input
                v-model="regForm.password"
                type="password"
                show-password
                placeholder="至少 6 位"
                size="large"
                :prefix-icon="Lock"
              />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirm">
              <el-input
                v-model="regForm.confirm"
                type="password"
                show-password
                placeholder="再次输入密码"
                size="large"
                :prefix-icon="Lock"
                @keyup.enter="onRegister"
              />
            </el-form-item>
            <el-form-item label="昵称（可选）" prop="display_name">
              <el-input v-model="regForm.display_name" placeholder="展示名，可留空" size="large" :prefix-icon="Avatar" />
            </el-form-item>
            <button type="button" class="auth-btn" :disabled="regLoading" @click="onRegister">
              <span v-if="!regLoading">注 册</span>
              <span v-else class="btn-loading" />
            </button>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { Avatar, Lock, User } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { register } from '@/api/auth'
import { useUserStore } from '@/store/user'

const store = useUserStore()
const router = useRouter()
const route = useRoute()

/** 当前模式：login 登录 / register 注册 */
const activeTab = ref<'login' | 'register'>('login')

// ---------------- 登录表单 ----------------
const loginRef = ref<FormInstance>()
const loginLoading = ref(false)
const loginForm = reactive({ username: '', password: '' })
const loginRules: FormRules = {
  username: [{ required: true, message: '请输入账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

async function onLogin() {
  if (loginLoading.value) return
  const valid = await loginRef.value?.validate().catch(() => false)
  if (!valid) return
  loginLoading.value = true
  try {
    await store.login({ username: loginForm.username.trim(), password: loginForm.password })
    ElMessage.success('登录成功')
    const redirect = (route.query.redirect as string) || '/dashboard'
    router.replace(redirect)
  } catch {
    // 拦截器已统一弹出错误消息（401/403/业务 code）
  } finally {
    loginLoading.value = false
  }
}

// ---------------- 注册表单（契约 8.1） ----------------
const regRef = ref<FormInstance>()
const regLoading = ref(false)
const regForm = reactive({ username: '', password: '', confirm: '', display_name: '' })

const validateRegUsername = (_rule: unknown, value: string, cb: (err?: Error) => void) => {
  if (!value) return cb(new Error('请输入用户名'))
  if (!/^[A-Za-z0-9_]{3,32}$/.test(value)) return cb(new Error('3~32 位字母、数字或下划线'))
  cb()
}
const validateRegPassword = (_rule: unknown, value: string, cb: (err?: Error) => void) => {
  if (!value) return cb(new Error('请输入密码'))
  if (value.length < 6) return cb(new Error('密码至少 6 位'))
  cb()
}
const validateConfirm = (_rule: unknown, value: string, cb: (err?: Error) => void) => {
  if (!value) return cb(new Error('请再次输入密码'))
  if (value !== regForm.password) return cb(new Error('两次输入的密码不一致'))
  cb()
}

const regRules: FormRules = {
  username: [{ validator: validateRegUsername, trigger: 'blur' }],
  password: [{ validator: validateRegPassword, trigger: 'blur' }],
  confirm: [{ validator: validateConfirm, trigger: 'blur' }],
  display_name: [{ max: 32, message: '昵称最长 32 个字符', trigger: 'blur' }]
}

async function onRegister() {
  if (regLoading.value) return
  const valid = await regRef.value?.validate().catch(() => false)
  if (!valid) return
  regLoading.value = true
  try {
    await register({
      username: regForm.username.trim(),
      password: regForm.password,
      display_name: regForm.display_name.trim() || undefined
    })
    ElMessage.success('注册成功，请登录')
    // 切回登录 Tab 并回填用户名；注册接口不签发 token（契约 8.1）
    activeTab.value = 'login'
    loginForm.username = regForm.username.trim()
    loginForm.password = ''
    regRef.value?.resetFields()
    regForm.password = ''
    regForm.confirm = ''
  } catch {
    // 拦截器透传后端 msg：用户名已存在 / 注册已关闭 / 参数校验失败等
  } finally {
    regLoading.value = false
  }
}

// ---------------- 粒子网络背景 ----------------
const canvasRef = ref<HTMLCanvasElement>()
let rafId = 0
let resizeHandler: (() => void) | null = null

interface Particle {
  x: number
  y: number
  vx: number
  vy: number
  r: number
}

function initCanvas() {
  const canvas = canvasRef.value
  if (!canvas) return
  const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  let w = (canvas.width = window.innerWidth)
  let h = (canvas.height = window.innerHeight)
  const count = Math.min(90, Math.floor((w * h) / 16000))
  const particles: Particle[] = Array.from({ length: count }, () => ({
    x: Math.random() * w,
    y: Math.random() * h,
    vx: (Math.random() - 0.5) * 0.35,
    vy: (Math.random() - 0.5) * 0.35,
    r: Math.random() * 1.6 + 0.6
  }))

  const LINK_DIST = 130

  function draw() {
    ctx.clearRect(0, 0, w, h)
    for (const p of particles) {
      p.x += p.vx
      p.y += p.vy
      if (p.x < 0 || p.x > w) p.vx *= -1
      if (p.y < 0 || p.y > h) p.vy *= -1
      ctx.beginPath()
      ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2)
      ctx.fillStyle = 'rgba(0, 212, 255, 0.55)'
      ctx.fill()
    }
    for (let i = 0; i < particles.length; i++) {
      for (let j = i + 1; j < particles.length; j++) {
        const a = particles[i]
        const b = particles[j]
        const dx = a.x - b.x
        const dy = a.y - b.y
        const dist = Math.hypot(dx, dy)
        if (dist < LINK_DIST) {
          ctx.beginPath()
          ctx.moveTo(a.x, a.y)
          ctx.lineTo(b.x, b.y)
          ctx.strokeStyle = `rgba(59, 130, 246, ${(1 - dist / LINK_DIST) * 0.22})`
          ctx.lineWidth = 1
          ctx.stroke()
        }
      }
    }
    rafId = requestAnimationFrame(draw)
  }

  if (reduceMotion) {
    draw()
    // 静态帧：画一帧即停，不做动画
    cancelAnimationFrame(rafId)
  } else {
    rafId = requestAnimationFrame(draw)
  }

  resizeHandler = () => {
    w = canvas.width = window.innerWidth
    h = canvas.height = window.innerHeight
  }
  window.addEventListener('resize', resizeHandler)
}

onMounted(initCanvas)
onBeforeUnmount(() => {
  cancelAnimationFrame(rafId)
  if (resizeHandler) window.removeEventListener('resize', resizeHandler)
})
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  padding: 24px 0;
  background:
    radial-gradient(ellipse 60% 45% at 70% 20%, rgba(59, 130, 246, 0.14), transparent),
    radial-gradient(ellipse 50% 40% at 20% 80%, rgba(0, 212, 255, 0.08), transparent),
    #0a1628;
}
.bg-canvas {
  position: absolute;
  inset: 0;
}
/* 弥散光斑，营造纵深 */
.glow {
  position: absolute;
  border-radius: 50%;
  filter: blur(90px);
  pointer-events: none;
}
.glow-1 {
  width: 420px;
  height: 420px;
  background: rgba(59, 130, 246, 0.16);
  top: -120px;
  right: -80px;
}
.glow-2 {
  width: 340px;
  height: 340px;
  background: rgba(0, 212, 255, 0.1);
  bottom: -120px;
  left: -60px;
}
@media (prefers-reduced-motion: reduce) {
  .glow {
    filter: none;
  }
}

.login-card {
  position: relative;
  z-index: 1;
  width: 420px;
  max-width: calc(100vw - 32px);
  padding: 30px 32px 22px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
}
.login-header {
  text-align: center;
  margin-bottom: 18px;
}
.logo-mark {
  width: 52px;
  height: 52px;
  margin: 0 auto;
  border-radius: 14px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  background: none;
}
.logo-mark img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.login-header h1 {
  font-size: 20px;
  margin: 16px 0 6px;
  color: #f1f5f9;
  font-weight: 700;
  letter-spacing: 0.5px;
}
.login-header p {
  margin: 0;
  color: #7d93b2;
  font-size: 13px;
}

/* 深色 Tabs 定制 */
.auth-tabs :deep(.el-tabs__nav-wrap::after) {
  background: rgba(255, 255, 255, 0.08);
  height: 1px;
}
.auth-tabs :deep(.el-tabs__item) {
  font-size: 15px;
  font-weight: 600;
  color: #64748b;
}
.auth-tabs :deep(.el-tabs__item.is-active),
.auth-tabs :deep(.el-tabs__item:hover) {
  color: #00d4ff;
}
.auth-tabs :deep(.el-tabs__active-bar) {
  background: linear-gradient(90deg, #3b82f6, #00d4ff);
  height: 2px;
  border-radius: 1px;
}

/* 深色输入框定制 */
.auth-tabs :deep(.el-form-item__label) {
  color: #8fa5c2;
  font-size: 13px;
}
.auth-tabs :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.06);
  border-radius: 10px;
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.12) inset;
}
.auth-tabs :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px rgba(0, 212, 255, 0.35) inset;
}
.auth-tabs :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #00d4ff inset;
  background: rgba(255, 255, 255, 0.08);
}
.auth-tabs :deep(.el-input__inner) {
  color: #e6edf6;
}
.auth-tabs :deep(.el-input__inner::placeholder) {
  color: #55688a;
}
.auth-tabs :deep(.el-input__prefix),
.auth-tabs :deep(.el-input__suffix) {
  color: #55688a;
}
.auth-tabs :deep(.el-form-item__error) {
  color: #ff7a90;
}

/* 渐变登录按钮 */
.auth-btn {
  width: 100%;
  height: 44px;
  margin-top: 6px;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 4px;
  color: #fff;
  background: linear-gradient(135deg, #3b82f6 0%, #00d4ff 100%);
  transition: filter 0.15s ease, transform 0.1s ease;
}
.auth-btn:hover:not(:disabled) {
  filter: brightness(1.1);
}
.auth-btn:active:not(:disabled) {
  transform: scale(0.985);
}
.auth-btn:disabled {
  cursor: wait;
  opacity: 0.75;
}
.btn-loading {
  display: inline-block;
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.35);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.login-tip {
  margin-top: 14px;
  text-align: center;
  font-size: 12px;
  color: #55688a;
  line-height: 1.6;
}
@media (prefers-reduced-motion: reduce) {
  .btn-loading {
    animation-duration: 2s;
  }
}

/* 窄屏自适应 */
@media (max-width: 480px) {
  .login-card {
    padding: 24px 20px 18px;
    border-radius: 14px;
  }
  .login-header h1 {
    font-size: 18px;
  }
}
</style>
