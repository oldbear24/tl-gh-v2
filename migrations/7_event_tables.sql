create table events(
        id SERIAL PRIMARY KEY,    
        guild BIGINT,
        channel BIGINT,
        messageId BIGINT,
        name VARCHAR(100) NOT NULL,
        description VARCHAR(1000),
        date TIMESTAMP NOT NULL,  
        FOREIGN KEY (guild) REFERENCES guilds(id) ON DELETE CASCADE
    );
