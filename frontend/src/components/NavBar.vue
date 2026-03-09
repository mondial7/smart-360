<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
</script>

<template>
  <nav class="navbar">
    <div class="nav-brand">
      <router-link to="/" class="brand-link">Smart 360</router-link>
    </div>
    
    <div class="nav-links">
      <router-link to="/" class="nav-link">Dashboard</router-link>
      <router-link to="/team" class="nav-link">Team</router-link>
      <router-link to="/my-feedback" class="nav-link">My Feedback</router-link>
    </div>
    
    <div class="nav-user">
      <img v-if="auth.user?.photoUrl" :src="auth.user.photoUrl" alt="Profile" class="user-avatar">
      <span class="user-name">{{ auth.user?.name }}</span>
      <button class="logout-btn" @click="auth.logout">Logout</button>
    </div>
  </nav>
</template>

<style scoped>
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 64px;
  background: white;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 2rem;
  z-index: 100;
}

.nav-brand .brand-link {
  font-size: 1.25rem;
  font-weight: 700;
  color: #667eea;
  text-decoration: none;
}

.nav-links {
  display: flex;
  gap: 2rem;
}

.nav-link {
  color: #666;
  text-decoration: none;
  font-weight: 500;
  padding: 0.5rem;
  border-radius: 6px;
  transition: all 0.2s;
}

.nav-link:hover,
.nav-link.router-link-active {
  color: #667eea;
  background: #f5f7fa;
}

.nav-user {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
}

.user-name {
  font-weight: 500;
  color: #333;
}

.logout-btn {
  padding: 0.5rem 1rem;
  border: 1px solid #ddd;
  border-radius: 6px;
  background: white;
  color: #666;
  cursor: pointer;
  transition: all 0.2s;
}

.logout-btn:hover {
  background: #f5f5f5;
  border-color: #ccc;
}
</style>
