-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feedfollows (id, created_at, updated_at, user_id, feed_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING *
)
SELECT 
    iff.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM inserted_feed_follow iff
INNER JOIN users ON iff.user_id = users.id
INNER JOIN feeds ON iff.feed_id = feeds.id;

-- name: GetFeedFollowsForUser :many
SELECT 
    feedfollows.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM feedfollows
INNER JOIN users ON feedfollows.user_id = users.id
INNER JOIN feeds ON feedfollows.feed_id = feeds.id
WHERE users.name = $1;


-- name: DeleteFeedFollowForUser :exec
DELETE FROM feedfollows 
WHERE user_id = $1 AND feed_id = $2;