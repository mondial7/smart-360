<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <aside class="sidebar">
    <div class="sidebar-header">
      <router-link to="/" class="brand-link">Smart 360</router-link>
    </div>

    <nav class="sidebar-nav">
      <router-link to="/" class="nav-link">
        <span class="nav-icon">📊</span>
        <span>Dashboard</span>
      </router-link>
      <router-link to="/team" class="nav-link">
        <span class="nav-icon">👥</span>
        <span>Team</span>
      </router-link>
      <router-link v-if="auth.isAdmin" to="/teams" class="nav-link">
        <span class="nav-icon">🏢</span>
        <span>Teams</span>
      </router-link>
      <router-link to="/rounds" class="nav-link">
        <span class="nav-icon">🔄</span>
        <span>Rounds</span>
      </router-link>
      <router-link v-if="auth.isAdmin" to="/audit-logs" class="nav-link">
        <span class="nav-icon">📋</span>
        <span>Audit Log</span>
      </router-link>
      <router-link to="/my-feedback" class="nav-link">
        <span class="nav-icon">💬</span>
        <span>My Feedback</span>
      </router-link>
    </nav>

    <div class="sidebar-footer">
      <div class="user-info">
        <img v-if="auth.user?.photoUrl" :src="auth.user.photoUrl" alt="Profile" class="user-avatar">
        <div class="user-details">
          <span class="user-name">{{ auth.user?.name }}</span>
          <button class="logout-btn" @click="handleLogout">Logout</button>
        </div>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: 250px;
  background: white;
  border-right: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;
  z-index: 100;
}

.sidebar-header {
  padding: 1.5rem;
  border-bottom: 1px solid #e0e0e0;
}

.brand-link {
  font-size: 1.5rem;
  font-weight: 700;
  color: #667eea;
  text-decoration: none;
}

.sidebar-nav {
  flex: 1;
  padding: 1rem 0;
  overflow-y: auto;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: #666;
  text-decoration: none;
  font-weight: 500;
  padding: 0.75rem 1.5rem;
  transition: all 0.2s;
}

.nav-link:hover {
  background: #f5f7fa;
  color: #667eea;
}

.nav-link.router-link-active {
  background: #f0f4ff;
  color: #667eea;
  border-right: 3px solid #667eea;
}

.nav-icon {
  font-size: 1.25rem;
  width: 24px;
  text-align: center;
}

.sidebar-footer {
  padding: 1rem;
  border-top: 1px solid #e0e0e0;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}

.user-details {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 0;
}

.user-name {
  font-weight: 500;
  color: #333;
  font-size: 0.875rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.logout-btn {
  padding: 0.375rem 0.75rem;
  border: 1px solid #ddd;
  border-radius: 6px;
  background: white;
  color: #666;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.875rem;
  width: 100%;
}

.logout-btn:hover {
  background: #f5f5f5;
  border-color: #ccc;
}
</style>
