import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { channelApi, playlistApi, videoApi } from '../shared/api/endpoints'

function resolveVideoMimeType(file: File): string {
  if (file.type) {
    return file.type
  }

  if (file.name.toLowerCase().endsWith('.mp4')) {
    return 'video/mp4'
  }

  return 'application/octet-stream'
}

export function StudioPage() {
  const queryClient = useQueryClient()
  const [message, setMessage] = useState('')
  const [createTitle, setCreateTitle] = useState('')
  const [createDescription, setCreateDescription] = useState('')
  const [createFileKey, setCreateFileKey] = useState('')
  const [uploadFile, setUploadFile] = useState<File | null>(null)
  const [editingVideoId, setEditingVideoId] = useState<number | null>(null)
  const [editingTitle, setEditingTitle] = useState('')
  const [editingDescription, setEditingDescription] = useState('')
  const [playlistName, setPlaylistName] = useState('')
  const [playlistDescription, setPlaylistDescription] = useState('')
  const [selectedPlaylistId, setSelectedPlaylistId] = useState<number | ''>('')
  const [selectedVideoId, setSelectedVideoId] = useState<number | ''>('')

  const myChannelQuery = useQuery({
    queryKey: ['my-channel'],
    queryFn: channelApi.getMine,
    retry: false,
  })

  const myVideosQuery = useQuery({
    queryKey: ['studio-my-videos'],
    queryFn: () => videoApi.listMine({ limit: 50, offset: 0 }),
    enabled: Boolean(myChannelQuery.data),
  })

  const currentChannelId = myChannelQuery.data?.id

  const playlistsQuery = useQuery({
    queryKey: ['channel-playlists', currentChannelId],
    queryFn: () => playlistApi.listByChannel(currentChannelId!, { limit: 100, offset: 0 }),
    enabled: Boolean(currentChannelId),
  })

  const uploadMutation = useMutation({
    mutationFn: async () => {
      if (!uploadFile || !currentChannelId) {
        throw new Error('Нужен канал и файл')
      }

      const uploadData = await videoApi.getUploadPresignedUrl({
        channel_id: currentChannelId,
        filename: uploadFile.name,
      })

      await fetch(uploadData.url, {
        method: 'PUT',
        body: uploadFile,
        headers: {
          'Content-Type': resolveVideoMimeType(uploadFile),
        },
      })

      return uploadData
    },
    onSuccess: (data) => {
      setCreateFileKey(data.file_key)
      setMessage('Файл загружен, можно публиковать видео')
    },
  })

  const createVideoMutation = useMutation({
    mutationFn: () => {
      if (!createFileKey) {
        throw new Error('Сначала загрузите файл')
      }
      return videoApi.create({
        channel_id: currentChannelId!,
        title: createTitle,
        description: createDescription,
        file_key: createFileKey,
      })
    },
    onSuccess: (video) => {
      setCreateTitle('')
      setCreateDescription('')
      setCreateFileKey('')
      setUploadFile(null)
      setMessage(`Видео «${video.title}» опубликовано`)
      queryClient.invalidateQueries({ queryKey: ['studio-my-videos'] })
      queryClient.invalidateQueries({ queryKey: ['videos'] })
      queryClient.invalidateQueries({ queryKey: ['channel-videos', currentChannelId] })
    },
  })

  const updateVideoMutation = useMutation({
    mutationFn: (videoId: number) =>
      videoApi.update(videoId, {
        title: editingTitle || undefined,
        description: editingDescription || undefined,
      }),
    onSuccess: (video) => {
      setMessage(`Видео «${video.title}» обновлено`)
      setEditingVideoId(null)
      setEditingTitle('')
      setEditingDescription('')
      queryClient.invalidateQueries({ queryKey: ['studio-my-videos'] })
      queryClient.invalidateQueries({ queryKey: ['video', video.id] })
    },
  })

  const deleteVideoMutation = useMutation({
    mutationFn: (videoId: number) => videoApi.remove(videoId),
    onSuccess: (data) => {
      setMessage(data.message)
      queryClient.invalidateQueries({ queryKey: ['studio-my-videos'] })
      queryClient.invalidateQueries({ queryKey: ['videos'] })
    },
  })

  const createPlaylistMutation = useMutation({
    mutationFn: () => playlistApi.create(currentChannelId!, { name: playlistName, description: playlistDescription }),
    onSuccess: (playlist) => {
      setPlaylistName('')
      setPlaylistDescription('')
      setMessage(`Плейлист «${playlist.name}» создан`)
      queryClient.invalidateQueries({ queryKey: ['channel-playlists', currentChannelId] })
    },
  })

  const deletePlaylistMutation = useMutation({
    mutationFn: (playlistId: number) => playlistApi.remove(playlistId),
    onSuccess: () => {
      setMessage('Плейлист удален')
      queryClient.invalidateQueries({ queryKey: ['channel-playlists', currentChannelId] })
    },
  })

  const addToPlaylistMutation = useMutation({
    mutationFn: (payload: { playlistId: number; videoId: number }) => playlistApi.addVideo(payload.playlistId, payload.videoId),
    onSuccess: () => {
      setMessage('Видео добавлено в плейлист')
      queryClient.invalidateQueries({ queryKey: ['channel-playlists', currentChannelId] })
    },
  })

  const removeFromPlaylistMutation = useMutation({
    mutationFn: (payload: { playlistId: number; videoId: number }) => playlistApi.removeVideo(payload.playlistId, payload.videoId),
    onSuccess: () => {
      setMessage('Видео удалено из плейлиста')
      queryClient.invalidateQueries({ queryKey: ['channel-playlists', currentChannelId] })
    },
  })

  const sortedVideos = useMemo(() => {
    return [...(myVideosQuery.data ?? [])].sort((a, b) => b.id - a.id)
  }, [myVideosQuery.data])

  if (myChannelQuery.isLoading) {
    return <p>Загружаем студию...</p>
  }

  if (myChannelQuery.isError) {
    return (
      <section className="page page--narrow">
        <h1>Студия</h1>
        <p className="page__lead">Чтобы публиковать видео, сначала создайте канал.</p>
        <Link to="/my-channel" className="app-button">
          Создать канал
        </Link>
      </section>
    )
  }

  return (
    <section className="page">
      <h1>Студия</h1>
      <p className="page__lead">Канал: {myChannelQuery.data?.name}</p>
      {message && <p className="video-card__meta">{message}</p>}

      <h2>1) Загрузите файл</h2>
      <form
        className="form page--narrow"
        onSubmit={(event) => {
          event.preventDefault()
          uploadMutation.mutate()
        }}
      >
        <label className="form__label">
          Видеофайл
          <input type="file" className="form__input" onChange={(event) => setUploadFile(event.target.files?.[0] ?? null)} required />
        </label>
        <button className="app-button app-button--ghost" type="submit" disabled={uploadMutation.isPending}>
          {uploadMutation.isPending ? 'Загрузка...' : 'Загрузить в хранилище'}
        </button>
      </form>

      <h2>2) Опубликуйте видео</h2>
      <form
        className="form page--narrow"
        onSubmit={(event) => {
          event.preventDefault()
          if (!createFileKey) {
            setMessage('Сначала загрузите файл в хранилище')
            return
          }
          createVideoMutation.mutate()
        }}
      >
        <label className="form__label">
          Название
          <input className="form__input" value={createTitle} onChange={(event) => setCreateTitle(event.target.value)} required />
        </label>
        <label className="form__label">
          Описание
          <textarea
            className="form__input form__textarea"
            rows={3}
            value={createDescription}
            onChange={(event) => setCreateDescription(event.target.value)}
            required
          />
        </label>
        <p className="video-card__meta">
          Ключ файла формируется автоматически после шага загрузки.
        </p>
        <button className="app-button" type="submit" disabled={createVideoMutation.isPending || uploadMutation.isPending || !createFileKey}>
          {createVideoMutation.isPending ? 'Публикуем...' : 'Опубликовать видео'}
        </button>
      </form>

      <h2>Мои видео</h2>
      {myVideosQuery.isLoading && <p>Загружаем видео...</p>}
      {!sortedVideos.length && <p className="page__lead">У вас пока нет видео.</p>}
      <ul className="subscription-list">
        {sortedVideos.map((video) => (
          <li key={video.id} className="subscription-list__item">
            <h3>{video.title}</h3>
            <p className="video-card__meta">{video.views} просмотров • {video.likes} лайков • {video.dislikes} дизлайков</p>
            <div className="comments__actions">
              <Link className="app-button app-button--ghost" to={`/videos/${video.id}`}>
                Открыть
              </Link>
              <button
                className="app-button app-button--ghost"
                type="button"
                onClick={() => {
                  setEditingVideoId(video.id)
                  setEditingTitle(video.title)
                  setEditingDescription(video.description)
                }}
              >
                Редактировать
              </button>
              <button className="app-button app-button--ghost" type="button" onClick={() => deleteVideoMutation.mutate(video.id)}>
                Удалить
              </button>
            </div>

            {editingVideoId === video.id && (
              <form
                className="form"
                onSubmit={(event) => {
                  event.preventDefault()
                  updateVideoMutation.mutate(video.id)
                }}
              >
                <label className="form__label">
                  Название
                  <input className="form__input" value={editingTitle} onChange={(event) => setEditingTitle(event.target.value)} />
                </label>
                <label className="form__label">
                  Описание
                  <textarea className="form__input form__textarea" rows={3} value={editingDescription} onChange={(event) => setEditingDescription(event.target.value)} />
                </label>
                <div className="comments__actions">
                  <button className="app-button app-button--ghost" type="submit" disabled={updateVideoMutation.isPending}>
                    Сохранить
                  </button>
                  <button
                    className="app-button app-button--ghost"
                    type="button"
                    onClick={() => {
                      setEditingVideoId(null)
                      setEditingTitle('')
                      setEditingDescription('')
                    }}
                  >
                    Отмена
                  </button>
                </div>
              </form>
            )}
          </li>
        ))}
      </ul>

      <h2>Плейлисты</h2>
      <form
        className="form page--narrow"
        onSubmit={(event) => {
          event.preventDefault()
          createPlaylistMutation.mutate()
        }}
      >
        <label className="form__label">
          Название плейлиста
          <input className="form__input" value={playlistName} onChange={(event) => setPlaylistName(event.target.value)} required />
        </label>
        <label className="form__label">
          Описание плейлиста
          <textarea className="form__input form__textarea" rows={2} value={playlistDescription} onChange={(event) => setPlaylistDescription(event.target.value)} />
        </label>
        <button className="app-button app-button--ghost" type="submit" disabled={createPlaylistMutation.isPending}>
          Создать плейлист
        </button>
      </form>

      <form
        className="form page--narrow"
        onSubmit={(event) => {
          event.preventDefault()
          if (!selectedPlaylistId || !selectedVideoId) {
            return
          }
          addToPlaylistMutation.mutate({ playlistId: selectedPlaylistId, videoId: selectedVideoId })
        }}
      >
        <label className="form__label">
          Плейлист
          <select className="form__input" value={selectedPlaylistId} onChange={(event) => setSelectedPlaylistId(event.target.value ? Number(event.target.value) : '')}>
            <option value="">Выберите плейлист</option>
            {(playlistsQuery.data ?? []).map((playlist) => (
              <option key={playlist.id} value={playlist.id}>
                {playlist.name}
              </option>
            ))}
          </select>
        </label>
        <label className="form__label">
          Видео
          <select className="form__input" value={selectedVideoId} onChange={(event) => setSelectedVideoId(event.target.value ? Number(event.target.value) : '')}>
            <option value="">Выберите видео</option>
            {sortedVideos.map((video) => (
              <option key={video.id} value={video.id}>
                {video.title}
              </option>
            ))}
          </select>
        </label>
        <button className="app-button app-button--ghost" type="submit" disabled={addToPlaylistMutation.isPending}>
          Добавить в плейлист
        </button>
      </form>

      <ul className="subscription-list">
        {(playlistsQuery.data ?? []).map((playlist) => (
          <li key={playlist.id} className="subscription-list__item">
            <h3>{playlist.name}</h3>
            <p className="page__lead">{playlist.description || 'Описание отсутствует'}</p>
            <p className="video-card__meta">Видео: {playlist.items.length}</p>
            <div className="comments__actions">
              <button className="app-button app-button--ghost" type="button" onClick={() => deletePlaylistMutation.mutate(playlist.id)}>
                Удалить плейлист
              </button>
            </div>
            <ul className="subscription-list">
              {playlist.items.map((item) => (
                <li key={item.videoId} className="subscription-list__item">
                  <span>{item.videoTitle || 'Видео'}</span>
                  <button
                    className="app-button app-button--ghost"
                    type="button"
                    onClick={() => removeFromPlaylistMutation.mutate({ playlistId: playlist.id, videoId: item.videoId })}
                  >
                    Убрать
                  </button>
                </li>
              ))}
            </ul>
          </li>
        ))}
      </ul>
    </section>
  )
}





