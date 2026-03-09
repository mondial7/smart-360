<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { DashboardStats } from '@/types/dashboard'
import type { FeedbackRound } from '@/types/round'

const auth = useAuthStore()
const stats = ref<DashboardStats | null>(null)
const activeRounds = ref<FeedbackRound[]>([])
const loading = ref(true)

onMounted(async () => {
  await loadDashboard()
})

async function loadDashboard() {
  try {
    const [statsRes, roundsRes] = await Promise.all([
      apiClient.get('/dashboard/stats'),
      apiClient.get('/dashboard/active-rounds')
    ])
    stats.value = statsRes.data
    activeRounds.value = roundsRes.data
  } catch (error) {
    console.error('Failed to load dashboard:', error)
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'No deadline'
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric'
  })
}
</script>

<template>
  <div class="dashboard">
    <header class="page-header">
      <h1>Dashboard</h1>
      <p class="welcome">Welcome back, {{ auth.user?.name }}</p>
    </header>

    <div v-if="loading" class="loading">Loading dashboard...</div>

    <template v-else>
      <!-- Admin Dashboard -->
      <template v-if="auth.isAdmin">
        <div class="stats-grid">
          <div class="stat-card">
            <span class="stat-value">{{ stats?.totalUsers || 0 }}</span>
            <span class="stat-label">Total Users</span>
            <div class="stat-breakdown">
              <span>{{ stats?.adminCount || 0 }} admins</span>
              <span>{{ stats?.memberCount || 0 }} members</span>
            </div>
          </div>
          <div class="stat-card">
            <span class="stat-value">{{ stats?.totalRounds || 0 }}</span>
            <span class="stat-label">Total Rounds</span>
          </div>
          <div class="stat-card accent">
            <span class="stat-value">{{ stats?.activeRounds || 0 }}</span>
            <span class="stat-label">Active Rounds</span>
          </div>
          <div class="stat-card" v-if="stats?.pendingReviews">
            <span class="stat-value">{{ stats.pendingReviews }}</span>
            <span class="stat-label">Pending Reviews</span>
          </div>
        </div>

        <div class="dashboard-section" v-if="activeRounds.length > 0">
          <h2>Active Rounds</h2>
          <div class="rounds-preview">
            <div v-for="round in activeRounds.slice(0, 3)" :key="round.id" class="round-item">
              <div class="round-subject">
                <img v-if="round.subject?.photoUrl" :src="round.subject.photoUrl" class="mini-avatar">
                <div v-else class="mini-avatar-placeholder">{{ round.subject?.name.charAt(0) }}</div>
                <span>{{ round.subject?.name }}</span>
              </div>
              <div class="round-meta">
                <span class="deadline">Due {{ formatDate(round.deadline) }}</span>
                <span class="reviewer-count">{{ round.reviewers?.length || 0 }} reviewers</span>
              </div>
            </div>
          </div>
          <router-link to="/rounds" class="view-all">View all rounds →</router-link>
        </div>

        <div class="quick-actions">
          <h2>Quick Actions</h2>
          <div class="action-grid">
            <router-link to="/team" class="action-card">
              <span class="action-icon">👥</span>
              <span class="action-text">Manage Team</span>
            </router-link>
            <router-link to="/rounds/new" class="action-card primary">
              <span class="action-icon">➕</span>
              <span class="action-text">Create Round</span>
            </router-link>
            <router-link to="/rounds" class="action-card">
              <span class="action-icon">📋</span>
              <span class="action-text">View Rounds</span>
            </router-link>
          </div>
        </div>
      </template>

      <!-- Team Member Dashboard -->
      <template v-else>
        <div class="welcome-card">
          <h2>Welcome to Smart 360 Feedback</h2>
          <p>This platform helps you receive anonymous feedback from your peers to support your professional growth.</p>
        </div>

        <div class="stats-grid member">
          <div class="stat-card accent" v-if="stats?.pendingReviews">
            <span class="stat-value">{{ stats.pendingReviews }}</span>
            <span class="stat-label">Pending Reviews</span>
            <router-link v-if="stats.pendingReviews > 0" to="/rounds" class="action-link">Review now →</router-link>
          </div>
          <div class="stat-card">
            <span class="stat-value">{{ stats?.myFeedbackCount || 0 }}</span>
            <span class="stat-label">My Feedback</span>
            <router-link to="/my-feedback" class="action-link">View →</router-link>
          </div>
        </div>

        <div class="quick-actions">
          <h2>Quick Links</h2>
          <div class="action-grid">
            <router-link to="/team" class="action-card">
              <span class="action-icon">👥</span>
              <span class="action-text">View Team</span>
            </router-link>
            <router-link to="/my-feedback" class="action-card">
              <span class="action-icon">📊</span>
              <span class="action-text">My Feedback</span>
            </router-link>
            <router-link to="/rounds" class="action-card" v-if="stats?.pendingReviews && stats.pendingReviews > 0">
              <span class="action-icon">✏️</span>
              <span class="action-text">Pending Reviews</span>
            </router-link>
          </div>
        </div>
      </template>
    </template>
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

.loading {
  text-align: center;
  color: #666;
  padding: 3rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.stats-grid.member {
  grid-template-columns: repeat(2, 1fr);
  max-width: 600px;
}

.stat-card {
  background: white;
  padding: 1.5rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  display: flex;
  flex-direction: column;
}

.stat-card.accent {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.stat-value {
  font-size: 2.5rem;
  font-weight: 700;
  line-height: 1;
}

.stat-label {
  font-size: 0.9rem;
  color: #666;
  margin-top: 0.5rem;
}

.stat-card.accent .stat-label {
  color: rgba(255,255,255,0.8);
}

.stat-breakdown {
  display: flex;
  gap: 1rem;
  margin-top: 0.75rem;
  font-size: 0.8rem;
  color: #888;
}

.dashboard-section {
  background: white;
  padding: 1.5rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  margin-bottom: 2rem;
}

.dashboard-section h2 {
  margin-bottom: 1rem;
  font-size: 1.25rem;
}

.rounds-preview {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.round-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 8px;
}

.round-subject {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.mini-avatar, .mini-avatar-placeholder {
  width: 32px;
  height: 32px;
  border-radius: 50%;
}

.mini-avatar {
  object-fit: cover;
}

.mini-avatar-placeholder {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  font-weight: 600;
}

.round-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.25rem;
}

.deadline {
  font-size: 0.85rem;
  color: #f44336;
  font-weight: 500;
}

.reviewer-count {
  font-size: 0.8rem;
  color: #666;
}

.view-all {
  display: block;
  margin-top: 1rem;
  text-align: center;
  color: #667eea;
  text-decoration: none;
  font-weight: 500;
}

.welcome-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 2rem;
  border-radius: 12px;
  margin-bottom: 2rem;
}

.welcome-card h2 {
  font-size: 1.5rem;
  margin-bottom: 0.5rem;
}

.welcome-card p {
  opacity: 0.9;
}

.quick-actions {
  margin-top: 2rem;
}

.quick-actions h2 {
  margin-bottom: 1rem;
  font-size: 1.25rem;
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1rem;
}

.action-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 1.5rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  text-decoration: none;
  color: #333;
  transition: transform 0.2s, box-shadow 0.2s;
}

.action-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.12);
}

.action-card.primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.action-icon {
  font-size: 1.5rem;
}

.action-text {
  font-size: 0.9rem;
  font-weight: 500;
}

.action-link {
  font-size: 0.85rem;
  color: #667eea;
  text-decoration: none;
  margin-top: 0.5rem;
}

.stat-card.accent .action-link {
  color: rgba(255,255,255,0.9);
}
</style>
