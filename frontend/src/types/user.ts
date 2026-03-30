export interface User {
  id: string  // Changed from number to string for ObjectID
  email: string
  name: string
  photoUrl: string
  role: 'admin' | 'team_admin' | 'member'
  teamId?: string | null
  createdAt: string
  lastLogin: string | null
}

export interface UserWithFeedbackStats extends User {
  lastFeedbackReceived: string | null
  activeRoundsAsSubject: number
  pendingReviews: number
  totalFeedbackReceived: number
}
