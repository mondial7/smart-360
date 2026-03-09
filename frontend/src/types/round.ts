import type { User } from './user'

export type RoundStatus = 'draft' | 'active' | 'closed' | 'shared'

export interface FeedbackRound {
  id: number
  subjectId: number
  subject?: User
  createdById: number
  createdBy?: User
  deadline: string | null
  status: RoundStatus
  createdAt: string
  updatedAt: string
  reviewers?: RoundReviewer[]
}

export interface RoundReviewer {
  id: number
  roundId: number
  reviewerId: number
  reviewer?: User
  createdAt: string
}

export interface CreateRoundRequest {
  subjectId: number
  reviewerIds: number[]
  deadline: string
}
