import { apiClient } from './client'
import type {
  AuthResponse,
  Channel,
  Comment,
  CommentListResponse,
  MessageResponse,
  StreamUrlResponse,
  SubscriptionChannel,
  Playlist,
  UploadPresignedResponse,
  User,
  Video,
} from './types'

type AuthApiResponse = {
  user: UserApiResponse
  access_token: string
}

type UserApiResponse = {
  id: number
  username: string
  email: string
  role: string
  notifications_enabled: boolean
}

type VideoApiResponse = {
  id: number
  channel_id: number
  channel_name: string
  title: string
  description: string
  views: number
  likes: number
  dislikes: number
  comments: number
  created_at: string
}

type CommentApiResponse = {
  id: number
  user_id: number
  username: string
  video_id: number
  content: string
  likes: number
  dislikes: number
  created_at: string
}

type SubscriptionChannelApiResponse = {
  id: number
  user_id: number
  name: string
  description: string
  subscribers_count: number
  new_videos_count: number
  subscribed_at: string
}

type ChannelApiResponse = {
  id: number
  user_id: number
  name: string
  description: string
  subscribers_count: number
}

type PlaylistItemApiResponse = {
  video_id: number
  video_title: string
  number: number
}

type PlaylistApiResponse = {
  id: number
  channel_id: number
  name: string
  description: string
  created_at: string
  items: PlaylistItemApiResponse[]
}

function mapUser(payload: UserApiResponse): User {
  return {
    id: payload.id,
    username: payload.username,
    email: payload.email,
    role: payload.role,
    notificationsEnabled: payload.notifications_enabled,
  }
}

function mapAuthResponse(payload: AuthApiResponse): AuthResponse {
  return { user: mapUser(payload.user), tokens: { accessToken: payload.access_token } }
}

function mapVideo(payload: VideoApiResponse): Video {
  return {
    id: payload.id,
    channelId: payload.channel_id,
    channelName: payload.channel_name,
    title: payload.title,
    description: payload.description,
    views: payload.views,
    likes: payload.likes,
    dislikes: payload.dislikes,
    comments: payload.comments,
    createdAt: payload.created_at,
  }
}

function mapComment(payload: CommentApiResponse): Comment {
  return {
    id: payload.id,
    userId: payload.user_id,
    username: payload.username,
    videoId: payload.video_id,
    content: payload.content,
    likes: payload.likes,
    dislikes: payload.dislikes,
    createdAt: payload.created_at,
  }
}

function mapChannel(payload: ChannelApiResponse): Channel {
  return {
    id: payload.id,
    userId: payload.user_id,
    name: payload.name,
    description: payload.description,
    subscribersCount: payload.subscribers_count,
  }
}

function mapSubscriptionChannel(payload: SubscriptionChannelApiResponse): SubscriptionChannel {
  return {
    id: payload.id,
    userId: payload.user_id,
    name: payload.name,
    description: payload.description,
    subscribersCount: payload.subscribers_count,
    newVideosCount: payload.new_videos_count,
    subscribedAt: payload.subscribed_at,
  }
}

function mapPlaylist(payload: PlaylistApiResponse): Playlist {
  return {
    id: payload.id,
    channelId: payload.channel_id,
    name: payload.name,
    description: payload.description,
    createdAt: payload.created_at,
    items: (payload.items ?? []).map((item) => ({
      videoId: item.video_id,
      videoTitle: item.video_title,
      number: item.number,
    })),
  }
}

export const authApi = {
  login: async (payload: { email: string; password: string }) => {
    const { data } = await apiClient.post<AuthApiResponse>('/login', payload)
    return mapAuthResponse(data)
  },
  register: async (payload: { username: string; email: string; password: string }) => {
    const { data } = await apiClient.post<MessageResponse>('/register', payload)
    return data
  },
  refresh: async () => {
    const { data } = await apiClient.post<AuthApiResponse>('/refresh')
    return mapAuthResponse(data)
  },
  logout: async () => {
    await apiClient.post('/logout')
  },
  getMe: async () => {
    const { data } = await apiClient.get<UserApiResponse>('/me')
    return mapUser(data)
  },
  updateNotifications: async (enabled: boolean) => {
    const { data } = await apiClient.patch<UserApiResponse>('/me/notifications', { enabled })
    return mapUser(data)
  },
}

export const channelApi = {
  getById: async (channelId: number) => {
    const { data } = await apiClient.get<ChannelApiResponse>(`/channels/${channelId}`)
    return mapChannel(data)
  },
  getMine: async () => {
    const { data } = await apiClient.get<ChannelApiResponse>('/channels/me')
    return mapChannel(data)
  },
  create: async (payload: { channel_name: string; description: string }) => {
    const { data } = await apiClient.post<MessageResponse>('/channels', payload)
    return data
  },
  update: async (channelId: number, payload: { channel_name?: string; description?: string }) => {
    const { data } = await apiClient.patch<MessageResponse>(`/channels/${channelId}`, payload)
    return data
  },
  remove: async (channelId: number) => {
    const { data } = await apiClient.delete<MessageResponse>(`/channels/${channelId}`)
    return data
  },
}

export const videoApi = {
  list: async (params?: { limit?: number; offset?: number }) => {
    const { data } = await apiClient.get<VideoApiResponse[]>('/videos', { params })
    return data.map(mapVideo)
  },
  listMine: async (params?: { limit?: number; offset?: number }) => {
    const { data } = await apiClient.get<VideoApiResponse[]>('/videos/me', { params })
    return data.map(mapVideo)
  },
  listByChannel: async (channelId: number, params?: { limit?: number; offset?: number }) => {
    const { data } = await apiClient.get<VideoApiResponse[]>(`/channels/${channelId}/videos`, { params })
    return data.map(mapVideo)
  },
  create: async (payload: { channel_id: number; title: string; description: string; file_key: string }) => {
    const { data } = await apiClient.post<VideoApiResponse>('/videos', payload)
    return mapVideo(data)
  },
  getById: async (videoId: number) => {
    const { data } = await apiClient.get<VideoApiResponse>(`/videos/${videoId}`)
    return mapVideo(data)
  },
  update: async (videoId: number, payload: { title?: string; description?: string }) => {
    const body: { title?: string; description?: string } = {}
    if (payload.title !== undefined) {
      body.title = payload.title
    }
    if (payload.description !== undefined) {
      body.description = payload.description
    }
    const { data } = await apiClient.patch<VideoApiResponse>(`/videos/${videoId}`, body)
    return mapVideo(data)
  },
  remove: async (videoId: number) => {
    const { data } = await apiClient.delete<MessageResponse>(`/videos/${videoId}`)
    return data
  },
  getUploadPresignedUrl: async (payload: { channel_id: number; filename: string }) => {
    const { data } = await apiClient.post<UploadPresignedResponse>('/videos/upload-url', payload)
    return data
  },
  listComments: async (videoId: number, params?: { limit?: number; offset?: number }) => {
    const { data } = await apiClient.get<{ comments: CommentApiResponse[]; total: number }>(`/videos/${videoId}/comments`, { params })
    const response: CommentListResponse = {
      comments: data.comments.map(mapComment),
      total: data.total,
    }
    return response
  },
  createComment: async (videoId: number, payload: { content: string }) => {
    const { data } = await apiClient.post<CommentApiResponse>(`/videos/${videoId}/comments`, payload)
    return mapComment(data)
  },
  updateComment: async (commentId: number, payload: { content: string }) => {
    const { data } = await apiClient.patch<CommentApiResponse>(`/comments/${commentId}`, payload)
    return mapComment(data)
  },
  deleteComment: async (commentId: number) => {
    const { data } = await apiClient.delete<MessageResponse>(`/comments/${commentId}`)
    return data
  },
  getStreamingUrl: async (videoId: number) => {
    const { data } = await apiClient.get<StreamUrlResponse>(`/videos/${videoId}/stream-url`)
    return data
  },
  likeVideo: async (videoId: number) => {
    const { data } = await apiClient.post<MessageResponse>(`/videos/${videoId}/like`)
    return data
  },
  dislikeVideo: async (videoId: number) => {
    const { data } = await apiClient.post<MessageResponse>(`/videos/${videoId}/dislike`)
    return data
  },
  removeVideoRating: async (videoId: number) => {
    const { data } = await apiClient.delete<MessageResponse>(`/videos/${videoId}/rating`)
    return data
  },
  likeComment: async (commentId: number) => {
    const { data } = await apiClient.post<MessageResponse>(`/comments/${commentId}/like`)
    return data
  },
  dislikeComment: async (commentId: number) => {
    const { data } = await apiClient.post<MessageResponse>(`/comments/${commentId}/dislike`)
    return data
  },
  removeCommentRating: async (commentId: number) => {
    const { data } = await apiClient.delete<MessageResponse>(`/comments/${commentId}/rating`)
    return data
  },
}

export const subscriptionApi = {
  listMySubscriptions: async (params?: { limit?: number; offset?: number }) => {
    const { data } = await apiClient.get<SubscriptionChannelApiResponse[]>('/subscriptions', { params })
    return data.map(mapSubscriptionChannel)
  },
  subscribe: async (channelId: number) => {
    const { data } = await apiClient.post<MessageResponse>(`/channels/${channelId}/subscribe`)
    return data
  },
  unsubscribe: async (channelId: number) => {
    const { data } = await apiClient.delete<MessageResponse>(`/channels/${channelId}/subscribe`)
    return data
  },
}

export const playlistApi = {
  listByChannel: async (channelId: number, params?: { limit?: number; offset?: number }) => {
    const { data } = await apiClient.get<PlaylistApiResponse[]>(`/channels/${channelId}/playlists`, { params })
    return data.map(mapPlaylist)
  },
  getById: async (playlistId: number) => {
    const { data } = await apiClient.get<PlaylistApiResponse>(`/playlists/${playlistId}`)
    return mapPlaylist(data)
  },
  create: async (channelId: number, payload: { name: string; description: string }) => {
    const { data } = await apiClient.post<PlaylistApiResponse>(`/channels/${channelId}/playlists`, payload)
    return mapPlaylist(data)
  },
  update: async (playlistId: number, payload: { name?: string; description?: string }) => {
    const { data } = await apiClient.patch<PlaylistApiResponse>(`/playlists/${playlistId}`, payload)
    return mapPlaylist(data)
  },
  remove: async (playlistId: number) => {
    const { data } = await apiClient.delete<MessageResponse>(`/playlists/${playlistId}`)
    return data
  },
  addVideo: async (playlistId: number, videoId: number) => {
    const { data } = await apiClient.post<MessageResponse>(`/playlists/${playlistId}/videos/${videoId}`)
    return data
  },
  removeVideo: async (playlistId: number, videoId: number) => {
    const { data } = await apiClient.delete<MessageResponse>(`/playlists/${playlistId}/videos/${videoId}`)
    return data
  },
}

export const adminApi = {
  banUser: async (userId: number) => {
    const { data } = await apiClient.post<MessageResponse>(`/admin/users/${userId}/ban`)
    return data
  },
  unbanUser: async (userId: number) => {
    const { data } = await apiClient.post<MessageResponse>(`/admin/users/${userId}/unban`)
    return data
  },
}




