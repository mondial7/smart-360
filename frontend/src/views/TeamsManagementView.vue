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
  <div class="teams">
    <header class="teams__header">
      <h1 class="teams__title">Teams</h1>
      <button v-if="auth.isAdmin" @click="createTeam" class="btn btn--primary">
        Create Team
      </button>
    </header>

    <div v-if="loading" class="teams__loading">Loading teams...</div>

    <div v-else-if="teams.length === 0" class="teams__empty">
      <p class="teams__empty-text">No teams have been created yet.</p>
      <button v-if="auth.isAdmin" @click="createTeam" class="btn btn--primary">
        Create Your First Team
      </button>
    </div>

    <div v-else class="teams__grid">
      <div v-for="team in teams" :key="team.id" class="team-card">
        <div class="team-card__header">
          <h3 class="team-card__name">{{ team.name }}</h3>
        </div>

        <div class="team-card__admin">
          <div class="team-card__admin-label">Team Admin</div>
          <div class="team-card__admin-info">
            <img
              v-if="team.teamAdmin?.photoUrl"
              :src="team.teamAdmin.photoUrl"
              :alt="team.teamAdmin.name"
              class="team-card__admin-photo"
            >
            <div v-else class="team-card__admin-photo-placeholder">
              {{ team.teamAdmin?.name.charAt(0).toUpperCase() }}
            </div>
            <div class="team-card__admin-details">
              <div class="team-card__admin-name">{{ team.teamAdmin?.name }}</div>
              <div class="team-card__admin-email">{{ team.teamAdmin?.email }}</div>
            </div>
          </div>
        </div>

        <div class="team-card__stats">
          <div class="stat">
            <span class="stat__value">{{ team.members?.length || 0 }}</span>
            <span class="stat__label">Members</span>
          </div>
        </div>

        <div class="team-card__actions">
          <button @click="viewTeam(team.id)" class="btn btn--secondary">
            View Team
          </button>
          <button @click="createTeamRound(team.id)" class="btn btn--primary">
            Create Team Round
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.teams {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1rem;

  @media (min-width: 768px) {
    padding: 2rem;
  }

  &__header {
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

  &__title {
    font-size: 1.5rem;
    color: var(--text-primary);
    margin: 0;

    @media (min-width: 768px) {
      font-size: 2rem;
    }
  }

  &__loading {
    text-align: center;
    padding: 2rem 1rem;
    color: var(--text-secondary);
    font-size: 1rem;

    @media (min-width: 768px) {
      padding: 3rem;
      font-size: 1.1rem;
    }
  }

  &__empty {
    text-align: center;
    padding: 2rem 1rem;
    background: var(--bg-primary);
    border-radius: 12px;
    border: 1px solid var(--border-color);

    @media (min-width: 768px) {
      padding: 4rem 2rem;
    }
  }

  &__empty-text {
    color: var(--text-secondary);
    font-size: 1rem;
    margin: 0 0 1.5rem 0;

    @media (min-width: 768px) {
      font-size: 1.1rem;
    }
  }

  &__grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 1.25rem;

    @media (min-width: 768px) {
      grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
      gap: 1.5rem;
    }
  }
}

.team-card {
  background: var(--bg-primary);
  border-radius: 12px;
  padding: 1.25rem;
  border: 1px solid var(--border-color);
  transition: box-shadow 0.2s;

  @media (min-width: 768px) {
    padding: 1.5rem;
  }

  &:hover {
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  }

  &__header {
    margin-bottom: 1rem;
  }

  &__name {
    margin: 0;
    font-size: 1.25rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1.5rem;
    }
  }

  &__admin {
    margin-bottom: 1rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid var(--border-color);
  }

  &__admin-label {
    font-size: 0.75rem;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 0.5rem;

    @media (min-width: 768px) {
      font-size: 0.85rem;
    }
  }

  &__admin-info {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  &__admin-photo {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;
  }

  &__admin-photo-placeholder {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.1rem;
    font-weight: bold;
    flex-shrink: 0;

    @media (min-width: 768px) {
      font-size: 1.2rem;
    }
  }

  &__admin-details {
    flex: 1;
  }

  &__admin-name {
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 0.25rem;
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__admin-email {
    font-size: 0.8rem;
    color: var(--text-secondary);

    @media (min-width: 768px) {
      font-size: 0.85rem;
    }
  }

  &__stats {
    display: flex;
    gap: 1.5rem;
    margin-bottom: 1rem;

    @media (min-width: 768px) {
      gap: 2rem;
    }
  }

  &__actions {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;

    @media (min-width: 768px) {
      flex-direction: row;
    }
  }
}

.stat {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;

  &__value {
    font-size: 1.25rem;
    font-weight: bold;
    color: var(--color-primary);

    @media (min-width: 768px) {
      font-size: 1.5rem;
    }
  }

  &__label {
    font-size: 0.75rem;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;

    @media (min-width: 768px) {
      font-size: 0.85rem;
    }
  }
}
</style>
