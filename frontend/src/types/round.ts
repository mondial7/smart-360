import type { User } from './user'

export type RoundStatus = 'draft' | 'active' | 'closed' | 'shared'

export interface FeedbackRound {
  id: string  // Changed from number to string for ObjectID
  subjectId: string  // Changed from number to string for ObjectID
  subject?: User
  createdById: string  // Changed from number to string for ObjectID
  createdBy?: User
  deadline: string | null
  status: RoundStatus
  createdAt: string
  updatedAt: string
  reviewers?: RoundReviewer[]
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
