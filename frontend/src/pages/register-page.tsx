import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { Link, useNavigate } from 'react-router-dom'
import { authApi } from '../shared/api/endpoints'
import { useAuthStore } from '../features/auth/store'

const registerSchema = z.object({
  username: z.string().min(3, 'Минимум 3 символа'),
  email: z.string().email('Введите корректный email'),
  password: z.string().min(6, 'Минимум 6 символов'),
})

type RegisterFormValues = z.infer<typeof registerSchema>

export function RegisterPage() {
  const navigate = useNavigate()
  const setSession = useAuthStore((state) => state.setSession)

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<RegisterFormValues>({ defaultValues: { username: '', email: '', password: '' } })

  const onSubmit = handleSubmit(async (values) => {
    const parsed = registerSchema.safeParse(values)
    if (!parsed.success) {
      const first = parsed.error.issues[0]
      setError(first.path[0] as 'username' | 'email' | 'password', { message: first.message })
      return
    }

    try {
      await authApi.register(parsed.data)
      const session = await authApi.login({ email: parsed.data.email, password: parsed.data.password })
      setSession(session)
      navigate('/videos', { replace: true })
    } catch {
      setError('root', { message: 'Не удалось зарегистрироваться. Попробуйте позже.' })
    }
  })

  return (
    <section className="page page--narrow">
      <h1>Регистрация</h1>
      <form className="form" onSubmit={onSubmit}>
        <label className="form__label">
          Имя пользователя
          <input className="form__input" {...register('username')} />
          {errors.username && <span className="form__error">{errors.username.message}</span>}
        </label>
        <label className="form__label">
          Email
          <input type="email" className="form__input" {...register('email')} />
          {errors.email && <span className="form__error">{errors.email.message}</span>}
        </label>
        <label className="form__label">
          Пароль
          <input type="password" className="form__input" {...register('password')} />
          {errors.password && <span className="form__error">{errors.password.message}</span>}
        </label>
        {errors.root && <p className="form__error">{errors.root.message}</p>}
        <button type="submit" className="app-button" disabled={isSubmitting}>
          {isSubmitting ? 'Создаем...' : 'Создать аккаунт'}
        </button>
      </form>
      <p>
        Уже зарегистрированы? <Link to="/login">Войти</Link>
      </p>
    </section>
  )
}



