CREATE
    OR REPLACE FUNCTION notify_subscribers_about_new_video(p_channel_id INT)
    RETURNS VOID
    LANGUAGE plpgsql
AS
$$
BEGIN
    UPDATE subscriptions s
    SET new_videos_count = s.new_videos_count + 1
    FROM users u
    WHERE s.channel_id = p_channel_id
      AND s.user_id = u.id
      AND u.notifications_enabled = TRUE;
END;
$$;
