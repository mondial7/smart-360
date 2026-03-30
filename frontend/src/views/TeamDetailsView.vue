<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { Team } from '@/types/team'
import type { User } from '@/types/user'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const team = ref<Team | null>(null)
const loading = ref(true)
const showEditModal = ref(false)
const editTeamName = ref('')

onMounted(async () => {
  await loadTeam()
})

async function loadTeam() {
  try {
    const response = await apiClient.get(`/teams/${route.params.id}`)
    team.value = response.data
    editTeamName.value = response.data.name
  } catch (error) {
    console.error('Failed to load team:', error)
  } finally {
    loading.value = false
  }
}

const canManage = computed(() => {
  if (!auth.user || !team.value) return false
  return auth.isAdmin || (auth.user.role === 'team_admin' && auth.user.teamId === team.value.id)
})

function createTeamRound() {
  router.push(`/teams/${team.value!.id}/create-round`)
}

function openEditModal() {
  showEditModal.value = true
}

function closeEditModal() {
  showEditModal.value = false
  editTeamName.value = team.value!.name
}

async function saveTeamName() {
  if (!team.value || !editTeamName.value.trim()) return

  try {
    await apiClient.put(`/teams/${team.value.id}`, {
      name: editTeamName.value.trim()
    })
    await loadTeam()
    closeEditModal()
  } catch (error: any) {
    console.error('Failed to update team name:', error)
    alert(error.response?.data?.error || 'Failed to update team name')
  }
}

async function removeMember(userId: string) {
  if (!team.value) return
  if (!confirm('Are you sure you want to remove this member from the team?')) return

  try {
    await apiClient.delete(`/teams/${team.value.id}/members/${userId}`)
    await loadTeam()
  } catch (error: any) {
    console.error('Failed to remove member:', error)
    alert(error.response?.data?.error || 'Failed to remove member')
  }
}

function goBack() {
  router.push('/teams')
}
</script>

<template>
  <div class="team-details-page">
    <div v-if="loading" class="loading">Loading team...</div>

    <div v-else-if="!team" class="error">Team not found</div>

    <div v-else>
      <header class="page-header">
        <button @click="goBack" class="back-btn">← Back to Teams</button>
        <div class="header-actions">
          <button v-if="canManage" @click="openEditModal" class="btn-secondary">
            Edit Name
          </button>
          <button @click="createTeamRound" class="btn-primary">
            Create Team Round
          </button>
        </div>
      </header>

      <div class="team-header">
        <h1>{{ team.name }}</h1>
        <div class="team-meta">
          <span>{{ team.members.length }} members</span>
        </div>
      </div>

      <div class="content-grid">
        <!-- Team Admin Card -->
        <div class="card">
          <h3>Team Admin</h3>
          <div v-if="team.teamAdmin" class="admin-card">
            <img
              v-if="team.teamAdmin.photoUrl"
              :src="team.teamAdmin.photoUrl"
              :alt="team.teamAdmin.name"
              class="admin-photo"
            >
            <div v-else class="admin-photo-placeholder">
              {{ team.teamAdmin.name.charAt(0).toUpperCase() }}
            </div>
            <div class="admin-info">
              <div class="admin-name">{{ team.teamAdmin.name }}</div>
              <div class="admin-email">{{ team.teamAdmin.email }}</div>
              <div class="admin-role">{{ team.teamAdmin.role }}</div>
            </div>
          </div>
        </div>

        <!-- Team Members Card -->
        <div class="card members-card">
          <div class="card-header">
            <h3>Team Members</h3>
            <span class="member-count">{{ team.members.length }}</span>
          </div>
          <div class="members-list">
            <div
              v-for="member in team.members"
              :key="member.id"
              class="member-item"
            >
              <img
                v-if="member.photoUrl"
                :src="member.photoUrl"
                :alt="member.name"
                class="member-photo"
              >
              <div v-else class="member-photo-placeholder">
                {{ member.name.charAt(0).toUpperCase() }}
              </div>
              <div class="member-info">
                <div class="member-name">{{ member.name }}</div>
                <div class="member-email">{{ member.email }}</div>
              </div>
              <span v-if="member.id === team.teamAdminId" class="admin-badge">Admin</span>
              <button
                v-else-if="canManage"
                @click="removeMember(member.id)"
                class="remove-btn"
                title="Remove member"
              >
                ×
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Edit Name Modal -->
    <div v-if="showEditModal" class="modal-overlay" @click="closeEditModal">
      <div class="modal" @click.stop>
        <h3>Edit Team Name</h3>
        <input
          v-model="editTeamName"
          type="text"
          placeholder="Team name"
          class="modal-input"
          @keyup.enter="saveTeamName"
          autofocus
        >
        <div class="modal-actions">
          <button @click="closeEditModal" class="btn-secondary">Cancel</button>
          <button @click="saveTeamName" class="btn-primary">Save</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.team-details-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

.loading,
.error {
  text-align: center;
  padding: 3rem;
  color: #666;
  font-size: 1.1rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.back-btn {
  background: none;
  border: none;
  color: #667eea;
  font-size: 1rem;
  cursor: pointer;
  padding: 0.5rem 0;
}

.back-btn:hover {
  text-decoration: underline;
}

.header-actions {
  display: flex;
  gap: 1rem;
}

.team-header {
  margin-bottom: 2rem;
}

.team-header h1 {
  font-size: 2.5rem;
  color: #333;
  margin-bottom: 0.5rem;
}

.team-meta {
  color: #666;
  font-size: 1.1rem;
}

.content-grid {
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: 2rem;
}

@media (max-width: 768px) {
  .content-grid {
    grid-template-columns: 1fr;
  }
}

.card {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.card h3 {
  font-size: 1.2rem;
  color: #333;
  margin-bottom: 1rem;
}

.admin-card {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.admin-photo {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  object-fit: cover;
}

.admin-photo-placeholder {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: bold;
}

.admin-info {
  flex: 1;
}

.admin-name {
  font-size: 1.2rem;
  font-weight: 600;
  color: #333;
  margin-bottom: 0.25rem;
}

.admin-email {
  color: #666;
  margin-bottom: 0.25rem;
}

.admin-role {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  background: #667eea;
  color: white;
  border-radius: 4px;
  font-size: 0.75rem;
  text-transform: uppercase;
  font-weight: 500;
}

.members-card {
  grid-column: span 1;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.card-header h3 {
  margin: 0;
}

.member-count {
  background: #f0f0f0;
  color: #666;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.9rem;
  font-weight: 600;
}

.members-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.member-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem;
  background: #f8f9fa;
  border-radius: 8px;
  transition: background 0.2s;
}

.member-item:hover {
  background: #f0f2f5;
}

.member-photo {
  width: 45px;
  height: 45px;
  border-radius: 50%;
  object-fit: cover;
}

.member-photo-placeholder {
  width: 45px;
  height: 45px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.1rem;
  font-weight: bold;
}

.member-info {
  flex: 1;
}

.member-name {
  font-weight: 600;
  color: #333;
  margin-bottom: 0.25rem;
}

.member-email {
  font-size: 0.85rem;
  color: #666;
}

.admin-badge {
  padding: 0.25rem 0.5rem;
  background: #667eea;
  color: white;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.remove-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: #fee;
  color: #c33;
  border-radius: 50%;
  font-size: 1.5rem;
  line-height: 1;
  cursor: pointer;
  transition: all 0.2s;
}

.remove-btn:hover {
  background: #fcc;
}

.btn-primary,
.btn-secondary {
  padding: 0.75rem 1.5rem;
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

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 12px;
  padding: 2rem;
  max-width: 500px;
  width: 90%;
}

.modal h3 {
  margin: 0 0 1rem 0;
  font-size: 1.5rem;
  color: #333;
}

.modal-input {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 1rem;
  margin-bottom: 1.5rem;
}

.modal-input:focus {
  outline: none;
  border-color: #667eea;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}
</style>
