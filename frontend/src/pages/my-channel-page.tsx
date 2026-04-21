import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import type { AxiosError } from 'axios'
import { Link } from 'react-router-dom'
import { channelApi } from '../shared/api/endpoints'
import type { ApiError } from '../shared/api/types'

function isNotFound(error: unknown): boolean {
  const axiosError = error as AxiosError<ApiError>
  return axiosError?.response?.status === 404
}

export function MyChannelPage() {
  const queryClient = useQueryClient()
  const [createName, setCreateName] = useState('')
  const [createDescription, setCreateDescription] = useState('')
  const [updateName, setUpdateName] = useState('')
  const [updateDescription, setUpdateDescription] = useState('')
  const [message, setMessage] = useState('')

  const myChannelQuery = useQuery({
    queryKey: ['my-channel'],
    queryFn: channelApi.getMine,
    retry: false,
  })

  const createMutation = useMutation({
    mutationFn: () => channelApi.create({ channel_name: createName, description: createDescription }),
    onSuccess: (data) => {
      setCreateName('')
      setCreateDescription('')
      setMessage(data.message)
      queryClient.invalidateQueries({ queryKey: ['my-channel'] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: () =>
      channelApi.update(myChannelQuery.data!.id, {
        channel_name: updateName || undefined,
        description: updateDescription || undefined,
      }),
    onSuccess: (data) => {
      setUpdateName('')
      setUpdateDescription('')
      setMessage(data.message)
      queryClient.invalidateQueries({ queryKey: ['my-channel'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: () => channelApi.remove(myChannelQuery.data!.id),
    onSuccess: (data) => {
      setMessage(data.message)
      queryClient.invalidateQueries({ queryKey: ['my-channel'] })
      queryClient.invalidateQueries({ queryKey: ['videos'] })
    },
  })

  if (myChannelQuery.isLoading) {
    return <p>Загружаем ваш канал...</p>
  }

  if (myChannelQuery.isError && !isNotFound(myChannelQuery.error)) {
    return <p>Не удалось загрузить канал. Попробуйте обновить страницу.</p>
  }

  if (myChannelQuery.isError && isNotFound(myChannelQuery.error)) {
    return (
      <section className="page page--narrow">
        <h1>Создать канал</h1>
        <p className="page__lead">У вас еще нет канала. Создайте его и начинайте публиковать видео.</p>
        <form
          className="form"
          onSubmit={(event) => {
            event.preventDefault()
            createMutation.mutate()
          }}
        >
          <label className="form__label">
            Название
            <input className="form__input" value={createName} onChange={(event) => setCreateName(event.target.value)} required />
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
          <button className="app-button" type="submit" disabled={createMutation.isPending}>
            {createMutation.isPending ? 'Создаем...' : 'Создать канал'}
          </button>
        </form>
        {message && <p className="video-card__meta">{message}</p>}
      </section>
    )
  }

  const channel = myChannelQuery.data
  if (!channel) {
    return null
  }

  return (
    <section className="page page--narrow">
      <h1>Мой канал</h1>
      <h2>{channel.name}</h2>
      <p className="page__lead">{channel.description}</p>
      <p className="video-card__meta">Подписчиков: {channel.subscribersCount}</p>
      <div className="page__actions">
        <Link to={`/channels/${channel.id}`} className="app-button app-button--ghost">
          Открыть публичную страницу
        </Link>
      </div>

      <form
        className="form"
        onSubmit={(event) => {
          event.preventDefault()
          updateMutation.mutate()
        }}
      >
        <label className="form__label">
          Новое название
          <input className="form__input" value={updateName} onChange={(event) => setUpdateName(event.target.value)} placeholder={channel.name} />
        </label>
        <label className="form__label">
          Новое описание
          <textarea
            className="form__input form__textarea"
            rows={3}
            value={updateDescription}
            onChange={(event) => setUpdateDescription(event.target.value)}
            placeholder={channel.description}
          />
        </label>
        <button className="app-button app-button--ghost" type="submit" disabled={updateMutation.isPending}>
          Сохранить изменения
        </button>
      </form>

      <button className="app-button app-button--ghost" type="button" onClick={() => deleteMutation.mutate()} disabled={deleteMutation.isPending}>
        Удалить канал
      </button>
      {message && <p className="video-card__meta">{message}</p>}
    </section>
  )
}

