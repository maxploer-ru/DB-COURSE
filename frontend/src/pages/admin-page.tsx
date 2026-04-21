import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import type { FormEvent } from 'react'
import { adminApi } from '../shared/api/endpoints'

export function AdminPage() {
  const [userId, setUserId] = useState('')
  const [message, setMessage] = useState('')

  const banMutation = useMutation({
    mutationFn: (id: number) => adminApi.banUser(id),
    onSuccess: (data) => setMessage(data.message),
  })

  const unbanMutation = useMutation({
    mutationFn: (id: number) => adminApi.unbanUser(id),
    onSuccess: (data) => setMessage(data.message),
  })

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!userId) {
      return
    }
    banMutation.mutate(Number(userId))
  }

  return (
    <section className="page page--narrow">
      <h1>Admin</h1>
      <p className="page__lead">Раздел использует ручки `/admin/users/{'{id}'}/ban` и `/admin/users/{'{id}'}/unban`.</p>
      {message && <p className="video-card__meta">{message}</p>}
      <form className="form" onSubmit={submit}>
        <label className="form__label">
          ID пользователя
          <input className="form__input" value={userId} onChange={(event) => setUserId(event.target.value)} />
        </label>
        <button className="app-button app-button--ghost" type="submit" disabled={banMutation.isPending}>
          Забанить
        </button>
        <button
          className="app-button app-button--ghost"
          type="button"
          disabled={unbanMutation.isPending}
          onClick={() => {
            if (!userId) {
              return
            }
            unbanMutation.mutate(Number(userId))
          }}
        >
          Разбанить
        </button>
      </form>
    </section>
  )
}


