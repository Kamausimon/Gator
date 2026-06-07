-- +goose Up
CREATE TABLE feedfollows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    user_id UUID NOT NULL,
    CONSTRAINT fk_users_feedfollows
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
        
    feed_id UUID NOT NULL,
    CONSTRAINT fk_feeds_feedfollows
        FOREIGN KEY (feed_id)
        REFERENCES feeds(id)
        ON DELETE CASCADE,
        
    -- Ensures a user cannot follow the same feed more than once
    CONSTRAINT uq_user_feed_pair UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feedfollows;
