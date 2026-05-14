import type { User } from './user'

export type RoundStatus = 'draft' | 'active' | 'closed' | 'shared'

export interface FeedbackRound {
  id: string  // Changed from number to string for ObjectID
  subjectId: string  // Changed from number to string for ObjectID
  subject?: User
  createdById: string  // Changed from number to string for ObjectID
  createdBy?: User
  templateId?: string
  deadline: string | null
  status: RoundStatus
  createdAt: string
  updatedAt: string
  reviewers?: RoundReviewer[]
}

export interface TemplateQuestion {
  key: string
  peerText: string
  selfText: string
  cardTitle: string
}

export interface TemplateCompetency {
  key: string
  name: string
  description: string
}

export interface RoundTemplate {
  id: string
  slug: string
  name: string
  description: string
  coachingPersona: string
  questions: TemplateQuestion[]
  competencies?: TemplateCompetency[]
  createdAt: string
  updatedAt: string
}

export interface CompetencyRating {
  key: string
  score: number // 1..5
  justification: string
}

export interface ManagerOnlyChannel {
  noteCount: number
  synthesis: string
  themes: string[]
  rawNotes?: string[]
}

export interface CompetencyRatingAggregate {
  key: string
  name: string
  description?: string
  selfScore?: number
  peerAverage?: number
  managerAverage?: number
  reportAverage?: number
  othersAverage?: number
  othersCount: number
  spread: number
}

export interface RoundReviewer {
  id: string  // Changed from number to string for ObjectID
  roundId: string  // Changed from number to string for ObjectID
  reviewerId: string  // Changed from number to string for ObjectID
  reviewer?: User
  createdAt: string
}

export interface CreateRoundRequest {
  subjectId: string  // Changed from number to string for ObjectID
  reviewerIds: string[]  // Changed from number[] to string[] for ObjectID
  deadline: string
}
