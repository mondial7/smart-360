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
      const subRes = await apiClient.get(`/submissions/${roundId}`)
      submissions.value = subRes.data || []
    } catch (error) {
      submissions.value = []
    }
    
    // Try to load existing consolidation
    try {
      const consRes = await apiClient.get(`/consolidations/${roundId}`)
      consolidation.value = consRes.data
      parseConsolidationFields(consolidation.value)
      adminNotes.value = consRes.data.adminNotes || ''
    } catch (error) {
      console.error('Error loading consolidation:', error)
      // No consolidation yet
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
    
    await apiClient.put(`/consolidations/${consolidation.value.id}`, updateData)
    
    // Update local consolidation data
    Object.assign(consolidation.value, updateData)
    
    editingSection.value = null
    alert('Changes saved successfully!')
  } catch (error) {
    console.error('Failed to save edits:', error)
    alert('Failed to save changes')
  }
}

function addArrayItem(array: string[], value: string) {
  if (value.trim()) {
    array.push(value.trim())
  }
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
</script>

<template>
  <div class="consolidation-page">
    <div v-if="loading" class="loading">Loading...</div>
    
    <template v-else-if="round">
      <header class="page-header">
        <router-link to="/rounds" class="back-link">← Back to Rounds</router-link>
        <h1>Feedback Consolidation</h1>
        <p class="subtitle">
          For <strong>{{ round.subject?.name }}</strong> • 
          {{ round.reviewers?.length || 0 }} reviewers • 
          Status: <span :class="['status-badge', round.status]">{{ round.status }}</span>
        </p>
      </header>

      <!-- Generate Button -->
      <div v-if="!consolidation && round.status === 'closed' && submissions.length > 0" class="generate-section">
        <p>This round is closed and ready for feedback consolidation.</p>
        <button 
          class="btn-primary" 
          @click="generateConsolidation"
          :disabled="generating"
        >
          {{ generating ? 'Generating...' : '🤖 Generate Consolidation' }}
        </button>
      </div>

      <div v-else-if="!consolidation && round.status === 'closed' && submissions.length === 0" class="info-section">
        <p>No feedback submissions found. Cannot generate consolidation.</p>
        <button class="btn-secondary" @click="router.push('/rounds')">
          Back to Rounds
        </button>
      </div>

      <div v-else-if="!consolidation && round.status !== 'closed'" class="info-section">
        <p>Round must be closed before generating consolidation.</p>
        <button class="btn-secondary" @click="router.push('/rounds')">
          Close Round First
        </button>
      </div>

      <!-- Consolidation View -->
      <div v-else-if="consolidation" class="consolidation-content">
        <div class="consolidation-header">
          <div class="meta">
            <span>Generated {{ formatDate(consolidation.createdAt) }}</span>
            <span v-if="consolidation.sharedAt" class="shared-badge">
              ✓ Shared {{ formatDate(consolidation.sharedAt) }}
            </span>
          </div>
          
          <div v-if="!consolidation.sharedAt" class="actions">
            <button 
              class="btn-primary" 
              @click="shareConsolidation"
              :disabled="sharing"
            >
              {{ sharing ? 'Sharing...' : '📤 Share with Subject' }}
            </button>
          </div>
        </div>

        <div class="consolidation-body">
          <!-- Executive Summary -->
          <section class="section">
            <div class="section-header">
              <h2>Executive Summary</h2>
              <button v-if="!consolidation.sharedAt && editingSection !== 'executive'" @click="startEditing('executive')" class="edit-btn">
                ✏️ Edit
              </button>
            </div>
            <div v-if="editingSection === 'executive'" class="edit-mode">
              <textarea
                v-model="editForm.executiveSummary"
                rows="4"
                placeholder="Enter executive summary..."
                class="edit-textarea"
              ></textarea>
              <div class="edit-actions">
                <button @click="saveEdits" class="btn-primary">Save</button>
                <button @click="cancelEditing" class="btn-secondary">Cancel</button>
              </div>
            </div>
            <p v-else class="summary-text">{{ consolidation.executiveSummary }}</p>
          </section>

          <!-- Strengths -->
          <section class="section">
            <div class="section-header">
              <h2>💪 Key Strengths</h2>
              <button v-if="!consolidation.sharedAt && editingSection !== 'strengths'" @click="startEditing('strengths')" class="edit-btn">
                ✏️ Edit
              </button>
            </div>
            <div v-if="editingSection === 'strengths'" class="edit-mode">
              <div class="array-edit">
                <div v-for="(strength, i) in editForm.strengths" :key="i" class="array-item">
                  <input v-model="editForm.strengths[i]" placeholder="Enter strength..." class="array-input">
                  <button @click="removeArrayItem(editForm.strengths, i)" class="remove-btn">×</button>
                </div>
                <button @click="addArrayItem(editForm.strengths, '')" class="add-btn">+ Add Strength</button>
              </div>
              <div class="edit-actions">
                <button @click="saveEdits" class="btn-primary">Save</button>
                <button @click="cancelEditing" class="btn-secondary">Cancel</button>
              </div>
            </div>
            <ul v-else class="styled-list positive">
              <li v-for="(strength, i) in consolidation.strengths" :key="i">
                {{ strength }}
              </li>
            </ul>
          </section>

          <!-- Areas for Improvement -->
          <section class="section">
            <div class="section-header">
              <h2>📈 Areas for Improvement</h2>
              <button v-if="!consolidation.sharedAt && editingSection !== 'improvements'" @click="startEditing('improvements')" class="edit-btn">
                ✏️ Edit
              </button>
            </div>
            <div v-if="editingSection === 'improvements'" class="edit-mode">
              <div class="array-edit">
                <div v-for="(improvement, i) in editForm.areasForImprovement" :key="i" class="array-item">
                  <input v-model="editForm.areasForImprovement[i]" placeholder="Enter improvement..." class="array-input">
                  <button @click="removeArrayItem(editForm.areasForImprovement, i)" class="remove-btn">×</button>
                </div>
                <button @click="addArrayItem(editForm.areasForImprovement, '')" class="add-btn">+ Add Improvement</button>
              </div>
              <div class="edit-actions">
                <button @click="saveEdits" class="btn-primary">Save</button>
                <button @click="cancelEditing" class="btn-secondary">Cancel</button>
              </div>
            </div>
            <ul v-else class="styled-list improvement">
              <li v-for="(area, i) in consolidation.areasForImprovement" :key="i">
                {{ area }}
              </li>
            </ul>
          </section>

          <!-- Actionable Insights -->
          <section class="section">
            <div class="section-header">
              <h2>🎯 Actionable Insights</h2>
              <button v-if="!consolidation.sharedAt && editingSection !== 'insights'" @click="startEditing('insights')" class="edit-btn">
                ✏️ Edit
              </button>
            </div>
            <div v-if="editingSection === 'insights'" class="edit-mode">
              <div class="array-edit">
                <div v-for="(insight, i) in editForm.actionableInsights" :key="i" class="array-item">
                  <input v-model="editForm.actionableInsights[i]" placeholder="Enter insight..." class="array-input">
                  <button @click="removeArrayItem(editForm.actionableInsights, i)" class="remove-btn">×</button>
                </div>
                <button @click="addArrayItem(editForm.actionableInsights, '')" class="add-btn">+ Add Insight</button>
              </div>
              <div class="edit-actions">
                <button @click="saveEdits" class="btn-primary">Save</button>
                <button @click="cancelEditing" class="btn-secondary">Cancel</button>
              </div>
            </div>
            <ul v-else class="styled-list action">
              <li v-for="(insight, i) in consolidation.actionableInsights" :key="i">
                {{ insight }}
              </li>
            </ul>
          </section>

          <!-- Question Summaries -->
          <section class="section question-summaries">
            <div class="section-header">
              <h2>📋 Detailed Question Analysis</h2>
              <button v-if="!consolidation.sharedAt && editingSection !== 'questions'" @click="startEditing('questions')" class="edit-btn">
                ✏️ Edit
              </button>
            </div>
            <div v-if="editingSection === 'questions'" class="edit-mode">
              <div class="question-edit">
                <div class="question-card">
                  <h4>1. Key Strengths</h4>
                  <textarea v-model="editForm.questionSummaries.a" placeholder="Summary of strengths..." class="edit-textarea"></textarea>
                </div>
                <div class="question-card">
                  <h4>2. Areas to Improve</h4>
                  <textarea v-model="editForm.questionSummaries.b" placeholder="Summary of improvements..." class="edit-textarea"></textarea>
                </div>
                <div class="question-card">
                  <h4>3. Observed Behaviors</h4>
                  <textarea v-model="editForm.questionSummaries.c" placeholder="Summary of behaviors..." class="edit-textarea"></textarea>
                </div>
                <div class="question-card">
                  <h4>4. Growth Advice</h4>
                  <textarea v-model="editForm.questionSummaries.d" placeholder="Summary of advice..." class="edit-textarea"></textarea>
                </div>
              </div>
              <div class="edit-actions">
                <button @click="saveEdits" class="btn-primary">Save</button>
                <button @click="cancelEditing" class="btn-secondary">Cancel</button>
              </div>
            </div>
            <div v-else class="question-cards">
              <div class="question-card">
                <h4>1. Key Strengths</h4>
                <p>{{ consolidation.questionSummaries?.a }}</p>
              </div>
              <div class="question-card">
                <h4>2. Areas to Improve</h4>
                <p>{{ consolidation.questionSummaries?.b }}</p>
              </div>
              <div class="question-card">
                <h4>3. Observed Behaviors</h4>
                <p>{{ consolidation.questionSummaries?.c }}</p>
              </div>
              <div class="question-card">
                <h4>4. Growth Advice</h4>
                <p>{{ consolidation.questionSummaries?.d }}</p>
              </div>
            </div>
          </section>

          <!-- Admin Notes -->
          <section v-if="!consolidation.sharedAt" class="section admin-notes">
            <h2>📝 Admin Notes (Optional)</h2>
            <p class="hint">Add context or guidance for the subject. These notes will be included when shared.</p>
            <textarea
              v-model="adminNotes"
              rows="4"
              placeholder="Add your notes here..."
            ></textarea>
            <button class="btn-secondary" @click="saveNotes">
              Save Notes
            </button>
          </section>
          
          <section v-else-if="consolidation.adminNotes" class="section">
            <h2>📝 Admin Notes</h2>
            <p class="admin-note-display">{{ consolidation.adminNotes }}</p>
          </section>
        </div>
      </div>
    </template>

    <div v-else class="error-state">
      <p>Round not found.</p>
      <router-link to="/rounds" class="btn-primary">Back to Rounds</router-link>
    </div>
  </div>
</template>

<style scoped>
.consolidation-page {
  padding: 2rem;
  max-width: 900px;
  margin: 0 auto;
}

.loading, .error-state {
  text-align: center;
  padding: 3rem;
}

.page-header {
  margin-bottom: 2rem;
}

.back-link {
  display: block;
  margin-bottom: 1rem;
  color: #667eea;
  text-decoration: none;
}

.page-header h1 {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

.subtitle {
  color: #666;
}

.subtitle strong {
  color: #333;
}

.status-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: capitalize;
}

.status-badge.closed {
  background: #ffebee;
  color: #c62828;
}

.status-badge.shared {
  background: #e3f2fd;
  color: #1976d2;
}

.generate-section, .info-section {
  text-align: center;
  padding: 3rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}

.generate-section p, .info-section p {
  color: #666;
  margin-bottom: 1.5rem;
}

.consolidation-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  overflow: hidden;
}

.consolidation-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  background: #f8f9fa;
  border-bottom: 1px solid #e0e0e0;
}

.meta {
  display: flex;
  gap: 1rem;
  align-items: center;
  color: #666;
  font-size: 0.9rem;
}

.shared-badge {
  color: #4caf50;
  font-weight: 500;
}

.consolidation-body {
  padding: 1.5rem;
}

.section {
  margin-bottom: 2rem;
  padding-bottom: 2rem;
  border-bottom: 1px solid #eee;
}

.section:last-child {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}

.section h2 {
  font-size: 1.25rem;
  margin-bottom: 1rem;
  color: #333;
}

.summary-text {
  font-size: 1.1rem;
  line-height: 1.6;
  color: #444;
}

.styled-list {
  list-style: none;
  padding: 0;
}

.styled-list li {
  padding: 1rem;
  margin-bottom: 0.75rem;
  border-radius: 8px;
  position: relative;
  padding-left: 2.5rem;
}

.styled-list li::before {
  position: absolute;
  left: 0.75rem;
  font-size: 1.25rem;
}

.styled-list.positive li {
  background: #e8f5e9;
}

.styled-list.positive li::before {
  content: '✓';
  color: #4caf50;
}

.styled-list.improvement li {
  background: #fff3e0;
}

.styled-list.improvement li::before {
  content: '↑';
  color: #ff9800;
}

.styled-list.action li {
  background: #e3f2fd;
}

.styled-list.action li::before {
  content: '→';
  color: #2196f3;
}

.question-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1rem;
}

.question-card {
  padding: 1.5rem;
  background: #f8f9fa;
  border-radius: 8px;
}

.question-card h4 {
  color: #667eea;
  font-size: 0.9rem;
  margin-bottom: 0.5rem;
}

.question-card p {
  color: #666;
  font-size: 0.9rem;
  line-height: 1.5;
}

.admin-notes .hint {
  color: #888;
  font-size: 0.9rem;
  margin-bottom: 1rem;
}

.admin-notes textarea {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-family: inherit;
  font-size: 1rem;
  resize: vertical;
  margin-bottom: 1rem;
}

.admin-note-display {
  background: #f5f5f5;
  padding: 1rem;
  border-radius: 8px;
  font-style: italic;
  color: #666;
}

.btn-primary, .btn-secondary {
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  text-decoration: none;
  display: inline-block;
}

.btn-primary {
  background: #667eea;
  color: white;
  border: none;
}

.btn-primary:hover:not(:disabled) {
  background: #5a6fd6;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: white;
  color: #666;
  border: 1px solid #ddd;
}

.btn-secondary:hover {
  background: #f5f5f5;
}

/* Edit Mode Styles */
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.edit-btn {
  background: none;
  border: 1px solid #667eea;
  color: #667eea;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
}

.edit-btn:hover {
  background: #667eea;
  color: white;
}

.edit-mode {
  background: #f8f9fa;
  padding: 1rem;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
  margin-bottom: 1rem;
}

.edit-textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: inherit;
  font-size: 1rem;
  resize: vertical;
  margin-bottom: 1rem;
}

.edit-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}

.array-edit {
  margin-bottom: 1rem;
}

.array-item {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  align-items: center;
}

.array-input {
  flex: 1;
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 0.9rem;
}

.remove-btn {
  background: #f44336;
  color: white;
  border: none;
  border-radius: 50%;
  width: 24px;
  height: 24px;
  cursor: pointer;
  font-size: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.remove-btn:hover {
  background: #d32f2f;
}

.add-btn {
  background: #4caf50;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
}

.add-btn:hover {
  background: #45a049;
}

.question-edit {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1rem;
  margin-bottom: 1rem;
}

.question-edit .question-card {
  background: white;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 1rem;
}

.question-edit .question-card h4 {
  margin-bottom: 0.5rem;
  color: #667eea;
}

.question-edit .edit-textarea {
  margin: 0;
  min-height: 80px;
}
</style>
