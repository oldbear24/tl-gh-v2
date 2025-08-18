create table notification_queue(
    id SERIAL PRIMARY KEY,
    player_id BIGINT,
    notification_text VARCHAR(4096),
    created timestamp DEFAULT CURRENT_TIMESTAMP
);

