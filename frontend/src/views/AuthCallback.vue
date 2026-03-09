<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

onMounted(() => {
  const token = route.query.token as string
  if (token) {
    auth.setToken(token)
    auth.fetchUser().then(() => {
      router.push('/')
    })
  } else {
    router.push('/login')
  }
})
</script>

<template>
  <div class="callback-container">
    <div class="loading-spinner"></div>
    <p>Authenticating...</p>
  </div>
</template>

<style scoped>
.callback-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #f3f3f3;
  border-top: 3px solid #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

p {
  color: #666;
}
</style>
