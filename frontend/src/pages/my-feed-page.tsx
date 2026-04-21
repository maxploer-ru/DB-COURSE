import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { subscriptionApi } from '../shared/api/endpoints'

export function MyFeedPage() {
  const queryClient = useQueryClient()
  const { data, isLoading, isError } = useQuery({
    queryKey: ['my-subscriptions'],
    queryFn: () => subscriptionApi.listMySubscriptions({ limit: 50, offset: 0 }),
  })

  const unsubscribeMutation = useMutation({
    mutationFn: (id: number) => subscriptionApi.unsubscribe(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['my-subscriptions'] }),
  })

  if (isLoading) {
    return <p>Загружаем подписки...</p>
  }

  if (isError) {
    return <p>Не удалось загрузить подписки.</p>
  }

  return (
    <section className="page">
      <h1>Мои подписки</h1>
      {!data?.length && <p className="page__lead">Подписок пока нет. Подписывайтесь прямо на странице видео или канала.</p>}
      <ul className="subscription-list">
        {data?.map((channel) => (
          <li key={channel.id} className="subscription-list__item">
            <h3>{channel.name}</h3>
            <p className="page__lead">{channel.description || 'Описание отсутствует'}</p>
            <p className="video-card__meta">
              Подписчиков: {channel.subscribersCount} • Новых видео: {channel.newVideosCount}
            </p>
            <div className="comments__actions">
              <Link className="app-button app-button--ghost" to={`/channels/${channel.id}`}>
                Перейти на канал
              </Link>
              <button
                className="app-button app-button--ghost"
                type="button"
                disabled={unsubscribeMutation.isPending}
                onClick={() => unsubscribeMutation.mutate(channel.id)}
              >
                Отписаться
              </button>
            </div>
          </li>
        ))}
      </ul>
    </section>
  )
}



