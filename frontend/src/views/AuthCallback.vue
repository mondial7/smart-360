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
  <div class="auth-callback">
    <div class="auth-callback__spinner"></div>
    <p class="auth-callback__text">Authenticating...</p>
  </div>
</template>

<style scoped lang="scss">
.auth-callback {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding: 1rem;

  &__spinner {
    width: 36px;
    height: 36px;
    border: 3px solid var(--border-color);
    border-top: 3px solid var(--color-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;

    @media (min-width: 768px) {
      width: 40px;
      height: 40px;
    }
  }

  &__text {
    color: var(--text-secondary);
    margin: 0;
    font-size: 0.95rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
</style>
