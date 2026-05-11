import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { channelApi } from '../shared/api/endpoints'
import { useAuthStore } from '../features/auth/store'

export function CommunityPage() {
  const queryClient = useQueryClient()
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const currentUserId = useAuthStore((state) => state.user?.id)
  const { channelId } = useParams()
  const parsedChannelId = Number(channelId)
  const [commentDrafts, setCommentDrafts] = useState<Record<number, string>>({})
  const [editingPostId, setEditingPostId] = useState<number | null>(null)
  const [editingPostValue, setEditingPostValue] = useState('')
  const [editingCommentId, setEditingCommentId] = useState<number | null>(null)
  const [editingCommentValue, setEditingCommentValue] = useState('')

  const communityQuery = useQuery({
    queryKey: ['channel-community', parsedChannelId],
    queryFn: () => channelApi.getCommunity(parsedChannelId),
    enabled: Number.isFinite(parsedChannelId),
  })

  const createCommentMutation = useMutation({
    mutationFn: (payload: { postId: number; content: string }) => channelApi.createCommunityComment(payload.postId, { content: payload.content }),
    onSuccess: (_, variables) => {
      setCommentDrafts((previous) => ({ ...previous, [variables.postId]: '' }))
      queryClient.invalidateQueries({ queryKey: ['channel-community', parsedChannelId] })
    },
  })

  const updatePostMutation = useMutation({
    mutationFn: (payload: { postId: number; content: string }) => channelApi.updateCommunityPost(payload.postId, { content: payload.content }),
    onSuccess: () => {
      setEditingPostId(null)
      setEditingPostValue('')
      queryClient.invalidateQueries({ queryKey: ['channel-community', parsedChannelId] })
    },
  })

  const updateCommentMutation = useMutation({
    mutationFn: (payload: { commentId: number; content: string }) => channelApi.updateCommunityComment(payload.commentId, { content: payload.content }),
    onSuccess: () => {
      setEditingCommentId(null)
      setEditingCommentValue('')
      queryClient.invalidateQueries({ queryKey: ['channel-community', parsedChannelId] })
    },
  })

  const deleteCommentMutation = useMutation({
    mutationFn: (commentId: number) => channelApi.deleteCommunityComment(commentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['channel-community', parsedChannelId] })
    },
  })

  if (!Number.isFinite(parsedChannelId)) {
    return <p>Некорректный id канала.</p>
  }

  if (communityQuery.isLoading) {
    return <p>Загружаем сообщество...</p>
  }

  if (communityQuery.isError || !communityQuery.data) {
    return <p>Не удалось загрузить сообщество канала.</p>
  }

  return (
    <section className="page">
      <h1>{communityQuery.data.channel.name} — сообщество</h1>
      <p className="page__lead">{communityQuery.data.channel.description || 'Описание отсутствует'}</p>
      <p className="video-card__meta">Подписчиков: {communityQuery.data.channel.subscribersCount}</p>
      <div className="page__actions">
        <Link className="app-button app-button--ghost" to={`/channels/${communityQuery.data.channel.id}`}>
          Назад к каналу
        </Link>
      </div>

      <h2>Публикации</h2>
      {!communityQuery.data.posts.length && <p className="page__lead">Пока нет публикаций.</p>}
      <ul className="comments__list">
        {communityQuery.data.posts.map((post) => (
          <li key={post.id} className="comments__item">
            {editingPostId === post.id ? (
              <form
                className="form comments__inline-form"
                onSubmit={(event) => {
                  event.preventDefault()
                  if (!editingPostValue.trim()) {
                    return
                  }
                  updatePostMutation.mutate({ postId: post.id, content: editingPostValue })
                }}
              >
                <textarea
                  className="form__input form__textarea"
                  rows={3}
                  value={editingPostValue}
                  onChange={(event) => setEditingPostValue(event.target.value)}
                />
                <div className="comments__actions">
                  <button className="app-button app-button--ghost" type="submit" disabled={updatePostMutation.isPending}>
                    Сохранить
                  </button>
                  <button
                    className="app-button app-button--ghost"
                    type="button"
                    onClick={() => {
                      setEditingPostId(null)
                      setEditingPostValue('')
                    }}
                  >
                    Отмена
                  </button>
                </div>
              </form>
            ) : (
              <p>{post.content}</p>
            )}
            <small>
                Автор {post.username || `Пользователь #${post.userId}`} • {new Date(post.createdAt).toLocaleString()} • комментариев: {post.comments.length}
            </small>
            {currentUserId === post.userId && editingPostId !== post.id && (
              <div className="comments__actions">
                <button
                  className="app-button app-button--ghost"
                  type="button"
                  onClick={() => {
                    setEditingPostId(post.id)
                    setEditingPostValue(post.content)
                  }}
                >
                  Редактировать пост
                </button>
              </div>
            )}

            {isAuthenticated ? (
              <form
                className="form comments__inline-form"
                onSubmit={(event) => {
                  event.preventDefault()
                  const content = commentDrafts[post.id]?.trim()
                  if (!content) {
                    return
                  }
                  createCommentMutation.mutate({ postId: post.id, content })
                }}
              >
                <textarea
                  className="form__input form__textarea"
                  rows={2}
                  placeholder="Написать комментарий"
                  value={commentDrafts[post.id] ?? ''}
                  onChange={(event) => setCommentDrafts((previous) => ({ ...previous, [post.id]: event.target.value }))}
                />
                <button className="app-button app-button--ghost" type="submit" disabled={createCommentMutation.isPending}>
                  Комментировать
                </button>
              </form>
            ) : null}

            {!!post.comments.length && (
              <ul className="subscription-list">
                {post.comments.map((comment) => (
                  <li key={comment.id} className="subscription-list__item">
                    {editingCommentId === comment.id ? (
                      <form
                        className="form comments__inline-form"
                        onSubmit={(event) => {
                          event.preventDefault()
                          if (!editingCommentValue.trim()) {
                            return
                          }
                          updateCommentMutation.mutate({ commentId: comment.id, content: editingCommentValue })
                        }}
                      >
                        <textarea
                          className="form__input form__textarea"
                          rows={2}
                          value={editingCommentValue}
                          onChange={(event) => setEditingCommentValue(event.target.value)}
                        />
                        <div className="comments__actions">
                          <button className="app-button app-button--ghost" type="submit" disabled={updateCommentMutation.isPending}>
                            Сохранить
                          </button>
                          <button
                            className="app-button app-button--ghost"
                            type="button"
                            onClick={() => {
                              setEditingCommentId(null)
                              setEditingCommentValue('')
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
                        Автор {comment.username || `Пользователь #${comment.userId}`} • {new Date(comment.createdAt).toLocaleString()}
                    </small>
                    {currentUserId === comment.userId && editingCommentId !== comment.id && (
                      <div className="comments__actions">
                        <button
                          className="app-button app-button--ghost"
                          type="button"
                          onClick={() => {
                            setEditingCommentId(comment.id)
                            setEditingCommentValue(comment.content)
                          }}
                        >
                          Редактировать
                        </button>
                        <button
                          className="app-button app-button--ghost"
                          type="button"
                          onClick={() => deleteCommentMutation.mutate(comment.id)}
                          disabled={deleteCommentMutation.isPending}
                        >
                          Удалить
                        </button>
                      </div>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </li>
        ))}
      </ul>
    </section>
  )
}

