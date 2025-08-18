create table player_configs(
        guild BIGINT,
        player BIGINT,
        events BOOLEAN NOT NULL DEFAULT FALSE,
        auctions BOOLEAN NOT NULL DEFAULT FALSE,
        FOREIGN KEY (player, guild) REFERENCES players(id, guild) ON DELETE CASCADE,
        PRIMARY KEY (guild,player)
    );

CREATE OR REPLACE FUNCTION create_player_config()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO player_configs(guild, player) VALUES (NEW.guild, NEW.id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER trg_create_player_config
AFTER INSERT ON players
FOR EACH ROW
EXECUTE FUNCTION create_player_config();