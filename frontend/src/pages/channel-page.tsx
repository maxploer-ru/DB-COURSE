import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link, useParams } from 'react-router-dom'
import { channelApi, playlistApi, subscriptionApi, videoApi } from '../shared/api/endpoints'

export function ChannelPage() {
  const queryClient = useQueryClient()
  const { channelId } = useParams()
  const parsedChannelId = Number(channelId)

  const channelQuery = useQuery({
    queryKey: ['channel', parsedChannelId],
    queryFn: () => channelApi.getById(parsedChannelId),
    enabled: Number.isFinite(parsedChannelId),
  })

  const videosQuery = useQuery({
    queryKey: ['channel-videos', parsedChannelId],
    queryFn: () => videoApi.listByChannel(parsedChannelId, { limit: 50, offset: 0 }),
    enabled: Number.isFinite(parsedChannelId),
  })

  const playlistsQuery = useQuery({
    queryKey: ['channel-playlists', parsedChannelId],
    queryFn: () => playlistApi.listByChannel(parsedChannelId, { limit: 100, offset: 0 }),
    enabled: Number.isFinite(parsedChannelId),
  })

  const mySubscriptionsQuery = useQuery({
    queryKey: ['my-subscriptions'],
    queryFn: () => subscriptionApi.listMySubscriptions({ limit: 100, offset: 0 }),
  })

  const subscribeMutation = useMutation({
    mutationFn: () => subscriptionApi.subscribe(parsedChannelId),
    onSuccess: () => {
      queryClient.setQueryData(['channel', parsedChannelId], (previous: { subscribersCount: number } | undefined) => {
        if (!previous) {
          return previous
        }
        return { ...previous, subscribersCount: previous.subscribersCount + 1 }
      })
      queryClient.invalidateQueries({ queryKey: ['my-subscriptions'] })
      queryClient.invalidateQueries({ queryKey: ['channel', parsedChannelId] })
    },
  })

  const unsubscribeMutation = useMutation({
    mutationFn: () => subscriptionApi.unsubscribe(parsedChannelId),
    onSuccess: () => {
      queryClient.setQueryData(['channel', parsedChannelId], (previous: { subscribersCount: number } | undefined) => {
        if (!previous) {
          return previous
        }
        return { ...previous, subscribersCount: Math.max(0, previous.subscribersCount - 1) }
      })
      queryClient.invalidateQueries({ queryKey: ['my-subscriptions'] })
      queryClient.invalidateQueries({ queryKey: ['channel', parsedChannelId] })
    },
  })

  const isSubscribed = Boolean(mySubscriptionsQuery.data?.some((channel) => channel.id === parsedChannelId))

  if (!Number.isFinite(parsedChannelId)) {
    return <p>Некорректный id канала.</p>
  }

  if (channelQuery.isLoading) {
    return <p>Загружаем канал...</p>
  }

  if (channelQuery.isError || !channelQuery.data) {
    return <p>Не удалось загрузить канал.</p>
  }

  return (
    <section className="page">
      <h1>{channelQuery.data.name}</h1>
      <p className="page__lead">{channelQuery.data.description || 'Описание отсутствует'}</p>
      <p className="video-card__meta">Подписчиков: {channelQuery.data.subscribersCount}</p>
      <div className="comments__actions">
        {isSubscribed ? (
          <button
            className="app-button app-button--ghost"
            type="button"
            onClick={() => unsubscribeMutation.mutate()}
            disabled={unsubscribeMutation.isPending}
          >
            Отписаться
          </button>
        ) : (
          <button className="app-button app-button--ghost" type="button" onClick={() => subscribeMutation.mutate()} disabled={subscribeMutation.isPending}>
            Подписаться
          </button>
        )}
      </div>

      <h2>Видео канала</h2>
      {videosQuery.isLoading && <p>Загружаем видео...</p>}
      {videosQuery.isError && <p>Не удалось загрузить видео канала.</p>}
      <ul className="subscription-list">
        {videosQuery.data?.map((video) => (
          <li key={video.id} className="subscription-list__item">
            <h3>{video.title}</h3>
            <p className="video-card__meta">{video.views} просмотров • {video.likes} лайков • {video.dislikes} дизлайков</p>
            <Link className="app-button app-button--ghost" to={`/videos/${video.id}`}>
              Открыть видео
            </Link>
          </li>
        ))}
      </ul>

      <h2>Плейлисты канала</h2>
      {playlistsQuery.isLoading && <p>Загружаем плейлисты...</p>}
      {playlistsQuery.isError && <p>Не удалось загрузить плейлисты.</p>}
      <ul className="subscription-list">
        {playlistsQuery.data?.map((playlist) => (
          <li key={playlist.id} className="subscription-list__item">
            <h3>{playlist.name}</h3>
            <p className="video-card__meta">Видео в плейлисте: {playlist.items.length}</p>
            <p className="page__lead">{playlist.description || 'Описание отсутствует'}</p>
            <Link className="app-button app-button--ghost" to={`/playlists/${playlist.id}`}>
              Открыть плейлист
            </Link>
          </li>
        ))}
      </ul>
    </section>
  )
}


