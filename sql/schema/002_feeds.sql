-- +goose Up
CREATE TABLE feeds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR (255) NOT NULL,
    user_id UUID NOT NULL,
    CONSTRAINT fk_users_feeds 
       FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    url VARCHAR (255) UNIQUE
);

-- +goose Down
DROP TABLE feeds;