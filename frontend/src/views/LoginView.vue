<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

const devUsers = [
  { email: 'admin@example.com', name: 'Emma Admin', role: 'Admin' },
  { email: 'alice@example.com', name: 'Alice Johnson', role: 'Team Admin' },
  { email: 'bob@example.com', name: 'Bob Smith', role: 'Member' },
  { email: 'carol@example.com', name: 'Carol Williams', role: 'Member' },
  { email: 'david@example.com', name: 'David Brown', role: 'Team Admin' },
  { email: 'eve@example.com', name: 'Eve Martinez', role: 'Member' },
]
</script>

<template>
  <div class="login">
    <div class="login__card">
      <h1 class="login__title">Smart 360 Feedback</h1>
      <p class="login__subtitle">Anonymous peer feedback platform</p>

      <button @click="auth.loginWithGoogle" class="login__google-btn">
        <svg class="login__google-icon" viewBox="0 0 24 24">
          <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
          <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
          <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
          <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
        </svg>
        Sign in with Google
      </button>

      <div class="login__divider">
        <span class="login__divider-text">or</span>
      </div>

      <p class="login__dev-title">Dev Login (No Google)</p>
      <div class="login__dev-users">
        <button
          v-for="user in devUsers"
          :key="user.email"
          @click="auth.devLogin(user.email)"
          :class="['login__dev-user', { 'login__dev-user--admin': user.role === 'Admin' }]"
        >
          <span class="login__dev-user-name">{{ user.name }}</span>
          <span class="login__dev-user-role">{{ user.role }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.login {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
  padding: 1rem;

  &__card {
    background: var(--bg-primary);
    padding: 2rem;
    border-radius: 16px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
    text-align: center;
    max-width: 400px;
    width: 100%;

    @media (min-width: 640px) {
      padding: 3rem;
    }
  }

  &__title {
    font-size: 1.75rem;
    margin-bottom: 0.5rem;
    color: var(--text-primary);

    @media (min-width: 640px) {
      font-size: 2rem;
    }
  }

  &__subtitle {
    color: var(--text-secondary);
    margin-bottom: 2rem;
  }

  &__google-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    width: 100%;
    padding: 14px 24px;
    border: 1px solid var(--border-color);
    border-radius: 8px;
    background: var(--bg-primary);
    font-size: 16px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    color: var(--text-primary);
    min-height: 44px;

    &:hover {
      background: var(--bg-secondary);
    }
  }

  &__google-icon {
    width: 20px;
    height: 20px;
  }

  &__divider {
    margin: 1.5rem 0;
    position: relative;
    text-align: center;

    &::before,
    &::after {
      content: '';
      position: absolute;
      top: 50%;
      width: calc(50% - 40px);
      height: 1px;
      background: var(--border-color);
    }

    &::before {
      left: 0;
    }

    &::after {
      right: 0;
    }
  }

  &__divider-text {
    color: var(--text-tertiary);
    font-size: 0.9rem;
    background: var(--bg-primary);
    padding: 0 0.5rem;
  }

  &__dev-title {
    font-size: 0.9rem;
    color: var(--text-secondary);
    margin-bottom: 0.75rem;
  }

  &__dev-users {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.5rem;
  }

  &__dev-user {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 0.75rem;
    border: 1px solid var(--border-color);
    border-radius: 8px;
    background: var(--bg-primary);
    cursor: pointer;
    transition: all 0.2s;
    min-height: 44px;

    &:hover {
      border-color: var(--color-primary);
      background: var(--bg-secondary);
    }

    &--admin {
      border-color: var(--color-primary);
      background: var(--bg-secondary);
    }
  }

  &__dev-user-name {
    font-size: 0.85rem;
    font-weight: 500;
    color: var(--text-primary);
  }

  &__dev-user-role {
    font-size: 0.7rem;
    color: var(--text-tertiary);
    text-transform: uppercase;
    margin-top: 0.25rem;
  }
}
</style>
