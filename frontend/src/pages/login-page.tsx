import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { authApi } from '../shared/api/endpoints'
import { useAuthStore } from '../features/auth/store'

const loginSchema = z.object({
  email: z.string().email('Введите корректный email'),
  password: z.string().min(6, 'Минимум 6 символов'),
})

type LoginFormValues = z.infer<typeof loginSchema>

export function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const setSession = useAuthStore((state) => state.setSession)

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormValues>({ defaultValues: { email: '', password: '' } })

  const onSubmit = handleSubmit(async (values) => {
    const parsed = loginSchema.safeParse(values)
    if (!parsed.success) {
      const first = parsed.error.issues[0]
      setError(first.path[0] as 'email' | 'password', { message: first.message })
      return
    }

    try {
      const response = await authApi.login(parsed.data)
      setSession(response)
      const redirectPath = (location.state as { from?: string } | null)?.from ?? '/videos'
      navigate(redirectPath, { replace: true })
    } catch {
      setError('root', { message: 'Не удалось войти. Проверьте логин и пароль.' })
    }
  })

  return (
    <section className="page page--narrow">
      <h1>Вход</h1>
      <form className="form" onSubmit={onSubmit}>
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
          {isSubmitting ? 'Входим...' : 'Войти'}
        </button>
      </form>
      <p>
        Еще нет аккаунта? <Link to="/register">Зарегистрируйтесь</Link>
      </p>
    </section>
  )
}


