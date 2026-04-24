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
  <div class="team-details">
    <div v-if="loading" class="team-details__loading">Loading team...</div>

    <div v-else-if="!team" class="team-details__error">Team not found</div>

    <div v-else>
      <header class="team-details__page-header">
        <button @click="goBack" class="team-details__back">← Back to Teams</button>
        <div class="team-details__header-actions">
          <button v-if="canManage" @click="openEditModal" class="btn btn--secondary">
            Edit Name
          </button>
          <button @click="createTeamRound" class="btn btn--primary">
            Create Team Round
          </button>
        </div>
      </header>

      <div class="team-details__header">
        <h1 class="team-details__title">{{ team.name }}</h1>
        <div class="team-details__meta">
          <span>{{ team.members.length }} members</span>
        </div>
      </div>

      <div class="team-details__content">
        <!-- Team Admin Card -->
        <div class="admin-card">
          <h3 class="admin-card__title">Team Admin</h3>
          <div v-if="team.teamAdmin" class="admin-card__content">
            <img
              v-if="team.teamAdmin.photoUrl"
              :src="team.teamAdmin.photoUrl"
              :alt="team.teamAdmin.name"
              class="admin-card__photo"
            >
            <div v-else class="admin-card__photo-placeholder">
              {{ team.teamAdmin.name.charAt(0).toUpperCase() }}
            </div>
            <div class="admin-card__info">
              <div class="admin-card__name">{{ team.teamAdmin.name }}</div>
              <div class="admin-card__email">{{ team.teamAdmin.email }}</div>
              <div class="admin-card__role">{{ team.teamAdmin.role }}</div>
            </div>
          </div>
        </div>

        <!-- Team Members Card -->
        <div class="members-card">
          <div class="members-card__header">
            <h3 class="members-card__title">Team Members</h3>
            <span class="members-card__count">{{ team.members.length }}</span>
          </div>
          <div class="members-card__list">
            <div
              v-for="member in team.members"
              :key="member.id"
              class="member"
            >
              <img
                v-if="member.photoUrl"
                :src="member.photoUrl"
                :alt="member.name"
                class="member__photo"
              >
              <div v-else class="member__photo-placeholder">
                {{ member.name.charAt(0).toUpperCase() }}
              </div>
              <div class="member__info">
                <div class="member__name">{{ member.name }}</div>
                <div class="member__email">{{ member.email }}</div>
              </div>
              <span v-if="member.id === team.teamAdminId" class="member__badge">Admin</span>
              <button
                v-else-if="canManage"
                @click="removeMember(member.id)"
                class="member__remove"
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
        <h3 class="modal__title">Edit Team Name</h3>
        <input
          v-model="editTeamName"
          type="text"
          placeholder="Team name"
          class="modal__input"
          @keyup.enter="saveTeamName"
          autofocus
        >
        <div class="modal__actions">
          <button @click="closeEditModal" class="btn btn--secondary">Cancel</button>
          <button @click="saveTeamName" class="btn btn--primary">Save</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.team-details {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1rem;

  @media (min-width: 768px) {
    padding: 2rem;
  }

  &__loading,
  &__error {
    text-align: center;
    padding: 2rem 1rem;
    color: var(--text-secondary);
    font-size: 1rem;

    @media (min-width: 768px) {
      padding: 3rem;
      font-size: 1.1rem;
    }
  }

  &__page-header {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-bottom: 1.5rem;

    @media (min-width: 768px) {
      flex-direction: row;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 2rem;
    }
  }

  &__back {
    background: none;
    border: none;
    color: var(--color-primary);
    font-size: 0.9rem;
    cursor: pointer;
    padding: 0.5rem 0;
    text-align: left;

    @media (min-width: 768px) {
      font-size: 1rem;
    }

    &:hover {
      text-decoration: underline;
    }
  }

  &__header-actions {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;

    @media (min-width: 768px) {
      flex-direction: row;
      gap: 1rem;
    }
  }

  &__header {
    margin-bottom: 1.5rem;

    @media (min-width: 768px) {
      margin-bottom: 2rem;
    }
  }

  &__title {
    font-size: 1.75rem;
    color: var(--text-primary);
    margin: 0 0 0.5rem 0;

    @media (min-width: 768px) {
      font-size: 2.5rem;
    }
  }

  &__meta {
    color: var(--text-secondary);
    font-size: 1rem;

    @media (min-width: 768px) {
      font-size: 1.1rem;
    }
  }

  &__content {
    display: grid;
    grid-template-columns: 1fr;
    gap: 1.25rem;

    @media (min-width: 768px) {
      grid-template-columns: 1fr 2fr;
      gap: 2rem;
    }
  }
}

.admin-card {
  background: var(--bg-primary);
  border-radius: 12px;
  padding: 1.25rem;
  border: 1px solid var(--border-color);

  @media (min-width: 768px) {
    padding: 1.5rem;
  }

  &__title {
    font-size: 1.1rem;
    color: var(--text-primary);
    margin: 0 0 1rem 0;

    @media (min-width: 768px) {
      font-size: 1.2rem;
    }
  }

  &__content {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  &__photo {
    width: 56px;
    height: 56px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 60px;
      height: 60px;
    }
  }

  &__photo-placeholder {
    width: 56px;
    height: 56px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.4rem;
    font-weight: bold;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 60px;
      height: 60px;
      font-size: 1.5rem;
    }
  }

  &__info {
    flex: 1;
  }

  &__name {
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 0.25rem;

    @media (min-width: 768px) {
      font-size: 1.2rem;
    }
  }

  &__email {
    color: var(--text-secondary);
    margin-bottom: 0.5rem;
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__role {
    display: inline-block;
    padding: 0.25rem 0.5rem;
    background: var(--color-primary);
    color: white;
    border-radius: 4px;
    font-size: 0.7rem;
    text-transform: uppercase;
    font-weight: 500;

    @media (min-width: 768px) {
      font-size: 0.75rem;
    }
  }
}

.members-card {
  background: var(--bg-primary);
  border-radius: 12px;
  padding: 1.25rem;
  border: 1px solid var(--border-color);

  @media (min-width: 768px) {
    padding: 1.5rem;
  }

  &__header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  &__title {
    margin: 0;
    font-size: 1.1rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1.2rem;
    }
  }

  &__count {
    background: var(--bg-secondary);
    color: var(--text-secondary);
    padding: 0.25rem 0.75rem;
    border-radius: 12px;
    font-size: 0.8rem;
    font-weight: 600;

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }
  }

  &__list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
}

.member {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--bg-secondary);
  border-radius: 8px;
  transition: background 0.2s;

  @media (min-width: 768px) {
    gap: 1rem;
  }

  &:hover {
    background: var(--bg-tertiary);
  }

  &__photo {
    width: 42px;
    height: 42px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 45px;
      height: 45px;
    }
  }

  &__photo-placeholder {
    width: 42px;
    height: 42px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1rem;
    font-weight: bold;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 45px;
      height: 45px;
      font-size: 1.1rem;
    }
  }

  &__info {
    flex: 1;
  }

  &__name {
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 0.25rem;
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__email {
    font-size: 0.8rem;
    color: var(--text-secondary);

    @media (min-width: 768px) {
      font-size: 0.85rem;
    }
  }

  &__badge {
    padding: 0.25rem 0.5rem;
    background: var(--color-primary);
    color: white;
    border-radius: 4px;
    font-size: 0.7rem;
    font-weight: 500;

    @media (min-width: 768px) {
      font-size: 0.75rem;
    }
  }

  &__remove {
    min-width: 44px;
    min-height: 44px;
    border: none;
    background: rgba(244, 67, 54, 0.1);
    color: var(--color-error);
    border-radius: 50%;
    font-size: 1.5rem;
    line-height: 1;
    cursor: pointer;
    transition: all 0.2s;
    flex-shrink: 0;

    &:hover {
      background: rgba(244, 67, 54, 0.2);
    }
  }
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
  padding: 1rem;
}

.modal {
  background: var(--bg-primary);
  border-radius: 12px;
  padding: 1.5rem;
  max-width: 500px;
  width: 100%;

  @media (min-width: 768px) {
    padding: 2rem;
  }

  &__title {
    margin: 0 0 1rem 0;
    font-size: 1.25rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1.5rem;
    }
  }

  &__input {
    width: 100%;
    padding: 0.75rem;
    border: 2px solid var(--border-color);
    border-radius: 8px;
    font-size: 0.9rem;
    margin-bottom: 1.5rem;
    background: var(--bg-primary);
    color: var(--text-primary);
    min-height: 44px;

    @media (min-width: 768px) {
      font-size: 1rem;
    }

    &:focus {
      outline: none;
      border-color: var(--color-primary);
    }
  }

  &__actions {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;

    @media (min-width: 768px) {
      flex-direction: row;
      gap: 1rem;
      justify-content: flex-end;
    }
  }
}
</style>
