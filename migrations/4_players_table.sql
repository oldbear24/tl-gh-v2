CREATE TABLE players (
    id BIGINT,   
    guild BIGINT REFERENCES guilds (id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    guild_nick VARCHAR(255),    
    combat_power integer DEFAULT 0,      
    role integer REFERENCES roles (id),
    weapon_1 integer REFERENCES weapons (id),
    weapon_2 integer REFERENCES weapons (id),     
    build_url VARCHAR(2048) NULL,
    active boolean  DEFAULT TRUE,
    PRIMARY KEY (id,guild)
);