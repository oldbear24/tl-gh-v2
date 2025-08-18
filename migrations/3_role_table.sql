CREATE TABLE roles(
        id SERIAL PRIMARY KEY,    
        name VARCHAR(255) NOT NULL,
        visible_name VARCHAR(255) NOT NULL,    
        emote VARCHAR(255) NOT NULL DEFAULT ''
);
INSERT INTO roles(name,visible_name,emote) VALUES ('tank','Tank','1369409105906634874'),
('dps','DPS','1369409122176471191'),
('healer','Healer','1369409090056486913') 