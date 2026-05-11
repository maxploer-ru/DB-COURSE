import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { authApi, subscriptionApi } from '../shared/api/endpoints'
import { useAuthStore } from '../features/auth/store'

export function NotificationsPage() {
  const queryClient = useQueryClient()
  const currentUser = useAuthStore((state) => state.user)
  const accessToken = useAuthStore((state) => state.accessToken)
  const setSession = useAuthStore((state) => state.setSession)

  const subscriptionsQuery = useQuery({
    queryKey: ['my-subscriptions'],
    queryFn: () => subscriptionApi.listMySubscriptions({ limit: 200, offset: 0 }),
  })

  const toggleMutation = useMutation({
    mutationFn: (enabled: boolean) => authApi.updateNotifications(enabled),
    onSuccess: (user) => {
      // ✅ Объект создаётся только внутри onSuccess – один раз, а не на каждый рендер
      setSession({
        user,
        tokens: {
          accessToken: accessToken ?? '',
        },
      })
      queryClient.invalidateQueries({ queryKey: ['my-subscriptions'] })
    },
  })

  const unreadTotal = (subscriptionsQuery.data ?? []).reduce((sum, item) => sum + item.newVideosCount, 0)

  return (
    <section className="page">
      <h1>Уведомления</h1>
      <p className="page__lead">Всего новых видео: {unreadTotal}</p>
      <div className="comments__actions">
        <button
          type="button"
          className="app-button app-button--ghost"
          onClick={() => toggleMutation.mutate(!(currentUser?.notificationsEnabled ?? true))}
          disabled={toggleMutation.isPending}
        >
          {currentUser?.notificationsEnabled ? 'Выключить уведомления' : 'Включить уведомления'}
        </button>
        <button
          type="button"
          className="app-button app-button--ghost"
          onClick={() => queryClient.invalidateQueries({ queryKey: ['my-subscriptions'] })}
        >
          Обновить
        </button>
      </div>

      <ul className="subscription-list">
        {(subscriptionsQuery.data ?? []).map((channel) => (
          <li key={channel.id} className="subscription-list__item">
            <h3>{channel.name}</h3>
            <p className="video-card__meta">Новых видео: {channel.newVideosCount}</p>
            <Link className="app-button app-button--ghost" to={`/channels/${channel.id}`}>
              Открыть канал (сбросить счетчик)
            </Link>
          </li>
        ))}
      </ul>
    </section>
  )
}

