<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { Team } from '@/types/team'

const router = useRouter()
const auth = useAuthStore()
const teams = ref<Team[]>([])
const loading = ref(true)

onMounted(async () => {
  await loadTeams()
})

async function loadTeams() {
  try {
    const response = await apiClient.get('/teams')
    teams.value = response.data
  } catch (error) {
    console.error('Failed to load teams:', error)
  } finally {
    loading.value = false
  }
}

function createTeam() {
  router.push('/teams/new')
}

function viewTeam(teamId: string) {
  router.push(`/teams/${teamId}`)
}

function createTeamRound(teamId: string) {
  router.push(`/teams/${teamId}/create-round`)
}
</script>

<template>
  <div class="teams-page">
    <header class="page-header">
      <h1>Teams</h1>
      <button v-if="auth.isAdmin" @click="createTeam" class="btn-primary">
        Create Team
      </button>
    </header>

    <div v-if="loading" class="loading">Loading teams...</div>

    <div v-else-if="teams.length === 0" class="empty-state">
      <p>No teams have been created yet.</p>
      <button v-if="auth.isAdmin" @click="createTeam" class="btn-primary">
        Create Your First Team
      </button>
    </div>

    <div v-else class="teams-grid">
      <div v-for="team in teams" :key="team.id" class="team-card">
        <div class="team-card-header">
          <h3>{{ team.name }}</h3>
        </div>

        <div class="team-admin">
          <div class="admin-label">Team Admin</div>
          <div class="admin-info">
            <img
              v-if="team.teamAdmin?.photoUrl"
              :src="team.teamAdmin.photoUrl"
              :alt="team.teamAdmin.name"
              class="admin-photo"
            >
            <div v-else class="admin-photo-placeholder">
              {{ team.teamAdmin?.name.charAt(0).toUpperCase() }}
            </div>
            <div class="admin-details">
              <div class="admin-name">{{ team.teamAdmin?.name }}</div>
              <div class="admin-email">{{ team.teamAdmin?.email }}</div>
            </div>
          </div>
        </div>

        <div class="team-stats">
          <div class="stat">
            <span class="stat-value">{{ team.members?.length || 0 }}</span>
            <span class="stat-label">Members</span>
          </div>
        </div>

        <div class="team-actions">
          <button @click="viewTeam(team.id)" class="btn-secondary">
            View Team
          </button>
          <button @click="createTeamRound(team.id)" class="btn-primary">
            Create Team Round
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.teams-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 2rem;
  color: #333;
  margin: 0;
}

.loading {
  text-align: center;
  padding: 3rem;
  color: #666;
  font-size: 1.1rem;
}

.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.empty-state p {
  color: #666;
  font-size: 1.1rem;
  margin-bottom: 1.5rem;
}

.teams-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1.5rem;
}

.team-card {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: box-shadow 0.2s;
}

.team-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
}

.team-card-header h3 {
  margin: 0 0 1rem 0;
  font-size: 1.5rem;
  color: #333;
}

.team-admin {
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #e0e0e0;
}

.admin-label {
  font-size: 0.85rem;
  color: #666;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 0.5rem;
}

.admin-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.admin-photo {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
}

.admin-photo-placeholder {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.2rem;
  font-weight: bold;
}

.admin-details {
  flex: 1;
}

.admin-name {
  font-weight: 600;
  color: #333;
  margin-bottom: 0.25rem;
}

.admin-email {
  font-size: 0.85rem;
  color: #666;
}

.team-stats {
  display: flex;
  gap: 2rem;
  margin-bottom: 1rem;
}

.stat {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: bold;
  color: #667eea;
}

.stat-label {
  font-size: 0.85rem;
  color: #666;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.team-actions {
  display: flex;
  gap: 0.75rem;
}

.btn-primary,
.btn-secondary {
  flex: 1;
  padding: 0.75rem 1rem;
  border: none;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #667eea;
  color: white;
}

.btn-primary:hover {
  background: #5568d3;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.btn-secondary {
  background: white;
  color: #667eea;
  border: 2px solid #667eea;
}

.btn-secondary:hover {
  background: #f5f7ff;
}
</style>
