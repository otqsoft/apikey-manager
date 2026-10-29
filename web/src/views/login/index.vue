<template>
  <div class="login-page">
    <el-card class="login-card">
      <div class="login-header">
        <el-icon class="logo-icon"><Key /></el-icon>
        <h2>多平台大模型 API-KEY 集中管理工具</h2>
        <p>统一纳管供应商、模型、密钥额度与对话测试</p>
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
            <el-button type="primary" size="large" class="auth-btn" :loading="loginLoading" @click="onLogin">
              登 录
            </el-button>
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
            <el-button type="primary" size="large" class="auth-btn" :loading="regLoading" @click="onRegister">
              注 册
            </el-button>
          </el-form>
        </el-tab-pane>
      </el-tabs>

      <div class="login-tip">注册成功后请使用新账号登录；内置管理员：admin / admin123（首次启动由后端初始化）</div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { Avatar, Key, Lock, User } from '@element-plus/icons-vue'
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
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1d4fc0 0%, #2563eb 55%, #60a5fa 100%);
  position: relative;
  overflow: hidden;
  padding: 24px 0;
}
.login-page::before {
  content: '';
  position: absolute;
  width: 480px;
  height: 480px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.08);
  top: -160px;
  right: -120px;
}
.login-page::after {
  content: '';
  position: absolute;
  width: 360px;
  height: 360px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.06);
  bottom: -140px;
  left: -100px;
}
.login-card {
  position: relative;
  z-index: 1;
  width: 420px;
  max-width: calc(100vw - 32px);
  padding: 22px 24px 20px;
  border-radius: 16px;
  border: none;
  box-shadow: 0 18px 48px rgba(15, 23, 42, 0.25);
}
.login-header {
  text-align: center;
  margin-bottom: 14px;
}
.logo-icon {
  font-size: 40px;
  color: #2563eb;
}
.login-header h2 {
  font-size: 19px;
  margin: 12px 0 6px;
  color: #1f2937;
  font-weight: 700;
  line-height: 1.4;
}
.login-header p {
  margin: 0;
  color: #94a3b8;
  font-size: 13px;
}
.auth-tabs {
  margin-bottom: 6px;
}
.auth-tabs :deep(.el-tabs__item) {
  font-size: 15px;
  font-weight: 600;
}
.auth-btn {
  width: 100%;
  margin-top: 8px;
  border-radius: 8px;
  border: none;
  background: linear-gradient(135deg, #2563eb 0%, #1d4fc0 100%);
}
.auth-btn:hover {
  background: linear-gradient(135deg, #1d4fc0 0%, #1a3fa0 100%);
}
/* 输入框聚焦蓝色边框（主色系） */
.auth-tabs :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #2563eb inset;
}
.login-tip {
  margin-top: 12px;
  text-align: center;
  font-size: 12px;
  color: #94a3b8;
  line-height: 1.6;
}
/* 窄屏自适应 */
@media (max-width: 480px) {
  .login-card {
    padding: 18px 16px 16px;
    border-radius: 14px;
  }
  .login-header h2 {
    font-size: 16px;
  }
  .login-header p {
    font-size: 12px;
  }
}
</style>
