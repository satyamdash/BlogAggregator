-- name: CreatePost :one
INSERT INTO posts (title, url, description,feed_id,published_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT p.*
FROM posts p
JOIN feeds f ON p.feed_id = f.id
WHERE f.user_id = $1
ORDER BY p.published_at DESC
LIMIT $2 OFFSET $3;

-- name: GetAllPosts :many
SELECT * FROM posts;

-- name: AddBookmark :one
INSERT INTO post_bookmarks (user_id, post_id) VALUES ($1, $2) RETURNING id;

-- name: RemoveBookmark :exec
DELETE FROM post_bookmarks WHERE user_id=$1 AND post_id=$2;

-- name: GetBookmarksForUser :many
SELECT posts.*
FROM posts
JOIN post_bookmarks ON posts.id = post_bookmarks.post_id
WHERE post_bookmarks.user_id=$1
ORDER BY post_bookmarks.created_at DESC;


-- name: AddLike :one
INSERT INTO post_likes (user_id, post_id) VALUES ($1, $2) RETURNING id;

-- name: RemoveLike :exec
DELETE FROM post_likes WHERE user_id=$1 AND post_id=$2;

-- name: GetLikesForUser :many
SELECT posts.*
FROM posts
JOIN post_likes ON posts.id = post_likes.post_id
WHERE post_likes.user_id=$1
ORDER BY post_likes.created_at DESC;

