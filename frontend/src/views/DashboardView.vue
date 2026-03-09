<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
</script>

<template>
  <div class="dashboard">
    <header class="page-header">
      <h1>Dashboard</h1>
      <p class="welcome">Welcome back, {{ auth.user?.name }}</p>
    </header>

    <div class="dashboard-grid">
      <div class="card stats-card">
        <h3>Your Role</h3>
        <span class="role-badge" :class="auth.user?.role">
          {{ auth.user?.role === 'admin' ? 'Administrator' : 'Team Member' }}
        </span>
      </div>
      
      <div class="card info-card">
        <h3>Quick Links</h3>
        <div class="links">
          <router-link to="/team" class="link">View Team</router-link>
          <router-link v-if="auth.isAdmin" to="/rounds/new" class="link">Create Feedback Round</router-link>
          <router-link to="/my-feedback" class="link">My Feedback</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

.welcome {
  color: #666;
}

.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1.5rem;
}

.card {
  background: white;
  padding: 1.5rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}

.card h3 {
  margin-bottom: 1rem;
  color: #333;
}

.role-badge {
  display: inline-block;
  padding: 0.5rem 1rem;
  border-radius: 20px;
  font-weight: 500;
  text-transform: capitalize;
}

.role-badge.admin {
  background: #e3f2fd;
  color: #1976d2;
}

.role-badge.member {
  background: #f3e5f5;
  color: #7b1fa2;
}

.links {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.link {
  color: #667eea;
  text-decoration: none;
  padding: 0.5rem;
  border-radius: 6px;
  transition: background 0.2s;
}

.link:hover {
  background: #f5f7fa;
}
</style>
