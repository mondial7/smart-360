<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import apiClient from '@/api/client'
import StatusDonut from '@/components/StatusDonut.vue'
import type { AdminAnalytics, CompletionPoint } from '@/types/adminAnalytics'
import {
  PhUsers,
  PhBuildings,
  PhArrowsClockwise,
  PhCheckCircle,
  PhPaperPlaneTilt,
  PhTimer,
  PhTrophy,
  PhTrendUp
} from '@phosphor-icons/vue'

const analytics = ref<AdminAnalytics | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

onMounted(async () => {
  await load()
})

async function load() {
  try {
    loading.value = true
    error.value = null
    const res = await apiClient.get<AdminAnalytics>('/analytics/admin')
    analytics.value = res.data
  } catch (err: any) {
    error.value = err?.response?.data?.error || 'Failed to load analytics'
  } finally {
    loading.value = false
  }
}

const completionPercent = computed(() => {
  const rate = analytics.value?.totals.completionRate ?? 0
  return Math.round(rate * 100)
})

const avgResponseLabel = computed(() => formatDuration(analytics.value?.totals.avgResponseSeconds ?? 0))

function formatDuration(seconds: number): string {
  if (!seconds || seconds <= 0) return '—'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  if (hours < 48) return `${hours}h`
  const days = Math.floor(hours / 24)
  return `${days}d`
}

const donutSlices = computed(() => {
  const r = analytics.value?.roundsByStatus
  if (!r) return []
  return [
    { label: 'Draft', value: r.draft, color: 'var(--text-tertiary)' },
    { label: 'Active', value: r.active, color: 'var(--color-primary)' },
    { label: 'Closed', value: r.closed, color: 'var(--color-warning)' },
    { label: 'Shared', value: r.shared, color: 'var(--color-success)' }
  ]
})

const trend = computed(() => analytics.value?.completionTrend ?? [])

const trendMaxExpected = computed(() =>
  Math.max(1, ...trend.value.map((p) => p.expected))
)

function trendBarHeight(value: number): string {
  return `${(value / trendMaxExpected.value) * 100}%`
}

function ratePercent(point: CompletionPoint): string {
  return `${Math.round(point.completionRate * 100)}%`
}

function formatRoundDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}

const teamActivity = computed(() => analytics.value?.teamActivity ?? [])
const topStrengths = computed(() => analytics.value?.topThemes.strengths ?? [])
const topImprovements = computed(() => analytics.value?.topThemes.improvements ?? [])

const themeMaxStrengths = computed(() => Math.max(1, ...topStrengths.value.map((t) => t.count)))
const themeMaxImprovements = computed(() => Math.max(1, ...topImprovements.value.map((t) => t.count)))

function themeBarWidth(count: number, max: number): string {
  return `${(count / max) * 100}%`
}
</script>

<template>
  <div class="analytics-page">
    <header class="analytics-page__header">
      <h1 class="analytics-page__title">Admin analytics</h1>
      <p class="analytics-page__subtitle">Org-wide engagement, completion, and theme insights.</p>
    </header>

    <div v-if="loading" class="analytics-page__placeholder">Loading analytics…</div>
    <div v-else-if="error" class="analytics-page__placeholder analytics-page__placeholder--error">{{ error }}</div>

    <template v-else-if="analytics">
      <section class="counters">
        <article class="counter">
          <div class="counter__icon counter__icon--primary"><PhUsers :size="22" weight="duotone" /></div>
          <div class="counter__value">{{ analytics.totals.users }}</div>
          <div class="counter__label">Users</div>
        </article>
        <article class="counter">
          <div class="counter__icon counter__icon--secondary"><PhBuildings :size="22" weight="duotone" /></div>
          <div class="counter__value">{{ analytics.totals.teams }}</div>
          <div class="counter__label">Teams</div>
        </article>
        <article class="counter">
          <div class="counter__icon counter__icon--info"><PhArrowsClockwise :size="22" weight="duotone" /></div>
          <div class="counter__value">{{ analytics.totals.rounds }}</div>
          <div class="counter__label">Rounds</div>
        </article>
        <article class="counter">
          <div class="counter__icon counter__icon--success"><PhCheckCircle :size="22" weight="duotone" /></div>
          <div class="counter__value">{{ completionPercent }}%</div>
          <div class="counter__label">Completion rate</div>
        </article>
        <article class="counter">
          <div class="counter__icon counter__icon--warning"><PhPaperPlaneTilt :size="22" weight="duotone" /></div>
          <div class="counter__value">{{ analytics.totals.consolidationsShared }}</div>
          <div class="counter__label">Shared with subjects</div>
        </article>
        <article class="counter">
          <div class="counter__icon counter__icon--neutral"><PhTimer :size="22" weight="duotone" /></div>
          <div class="counter__value">{{ avgResponseLabel }}</div>
          <div class="counter__label">Avg response time</div>
        </article>
      </section>

      <section class="row">
        <article class="card row__card">
          <h2 class="card__title">Round status breakdown</h2>
          <StatusDonut :slices="donutSlices" :size="180" :thickness="22" />
        </article>

        <article class="card row__card row__card--wide">
          <h2 class="card__title">Completion rate by round</h2>
          <p v-if="trend.length === 0" class="card__empty">No completed rounds yet.</p>
          <div v-else class="trend">
            <div
              v-for="point in trend"
              :key="point.roundId"
              class="trend__bar"
              :title="`${point.subjectName}: ${point.received}/${point.expected} (${ratePercent(point)})`"
            >
              <div class="trend__col">
                <div class="trend__received" :style="{ height: trendBarHeight(point.received) }">
                  <span class="trend__rate">{{ ratePercent(point) }}</span>
                </div>
                <div class="trend__expected" :style="{ height: trendBarHeight(point.expected) }"></div>
              </div>
              <div class="trend__caption">
                <span class="trend__name">{{ point.subjectName || '—' }}</span>
                <span class="trend__date">{{ formatRoundDate(point.createdAt) }}</span>
              </div>
            </div>
          </div>
          <div v-if="trend.length > 0" class="trend__legend">
            <span class="trend__chip trend__chip--received"></span> Received
            <span class="trend__chip trend__chip--expected"></span> Expected
          </div>
        </article>
      </section>

      <section class="card">
        <h2 class="card__title">Team activity</h2>
        <p v-if="teamActivity.length === 0" class="card__empty">No teams yet.</p>
        <div v-else class="team-table-wrap">
          <table class="team-table">
            <thead>
              <tr>
                <th>Team</th>
                <th class="team-table__num">Members</th>
                <th class="team-table__num">Active rounds</th>
                <th class="team-table__num">Submissions</th>
                <th class="team-table__num">Avg response</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in teamActivity" :key="row.teamId">
                <td>{{ row.teamName }}</td>
                <td class="team-table__num">{{ row.memberCount }}</td>
                <td class="team-table__num">{{ row.activeRounds }}</td>
                <td class="team-table__num">{{ row.totalSubmissions }}</td>
                <td class="team-table__num">{{ formatDuration(row.avgResponseSeconds) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="row">
        <article class="card row__card">
          <h2 class="card__title">
            <PhTrophy :size="18" weight="duotone" />
            <span>Top strengths mentioned</span>
          </h2>
          <p v-if="topStrengths.length === 0" class="card__empty">No themes yet — share a few rounds to populate this.</p>
          <ul v-else class="theme-list">
            <li v-for="t in topStrengths" :key="t.phrase" class="theme-row">
              <span class="theme-row__phrase">{{ t.phrase }}</span>
              <span class="theme-row__bar">
                <span class="theme-row__fill theme-row__fill--strength" :style="{ width: themeBarWidth(t.count, themeMaxStrengths) }"></span>
              </span>
              <span class="theme-row__count">{{ t.count }}</span>
            </li>
          </ul>
        </article>

        <article class="card row__card">
          <h2 class="card__title">
            <PhTrendUp :size="18" weight="duotone" />
            <span>Top improvement themes</span>
          </h2>
          <p v-if="topImprovements.length === 0" class="card__empty">No themes yet.</p>
          <ul v-else class="theme-list">
            <li v-for="t in topImprovements" :key="t.phrase" class="theme-row">
              <span class="theme-row__phrase">{{ t.phrase }}</span>
              <span class="theme-row__bar">
                <span class="theme-row__fill theme-row__fill--improvement" :style="{ width: themeBarWidth(t.count, themeMaxImprovements) }"></span>
              </span>
              <span class="theme-row__count">{{ t.count }}</span>
            </li>
          </ul>
        </article>
      </section>
    </template>
  </div>
</template>

<style scoped lang="scss">
.analytics-page {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;

  @media (max-width: 767px) {
    padding: 1rem;
  }

  &__header {
    margin-bottom: 1.5rem;
  }

  &__title {
    font-size: 1.75rem;
    margin: 0 0 0.25rem;
    color: var(--text-primary);
  }

  &__subtitle {
    margin: 0;
    color: var(--text-secondary);
  }

  &__placeholder {
    background: var(--bg-primary);
    padding: 2rem;
    border-radius: 12px;
    text-align: center;
    color: var(--text-secondary);

    &--error {
      color: var(--color-error);
    }
  }
}

.counters {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(170px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.counter {
  background: var(--bg-primary);
  padding: 1rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  display: flex;
  flex-direction: column;
  gap: 0.25rem;

  &__icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 38px;
    height: 38px;
    border-radius: 10px;
    margin-bottom: 0.4rem;
    color: white;

    &--primary { background: var(--color-primary); }
    &--secondary { background: var(--color-secondary); }
    &--info { background: var(--color-info); }
    &--success { background: var(--color-success); }
    &--warning { background: var(--color-warning); }
    &--neutral { background: var(--text-tertiary); }
  }

  &__value {
    font-size: 1.65rem;
    font-weight: 700;
    color: var(--text-primary);
    line-height: 1.05;
  }

  &__label {
    font-size: 0.85rem;
    color: var(--text-secondary);
  }
}

.row {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.5rem;
  margin-bottom: 1.5rem;

  @media (min-width: 900px) {
    grid-template-columns: 1fr 1fr;

    &__card--wide {
      grid-column: span 1;
    }
  }
}

.card {
  background: var(--bg-primary);
  padding: 1.25rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);

  &__title {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 1rem;
    margin: 0 0 1rem;
    color: var(--text-primary);
  }

  &__empty {
    color: var(--text-secondary);
    font-size: 0.9rem;
    margin: 0;
  }
}

.trend {
  display: flex;
  align-items: flex-end;
  gap: 0.6rem;
  height: 200px;
  padding: 0.5rem 0 0.25rem;
  border-bottom: 1px solid var(--border-color);
  overflow-x: auto;

  &__bar {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.4rem;
    min-width: 60px;
  }

  &__col {
    position: relative;
    height: 160px;
    width: 28px;
    display: flex;
    align-items: flex-end;
    justify-content: center;
  }

  &__expected,
  &__received {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    border-radius: 4px 4px 0 0;
  }

  &__expected {
    background: var(--bg-secondary);
    border: 1px dashed var(--border-color);
  }

  &__received {
    background: linear-gradient(180deg, var(--color-primary) 0%, var(--color-secondary) 100%);
    z-index: 1;
    transition: height 0.3s ease;
    display: flex;
    justify-content: center;
  }

  &__rate {
    position: absolute;
    top: -1.1rem;
    font-size: 0.7rem;
    color: var(--text-secondary);
    white-space: nowrap;
  }

  &__caption {
    display: flex;
    flex-direction: column;
    align-items: center;
    line-height: 1.1;
  }

  &__name {
    font-size: 0.75rem;
    color: var(--text-primary);
    text-align: center;
    max-width: 80px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &__date {
    font-size: 0.7rem;
    color: var(--text-tertiary);
  }

  &__legend {
    display: inline-flex;
    align-items: center;
    gap: 0.6rem;
    margin-top: 0.5rem;
    font-size: 0.8rem;
    color: var(--text-secondary);
  }

  &__chip {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 3px;
    margin-right: 0.25rem;

    &--received {
      background: var(--color-primary);
    }

    &--expected {
      background: var(--bg-secondary);
      border: 1px dashed var(--border-color);
    }
  }
}

.team-table-wrap {
  overflow-x: auto;
}

.team-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;

  th,
  td {
    padding: 0.6rem 0.75rem;
    text-align: left;
    border-bottom: 1px solid var(--border-color);
  }

  th {
    color: var(--text-secondary);
    font-weight: 600;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  &__num {
    text-align: right;
    font-variant-numeric: tabular-nums;
  }
}

.theme-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.45rem;
}

.theme-row {
  display: grid;
  grid-template-columns: minmax(80px, 130px) 1fr 32px;
  align-items: center;
  gap: 0.6rem;

  &__phrase {
    color: var(--text-primary);
    font-size: 0.9rem;
    text-transform: capitalize;
  }

  &__bar {
    background: var(--bg-secondary);
    height: 8px;
    border-radius: 4px;
    overflow: hidden;
  }

  &__fill {
    display: block;
    height: 100%;
    transition: width 0.3s ease;

    &--strength {
      background: var(--color-success);
    }

    &--improvement {
      background: var(--color-warning);
    }
  }

  &__count {
    font-variant-numeric: tabular-nums;
    color: var(--text-secondary);
    font-size: 0.85rem;
    text-align: right;
  }
}
</style>
