export interface User {
  id: string  // Changed from number to string for ObjectID
  email: string
  name: string
  photoUrl: string
  role: 'admin' | 'member'
  createdAt: string
  lastLogin: string | null
}
