import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { videoApi } from '../shared/api/endpoints'
import { VideoCard } from '../entities/video/ui/video-card'

export function VideosPage() {
  const [page, setPage] = useState(0)
  const limit = 12

  const { data, isLoading, isError } = useQuery({
    queryKey: ['videos', page],
    queryFn: () => videoApi.list({ limit, offset: page * limit }),
  })

  if (isLoading) {
    return <p>Загружаем список видео...</p>
  }

  if (isError) {
    return <p>Не удалось загрузить видео. Проверьте API и повторите попытку.</p>
  }

  return (
    <section className="page">
      <h1>Каталог видео</h1>
      {!data?.length && <p>Пока нет опубликованных видео.</p>}
      <div className="video-grid">
        {data?.map((video) => (
          <VideoCard key={video.id} video={video} />
        ))}
      </div>
      <div className="page__actions">
        <button className="app-button app-button--ghost" type="button" onClick={() => setPage((p) => Math.max(0, p - 1))} disabled={page === 0}>
          Назад
        </button>
        <button className="app-button app-button--ghost" type="button" onClick={() => setPage((p) => p + 1)} disabled={!data?.length || data.length < limit}>
          Далее
        </button>
      </div>
    </section>
  )
}


