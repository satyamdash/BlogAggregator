-- name: CreateFeed :one
INSERT INTO feeds (name,url,user_id)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetFeedNameByUrl :one
SELECT * 
FROM feeds
WHERE url=$1;

-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (user_id, feed_id)
    VALUES ($1, $2)
    RETURNING *
)
SELECT
    iff.id,
    iff.created_at,
    iff.updated_at,
    iff.user_id,
    iff.feed_id,
    u.name AS user_name,
    f.name AS feed_name
FROM inserted_feed_follow iff
INNER JOIN users u ON iff.user_id = u.id
INNER JOIN feeds f ON iff.feed_id = f.id;

-- name: GetFeedFollowsForUser :many
SELECT
    ff.id,
    ff.created_at,
    ff.updated_at,
    ff.user_id,
    ff.feed_id,
    u.name AS user_name,
    f.name AS feed_name
FROM feed_follows ff
INNER JOIN users u ON ff.user_id = u.id
INNER JOIN feeds f ON ff.feed_id = f.id
WHERE ff.user_id = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows ff
WHERE ff.user_id = $1 and ff.feed_id=$2;

-- name: MarkFeedFetched :one
UPDATE feeds
SET 
    last_fetched_at = $1,
    updated_at = $2
WHERE id = $3
RETURNING *;

-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds
ORDER BY last_fetched_at NULLS FIRST
LIMIT 1;

