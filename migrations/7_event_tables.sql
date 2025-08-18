create table events(
        id SERIAL PRIMARY KEY,    
        guild BIGINT,
        channel BIGINT,
        message_id BIGINT,
        name VARCHAR(100) NOT NULL,
        description VARCHAR(1000),
        date TIMESTAMP NOT NULL,  
        state VARCHAR(20) NOT NULL CHECK (state IN ('upcoming', 'ended', 'cancelled')) DEFAULT 'upcoming',
        created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (guild) REFERENCES guilds(id) ON DELETE CASCADE
    );
create table event_participants(       
        event INT NOT NULL,
        player BIGINT NOT NULL,
        guild BIGINT NOT NULL,
        status VARCHAR(20) NOT NULL CHECK (status IN ('going', 'not_going', 'tentative')) DEFAULT 'going',
        created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (event) REFERENCES events(id) ON DELETE CASCADE,
        FOREIGN KEY (player, guild) REFERENCES players(id, guild) ON DELETE CASCADE,
        PRIMARY KEY (event, player,guild)
    );