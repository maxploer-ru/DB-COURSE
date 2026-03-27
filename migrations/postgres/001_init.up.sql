CREATE TABLE IF NOT EXISTS roles
(
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(32) UNIQUE NOT NULL,
    is_default BOOLEAN            NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS users
(
    id                    SERIAL PRIMARY KEY,
    role_id               INT                NOT NULL REFERENCES roles (id) ON DELETE RESTRICT,
    username              VARCHAR(32) UNIQUE NOT NULL,
    email                 VARCHAR(64) UNIQUE NOT NULL,
    password_hash         TEXT               NOT NULL,
    notifications_enabled BOOLEAN            NOT NULL DEFAULT TRUE,
    is_active             BOOLEAN            NOT NULL DEFAULT TRUE,
    created_at            TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS channels
(
    id          SERIAL PRIMARY KEY,
    user_id     INT                NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name        VARCHAR(32) UNIQUE NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS videos
(
    id               SERIAL PRIMARY KEY,
    channel_id       INT         NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    title            VARCHAR(64) NOT NULL,
    description      TEXT,
    filepath         TEXT        NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS playlists
(
    id          SERIAL PRIMARY KEY,
    channel_id  INT         NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    name        VARCHAR(32) NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS viewings
(
    id         SERIAL PRIMARY KEY,
    user_id    INT         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    video_id   INT         NOT NULL REFERENCES videos (id) ON DELETE CASCADE,
    watched_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS comments
(
    id         SERIAL PRIMARY KEY,
    user_id    INT         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    video_id   INT         NOT NULL REFERENCES videos (id) ON DELETE CASCADE,
    content    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS subscriptions
(
    user_id          INT         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    channel_id       INT         NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, channel_id),
    new_videos_count INT         NOT NULL DEFAULT 0,
    subscribed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS playlist_items
(
    playlist_id INT         NOT NULL REFERENCES playlists (id) ON DELETE CASCADE,
    video_id    INT         NOT NULL REFERENCES videos (id) ON DELETE CASCADE,
    PRIMARY KEY (playlist_id, video_id),
    number      INT         NOT NULL,
    added_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS video_ratings
(
    user_id  INT         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    video_id INT         NOT NULL REFERENCES videos (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, video_id),
    liked    BOOLEAN     NOT NULL,
    rated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS comment_ratings
(
    user_id    INT         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    comment_id INT         NOT NULL REFERENCES comments (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, comment_id),
    liked      BOOLEAN     NOT NULL,
    rated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
