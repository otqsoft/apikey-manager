<template>
  <div class="login-page">
    <!-- 浅色粒子网络背景（尊重 reduced-motion：只画一帧不做动画） -->
    <canvas ref="canvasRef" class="bg-canvas" />

    <div class="login-shell">
      <!-- 左侧品牌区（窄屏隐藏） -->
      <section class="brand-panel">
        <div class="brand-top">
          <div class="logo-mark"><img src="/logo.png" alt="logo" /></div>
          <span class="brand-name">API Key 控制台</span>
        </div>

        <div class="brand-hero">
          <h1>统一纳管多平台<br />大模型 API 密钥</h1>
          <p>供应商 · 模型 · 密钥 · 额度 · 对话测试，一个控制台全部搞定</p>
          <ul class="brand-features">
            <li>
              <el-icon :size="18"><Connection /></el-icon>
              <div><b>多平台密钥纳管</b><span>OpenAI / Anthropic 兼容协议统一接入</span></div>
            </li>
            <li>
              <el-icon :size="18"><DataLine /></el-icon>
              <div><b>额度费用自动统计</b><span>Token 用量与费用折算实时聚合</span></div>
            </li>
            <li>
              <el-icon :size="18"><ChatDotRound /></el-icon>
              <div><b>流式对话测试</b><span>SSE 逐字渲染，一键验证密钥可用性</span></div>
            </li>
          </ul>
        </div>

        <p class="brand-foot">轻量 · 安全 · 多用户隔离</p>
      </section>

      <!-- 右侧表单区 -->
      <section class="form-panel">
        <div class="form-inner">
          <h2>{{ activeTab === 'login' ? '欢迎回来' : '创建账号' }}</h2>
          <p class="form-sub">{{ activeTab === 'login' ? '登录后开始管理你的密钥' : '注册即开通普通用户权限' }}</p>

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
      </section>
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

// ---------------- 浅色粒子网络背景 ----------------
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
  const count = Math.min(80, Math.floor((w * h) / 20000))
  const particles: Particle[] = Array.from({ length: count }, () => ({
    x: Math.random() * w,
    y: Math.random() * h,
    vx: (Math.random() - 0.5) * 0.3,
    vy: (Math.random() - 0.5) * 0.3,
    r: Math.random() * 1.5 + 0.6
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
      ctx.fillStyle = 'rgba(37, 99, 235, 0.22)'
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
          ctx.strokeStyle = `rgba(59, 130, 246, ${(1 - dist / LINK_DIST) * 0.09})`
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
/* 页面底：浅色 + 两处淡蓝弥散光 */
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  padding: 24px;
  background:
    radial-gradient(ellipse 55% 40% at 78% 12%, rgba(59, 130, 246, 0.10), transparent),
    radial-gradient(ellipse 45% 38% at 12% 88%, rgba(14, 165, 233, 0.08), transparent),
    #f4f7fc;
}
.bg-canvas {
  position: absolute;
  inset: 0;
}

/* 主体：左品牌 + 右表单 的整体容器（无边框、无阴影轻量风） */
.login-shell {
  position: relative;
  z-index: 1;
  display: flex;
  width: min(980px, 100%);
  min-height: 560px;
  border-radius: 24px;
  overflow: hidden;
}

/* ---------- 左侧品牌区 ---------- */
.brand-panel {
  width: 46%;
  flex-shrink: 0;
  padding: 36px 40px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  background: linear-gradient(165deg, #eff6ff 0%, #dbeafe 55%, #e0f2fe 100%);
  color: #1e3a5f;
  text-align: left;
}
.brand-top {
  display: flex;
  align-items: center;
  gap: 12px;
}
.logo-mark {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}
.logo-mark img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.brand-name {
  font-size: 15px;
  font-weight: 700;
  color: #1d4ed8;
  letter-spacing: 0.5px;
}
.brand-hero h1 {
  margin: 0 0 12px;
  font-size: 27px;
  line-height: 1.4;
  font-weight: 700;
  color: #1e3a8a;
  letter-spacing: 0.5px;
}
.brand-hero > p {
  margin: 0 0 26px;
  font-size: 13.5px;
  color: #5b7aa5;
  line-height: 1.7;
}
.brand-features {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.brand-features li {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}
.brand-features .el-icon {
  flex-shrink: 0;
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #2563eb;
  background: rgba(255, 255, 255, 0.72);
}
.brand-features b {
  display: block;
  font-size: 14px;
  color: #1e3a8a;
  margin-bottom: 2px;
}
.brand-features span {
  font-size: 12.5px;
  color: #64809f;
}
.brand-foot {
  margin: 0;
  font-size: 12px;
  color: #8aa6c4;
  letter-spacing: 1px;
}

/* ---------- 右侧表单区 ---------- */
.form-panel {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #ffffff;
  padding: 36px 44px;
}
.form-inner {
  width: 100%;
  max-width: 340px;
}
.form-inner h2 {
  margin: 0 0 6px;
  font-size: 22px;
  font-weight: 700;
  color: #1f2937;
}
.form-sub {
  margin: 0 0 14px;
  font-size: 13px;
  color: #94a3b8;
}

/* Tabs 定制（浅色） */
.auth-tabs :deep(.el-tabs__nav-wrap::after) {
  height: 1px;
  background: #eef2f7;
}
.auth-tabs :deep(.el-tabs__item) {
  font-size: 15px;
  font-weight: 600;
  color: #94a3b8;
}
.auth-tabs :deep(.el-tabs__item.is-active),
.auth-tabs :deep(.el-tabs__item:hover) {
  color: #2563eb;
}
.auth-tabs :deep(.el-tabs__active-bar) {
  background: #2563eb;
  height: 2.5px;
  border-radius: 2px;
}

/* 输入框定制（浅色、轻边框） */
.auth-tabs :deep(.el-form-item__label) {
  color: #64748b;
  font-size: 13px;
  padding-bottom: 4px;
}
.auth-tabs :deep(.el-input__wrapper) {
  background: #f8fafc;
  border-radius: 10px;
  box-shadow: 0 0 0 1px #e5eaf1 inset;
  padding: 2px 12px;
}
.auth-tabs :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px #c3d4ee inset;
}
.auth-tabs :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1.5px #2563eb inset;
  background: #ffffff;
}
.auth-tabs :deep(.el-input__inner) {
  color: #1f2937;
}
.auth-tabs :deep(.el-input__inner::placeholder) {
  color: #b3c0d1;
}
.auth-tabs :deep(.el-input__prefix),
.auth-tabs :deep(.el-input__suffix) {
  color: #9aa9bd;
}
.auth-tabs :deep(.el-form-item__error) {
  color: #ef4444;
}
.auth-tabs :deep(.el-form-item) {
  margin-bottom: 18px;
}

/* 主按钮：实心蓝 */
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
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  transition: filter 0.15s ease, transform 0.1s ease;
}
.auth-btn:hover:not(:disabled) {
  filter: brightness(1.08);
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
  border: 2px solid rgba(255, 255, 255, 0.4);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
@media (prefers-reduced-motion: reduce) {
  .btn-loading {
    animation-duration: 2s;
  }
}

/* ---------- 响应式 ---------- */
@media (max-width: 880px) {
  /* 窄屏：隐藏品牌区，表单居中铺满 */
  .brand-panel {
    display: none;
  }
  .login-shell {
    width: 100%;
    max-width: 440px;
    min-height: auto;
    border-radius: 20px;
  }
  .form-panel {
    padding: 32px 26px 26px;
  }
}
@media (max-width: 480px) {
  .login-page {
    padding: 14px;
  }
  .form-inner {
    max-width: none;
  }
  .form-inner h2 {
    font-size: 20px;
  }
}
</style>
