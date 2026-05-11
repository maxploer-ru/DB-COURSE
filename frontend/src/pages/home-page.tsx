import { Link } from 'react-router-dom'

export function HomePage() {
  return (
    <section className="page">
      <h1>ZVideo Frontend</h1>
      <p className="page__lead">Каркас SPA готов: роутинг, авторизация, список видео и страница видео.</p>
      <div className="page__actions">
        <Link to="/videos" className="app-button">
          Перейти к видео
        </Link>
        <Link to="/my-channel" className="app-button app-button--ghost">
          Создать свой канал
        </Link>
        <Link to="/login" className="app-button app-button--ghost">
          Войти
        </Link>
      </div>
    </section>
  )
}


