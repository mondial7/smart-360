<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import apiClient from '@/api/client'
import RadarChart from '@/components/RadarChart.vue'
import type { MyAnalytics } from '@/types/analytics'

const analytics = ref<MyAnalytics | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

onMounted(async () => {
  await load()
})

async function load() {
  try {
    loading.value = true
    error.value = null
    const res = await apiClient.get<MyAnalytics>('/analytics/me')
    analytics.value = res.data
  } catch (err: any) {
    error.value = err?.response?.data?.error || 'Failed to load analytics'
  } finally {
    loading.value = false
  }
}

const radarAxes = computed(() => {
  const r = analytics.value?.latestRadar
  if (!r) return []
  return [
    { label: 'Strengths', value: r.strengths },
    { label: 'Improvements', value: r.improvements },
    { label: 'Behaviors', value: r.behaviors },
    { label: 'Growth', value: r.growth }
  ]
})

const trendMax = computed(() => {
  const rounds = analytics.value?.rounds ?? []
  if (rounds.length === 0) return 1
  return Math.max(
    1,
    ...rounds.flatMap((r) => [r.strengthsCount, r.improvementsCount, r.insightsCount])
  )
})

function barHeight(count: number): string {
  return `${(count / trendMax.value) * 100}%`
}

function formatRoundLabel(iso: string): string {
  if (!iso) return ''
  return new Date(iso).toLocaleDateString('en-US', { month: 'short', year: '2-digit' })
}
</script>

<template>
  <section class="analytics card">
    <header class="analytics__header">
      <h2 class="analytics__title">My feedback insights</h2>
      <p class="analytics__subtitle">Personal stats based on the rounds shared with you.</p>
    </header>

    <div v-if="loading" class="analytics__placeholder">Loading insights...</div>
    <div v-else-if="error" class="analytics__placeholder analytics__placeholder--error">{{ error }}</div>

    <template v-else-if="analytics">
      <div class="analytics__stats">
        <div class="analytics__stat">
          <div class="analytics__stat-value">{{ analytics.feedbackReceivedCount }}</div>
          <div class="analytics__stat-label">Rounds received</div>
        </div>
        <div class="analytics__stat">
          <div class="analytics__stat-value">{{ analytics.feedbackGivenCount }}</div>
          <div class="analytics__stat-label">Reviews submitted</div>
        </div>
        <div class="analytics__stat">
          <div class="analytics__stat-value">{{ analytics.pendingReviewsCount }}</div>
          <div class="analytics__stat-label">Pending reviews</div>
        </div>
      </div>

      <div v-if="analytics.feedbackReceivedCount === 0" class="analytics__empty">
        No shared feedback yet. Once an admin shares a round with you, your radar appears here.
      </div>

      <div v-else class="analytics__charts">
        <div class="analytics__radar">
          <h3 class="analytics__chart-title">Latest round breakdown</h3>
          <RadarChart :axes="radarAxes" :size="240" />
        </div>

        <div class="analytics__trend">
          <h3 class="analytics__chart-title">Strengths vs. improvements over time</h3>
          <div class="bars" role="list">
            <div
              v-for="round in analytics.rounds"
              :key="round.roundId"
              class="bars__group"
              role="listitem"
              :title="`${round.strengthsCount} strengths, ${round.improvementsCount} improvements, ${round.insightsCount} insights`"
            >
              <div class="bars__col">
                <div class="bars__bar bars__bar--strengths" :style="{ height: barHeight(round.strengthsCount) }"></div>
                <div class="bars__bar bars__bar--improvements" :style="{ height: barHeight(round.improvementsCount) }"></div>
                <div class="bars__bar bars__bar--insights" :style="{ height: barHeight(round.insightsCount) }"></div>
              </div>
              <div class="bars__label">{{ formatRoundLabel(round.sharedAt) }}</div>
            </div>
          </div>
          <div class="legend">
            <span class="legend__chip legend__chip--strengths"></span> Strengths
            <span class="legend__chip legend__chip--improvements"></span> Improvements
            <span class="legend__chip legend__chip--insights"></span> Insights
          </div>
        </div>
      </div>
    </template>
  </section>
</template>

<style scoped lang="scss">
.analytics {
  padding: 1.5rem;
  background: var(--bg-primary);
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  margin-bottom: 2rem;

  &__header {
    margin-bottom: 1rem;
  }

  &__title {
    font-size: 1.25rem;
    margin: 0 0 0.25rem;
    color: var(--text-primary);
  }

  &__subtitle {
    margin: 0;
    color: var(--text-secondary);
    font-size: 0.9rem;
  }

  &__placeholder {
    padding: 1.5rem;
    text-align: center;
    color: var(--text-secondary);

    &--error {
      color: var(--color-error);
    }
  }

  &__stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  &__stat {
    background: var(--bg-secondary);
    padding: 1rem;
    border-radius: 8px;
    text-align: center;
  }

  &__stat-value {
    font-size: 1.75rem;
    font-weight: 700;
    color: var(--text-primary);
    line-height: 1.1;
  }

  &__stat-label {
    font-size: 0.8rem;
    color: var(--text-secondary);
    margin-top: 0.25rem;
  }

  &__empty {
    background: var(--bg-secondary);
    padding: 1.5rem;
    border-radius: 8px;
    text-align: center;
    color: var(--text-secondary);
  }

  &__charts {
    display: grid;
    grid-template-columns: 1fr;
    gap: 1.5rem;

    @media (min-width: 768px) {
      grid-template-columns: minmax(240px, 280px) 1fr;
      align-items: start;
    }
  }

  &__chart-title {
    font-size: 0.95rem;
    color: var(--text-primary);
    margin: 0 0 0.75rem;
    font-weight: 600;
  }

  &__radar {
    text-align: center;
  }
}

.bars {
  display: flex;
  align-items: flex-end;
  gap: 0.75rem;
  height: 160px;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--border-color);
  overflow-x: auto;

  &__group {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.4rem;
    min-width: 44px;
  }

  &__col {
    display: flex;
    align-items: flex-end;
    gap: 2px;
    height: 130px;
  }

  &__bar {
    width: 10px;
    border-radius: 3px 3px 0 0;
    min-height: 2px;
    transition: height 0.3s ease;

    &--strengths {
      background: var(--color-success);
    }

    &--improvements {
      background: var(--color-warning);
    }

    &--insights {
      background: var(--color-primary);
    }
  }

  &__label {
    font-size: 0.7rem;
    color: var(--text-secondary);
  }
}

.legend {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 0.75rem;
  font-size: 0.8rem;
  color: var(--text-secondary);
  align-items: center;

  &__chip {
    display: inline-block;
    width: 12px;
    height: 12px;
    border-radius: 3px;
    margin-right: 0.25rem;

    &--strengths {
      background: var(--color-success);
    }

    &--improvements {
      background: var(--color-warning);
    }

    &--insights {
      background: var(--color-primary);
    }
  }
}
</style>
