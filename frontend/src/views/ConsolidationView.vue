<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { FeedbackRound } from '@/types/round'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const roundId = route.params.roundId as string  // Changed from parseInt to string
const round = ref<FeedbackRound | null>(null)
const consolidation = ref<any>(null)
const loading = ref(true)
const generating = ref(false)
const sharing = ref(false)
const adminNotes = ref('')
const submissions = ref<any[]>([])
const editingSection = ref<string | null>(null)
const editForm = ref({
  executiveSummary: '',
  strengths: [] as string[],
  areasForImprovement: [] as string[],
  actionableInsights: [] as string[],
  questionSummaries: {} as Record<string, string>
})
const showConsolidationContent = ref(true)

onMounted(async () => {
  await loadData()
})

async function loadData() {
  try {
    // Load round details
    const roundRes = await apiClient.get(`/rounds/${roundId}`)
    round.value = roundRes.data
    
    // Load submissions to check if any exist
    try {
      const subRes = await apiClient.get(`/submissions/round/${roundId}`)
      console.log('Submissions API response:', subRes.data)
      submissions.value = subRes.data || []
      console.log('Submissions count:', submissions.value.length)
    } catch (error) {
      console.error('Error loading submissions:', error)
      submissions.value = []
    }
    
    // Try to load existing consolidation
    try {
      const consRes = await apiClient.get(`/consolidations/${roundId}`)
      consolidation.value = consRes.data
      parseConsolidationFields(consolidation.value)
      adminNotes.value = consRes.data.adminNotes || ''
      showConsolidationContent.value = false // Hide content by default when consolidation exists
    } catch (error) {
      console.error('Error loading consolidation:', error)
      // No consolidation yet
      showConsolidationContent.value = true
    }
  } catch (error) {
    console.error('Failed to load data:', error)
  } finally {
    loading.value = false
  }
}

async function generateConsolidation() {
  generating.value = true
  try {
    const res = await apiClient.post(`/rounds/${roundId}/consolidate`)
    consolidation.value = res.data
    parseConsolidationFields(consolidation.value)
    adminNotes.value = res.data.adminNotes || ''
    showConsolidationContent.value = true // Show content after generation
  } catch (error: any) {
    if (error.response?.status === 409) {
      alert('Consolidation already exists')
      await loadData()
    } else {
      alert(error.response?.data?.error || 'Failed to generate consolidation')
    }
  } finally {
    generating.value = false
  }
}

async function saveNotes() {
  if (!consolidation.value) return
  try {
    await apiClient.put(`/consolidations/${consolidation.value.id}/notes`, {
      adminNotes: adminNotes.value
    })
    alert('Notes saved')
  } catch (error) {
    alert('Failed to save notes')
  }
}

async function shareConsolidation() {
  if (!consolidation.value) return
  if (!confirm('Share this consolidated feedback with the subject?')) return
  
  sharing.value = true
  try {
    await apiClient.post(`/consolidations/${consolidation.value.id}/share`)
    
    // Update local consolidation data
    consolidation.value.sharedAt = new Date().toISOString()
    
    // Update round status to 'shared'
    if (round.value) {
      try {
        await apiClient.put(`/rounds/${roundId}`, { status: 'shared' })
        round.value.status = 'shared'
      } catch (error) {
        console.error('Failed to update round status:', error)
      }
    }
    
    alert('Consolidation shared successfully!')
    await loadData()
  } catch (error) {
    alert('Failed to share consolidation')
  } finally {
    sharing.value = false
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

function startEditing(section: string) {
  if (!consolidation.value) return
  
  editingSection.value = section
  editForm.value = {
    executiveSummary: consolidation.value.executiveSummary || '',
    strengths: [...(consolidation.value.strengths || [])],
    areasForImprovement: [...(consolidation.value.areasForImprovement || [])],
    actionableInsights: [...(consolidation.value.actionableInsights || [])],
    questionSummaries: {...(consolidation.value.questionSummaries || {})}
  }
}

function cancelEditing() {
  editingSection.value = null
  editForm.value = {
    executiveSummary: '',
    strengths: [],
    areasForImprovement: [],
    actionableInsights: [],
    questionSummaries: {}
  }
}

async function saveEdits() {
  if (!consolidation.value || !editingSection.value) return
  
  try {
    const updateData: any = {}
    
    // Only update the section being edited
    switch (editingSection.value) {
      case 'executive':
        updateData.executiveSummary = editForm.value.executiveSummary
        break
      case 'strengths':
        updateData.strengths = editForm.value.strengths
        break
      case 'improvements':
        updateData.areasForImprovement = editForm.value.areasForImprovement
        break
      case 'insights':
        updateData.actionableInsights = editForm.value.actionableInsights
        break
      case 'questions':
        updateData.questionSummaries = editForm.value.questionSummaries
        break
    }
    
    console.log('Saving consolidation edits:', updateData)
    console.log('Consolidation ID:', consolidation.value.id)
    
    await apiClient.put(`/consolidations/${consolidation.value.id}`, updateData)
    
    console.log('Save successful, updating local data')
    
    // Update local consolidation data
    Object.assign(consolidation.value, updateData)
    
    editingSection.value = null
    alert('Changes saved successfully!')
  } catch (error: any) {
    console.error('Failed to save edits:', error)
    console.error('Error response:', error.response?.data)
    alert(error.response?.data?.error || 'Failed to save changes')
  }
}

function addArrayItem(array: string[]) {
  // Add an empty string to create a new input field
  array.push('')
}

function removeArrayItem(array: string[], index: number) {
  array.splice(index, 1)
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'Not shared yet'
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function parseResponses(responsesStr: string): Record<string, string> {
  try {
    return JSON.parse(responsesStr)
  } catch (error) {
    console.error('Error parsing responses:', error)
    return {}
  }
}
</script>

<template>
  <div class="consolidation">
    <div v-if="loading" class="consolidation__loading">Loading...</div>

    <template v-else-if="round">
      <header class="consolidation__header">
        <router-link to="/rounds" class="consolidation__back">← Back to Rounds</router-link>
        <h1 class="consolidation__title">Feedback Consolidation</h1>
        <p class="consolidation__subtitle">
          For <strong class="consolidation__subject">{{ round.subject?.name }}</strong> •
          {{ round.reviewers?.length || 0 }} reviewers •
          Status: <span class="badge" :class="[`badge--${round.status}`]">{{ round.status }}</span>
        </p>
      </header>

      <!-- Generate Button -->
      <div v-if="!consolidation && round.status === 'closed' && submissions.length > 0" class="consolidation__prompt">
        <p class="consolidation__prompt-text">This round is closed and ready for feedback consolidation.</p>
        <button
          class="btn btn--primary"
          @click="generateConsolidation"
          :disabled="generating"
        >
          {{ generating ? 'Generating...' : '🤖 Generate Consolidation' }}
        </button>
      </div>

      <div v-else-if="!consolidation && round.status === 'closed' && submissions.length === 0" class="consolidation__prompt">
        <p class="consolidation__prompt-text">No feedback submissions found. Cannot generate consolidation.</p>
        <button class="btn btn--secondary" @click="router.push('/rounds')">
          Back to Rounds
        </button>
      </div>

      <div v-else-if="!consolidation && round.status !== 'closed'" class="consolidation__prompt">
        <p class="consolidation__prompt-text">Round must be closed before generating consolidation.</p>
        <button class="btn btn--secondary" @click="router.push('/rounds')">
          Close Round First
        </button>
      </div>

      <!-- View Feedback Button -->
      <div v-else-if="consolidation && !showConsolidationContent" class="consolidation__prompt">
        <p class="consolidation__prompt-text">Feedback consolidation has been generated for this round.</p>
        <button
          class="btn btn--primary"
          @click="showConsolidationContent = true"
        >
          👁️ View Consolidated Feedback
        </button>
      </div>

      <!-- Consolidation View -->
      <div v-else-if="consolidation" class="consolidation__content">
        <div class="consolidation__meta">
          <div class="consolidation__meta-info">
            <span class="consolidation__meta-item">Generated {{ formatDate(consolidation.createdAt) }}</span>
            <span v-if="consolidation.sharedAt" class="consolidation__meta-item consolidation__meta-item--shared">
              ✓ Shared {{ formatDate(consolidation.sharedAt) }}
            </span>
            <span class="consolidation__meta-item">Round Status: <span class="badge" :class="[`badge--${round.status}`]">{{ round.status }}</span></span>
          </div>

          <div v-if="!consolidation.sharedAt" class="consolidation__meta-actions">
            <button
              class="btn btn--primary"
              @click="shareConsolidation"
              :disabled="sharing"
            >
              {{ sharing ? 'Sharing...' : '📤 Share with Subject' }}
            </button>
          </div>
        </div>

        <div class="consolidation__body">
          <!-- Executive Summary -->
          <section class="section">
            <div class="section__header">
              <h2 class="section__title">Executive Summary</h2>
              <button v-if="!consolidation.sharedAt && editingSection !== 'executive'" @click="startEditing('executive')" class="btn btn--secondary section__edit-btn">
                ✏️ Edit
              </button>
            </div>
            <div v-if="editingSection === 'executive'" class="section__edit">
              <textarea
                v-model="editForm.executiveSummary"
                rows="4"
                placeholder="Enter executive summary..."
                class="section__textarea"
              ></textarea>
              <div class="section__edit-actions">
                <button @click="saveEdits" class="btn btn--primary">Save</button>
                <button @click="cancelEditing" class="btn btn--secondary">Cancel</button>
              </div>
            </div>
            <p v-else class="section__summary">{{ consolidation.executiveSummary }}</p>
          </section>

          <!-- Strengths -->
          <section class="section">
            <div class="section__header">
              <h2 class="section__title">💪 Key Strengths</h2>
              <button v-if="!consolidation.sharedAt && editingSection !== 'strengths'" @click="startEditing('strengths')" class="btn btn--secondary section__edit-btn">
                ✏️ Edit
              </button>
            </div>
            <div v-if="editingSection === 'strengths'" class="section__edit">
              <div class="array-editor">
                <div v-for="(strength, i) in editForm.strengths" :key="i" class="array-editor__item">
                  <input v-model="editForm.strengths[i]" placeholder="Enter strength..." class="array-editor__input">
                  <button @click="removeArrayItem(editForm.strengths, i)" class="array-editor__remove">×</button>
                </div>
                <button @click="addArrayItem(editForm.strengths)" class="array-editor__add">+ Add Strength</button>
              </div>
              <div class="section__edit-actions">
                <button @click="saveEdits" class="btn btn--primary">Save</button>
                <button @click="cancelEditing" class="btn btn--secondary">Cancel</button>
              </div>
            </div>
            <ul v-else class="feedback-list feedback-list--positive">
              <li v-for="(strength, i) in consolidation.strengths" :key="i" class="feedback-list__item">
                {{ strength }}
              </li>
            </ul>
          </section>

          <!-- Areas for Improvement -->
          <section class="section">
            <div class="section__header">
              <h2 class="section__title">📈 Areas for Improvement</h2>
              <button v-if="!consolidation.sharedAt && editingSection !== 'improvements'" @click="startEditing('improvements')" class="btn btn--secondary section__edit-btn">
                ✏️ Edit
              </button>
            </div>
            <div v-if="editingSection === 'improvements'" class="section__edit">
              <div class="array-editor">
                <div v-for="(improvement, i) in editForm.areasForImprovement" :key="i" class="array-editor__item">
                  <input v-model="editForm.areasForImprovement[i]" placeholder="Enter improvement..." class="array-editor__input">
                  <button @click="removeArrayItem(editForm.areasForImprovement, i)" class="array-editor__remove">×</button>
                </div>
                <button @click="addArrayItem(editForm.areasForImprovement)" class="array-editor__add">+ Add Improvement</button>
              </div>
              <div class="section__edit-actions">
                <button @click="saveEdits" class="btn btn--primary">Save</button>
                <button @click="cancelEditing" class="btn btn--secondary">Cancel</button>
              </div>
            </div>
            <ul v-else class="feedback-list feedback-list--improvement">
              <li v-for="(area, i) in consolidation.areasForImprovement" :key="i" class="feedback-list__item">
                {{ area }}
              </li>
            </ul>
          </section>

          <!-- Actionable Insights -->
          <section class="section">
            <div class="section__header">
              <h2 class="section__title">🎯 Actionable Insights</h2>
              <button v-if="!consolidation.sharedAt && editingSection !== 'insights'" @click="startEditing('insights')" class="btn btn--secondary section__edit-btn">
                ✏️ Edit
              </button>
            </div>
            <div v-if="editingSection === 'insights'" class="section__edit">
              <div class="array-editor">
                <div v-for="(insight, i) in editForm.actionableInsights" :key="i" class="array-editor__item">
                  <input v-model="editForm.actionableInsights[i]" placeholder="Enter insight..." class="array-editor__input">
                  <button @click="removeArrayItem(editForm.actionableInsights, i)" class="array-editor__remove">×</button>
                </div>
                <button @click="addArrayItem(editForm.actionableInsights)" class="array-editor__add">+ Add Insight</button>
              </div>
              <div class="section__edit-actions">
                <button @click="saveEdits" class="btn btn--primary">Save</button>
                <button @click="cancelEditing" class="btn btn--secondary">Cancel</button>
              </div>
            </div>
            <ul v-else class="feedback-list feedback-list--action">
              <li v-for="(insight, i) in consolidation.actionableInsights" :key="i" class="feedback-list__item">
                {{ insight }}
              </li>
            </ul>
          </section>

          <!-- Question Summaries -->
          <section class="section">
            <div class="section__header">
              <h2 class="section__title">📋 Detailed Question Analysis</h2>
              <button v-if="!consolidation.sharedAt && editingSection !== 'questions'" @click="startEditing('questions')" class="btn btn--secondary section__edit-btn">
                ✏️ Edit
              </button>
            </div>
            <div v-if="editingSection === 'questions'" class="section__edit">
              <div class="questions-editor">
                <div class="question-card">
                  <h4 class="question-card__title">1. Key Strengths</h4>
                  <textarea v-model="editForm.questionSummaries.a" placeholder="Summary of strengths..." class="section__textarea"></textarea>
                </div>
                <div class="question-card">
                  <h4 class="question-card__title">2. Areas to Improve</h4>
                  <textarea v-model="editForm.questionSummaries.b" placeholder="Summary of improvements..." class="section__textarea"></textarea>
                </div>
                <div class="question-card">
                  <h4 class="question-card__title">3. Observed Behaviors</h4>
                  <textarea v-model="editForm.questionSummaries.c" placeholder="Summary of behaviors..." class="section__textarea"></textarea>
                </div>
                <div class="question-card">
                  <h4 class="question-card__title">4. Growth Advice</h4>
                  <textarea v-model="editForm.questionSummaries.d" placeholder="Summary of advice..." class="section__textarea"></textarea>
                </div>
              </div>
              <div class="section__edit-actions">
                <button @click="saveEdits" class="btn btn--primary">Save</button>
                <button @click="cancelEditing" class="btn btn--secondary">Cancel</button>
              </div>
            </div>
            <div v-else class="questions">
              <div class="question-card">
                <h4 class="question-card__title">1. Key Strengths</h4>
                <p class="question-card__text">{{ consolidation.questionSummaries?.a }}</p>
              </div>
              <div class="question-card">
                <h4 class="question-card__title">2. Areas to Improve</h4>
                <p class="question-card__text">{{ consolidation.questionSummaries?.b }}</p>
              </div>
              <div class="question-card">
                <h4 class="question-card__title">3. Observed Behaviors</h4>
                <p class="question-card__text">{{ consolidation.questionSummaries?.c }}</p>
              </div>
              <div class="question-card">
                <h4 class="question-card__title">4. Growth Advice</h4>
                <p class="question-card__text">{{ consolidation.questionSummaries?.d }}</p>
              </div>
            </div>
          </section>

          <!-- Individual Reviews -->
          <section v-if="submissions.length > 0" class="section">
            <h2 class="section__title">📝 Individual Reviews ({{ submissions.length }})</h2>
            <p class="section__hint">View all individual feedback submissions below</p>
            <div class="reviews">
              <div v-for="(submission, idx) in submissions" :key="submission.id" class="review">
                <div class="review__header">
                  <span class="review__label">Reviewer {{ idx + 1 }}</span>
                  <span class="review__date">{{ formatDate(submission.submittedAt) }}</span>
                </div>
                <div class="review__content">
                  <div class="review__item">
                    <h4 class="review__question">What are this person's key strengths?</h4>
                    <p class="review__answer">{{ parseResponses(submission.responses).a || 'No response' }}</p>
                  </div>
                  <div class="review__item">
                    <h4 class="review__question">What areas could they improve?</h4>
                    <p class="review__answer">{{ parseResponses(submission.responses).b || 'No response' }}</p>
                  </div>
                  <div class="review__item">
                    <h4 class="review__question">What should they continue doing?</h4>
                    <p class="review__answer">{{ parseResponses(submission.responses).c || 'No response' }}</p>
                  </div>
                  <div class="review__item">
                    <h4 class="review__question">What should they start or stop doing?</h4>
                    <p class="review__answer">{{ parseResponses(submission.responses).d || 'No response' }}</p>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <!-- Admin Notes -->
          <section v-if="!consolidation.sharedAt" class="section">
            <h2 class="section__title">📝 Admin Notes (Optional)</h2>
            <p class="section__hint">Add context or guidance for the subject. These notes will be included when shared.</p>
            <textarea
              v-model="adminNotes"
              rows="4"
              placeholder="Add your notes here..."
              class="section__textarea"
            ></textarea>
            <button class="btn btn--secondary" @click="saveNotes">
              Save Notes
            </button>
          </section>

          <section v-else-if="consolidation.adminNotes" class="section">
            <h2 class="section__title">📝 Admin Notes</h2>
            <p class="admin-notes">{{ consolidation.adminNotes }}</p>
          </section>
        </div>
      </div>
    </template>

    <div v-else class="consolidation__error">
      <p class="consolidation__error-text">Round not found.</p>
      <router-link to="/rounds" class="btn btn--primary">Back to Rounds</router-link>
    </div>
  </div>
</template>

<style scoped lang="scss">
.consolidation {
  padding: 1rem;
  max-width: 900px;
  margin: 0 auto;

  @media (min-width: 768px) {
    padding: 2rem;
  }

  &__loading,
  &__error {
    text-align: center;
    padding: 2rem 1rem;

    @media (min-width: 768px) {
      padding: 3rem;
    }
  }

  &__error-text {
    color: var(--color-error);
    margin-bottom: 1.5rem;
  }

  &__header {
    margin-bottom: 2rem;
  }

  &__back {
    display: block;
    margin-bottom: 1rem;
    color: var(--color-primary);
    text-decoration: none;
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }

    &:hover {
      text-decoration: underline;
    }
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
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__subject {
    color: var(--text-primary);
  }

  &__prompt {
    text-align: center;
    padding: 2rem 1rem;
    background: var(--bg-primary);
    border-radius: 12px;
    border: 1px solid var(--border-color);

    @media (min-width: 768px) {
      padding: 3rem;
    }
  }

  &__prompt-text {
    color: var(--text-secondary);
    margin-bottom: 1.5rem;
  }

  &__content {
    background: var(--bg-primary);
    border-radius: 12px;
    border: 1px solid var(--border-color);
    overflow: hidden;
  }

  &__meta {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 1.25rem;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);

    @media (min-width: 768px) {
      flex-direction: row;
      justify-content: space-between;
      align-items: center;
      padding: 1.5rem;
    }
  }

  &__meta-info {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    color: var(--text-secondary);
    font-size: 0.85rem;

    @media (min-width: 768px) {
      flex-direction: row;
      gap: 1rem;
      align-items: center;
      font-size: 0.9rem;
    }
  }

  &__meta-item {
    &--shared {
      color: var(--color-success);
      font-weight: 500;
    }
  }

  &__meta-actions {
    display: flex;
    gap: 0.5rem;
  }

  &__body {
    padding: 1.25rem;

    @media (min-width: 768px) {
      padding: 1.5rem;
    }
  }
}

.section {
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

  &__header {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-bottom: 1rem;

    @media (min-width: 768px) {
      flex-direction: row;
      justify-content: space-between;
      align-items: center;
    }
  }

  &__title {
    font-size: 1.125rem;
    margin: 0;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1.25rem;
    }
  }

  &__edit-btn {
    font-size: 0.75rem;
    padding: 0.375rem 0.75rem;
    align-self: flex-start;

    @media (min-width: 768px) {
      font-size: 0.8rem;
      align-self: auto;
    }
  }

  &__summary {
    font-size: 1rem;
    line-height: 1.6;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1.1rem;
    }
  }

  &__hint {
    color: var(--text-tertiary);
    font-size: 0.85rem;
    margin-bottom: 1rem;

    @media (min-width: 768px) {
      font-size: 0.9rem;
      margin-bottom: 1.5rem;
    }
  }

  &__edit {
    background: var(--bg-secondary);
    padding: 1rem;
    border-radius: 8px;
    border: 1px solid var(--border-color);
    margin-bottom: 1rem;
  }

  &__textarea {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid var(--border-color);
    border-radius: 8px;
    font-family: inherit;
    font-size: 0.9rem;
    resize: vertical;
    margin-bottom: 1rem;
    background: var(--bg-primary);
    color: var(--text-primary);
    min-height: 80px;

    @media (min-width: 768px) {
      font-size: 1rem;
    }

    &:focus {
      outline: none;
      border-color: var(--color-primary);
    }
  }

  &__edit-actions {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;

    @media (min-width: 768px) {
      flex-direction: row;
      justify-content: flex-end;
    }
  }
}

.array-editor {
  margin-bottom: 1rem;

  &__item {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
    align-items: center;
  }

  &__input {
    flex: 1;
    padding: 0.5rem;
    border: 1px solid var(--border-color);
    border-radius: 8px;
    font-size: 0.85rem;
    background: var(--bg-primary);
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }

    &:focus {
      outline: none;
      border-color: var(--color-primary);
    }
  }

  &__remove {
    background: var(--color-error);
    color: white;
    border: none;
    border-radius: 50%;
    width: 32px;
    height: 32px;
    cursor: pointer;
    font-size: 1.25rem;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    min-height: 44px;
    min-width: 44px;

    &:hover {
      opacity: 0.9;
    }
  }

  &__add {
    background: var(--color-success);
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 8px;
    cursor: pointer;
    font-size: 0.85rem;
    min-height: 44px;

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }

    &:hover {
      opacity: 0.9;
    }
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

.questions-editor {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;
  margin-bottom: 1rem;

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
    font-size: 0.85rem;
    line-height: 1.6;
    margin: 0;

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }
  }
}

.reviews {
  display: flex;
  flex-direction: column;
  gap: 1rem;

  @media (min-width: 768px) {
    gap: 1.5rem;
  }
}

.review {
  background: var(--bg-secondary);
  border-radius: 8px;
  padding: 1.25rem;
  border: 1px solid var(--border-color);

  @media (min-width: 768px) {
    padding: 1.5rem;
  }

  &__header {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 1rem;
    padding-bottom: 0.75rem;
    border-bottom: 2px solid var(--border-color);

    @media (min-width: 768px) {
      flex-direction: row;
      justify-content: space-between;
      align-items: center;
    }
  }

  &__label {
    font-weight: 600;
    color: var(--color-primary);
    font-size: 0.95rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__date {
    font-size: 0.8rem;
    color: var(--text-tertiary);

    @media (min-width: 768px) {
      font-size: 0.85rem;
    }
  }

  &__content {
    display: flex;
    flex-direction: column;
    gap: 1rem;

    @media (min-width: 768px) {
      gap: 1.25rem;
    }
  }

  &__item {
    background: var(--bg-primary);
    padding: 1rem;
    border-radius: 6px;
    border-left: 3px solid var(--color-primary);
  }

  &__question {
    color: var(--color-primary);
    font-size: 0.85rem;
    margin: 0 0 0.5rem 0;
    font-weight: 600;

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }
  }

  &__answer {
    color: var(--text-primary);
    line-height: 1.6;
    margin: 0;
    white-space: pre-wrap;
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 0.95rem;
    }
  }
}

.admin-notes {
  background: var(--bg-secondary);
  padding: 1rem;
  border-radius: 8px;
  font-style: italic;
  color: var(--text-primary);
  line-height: 1.6;
  border-left: 4px solid var(--color-primary);
}
</style>
