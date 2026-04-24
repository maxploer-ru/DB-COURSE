export type Role = 'admin' | 'author' | 'user'

export type User = {
  id: number
  username: string
  email: string
  role: Role | string
  notificationsEnabled: boolean
}

export type AuthTokens = {
  accessToken: string
}

export type AuthResponse = {
  user: User
  tokens: AuthTokens
}

export type MessageResponse = {
  message: string
}

export type Video = {
  id: number
  channelId: number
  channelName: string
  title: string
  description: string
  views: number
  likes: number
  dislikes: number
  comments: number
  createdAt: string
}

export type Comment = {
  id: number
  userId: number
  username: string
  videoId: number
  content: string
  likes: number
  dislikes: number
  createdAt: string
}

export type Channel = {
  id: number
  userId: number
  name: string
  description: string
  subscribersCount: number
}

export type CommentListResponse = {
  comments: Comment[]
  total: number
}

export type SubscriptionChannel = {
  id: number
  userId: number
  name: string
  description: string
  subscribersCount: number
  newVideosCount: number
  subscribedAt: string
}

export type UploadPresignedResponse = {
  url: string
  file_key: string
}

export type StreamUrlResponse = {
  url: string
}

export type PlaylistItem = {
  videoId: number
  videoTitle: string
  number: number
}

export type Playlist = {
  id: number
  channelId: number
  name: string
  description: string
  createdAt: string
  items: PlaylistItem[]
}

export type ApiError = {
  error: {
    code: string
    message: string
  }
}



