<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { User } from '@/types/user'

const auth = useAuthStore()
const teamMembers = ref<User[]>([])
const loading = ref(true)
const updating = ref<number | null>(null)

onMounted(async () => {
  await loadTeam()
})

async function loadTeam() {
  try {
    const response = await apiClient.get('/users')
    teamMembers.value = response.data
  } catch (error) {
    console.error('Failed to load team:', error)
  } finally {
    loading.value = false
  }
}

async function toggleRole(member: User) {
  if (!auth.isAdmin || updating.value === member.id) return

  const newRole = member.role === 'admin' ? 'member' : 'admin'
  updating.value = member.id

  try {
    await apiClient.put(`/users/${member.id}/role`, { role: newRole })
    member.role = newRole
  } catch (error) {
    console.error('Failed to update role:', error)
    alert('Failed to update role. Please try again.')
  } finally {
    updating.value = null
  }
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'Never'
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

function formatDateTime(dateStr: string | null): string {
  if (!dateStr) return 'Never'
  return new Date(dateStr).toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<template>
  <div class="team-page">
    <header class="page-header">
      <h1>Team</h1>
      <p>Organization roster and member details</p>
    </header>

    <div v-if="loading" class="loading">Loading team members...</div>
    
    <div v-else class="team-grid">
      <div v-for="member in teamMembers" :key="member.id" class="member-card">
        <img v-if="member.photoUrl" :src="member.photoUrl" :alt="member.name" class="member-photo">
        <div v-else class="member-photo-placeholder">{{ member.name.charAt(0).toUpperCase() }}</div>
        
        <div class="member-info">
          <h3>{{ member.name }}</h3>
          <p class="email">{{ member.email }}</p>
          
          <div class="meta">
            <span class="join-date">Joined {{ formatDate(member.createdAt) }}</span>
            <span class="last-login">Last login: {{ formatDateTime(member.lastLogin) }}</span>
          </div>
          
          <div class="role-section">
            <span class="role-badge" :class="member.role">{{ member.role }}</span>
            
            <button 
              v-if="auth.isAdmin && member.id !== auth.user?.id" 
              class="role-toggle-btn"
              :class="{ 'loading': updating === member.id }"
              @click="toggleRole(member)"
              :disabled="updating === member.id"
            >
              {{ member.role === 'admin' ? 'Demote to Member' : 'Promote to Admin' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.team-page {
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

.page-header p {
  color: #666;
}

.loading {
  text-align: center;
  color: #666;
  padding: 3rem;
}

.team-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.5rem;
}

.member-card {
  background: white;
  padding: 1.5rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  display: flex;
  align-items: flex-start;
  gap: 1rem;
}

.member-photo {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  object-fit: cover;
}

.member-photo-placeholder {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: 600;
  flex-shrink: 0;
}

.member-info {
  flex: 1;
  min-width: 0;
}

.member-info h3 {
  font-size: 1.1rem;
  margin-bottom: 0.25rem;
}

.email {
  color: #666;
  font-size: 0.9rem;
  margin-bottom: 0.5rem;
}

.meta {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 0.75rem;
}

.join-date, .last-login {
  font-size: 0.8rem;
  color: #888;
}

.role-section {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.role-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
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

.role-toggle-btn {
  font-size: 0.75rem;
  padding: 0.4rem 0.75rem;
  border: 1px solid #ddd;
  border-radius: 6px;
  background: white;
  color: #666;
  cursor: pointer;
  transition: all 0.2s;
}

.role-toggle-btn:hover {
  background: #f5f5f5;
  border-color: #667eea;
  color: #667eea;
}

.role-toggle-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.role-toggle-btn.loading {
  opacity: 0.6;
}
</style>
