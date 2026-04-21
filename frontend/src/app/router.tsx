import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AdminPage } from '../pages/admin-page'
import { AppShell } from './ui/app-shell'
import { ChannelPage } from '../pages/channel-page'
import { ProtectedRoute } from '../features/auth/protected-route'
import { HomePage } from '../pages/home-page'
import { LoginPage } from '../pages/login-page'
import { MyChannelPage } from '../pages/my-channel-page'
import { MyFeedPage } from '../pages/my-feed-page'
import { NotFoundPage } from '../pages/not-found-page'
import { NotificationsPage } from '../pages/notifications-page'
import { PlaylistPage } from '../pages/playlist-page'
import { RegisterPage } from '../pages/register-page'
import { StudioPage } from '../pages/studio-page'
import { VideoPage } from '../pages/video-page'
import { VideosPage } from '../pages/videos-page'

export function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AppShell />}>
          <Route index element={<HomePage />} />
          <Route path="login" element={<LoginPage />} />
          <Route path="register" element={<RegisterPage />} />
          <Route element={<ProtectedRoute />}>
            <Route path="videos" element={<VideosPage />} />
            <Route path="videos/:videoId" element={<VideoPage />} />
            <Route path="channels/:channelId" element={<ChannelPage />} />
            <Route path="playlists/:playlistId" element={<PlaylistPage />} />
            <Route path="my-channel" element={<MyChannelPage />} />
            <Route path="my-feed" element={<MyFeedPage />} />
            <Route path="notifications" element={<NotificationsPage />} />
            <Route path="studio" element={<StudioPage />} />
            <Route path="admin" element={<AdminPage />} />
          </Route>
          <Route path="404" element={<NotFoundPage />} />
          <Route path="*" element={<Navigate to="/404" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}




