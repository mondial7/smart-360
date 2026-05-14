<script setup lang="ts">
import { ref, onMounted } from 'vue'
import apiClient from '@/api/client'
import {
  PhDownloadSimple,
  PhFileText,
  PhTrophy,
  PhTrendUp,
  PhTarget,
  PhClipboardText,
  PhNotePencil,
  PhPlus,
  PhMinus,
  PhCheck,
  PhArrowUp,
  PhArrowRight,
  PhEye,
  PhEyeSlash,
  PhScales,
  PhUsersThree
} from '@phosphor-icons/vue'

const consolidations = ref<any[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const expandedIds = ref<Set<string>>(new Set())
const downloadingIds = ref<Set<string>>(new Set())

onMounted(async () => {
  await loadConsolidations()
})

async function loadConsolidations() {
  try {
    loading.value = true
    error.value = null
    const response = await apiClient.get('/dashboard/my-consolidations')
    consolidations.value = response.data || []

    // Parse JSON fields for each consolidation
    consolidations.value.forEach(parseConsolidationFields)
  } catch (err: any) {
    console.error('Failed to load consolidations:', err)
    error.value = err.response?.data?.error || 'Failed to load feedback'
  } finally {
    loading.value = false
  }
}

function parseConsolidationFields(consolidation: any) {
  try {
    if (consolidation.strengths && typeof consolidation.strengths === 'string') {
      consolidation.strengths = JSON.parse(consolidation.strengths)
    }
    if (consolidation.areasForImprovement && typeof consolidation.areasForImprovement === 'string') {
      consolidation.areasForImprovement = JSON.parse(consolidation.areasForImprovement)
    }
    if (consolidation.actionableInsights && typeof consolidation.actionableInsights === 'string') {
      consolidation.actionableInsights = JSON.parse(consolidation.actionableInsights)
    }
    if (consolidation.questionSummaries && typeof consolidation.questionSummaries === 'string') {
      consolidation.questionSummaries = JSON.parse(consolidation.questionSummaries)
    }
  } catch (error) {
    console.error('Error parsing consolidation fields:', error)
  }
}

function toggleExpanded(id: string) {
  if (expandedIds.value.has(id)) {
    expandedIds.value.delete(id)
  } else {
    expandedIds.value.add(id)
  }
}

function cardTitle(consolidation: any, key: string, fallback: string): string {
  const fromTemplate = consolidation?.questionLabels?.[key]
  return (fromTemplate && fromTemplate.trim()) ? fromTemplate : fallback
}

function isExpanded(id: string): boolean {
  return expandedIds.value.has(id)
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'Unknown date'
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'long',
    day: 'numeric',
    year: 'numeric'
  })
}

async function downloadPDF(consolidation: any) {
  const roundId = consolidation.roundId
  if (!roundId || downloadingIds.value.has(roundId)) return
  downloadingIds.value.add(roundId)
  try {
    const response = await apiClient.get(`/consolidations/${roundId}/pdf`, {
      responseType: 'blob'
    })
    const blob = new Blob([response.data], { type: 'application/pdf' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    const disposition = response.headers['content-disposition'] as string | undefined
    link.download = parseFilename(disposition) || `feedback-${roundId}.pdf`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  } catch (err) {
    console.error('PDF download failed:', err)
    error.value = 'Could not download PDF. Please try again.'
  } finally {
    downloadingIds.value.delete(roundId)
  }
}

function parseFilename(disposition: string | undefined): string | null {
  if (!disposition) return null
  const match = disposition.match(/filename="?([^";]+)"?/)
  return match ? match[1] : null
}
</script>

<template>
  <div class="feedback">
    <header class="feedback__header">
      <h1 class="feedback__title">My Feedback</h1>
      <p class="feedback__subtitle">View your consolidated feedback and development insights</p>
    </header>

    <div v-if="loading" class="feedback__loading">
      <div class="feedback__spinner"></div>
      <p class="feedback__loading-text">Loading your feedback...</p>
    </div>

    <div v-else-if="error" class="feedback__error">
      <p class="feedback__error-message">{{ error }}</p>
      <button @click="loadConsolidations" class="btn btn--secondary">Try Again</button>
    </div>

    <div v-else-if="consolidations.length === 0" class="feedback__empty">
      <p class="feedback__empty-text">No feedback has been shared with you yet.</p>
      <p class="feedback__empty-subtext">When an administrator shares consolidated feedback, it will appear here.</p>
    </div>

    <div v-else class="feedback__list">
      <div
        v-for="consolidation in consolidations"
        :key="consolidation.id"
        class="feedback-card"
      >
        <div class="feedback-card__header" @click="toggleExpanded(consolidation.id)">
          <div class="feedback-card__header-content">
            <h2 class="feedback-card__title">Feedback Received</h2>
            <p class="feedback-card__date">Shared {{ formatDate(consolidation.sharedAt) }}</p>
          </div>
          <div class="feedback-card__header-actions">
            <button
              class="feedback-card__pdf"
              :disabled="downloadingIds.has(consolidation.roundId)"
              @click.stop="downloadPDF(consolidation)"
              type="button"
              title="Download as PDF"
            >
              <PhDownloadSimple :size="16" weight="bold" />
              <span>{{ downloadingIds.has(consolidation.roundId) ? 'Preparing…' : 'PDF' }}</span>
            </button>
            <button class="feedback-card__toggle" :class="{ 'feedback-card__toggle--expanded': isExpanded(consolidation.id) }">
              <PhMinus v-if="isExpanded(consolidation.id)" :size="18" weight="bold" />
              <PhPlus v-else :size="18" weight="bold" />
            </button>
          </div>
        </div>

        <div v-if="isExpanded(consolidation.id)" class="feedback-card__body">
          <!-- Executive Summary -->
          <section class="feedback-section">
            <h3 class="feedback-section__title">
              <PhFileText :size="20" weight="duotone" />
              <span>Executive Summary</span>
            </h3>
            <p class="feedback-section__summary">{{ consolidation.executiveSummary }}</p>
          </section>

          <!-- Strengths -->
          <section class="feedback-section">
            <h3 class="feedback-section__title">
              <PhTrophy :size="20" weight="duotone" />
              <span>Key Strengths</span>
            </h3>
            <ul class="feedback-list feedback-list--positive">
              <li v-for="(strength, i) in consolidation.strengths" :key="i" class="feedback-list__item">
                <PhCheck class="feedback-list__icon" :size="18" weight="bold" />
                <span>{{ strength }}</span>
              </li>
            </ul>
          </section>

          <!-- Areas for Improvement -->
          <section class="feedback-section">
            <h3 class="feedback-section__title">
              <PhTrendUp :size="20" weight="duotone" />
              <span>Areas for Improvement</span>
            </h3>
            <ul class="feedback-list feedback-list--improvement">
              <li v-for="(area, i) in consolidation.areasForImprovement" :key="i" class="feedback-list__item">
                <PhArrowUp class="feedback-list__icon" :size="18" weight="bold" />
                <span>{{ area }}</span>
              </li>
            </ul>
          </section>

          <!-- Actionable Insights -->
          <section class="feedback-section">
            <h3 class="feedback-section__title">
              <PhTarget :size="20" weight="duotone" />
              <span>Actionable Insights</span>
            </h3>
            <ul class="feedback-list feedback-list--action">
              <li v-for="(insight, i) in consolidation.actionableInsights" :key="i" class="feedback-list__item">
                <PhArrowRight class="feedback-list__icon" :size="18" weight="bold" />
                <span>{{ insight }}</span>
              </li>
            </ul>
          </section>

          <!-- Question Summaries -->
          <section class="feedback-section">
            <h3 class="feedback-section__title">
              <PhClipboardText :size="20" weight="duotone" />
              <span>Detailed Question Analysis</span>
            </h3>
            <div class="questions">
              <div class="question-card">
                <h4 class="question-card__title">1. {{ cardTitle(consolidation, 'a', 'What to continue') }}</h4>
                <p class="question-card__text">{{ consolidation.questionSummaries?.a || 'No summary available' }}</p>
              </div>
              <div class="question-card">
                <h4 class="question-card__title">2. {{ cardTitle(consolidation, 'b', "What's blocking growth") }}</h4>
                <p class="question-card__text">{{ consolidation.questionSummaries?.b || 'No summary available' }}</p>
              </div>
              <div class="question-card">
                <h4 class="question-card__title">3. {{ cardTitle(consolidation, 'c', 'Where to double down') }}</h4>
                <p class="question-card__text">{{ consolidation.questionSummaries?.c || 'No summary available' }}</p>
              </div>
              <div class="question-card">
                <h4 class="question-card__title">4. {{ cardTitle(consolidation, 'd', 'One experiment to try') }}</h4>
                <p class="question-card__text">{{ consolidation.questionSummaries?.d || 'No summary available' }}</p>
              </div>
            </div>
          </section>

          <!-- Voice breakdown -->
          <section v-if="consolidation.voiceBreakdown" class="feedback-section">
            <h3 class="feedback-section__title">
              <PhUsersThree :size="20" weight="duotone" />
              <span>Voices — Different Vantages, Distinct Signals</span>
            </h3>
            <div class="voices">
              <article v-if="consolidation.voiceBreakdown.managerVoice" class="voice voice--manager">
                <header class="voice__header">
                  <h4 class="voice__title">Manager voice</h4>
                  <span class="voice__count">{{ consolidation.voiceBreakdown.managerVoice.reviewerCount }} reviewer(s)</span>
                </header>
                <p v-if="consolidation.voiceBreakdown.managerVoice.summary" class="voice__summary">
                  {{ consolidation.voiceBreakdown.managerVoice.summary }}
                </p>
                <ul v-if="consolidation.voiceBreakdown.managerVoice.themes?.length" class="voice__themes">
                  <li v-for="(t, i) in consolidation.voiceBreakdown.managerVoice.themes" :key="`mv-${i}`">{{ t }}</li>
                </ul>
              </article>
              <article v-if="consolidation.voiceBreakdown.peerVoice" class="voice voice--peer">
                <header class="voice__header">
                  <h4 class="voice__title">Peer voice</h4>
                  <span class="voice__count">{{ consolidation.voiceBreakdown.peerVoice.reviewerCount }} reviewer(s)</span>
                </header>
                <p v-if="consolidation.voiceBreakdown.peerVoice.summary" class="voice__summary">
                  {{ consolidation.voiceBreakdown.peerVoice.summary }}
                </p>
                <ul v-if="consolidation.voiceBreakdown.peerVoice.themes?.length" class="voice__themes">
                  <li v-for="(t, i) in consolidation.voiceBreakdown.peerVoice.themes" :key="`pv-${i}`">{{ t }}</li>
                </ul>
              </article>
              <article v-if="consolidation.voiceBreakdown.reportVoice" class="voice voice--report">
                <header class="voice__header">
                  <h4 class="voice__title">From people you manage</h4>
                  <span class="voice__count">{{ consolidation.voiceBreakdown.reportVoice.reviewerCount }} reviewer(s)</span>
                </header>
                <p v-if="consolidation.voiceBreakdown.reportVoice.summary" class="voice__summary">
                  {{ consolidation.voiceBreakdown.reportVoice.summary }}
                </p>
                <ul v-if="consolidation.voiceBreakdown.reportVoice.themes?.length" class="voice__themes">
                  <li v-for="(t, i) in consolidation.voiceBreakdown.reportVoice.themes" :key="`rv-${i}`">{{ t }}</li>
                </ul>
              </article>
            </div>
          </section>

          <!-- Self vs Peers delta -->
          <section v-if="consolidation.selfVsOthersDelta?.selfSubmitted" class="feedback-section">
            <h3 class="feedback-section__title">
              <PhScales :size="20" weight="duotone" />
              <span>Self vs Peers — Where You and Your Team See Things Differently</span>
            </h3>
            <p v-if="consolidation.selfVsOthersDelta.summary" class="feedback-section__summary">
              {{ consolidation.selfVsOthersDelta.summary }}
            </p>
            <div class="delta">
              <div v-if="consolidation.selfVsOthersDelta.blindSpots?.length" class="delta__group">
                <h4 class="delta__title">
                  <PhEyeSlash :size="16" weight="bold" />
                  <span>Blind spots</span>
                </h4>
                <p class="delta__hint">Things peers consistently flagged that your self-assessment didn't surface.</p>
                <ul class="delta__list">
                  <li v-for="(item, idx) in consolidation.selfVsOthersDelta.blindSpots" :key="`bs-${idx}`">{{ item }}</li>
                </ul>
              </div>
              <div v-if="consolidation.selfVsOthersDelta.hiddenStrengths?.length" class="delta__group">
                <h4 class="delta__title">
                  <PhEye :size="16" weight="bold" />
                  <span>Hidden strengths</span>
                </h4>
                <p class="delta__hint">Things peers value highly that you may underestimate.</p>
                <ul class="delta__list">
                  <li v-for="(item, idx) in consolidation.selfVsOthersDelta.hiddenStrengths" :key="`hs-${idx}`">{{ item }}</li>
                </ul>
              </div>
              <div v-if="consolidation.selfVsOthersDelta.aligned?.length" class="delta__group">
                <h4 class="delta__title">
                  <PhCheck :size="16" weight="bold" />
                  <span>Aligned themes</span>
                </h4>
                <p class="delta__hint">Where your self-view and peer feedback clearly agree.</p>
                <ul class="delta__list">
                  <li v-for="(item, idx) in consolidation.selfVsOthersDelta.aligned" :key="`al-${idx}`">{{ item }}</li>
                </ul>
              </div>
            </div>
          </section>

          <!-- Admin Notes -->
          <section v-if="consolidation.adminNotes" class="feedback-section">
            <h3 class="feedback-section__title feedback-section__title--admin">
              <PhNotePencil :size="20" weight="duotone" />
              <span>Additional Notes</span>
            </h3>
            <div class="admin-notes">{{ consolidation.adminNotes }}</div>
          </section>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.feedback {
  padding: 1rem;
  max-width: 900px;
  margin: 0 auto;

  @media (min-width: 768px) {
    padding: 2rem;
  }

  &__header {
    margin-bottom: 2rem;
  }

  &__title {
    font-size: 1.5rem;
    margin-bottom: 0.5rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 2rem;
    }
  }

  &__subtitle {
    color: var(--text-secondary);
    margin: 0;
  }

  &__loading,
  &__error,
  &__empty {
    background: var(--bg-primary);
    padding: 2rem 1rem;
    border-radius: 12px;
    border: 1px solid var(--border-color);
    text-align: center;

    @media (min-width: 768px) {
      padding: 3rem;
    }
  }

  &__loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
  }

  &__spinner {
    width: 36px;
    height: 36px;
    border: 4px solid var(--bg-tertiary);
    border-top: 4px solid var(--color-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;

    @media (min-width: 768px) {
      width: 40px;
      height: 40px;
    }
  }

  &__loading-text {
    color: var(--text-secondary);
    margin: 0;
  }

  &__error-message {
    color: var(--color-error);
    margin: 0 0 1rem 0;
  }

  &__empty-text {
    color: var(--text-secondary);
    margin: 0 0 0.5rem 0;
  }

  &__empty-subtext {
    font-size: 0.85rem;
    color: var(--text-tertiary);
    margin: 0.5rem 0 0 0;

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }
  }

  &__list {
    display: flex;
    flex-direction: column;
    gap: 1rem;

    @media (min-width: 768px) {
      gap: 1.5rem;
    }
  }
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.feedback-card {
  background: var(--bg-primary);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  overflow: hidden;
  transition: box-shadow 0.2s;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  }

  &__header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.25rem;
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
    color: white;
    cursor: pointer;
    user-select: none;

    @media (min-width: 768px) {
      padding: 1.5rem;
    }
  }

  &__header-content {
    flex: 1;
  }

  &__title {
    font-size: 1.125rem;
    margin: 0 0 0.25rem 0;

    @media (min-width: 768px) {
      font-size: 1.25rem;
    }
  }

  &__date {
    font-size: 0.85rem;
    opacity: 0.9;
    margin: 0;

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }
  }

  &__header-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-shrink: 0;
  }

  &__pdf {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    background: rgba(255, 255, 255, 0.18);
    color: white;
    border: 1px solid rgba(255, 255, 255, 0.4);
    padding: 0.4rem 0.75rem;
    border-radius: 6px;
    font-size: 0.85rem;
    cursor: pointer;
    transition: background 0.2s;

    &:hover:not(:disabled) {
      background: rgba(255, 255, 255, 0.3);
    }

    &:disabled {
      opacity: 0.7;
      cursor: not-allowed;
    }
  }

  &__toggle {
    background: rgba(255, 255, 255, 0.2);
    border: 2px solid rgba(255, 255, 255, 0.5);
    color: white;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    font-size: 1.5rem;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
    flex-shrink: 0;
    min-height: 44px;
    min-width: 44px;

    &:hover {
      background: rgba(255, 255, 255, 0.3);
      transform: scale(1.1);
    }

    &--expanded {
      transform: rotate(180deg);

      &:hover {
        transform: rotate(180deg) scale(1.1);
      }
    }
  }

  &__body {
    padding: 1.5rem;
    animation: slideDown 0.3s ease-out;

    @media (min-width: 768px) {
      padding: 2rem;
    }
  }
}

.feedback-section {
  margin-bottom: 1.5rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid var(--border-color);

  @media (min-width: 768px) {
    margin-bottom: 2rem;
    padding-bottom: 2rem;
  }

  &:last-child {
    margin-bottom: 0;
    padding-bottom: 0;
    border-bottom: none;
  }

  &__title {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 1rem;
    margin-bottom: 1rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1.125rem;
    }

    &--admin {
      color: var(--color-primary);
    }
  }

  &__summary {
    font-size: 1rem;
    line-height: 1.7;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1.05rem;
    }
  }
}

.feedback-list {
  list-style: none;
  padding: 0;
  margin: 0;

  &__item {
    display: flex;
    align-items: flex-start;
    gap: 0.6rem;
    padding: 0.875rem 1rem;
    margin-bottom: 0.75rem;
    border-radius: 8px;
    line-height: 1.5;
    color: var(--text-primary);

    @media (min-width: 768px) {
      padding: 1rem;
    }
  }

  &__icon {
    flex-shrink: 0;
    margin-top: 0.15rem;
  }

  &--positive &__item {
    background: rgba(76, 175, 80, 0.1);
    border-left: 3px solid var(--color-success);
  }

  &--positive &__icon {
    color: var(--color-success);
  }

  &--improvement &__item {
    background: rgba(255, 152, 0, 0.1);
    border-left: 3px solid var(--color-warning);
  }

  &--improvement &__icon {
    color: var(--color-warning);
  }

  &--action &__item {
    background: rgba(33, 150, 243, 0.1);
    border-left: 3px solid #2196f3;
  }

  &--action &__icon {
    color: #2196f3;
  }
}

.questions {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;

  @media (min-width: 640px) {
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  }
}

.question-card {
  padding: 1.25rem;
  background: var(--bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--border-color);

  @media (min-width: 768px) {
    padding: 1.5rem;
  }

  &__title {
    color: var(--color-primary);
    font-size: 0.85rem;
    margin: 0 0 0.75rem 0;
    font-weight: 600;

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }
  }

  &__text {
    color: var(--text-primary);
    font-size: 0.9rem;
    line-height: 1.6;
    margin: 0;

    @media (min-width: 768px) {
      font-size: 0.95rem;
    }
  }
}

.admin-notes {
  background: var(--bg-secondary);
  padding: 1.125rem;
  border-radius: 8px;
  font-style: italic;
  color: var(--text-primary);
  line-height: 1.6;
  border-left: 4px solid var(--color-primary);

  @media (min-width: 768px) {
    padding: 1.25rem;
  }
}

.voices {
  display: grid;
  gap: 1rem;
  margin-top: 0.75rem;
}

.voice {
  background: var(--bg-secondary);
  border-radius: 10px;
  padding: 1rem 1.125rem;
  border-left: 4px solid var(--color-primary);

  &__header {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 0.4rem;
  }

  &__title {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  &__count {
    font-size: 0.75rem;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  &__summary {
    margin: 0 0 0.5rem;
    color: var(--text-primary);
    line-height: 1.55;
    font-style: italic;
  }

  &__themes {
    margin: 0;
    padding-left: 1.1rem;
    color: var(--text-primary);
    font-size: 0.92rem;
    line-height: 1.5;

    li {
      margin-bottom: 0.3rem;
    }
  }
}

.delta {
  display: grid;
  gap: 1rem;
  margin-top: 0.75rem;

  @media (min-width: 768px) {
    grid-template-columns: repeat(3, 1fr);
  }

  &__group {
    background: var(--bg-secondary);
    border-radius: 8px;
    padding: 1rem;
    border-left: 3px solid var(--color-primary);
  }

  &__title {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 0.25rem;
  }

  &__hint {
    font-size: 0.8rem;
    color: var(--text-secondary);
    margin: 0 0 0.5rem;
    line-height: 1.4;
  }

  &__list {
    margin: 0;
    padding-left: 1.1rem;
    color: var(--text-primary);
    font-size: 0.9rem;
    line-height: 1.5;

    li {
      margin-bottom: 0.35rem;
    }
  }
}
</style>
