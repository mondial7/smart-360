import type { User } from './user'

export interface Team {
  id: string
  name: string
  teamAdminId: string
  teamAdmin?: User
  memberIds: string[]
  members: User[]
  createdAt: string
  updatedAt: string
}

export interface CreateTeamRequest {
  name: string
  teamAdminId: string
  memberIds: string[]
}

export interface UpdateTeamRequest {
  name?: string
  teamAdminId?: string
  memberIds?: string[]
}

export interface TeamRoundSubject {
  subjectId: string
  deadline: string
}

export interface CreateTeamRoundsRequest {
  subjects: TeamRoundSubject[]
  templateId?: string
}

export interface CreateTeamRoundsResponse {
  createdRounds: string[]
  successCount: number
  failedSubjects?: string[]
}
