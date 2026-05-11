export interface RoundAnalytics {
  roundId: string
  createdAt: string
  sharedAt: string
  strengthsCount: number
  improvementsCount: number
  insightsCount: number
  behaviorsHasSummary: boolean
  growthHasSummary: boolean
}

export interface RadarAxes {
  strengths: number
  improvements: number
  behaviors: number
  growth: number
}

export interface MyAnalytics {
  feedbackReceivedCount: number
  feedbackGivenCount: number
  pendingReviewsCount: number
  rounds: RoundAnalytics[]
  latestRadar: RadarAxes
}
