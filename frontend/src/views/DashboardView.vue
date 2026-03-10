<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { DashboardStats } from '@/types/dashboard'
import type { FeedbackRound } from '@/types/round'

const auth = useAuthStore()
const stats = ref<DashboardStats | null>(null)
const activeRounds = ref<FeedbackRound[]>([])
const myRounds = ref<FeedbackRound[]>([])
const submissionStatus = ref<Record<string, boolean>>({})
const loading = ref(true)

onMounted(async () => {
  await loadDashboard()
})

async function loadDashboard() {
  try {
    if (auth.isAdmin) {
      const [statsRes, roundsRes] = await Promise.all([
        apiClient.get('/dashboard/stats'),
        apiClient.get('/dashboard/active-rounds')
      ])
      stats.value = statsRes.data
      activeRounds.value = roundsRes.data
    } else {
      // Member dashboard - load rounds where they're reviewers
      const [statsRes, roundsRes] = await Promise.all([
        apiClient.get('/dashboard/stats'),
        apiClient.get('/rounds-for-me')
      ])
      stats.value = statsRes.data
      myRounds.value = roundsRes.data
      
      // Check submission status for each round
      for (const round of myRounds.value) {
        try {
          const checkRes = await apiClient.get(`/submissions/check/${round.id}`)
          submissionStatus.value[round.id] = checkRes.data.submitted
        } catch {
          submissionStatus.value[round.id] = false
        }
      }
    }
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

function getDeadlineStatus(deadline: string | null): string {
  if (!deadline) return 'no-deadline'
  const deadlineDate = new Date(deadline)
  const today = new Date()
  const daysLeft = Math.ceil((deadlineDate.getTime() - today.getTime()) / (1000 * 60 * 60 * 24))
  
  if (daysLeft < 0) return 'overdue'
  if (daysLeft === 0) return 'due-today'
  if (daysLeft <= 3) return 'due-soon'
  return 'on-time'
}

function getRoundStatusText(round: FeedbackRound): string {
  if (submissionStatus.value[round.id]) return 'Submitted'
  if (round.status !== 'active') return 'Closed'
  if (getDeadlineStatus(round.deadline) === 'overdue') return 'Overdue'
  return 'Pending'
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
        <!-- Stats Overview -->
        <div class="stats-grid member">
          <div class="stat-card accent" v-if="myRounds.filter(r => r.status === 'active' && !submissionStatus[r.id]).length > 0">
            <span class="stat-value">{{ myRounds.filter(r => r.status === 'active' && !submissionStatus[r.id]).length }}</span>
            <span class="stat-label">Pending Reviews</span>
            <router-link to="/rounds" class="action-link">Review now →</router-link>
          </div>
          <div class="stat-card">
            <span class="stat-value">{{ myRounds.filter(r => submissionStatus[r.id]).length }}</span>
            <span class="stat-label">Completed Reviews</span>
          </div>
          <div class="stat-card">
            <span class="stat-value">{{ stats?.myFeedbackCount || 0 }}</span>
            <span class="stat-label">My Feedback</span>
            <router-link to="/my-feedback" class="action-link">View →</router-link>
          </div>
        </div>

        <!-- Feedback Requests -->
        <div class="feedback-section" v-if="myRounds.length > 0">
          <h2>Feedback Requests</h2>
          <div class="feedback-list">
            <div v-for="round in myRounds" :key="round.id" class="feedback-item">
              <div class="feedback-header">
                <div class="person-info">
                  <img v-if="round.subject?.photoUrl" :src="round.subject.photoUrl" class="avatar">
                  <div v-else class="avatar-placeholder">{{ round.subject?.name.charAt(0) }}</div>
                  <div>
                    <h3>{{ round.subject?.name }}</h3>
                    <span :class="['status', getRoundStatusText(round).toLowerCase()]">
                      {{ getRoundStatusText(round) }}
                    </span>
                  </div>
                </div>
                <div class="deadline">
                  <span class="label">Due</span>
                  <span :class="getDeadlineStatus(round.deadline)">
                    {{ formatDate(round.deadline) }}
                  </span>
                </div>
              </div>
              
              <div class="feedback-actions">
                <template v-if="!submissionStatus[round.id] && round.status === 'active'">
                  <router-link :to="`/rounds/${round.id}/submit`" class="btn-primary">
                    Submit Feedback
                  </router-link>
                </template>
                <template v-else-if="submissionStatus[round.id]">
                  <router-link :to="`/rounds/${round.id}/submission`" class="btn-secondary">
                    View Submission
                  </router-link>
                </template>
                <template v-else>
                  <span class="closed-text">Round Closed</span>
                </template>
              </div>
            </div>
          </div>
        </div>

        <!-- No Feedback Requests -->
        <div class="empty-state" v-else>
          <div class="empty-icon">📋</div>
          <h2>No Feedback Requests</h2>
          <p>You don't have any pending feedback requests at the moment.</p>
          <router-link to="/team" class="btn-primary">View Team</router-link>
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

.view-all:hover {
  color: #667eea;
  text-decoration: underline;
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
  font-size: 0.85rem;
  font-weight: 500;
}

/* Member Dashboard Styles */
.feedback-section {
  margin-top: 2rem;
}

.feedback-section h2 {
  font-size: 1.5rem;
  margin-bottom: 1.5rem;
  color: #333;
}

.feedback-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.feedback-item {
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 1.5rem;
  transition: border-color 0.2s;
}

.feedback-item:hover {
  border-color: #667eea;
}

.feedback-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.person-info {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
}

.avatar-placeholder {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #667eea;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
}

.person-info h3 {
  margin: 0 0 0.25rem 0;
  font-size: 1rem;
  color: #333;
}

.status {
  font-size: 0.85rem;
  font-weight: 500;
  text-transform: capitalize;
}

.status.pending {
  color: #ff6b35;
}

.status.submitted {
  color: #51cf66;
}

.status.overdue {
  color: #ff6348;
}

.status.closed {
  color: #868e96;
}

.deadline {
  text-align: right;
}

.deadline .label {
  font-size: 0.75rem;
  color: #868e96;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.deadline span:last-child {
  font-size: 0.9rem;
  font-weight: 500;
  display: block;
  margin-top: 0.25rem;
}

.deadline .overdue {
  color: #ff6348;
}

.deadline .due-today {
  color: #ff6b35;
}

.deadline .due-soon {
  color: #ff9800;
}

.deadline .on-time {
  color: #51cf66;
}

.feedback-actions {
  display: flex;
  justify-content: flex-end;
}

.btn-primary {
  display: inline-block;
  padding: 0.5rem 1rem;
  background: #667eea;
  color: white;
  text-decoration: none;
  border-radius: 6px;
  font-weight: 500;
  font-size: 0.9rem;
  transition: background 0.2s;
  border: none;
}

.btn-primary:hover {
  background: #5a6fd6;
}

.btn-secondary {
  display: inline-block;
  padding: 0.5rem 1rem;
  background: white;
  color: #666;
  text-decoration: none;
  border-radius: 6px;
  font-weight: 500;
  font-size: 0.9rem;
  transition: background 0.2s;
  border: 1px solid #ddd;
}

.btn-secondary:hover {
  background: #f8f9fa;
}

.closed-text {
  color: #868e96;
  font-size: 0.9rem;
  font-style: italic;
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
