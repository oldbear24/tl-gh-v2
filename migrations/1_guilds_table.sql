CREATE TABLE guilds (
    id BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    api_user VARCHAR(255) NULL,
    api_key VARCHAR(255) NULL,
    game_role VARCHAR(255) NULL,
    game_leader_role VARCHAR(255) NULL
);