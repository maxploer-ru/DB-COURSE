import { useQuery } from '@tanstack/react-query'
import { Link, useParams } from 'react-router-dom'
import { playlistApi } from '../shared/api/endpoints'

export function PlaylistPage() {
  const { playlistId } = useParams()
  const parsedPlaylistId = Number(playlistId)

  const playlistQuery = useQuery({
    queryKey: ['playlist', parsedPlaylistId],
    queryFn: () => playlistApi.getById(parsedPlaylistId),
    enabled: Number.isFinite(parsedPlaylistId),
  })

  if (!Number.isFinite(parsedPlaylistId)) {
    return <p>Некорректный id плейлиста.</p>
  }

  if (playlistQuery.isLoading) {
    return <p>Загружаем плейлист...</p>
  }

  if (playlistQuery.isError || !playlistQuery.data) {
    return <p>Не удалось загрузить плейлист.</p>
  }

  const items = [...(playlistQuery.data.items ?? [])].sort((a, b) => a.number - b.number)

  return (
    <section className="page">
      <h1>{playlistQuery.data.name}</h1>
      <p className="page__lead">{playlistQuery.data.description || 'Описание отсутствует'}</p>
      <p className="video-card__meta">Видео в плейлисте: {items.length}</p>

      {!items.length && <p className="page__lead">Плейлист пока пуст.</p>}
      <ul className="subscription-list">
        {items.map((item) => (
          <li key={item.videoId} className="subscription-list__item">
            <h3>#{item.number} • Видео {item.videoId}</h3>
            <Link className="app-button app-button--ghost" to={`/videos/${item.videoId}`}>
              Открыть видео
            </Link>
          </li>
        ))}
      </ul>
    </section>
  )
}

