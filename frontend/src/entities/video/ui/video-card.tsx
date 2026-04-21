import { Link } from 'react-router-dom'
import type { Video } from '../../../shared/api/types'

type VideoCardProps = {
  video: Video
}

export function VideoCard({ video }: VideoCardProps) {
  return (
    <article className="video-card">
      <Link to={`/videos/${video.id}`} className="video-card__link">
        <div className="video-card__preview" />
        <h3>{video.title}</h3>
      </Link>
      <p className="video-card__meta">
        <Link to={`/channels/${video.channelId}`}>Канал #{video.channelId}</Link> • {video.views} просмотров • {video.likes} лайков • {video.dislikes} дизлайков
      </p>
    </article>
  )
}




