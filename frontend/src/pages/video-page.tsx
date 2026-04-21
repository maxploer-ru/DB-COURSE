import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useParams } from 'react-router-dom'
import { channelApi, subscriptionApi, videoApi } from '../shared/api/endpoints'
import { useAuthStore } from '../features/auth/store'

export function VideoPage() {
  const queryClient = useQueryClient()
  const { videoId } = useParams()
  const parsedVideoId = Number(videoId)
  const [commentText, setCommentText] = useState('')
  const [editingCommentId, setEditingCommentId] = useState<number | null>(null)
  const [editingValue, setEditingValue] = useState('')
  const [statusMessage, setStatusMessage] = useState('')
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const currentUserId = useAuthStore((state) => state.user?.id)

  const videoQuery = useQuery({
    queryKey: ['video', parsedVideoId],
    queryFn: () => videoApi.getById(parsedVideoId),
    enabled: Number.isFinite(parsedVideoId),
  })

  const commentsQuery = useQuery({
    queryKey: ['video-comments', parsedVideoId],
    queryFn: () => videoApi.listComments(parsedVideoId, { limit: 50, offset: 0 }),
    enabled: Number.isFinite(parsedVideoId),
  })

  const channelQuery = useQuery({
    queryKey: ['channel', videoQuery.data?.channelId],
    queryFn: () => channelApi.getById(videoQuery.data!.channelId),
    enabled: Boolean(videoQuery.data?.channelId),
  })

  const mySubscriptionsQuery = useQuery({
    queryKey: ['my-subscriptions'],
    queryFn: () => subscriptionApi.listMySubscriptions({ limit: 100, offset: 0 }),
    enabled: isAuthenticated,
  })

  const streamQuery = useQuery({
    queryKey: ['video-stream-url', parsedVideoId],
    queryFn: () => videoApi.getStreamingUrl(parsedVideoId),
    enabled: Number.isFinite(parsedVideoId),
    retry: 1,
  })

  const commentMutation = useMutation({
    mutationFn: (content: string) => videoApi.createComment(parsedVideoId, { content }),
    onSuccess: () => {
      setCommentText('')
      queryClient.invalidateQueries({ queryKey: ['video-comments', parsedVideoId] })
    },
  })

  const updateCommentMutation = useMutation({
    mutationFn: (payload: { id: number; content: string }) => videoApi.updateComment(payload.id, { content: payload.content }),
    onSuccess: () => {
      setEditingCommentId(null)
      setEditingValue('')
      queryClient.invalidateQueries({ queryKey: ['video-comments', parsedVideoId] })
    },
  })

  const deleteCommentMutation = useMutation({
    mutationFn: (commentId: number) => videoApi.deleteComment(commentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['video-comments', parsedVideoId] })
    },
  })

  const videoLikeMutation = useMutation({
    mutationFn: () => videoApi.likeVideo(parsedVideoId),
    onSuccess: (data) => {
      setStatusMessage(data.message)
      queryClient.invalidateQueries({ queryKey: ['video', parsedVideoId] })
    },
  })

  const videoDislikeMutation = useMutation({
    mutationFn: () => videoApi.dislikeVideo(parsedVideoId),
    onSuccess: (data) => {
      setStatusMessage(data.message)
      queryClient.invalidateQueries({ queryKey: ['video', parsedVideoId] })
    },
  })

  const videoRemoveRatingMutation = useMutation({
    mutationFn: () => videoApi.removeVideoRating(parsedVideoId),
    onSuccess: (data) => {
      setStatusMessage(data.message)
      queryClient.invalidateQueries({ queryKey: ['video', parsedVideoId] })
    },
  })

  const subscribeMutation = useMutation({
    mutationFn: (channelId: number) => subscriptionApi.subscribe(channelId),
    onSuccess: (data) => {
      setStatusMessage(data.message)
      queryClient.setQueryData(['channel', videoQuery.data?.channelId], (previous: { subscribersCount: number } | undefined) => {
        if (!previous) {
          return previous
        }
        return { ...previous, subscribersCount: previous.subscribersCount + 1 }
      })
      queryClient.invalidateQueries({ queryKey: ['my-subscriptions'] })
    },
  })

  const unsubscribeMutation = useMutation({
    mutationFn: (channelId: number) => subscriptionApi.unsubscribe(channelId),
    onSuccess: (data) => {
      setStatusMessage(data.message)
      queryClient.setQueryData(['channel', videoQuery.data?.channelId], (previous: { subscribersCount: number } | undefined) => {
        if (!previous) {
          return previous
        }
        return { ...previous, subscribersCount: Math.max(0, previous.subscribersCount - 1) }
      })
      queryClient.invalidateQueries({ queryKey: ['my-subscriptions'] })
    },
  })

  const commentLikeMutation = useMutation({
    mutationFn: (commentId: number) => videoApi.likeComment(commentId),
    onSuccess: (data) => {
      setStatusMessage(data.message)
      queryClient.invalidateQueries({ queryKey: ['video-comments', parsedVideoId] })
    },
  })

  const commentDislikeMutation = useMutation({
    mutationFn: (commentId: number) => videoApi.dislikeComment(commentId),
    onSuccess: (data) => {
      setStatusMessage(data.message)
      queryClient.invalidateQueries({ queryKey: ['video-comments', parsedVideoId] })
    },
  })

  const commentRemoveRatingMutation = useMutation({
    mutationFn: (commentId: number) => videoApi.removeCommentRating(commentId),
    onSuccess: (data) => {
      setStatusMessage(data.message)
      queryClient.invalidateQueries({ queryKey: ['video-comments', parsedVideoId] })
    },
  })

  if (videoQuery.isLoading) {
    return <p>Загружаем видео...</p>
  }

  if (videoQuery.isError || !videoQuery.data) {
    return <p>Видео не найдено.</p>
  }

  const handleSubmitComment = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!commentText.trim()) {
      return
    }
    commentMutation.mutate(commentText)
  }

  const isSubscribed = Boolean(mySubscriptionsQuery.data?.some((channel) => channel.id === videoQuery.data.channelId))

  return (
    <section className="page">
      <h1>{videoQuery.data.title}</h1>
      <p className="page__lead">{videoQuery.data.description}</p>
      <p className="video-card__meta">
        {channelQuery.data ? channelQuery.data.name : `Канал #${videoQuery.data.channelId}`} • {videoQuery.data.views} просмотров • {videoQuery.data.likes} лайков • {videoQuery.data.dislikes} дизлайков • {videoQuery.data.comments} комментариев
      </p>
      <div className="comments__actions">
        <Link className="app-button app-button--ghost" to={`/channels/${videoQuery.data.channelId}`}>
          Перейти на канал
        </Link>
        {isSubscribed ? (
          <button
            className="app-button app-button--ghost"
            type="button"
            onClick={() => unsubscribeMutation.mutate(videoQuery.data.channelId)}
            disabled={unsubscribeMutation.isPending}
          >
            Отписаться
          </button>
        ) : (
          <button
            className="app-button app-button--ghost"
            type="button"
            onClick={() => subscribeMutation.mutate(videoQuery.data.channelId)}
            disabled={subscribeMutation.isPending}
          >
            Подписаться
          </button>
        )}
      </div>

      <section className="video-player">
        {streamQuery.isLoading && <p>Готовим поток для воспроизведения...</p>}
        {streamQuery.isError && <p>Не удалось получить поток для воспроизведения.</p>}
        {streamQuery.data?.url && <video key={streamQuery.data.url} controls className="video-player__element" src={streamQuery.data.url} />}
      </section>

      <div className="comments__actions">
        <button className="app-button app-button--ghost" type="button" onClick={() => videoLikeMutation.mutate()}>
          Лайк
        </button>
        <button className="app-button app-button--ghost" type="button" onClick={() => videoDislikeMutation.mutate()}>
          Дизлайк
        </button>
        <button className="app-button app-button--ghost" type="button" onClick={() => videoRemoveRatingMutation.mutate()}>
          Убрать реакцию
        </button>
      </div>
      {statusMessage && <p className="video-card__meta">{statusMessage}</p>}

      <section className="comments">
        <h2>Комментарии</h2>
        {!isAuthenticated && <p>Войдите, чтобы оставить комментарий.</p>}
        {isAuthenticated && (
          <form className="form comments__form" onSubmit={handleSubmitComment}>
            <textarea
              className="form__input form__textarea"
              value={commentText}
              placeholder="Напишите комментарий"
              onChange={(event) => setCommentText(event.target.value)}
              rows={3}
            />
            <button type="submit" className="app-button" disabled={commentMutation.isPending}>
              {commentMutation.isPending ? 'Публикуем...' : 'Опубликовать'}
            </button>
          </form>
        )}

        {commentsQuery.isLoading && <p>Загружаем комментарии...</p>}
        {commentsQuery.isError && <p>Не удалось загрузить комментарии.</p>}
        <ul className="comments__list">
          {commentsQuery.data?.comments.map((comment) => (
            <li key={comment.id} className="comments__item">
              {editingCommentId === comment.id ? (
                <form
                  className="form comments__inline-form"
                  onSubmit={(event) => {
                    event.preventDefault()
                    if (!editingValue.trim()) {
                      return
                    }
                    updateCommentMutation.mutate({ id: comment.id, content: editingValue })
                  }}
                >
                  <textarea className="form__input form__textarea" rows={2} value={editingValue} onChange={(event) => setEditingValue(event.target.value)} />
                  <div className="comments__actions">
                    <button type="submit" className="app-button app-button--ghost" disabled={updateCommentMutation.isPending}>
                      Сохранить
                    </button>
                    <button
                      type="button"
                      className="app-button app-button--ghost"
                      onClick={() => {
                        setEditingCommentId(null)
                        setEditingValue('')
                      }}
                    >
                      Отмена
                    </button>
                  </div>
                </form>
              ) : (
                <p>{comment.content}</p>
              )}
              <small>
                Автор #{comment.userId} • {new Date(comment.createdAt).toLocaleString()} • {comment.likes}/{comment.dislikes}
              </small>
              {isAuthenticated && currentUserId === comment.userId && editingCommentId !== comment.id && (
                <div className="comments__actions">
                  <button
                    type="button"
                    className="app-button app-button--ghost"
                    onClick={() => {
                      setEditingCommentId(comment.id)
                      setEditingValue(comment.content)
                    }}
                  >
                    Редактировать
                  </button>
                  <button
                    type="button"
                    className="app-button app-button--ghost"
                    onClick={() => deleteCommentMutation.mutate(comment.id)}
                    disabled={deleteCommentMutation.isPending}
                  >
                    Удалить
                  </button>
                </div>
              )}
              {isAuthenticated && (
                <div className="comments__actions">
                  <button type="button" className="app-button app-button--ghost" onClick={() => commentLikeMutation.mutate(comment.id)}>
                    Лайк
                  </button>
                  <button type="button" className="app-button app-button--ghost" onClick={() => commentDislikeMutation.mutate(comment.id)}>
                    Дизлайк
                  </button>
                  <button
                    type="button"
                    className="app-button app-button--ghost"
                    onClick={() => commentRemoveRatingMutation.mutate(comment.id)}
                  >
                    Убрать реакцию
                  </button>
                </div>
              )}
            </li>
          ))}
        </ul>
      </section>
    </section>
  )
}







