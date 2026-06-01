import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import type { FormEvent } from 'react'
import { useAuthStore } from '../features/auth/store'
import { adminApi } from '../shared/api/endpoints'

export function AdminPage() {
  const { user } = useAuthStore()
  const [userId, setUserId] = useState('')
  const [role, setRole] = useState('user')
  const [message, setMessage] = useState('')

  const banMutation = useMutation({
    mutationFn: (id: number) => adminApi.banUser(id),
    onSuccess: (data) => setMessage(data.message),
  })

  const unbanMutation = useMutation({
    mutationFn: (id: number) => adminApi.unbanUser(id),
    onSuccess: (data) => setMessage(data.message),
  })

  const roleMutation = useMutation({
    mutationFn: (payload: { id: number; role: string }) => adminApi.changeUserRole(payload.id, { role: payload.role }),
    onSuccess: (data) => setMessage(data.message),
  })

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!userId) {
      return
    }
    banMutation.mutate(Number(userId))
  }

  const changeRole = () => {
    if (!userId) {
      return
    }
    const id = Number(userId)
    if (user && user.id === id) {
      setMessage('Нельзя менять роль самому себе.')
      return
    }
    roleMutation.mutate({ id, role })
  }

  return (
    <section className="page page--narrow">
      <h1>Admin</h1>
      {message && <p className="video-card__meta">{message}</p>}
      <form className="form" onSubmit={submit}>
        <label className="form__label">
          ID пользователя
          <input className="form__input" value={userId} onChange={(event) => setUserId(event.target.value)} />
        </label>
        <label className="form__label">
          Роль
          <select className="form__input" value={role} onChange={(event) => setRole(event.target.value)}>
            <option value="user">user</option>
            <option value="moderator">moderator</option>
            <option value="admin">admin</option>
          </select>
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
        <button
          className="app-button app-button--ghost"
          type="button"
          disabled={roleMutation.isPending}
          onClick={changeRole}
        >
          Изменить роль
        </button>
      </form>
    </section>
  )
}


