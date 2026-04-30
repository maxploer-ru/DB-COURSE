CREATE TABLE IF NOT EXISTS community_posts
(
    id          SERIAL PRIMARY KEY,
    channel_id  INT         NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    user_id     INT         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    content     TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS community_post_comments
(
    id          SERIAL PRIMARY KEY,
    post_id     INT         NOT NULL REFERENCES community_posts (id) ON DELETE CASCADE,
    user_id     INT         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    content     TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
