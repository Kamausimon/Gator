-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPostsForUser :many
SELECT posts.*
FROM posts
INNER JOIN feedfollows ON posts.feed_id = feedfollows.feed_id
WHERE feedfollows.user_id = $1
ORDER BY posts.published_at DESC NULLS LAST
LIMIT $2;
