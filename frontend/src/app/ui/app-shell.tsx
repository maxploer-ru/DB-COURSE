import { Link, NavLink, Outlet } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useAuthStore } from '../../features/auth/store'
import { authApi, subscriptionApi } from '../../shared/api/endpoints'

const navLinkClassName = ({ isActive }: { isActive: boolean }) =>
  isActive ? 'app-nav__link app-nav__link--active' : 'app-nav__link'

export function AppShell() {
  const { isAuthenticated, logout, user } = useAuthStore()
  const subscriptionsQuery = useQuery({
    queryKey: ['my-subscriptions'],
    queryFn: () => subscriptionApi.listMySubscriptions({ limit: 200, offset: 0 }),
    enabled: isAuthenticated,
  })

  const unreadCount = (subscriptionsQuery.data ?? []).reduce((sum, item) => sum + item.newVideosCount, 0)

  const handleLogout = async () => {
    try {
      await authApi.logout()
    } finally {
      logout()
    }
  }

  return (
    <div className="app-shell">
      <header className="app-header">
        <Link to="/" className="app-logo">
          ZVideo
        </Link>
        <nav className="app-nav">
          {isAuthenticated ? (
            <>
              <NavLink to="/videos" className={navLinkClassName}>
                Видео
              </NavLink>
              <NavLink to="/my-feed" className={navLinkClassName}>
                Подписки
              </NavLink>
              <NavLink to="/notifications" className={navLinkClassName}>
                Уведомления{unreadCount > 0 ? ` (${unreadCount})` : ''}
              </NavLink>
              <NavLink to="/my-channel" className={navLinkClassName}>
                Мой канал
              </NavLink>
              <NavLink to="/my-community" className={navLinkClassName}>
                Сообщество
              </NavLink>
              <NavLink to="/studio" className={navLinkClassName}>
                Студия
              </NavLink>
              {user?.role === 'admin' && (
                <NavLink to="/admin" className={navLinkClassName}>
                  Admin
                </NavLink>
              )}
              <button type="button" className="app-button app-button--ghost" onClick={handleLogout}>
                Выйти
              </button>
            </>
          ) : (
            <>
              <NavLink to="/login" className={navLinkClassName}>
                Вход
              </NavLink>
              <NavLink to="/register" className={navLinkClassName}>
                Регистрация
              </NavLink>
            </>
          )}
        </nav>
      </header>
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  )
}
