export interface AdminTotals {
  users: number
  teams: number
  rounds: number
  submissions: number
  consolidationsShared: number
  completionRate: number
  avgResponseSeconds: number
}

export interface RoundsByStatus {
  draft: number
  active: number
  closed: number
  shared: number
}

export type RoundStatus = keyof RoundsByStatus

export interface CompletionPoint {
  roundId: string
  subjectName: string
  createdAt: string
  expected: number
  received: number
  completionRate: number
  status: RoundStatus
}

export interface TeamActivity {
  teamId: string
  teamName: string
  memberCount: number
  activeRounds: number
  totalSubmissions: number
  avgResponseSeconds: number
}

export interface ThemePhrase {
  phrase: string
  count: number
}

export interface AdminThemes {
  strengths: ThemePhrase[]
  improvements: ThemePhrase[]
}

export interface AdminAnalytics {
  totals: AdminTotals
  roundsByStatus: RoundsByStatus
  completionTrend: CompletionPoint[]
  teamActivity: TeamActivity[]
  topThemes: AdminThemes
}
