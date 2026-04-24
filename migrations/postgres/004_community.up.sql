CREATE TABLE IF NOT EXISTS community_posts
(
    id          SERIAL PRIMARY KEY,
    channel_id  INT         NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    user_id     INT         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    content     TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_community_posts_channel_created_at
    ON community_posts (channel_id, created_at DESC);

CREATE TABLE IF NOT EXISTS community_post_comments
(
    id          SERIAL PRIMARY KEY,
    post_id     INT         NOT NULL REFERENCES community_posts (id) ON DELETE CASCADE,
    user_id     INT         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    content     TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_community_post_comments_post_created_at
    ON community_post_comments (post_id, created_at DESC);

