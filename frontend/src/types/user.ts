export interface User {
  id: number
  email: string
  name: string
  photoUrl: string
  role: 'admin' | 'member'
  createdAt: string
  lastLogin: string | null
}
