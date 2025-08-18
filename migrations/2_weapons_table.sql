CREATE TABLE weapons(
        id SERIAL PRIMARY KEY,    
        name VARCHAR(255) NOT NULL,
        visible_name VARCHAR(255) NOT NULL,    
        emote VARCHAR(255) NOT NULL DEFAULT ''
);
INSERT INTO weapons(name,visible_name,emote) VALUES ('sns','Sword and shield','1339399450564628501'),
('staff','Staff','1339399440011624540'),
('crossbow','Crossbow','1339399426816610356'),
('daggers','Daggers','1339399413210021899'),
('wand','Wand','1339399410031005800'),
('spear','Spear','1339399395141226516'),
('longbow','Longbow','1339399379127369749'),
('greatsword','Greatsword','1339399348970324019');