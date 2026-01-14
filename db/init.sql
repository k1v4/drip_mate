-- Тип сезона
CREATE TYPE season AS ENUM ('summer', 'winter', 'autumn', 'spring');

-- Таблицы
CREATE TABLE access_level (
                              id SERIAL PRIMARY KEY,
                              name TEXT NOT NULL UNIQUE
);

CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       email TEXT NOT NULL UNIQUE,
                       password TEXT NOT NULL,
                       username TEXT UNIQUE,
                       name TEXT,
                       surname TEXT,
                       city TEXT,
                       access_id INT NOT NULL default 1
);

CREATE TABLE catalog (
                         id SERIAL PRIMARY KEY,
                         name TEXT NOT NULL UNIQUE,
                         category TEXT,
                         subcategory TEXT,
                         gender TEXT,
                         season season[],
                         formality_level SMALLINT,
                         material TEXT,
                         image_url TEXT,
                         created_at TIMESTAMP DEFAULT now(),
                         updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE music (
                       id SERIAL PRIMARY KEY,
                       name TEXT NOT NULL UNIQUE
);

CREATE TABLE music_user (
                            id SERIAL PRIMARY KEY,
                            user_id INT,
                            music_id INT
);

CREATE TABLE style_types (
                             id SERIAL PRIMARY KEY,
                             name TEXT NOT NULL UNIQUE
);

CREATE TABLE style_user (
                            id SERIAL PRIMARY KEY,
                            user_id INT,
                            style_id INT
);

CREATE TABLE color_types (
                             id SERIAL PRIMARY KEY,
                             name TEXT NOT NULL UNIQUE
);

CREATE TABLE color_user (
                            id SERIAL PRIMARY KEY,
                            user_id INT,
                            color_id INT
);

CREATE TABLE style_catalog (
                               id SERIAL PRIMARY KEY,
                               catalog_id INT,
                               style_id INT
);

CREATE TABLE color_catalog (
                               id SERIAL PRIMARY KEY,
                               catalog_id INT,
                               color_id INT
);

CREATE TABLE saved_outfits (
                               id SERIAL PRIMARY KEY,
                               user_id INT,
                               catalog_item_id INT,
                               created_at TIMESTAMP DEFAULT now()
);

-- Внешние ключи
ALTER TABLE users
    ADD CONSTRAINT fk_users_access
        FOREIGN KEY (access_id) REFERENCES access_level(id);

ALTER TABLE music_user
    ADD CONSTRAINT fk_music_user_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE music_user
    ADD CONSTRAINT fk_music_user_music
        FOREIGN KEY (music_id) REFERENCES music(id) ON DELETE CASCADE;

ALTER TABLE style_user
    ADD CONSTRAINT fk_style_user_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE style_user
    ADD CONSTRAINT fk_style_user_style
        FOREIGN KEY (style_id) REFERENCES style_types(id) ON DELETE CASCADE;

ALTER TABLE color_user
    ADD CONSTRAINT fk_color_user_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE color_user
    ADD CONSTRAINT fk_color_user_color
        FOREIGN KEY (color_id) REFERENCES color_types(id) ON DELETE CASCADE;

ALTER TABLE style_catalog
    ADD CONSTRAINT fk_style_catalog_catalog
        FOREIGN KEY (catalog_id) REFERENCES catalog(id) ON DELETE CASCADE;

ALTER TABLE style_catalog
    ADD CONSTRAINT fk_style_catalog_style
        FOREIGN KEY (style_id) REFERENCES style_types(id) ON DELETE CASCADE;

ALTER TABLE color_catalog
    ADD CONSTRAINT fk_color_catalog_catalog
        FOREIGN KEY (catalog_id) REFERENCES catalog(id) ON DELETE CASCADE;

ALTER TABLE color_catalog
    ADD CONSTRAINT fk_color_catalog_color
        FOREIGN KEY (color_id) REFERENCES color_types(id) ON DELETE CASCADE;

ALTER TABLE saved_outfits
    ADD CONSTRAINT fk_saved_outfits_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE saved_outfits
    ADD CONSTRAINT fk_saved_outfits_catalog
        FOREIGN KEY (catalog_item_id) REFERENCES catalog(id) ON DELETE CASCADE;


-- Справочные данные
INSERT INTO access_level(name) VALUES ('user'), ('admin');

INSERT INTO color_types (name) VALUES
                                       ('black'),
                                       ('white'),
                                       ('red'),
                                       ('blue'),
                                       ('green'),
                                       ('grey'),
                                       ('brown'),
                                       ('beige'),
                                       ('yellow');

INSERT INTO style_types (name) VALUES
                                       ('casual'),
                                       ('sport'),
                                       ('streetwear'),
                                       ('classic'),
                                       ('running'),
                                       ('basketball'),
                                       ('training'),
                                       ('lifestyle'),
                                       ('outdoor'),
                                       ('formal');

INSERT INTO music (name) VALUES
                                 ('hip-hop'),
                                 ('rap'),
                                 ('rock'),
                                 ('pop'),
                                 ('electronic'),
                                 ('house'),
                                 ('techno'),
                                 ('jazz'),
                                 ('rnb'),
                                 ('indie');

