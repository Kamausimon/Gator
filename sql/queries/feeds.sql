-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, user_id, url) 
VALUES ($1, $2,$3,$4,$5,$6) RETURNING *;

-- name: ListFeedsWithUser :many
SELECT 
    feeds.name AS feed_name,
    feeds.url,
    users.name AS user_name
FROM feeds
INNER JOIN users ON feeds.user_id = users.id;


-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = $1;