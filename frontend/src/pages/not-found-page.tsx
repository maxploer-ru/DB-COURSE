import { Link } from 'react-router-dom'

export function NotFoundPage() {
  return (
    <section className="page">
      <h1>Страница не найдена</h1>
      <p className="page__lead">Проверьте адрес или вернитесь к списку видео.</p>
      <Link to="/videos" className="app-button">
        Открыть каталог
      </Link>
    </section>
  )
}

