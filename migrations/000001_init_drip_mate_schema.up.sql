CREATE TABLE access_level (
                              id   SERIAL      PRIMARY KEY,
                              name TEXT        NOT NULL
);

CREATE TABLE users (
                       id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
                       email      TEXT        NOT NULL UNIQUE,
                       password   TEXT        NOT NULL,
                       username   TEXT,
                       name       TEXT,
                       surname    TEXT,
                       city       TEXT,
                       gender     VARCHAR(16),
                       access_id  INT         NOT NULL DEFAULT 1
);

CREATE TABLE season (
                        id   SERIAL  PRIMARY KEY,
                        name TEXT    NOT NULL
);

CREATE TABLE category (
                          id   SERIAL  PRIMARY KEY,
                          name TEXT    NOT NULL
);

CREATE TABLE style_types (
                             id   SERIAL  PRIMARY KEY,
                             name TEXT    NOT NULL
);

CREATE TABLE color_types (
                             id   SERIAL  PRIMARY KEY,
                             name TEXT    NOT NULL,
                             hex varchar(7) NOT NULL
);

CREATE TABLE music (
                       id   SERIAL  PRIMARY KEY,
                       name TEXT    NOT NULL
);

CREATE TABLE catalog (
                         id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
                         name            TEXT        NOT NULL,
                         category_id     INT         NOT NULL,
                         gender          VARCHAR(16),
                         season_id       INT         NOT NULL,
                         formality_level SMALLINT,
                         material        VARCHAR(64),
                         image_url       TEXT,
                         created_at      TIMESTAMP   NOT NULL DEFAULT now(),
                         updated_at      TIMESTAMP   NOT NULL DEFAULT now(),
                         is_deleted      boolean     NOT NULL DEFAULT false
);

CREATE TABLE music_user (
                            id       SERIAL  PRIMARY KEY,
                            user_id  UUID    NOT NULL,
                            music_id INT     NOT NULL
);

CREATE TABLE style_user (
                            id       SERIAL  PRIMARY KEY,
                            user_id  UUID    NOT NULL,
                            style_id INT     NOT NULL
);

CREATE TABLE color_user (
                            id       SERIAL  PRIMARY KEY,
                            user_id  UUID    NOT NULL,
                            color_id INT     NOT NULL
);

CREATE TABLE style_catalog (
                               id         SERIAL  PRIMARY KEY,
                               catalog_id UUID    NOT NULL,
                               style_id   INT     NOT NULL
);

CREATE TABLE color_catalog (
                               id         SERIAL  PRIMARY KEY,
                               catalog_id UUID    NOT NULL,
                               color_id   INT     NOT NULL
);

CREATE TABLE saved_outfits_name (
                                    id      UUID  PRIMARY KEY DEFAULT gen_random_uuid(),
                                    name    TEXT    NOT NULL,
                                    user_id UUID    NOT NULL
);

CREATE TABLE saved_outfits (
                               id              SERIAL      PRIMARY KEY,
                               outfit_id       UUID         NOT NULL,
                               catalog_item_id UUID        NOT NULL,
                               created_at      TIMESTAMP   NOT NULL DEFAULT now()
);

CREATE TABLE recommendation_log (
                                    id              SERIAL      PRIMARY KEY,
                                    user_id         UUID        NOT NULL,
    -- [[item_id, ...], [...]] — список показанных аутфитов
                                    outfits_shown   JSONB       NOT NULL,
    -- cold_start | fm
                                    model_phase     VARCHAR(32) NOT NULL,
    -- параметры запроса: season, style, occasion
                                    request_context JSONB,
                                    created_at      TIMESTAMP   NOT NULL DEFAULT now()
);

CREATE TABLE user_interactions (
                                   id                    SERIAL      PRIMARY KEY,
                                   user_id               UUID        NOT NULL,
                                   recommendation_log_id INT         NOT NULL,
    -- [item_id, ...] — состав аутфита
                                   outfit_items          JSONB       NOT NULL,
    -- save | skip
                                   event_type            VARCHAR(16) NOT NULL,
    -- season, style, occasion
                                   context_snapshot      JSONB,
                                   created_at            TIMESTAMP   NOT NULL DEFAULT now()
);



ALTER TABLE users
    ADD CONSTRAINT fk_users_access_level
        FOREIGN KEY (access_id) REFERENCES access_level (id);

ALTER TABLE catalog
    ADD CONSTRAINT fk_catalog_category
        FOREIGN KEY (category_id) REFERENCES category (id);

ALTER TABLE catalog
    ADD CONSTRAINT fk_catalog_season
        FOREIGN KEY (season_id) REFERENCES season (id);

ALTER TABLE music_user
    ADD CONSTRAINT fk_music_user_user
        FOREIGN KEY (user_id) REFERENCES users (id);

ALTER TABLE music_user
    ADD CONSTRAINT fk_music_user_music
        FOREIGN KEY (music_id) REFERENCES music (id);

ALTER TABLE style_user
    ADD CONSTRAINT fk_style_user_user
        FOREIGN KEY (user_id) REFERENCES users (id);

ALTER TABLE style_user
    ADD CONSTRAINT fk_style_user_style
        FOREIGN KEY (style_id) REFERENCES style_types (id);

ALTER TABLE color_user
    ADD CONSTRAINT fk_color_user_user
        FOREIGN KEY (user_id) REFERENCES users (id);

ALTER TABLE color_user
    ADD CONSTRAINT fk_color_user_color
        FOREIGN KEY (color_id) REFERENCES color_types (id);

ALTER TABLE style_catalog
    ADD CONSTRAINT fk_style_catalog_catalog
        FOREIGN KEY (catalog_id) REFERENCES catalog (id);

ALTER TABLE style_catalog
    ADD CONSTRAINT fk_style_catalog_style
        FOREIGN KEY (style_id) REFERENCES style_types (id);

ALTER TABLE color_catalog
    ADD CONSTRAINT fk_color_catalog_catalog
        FOREIGN KEY (catalog_id) REFERENCES catalog (id);

ALTER TABLE color_catalog
    ADD CONSTRAINT fk_color_catalog_color
        FOREIGN KEY (color_id) REFERENCES color_types (id);

ALTER TABLE saved_outfits_name
    ADD CONSTRAINT fk_saved_outfits_name_user
        FOREIGN KEY (user_id) REFERENCES users (id);

ALTER TABLE saved_outfits
    ADD CONSTRAINT fk_saved_outfits_outfit_name
        FOREIGN KEY (outfit_id) REFERENCES saved_outfits_name (id) ON DELETE CASCADE;

ALTER TABLE saved_outfits
    ADD CONSTRAINT fk_saved_outfits_catalog
        FOREIGN KEY (catalog_item_id) REFERENCES catalog (id);

ALTER TABLE recommendation_log
    ADD CONSTRAINT fk_recommendation_log_user
        FOREIGN KEY (user_id) REFERENCES users (id);

ALTER TABLE user_interactions
    ADD CONSTRAINT fk_user_interactions_user
        FOREIGN KEY (user_id) REFERENCES users (id);

ALTER TABLE user_interactions
    ADD CONSTRAINT fk_user_interactions_recommendation_log
        FOREIGN KEY (recommendation_log_id) REFERENCES recommendation_log (id);

INSERT INTO access_level (name) VALUES ('user'), ('admin');

-- Сезоны
INSERT INTO season (name) VALUES
                              ('spring'),
                              ('summer'),
                              ('autumn'),
                              ('winter');

-- Категории одежды (важно для сборки аутфита по типу)
INSERT INTO category (name) VALUES
                                ('top'),        -- 1: верх (футболки, рубашки, свитера)
                                ('bottom'),     -- 2: низ (брюки, юбки, шорты)
                                ('shoes'),      -- 3: обувь
                                ('outerwear'),  -- 4: верхняя одежда (куртки, пальто)
                                ('dress'),      -- 5: платья/комбинезоны (верх+низ в одном)
                                ('bag'),        -- 6: сумки
                                ('accessories');-- 7: аксессуары (шарфы, шапки, ремни)

-- Стили
INSERT INTO style_types (name) VALUES
                                   ('casual'),
                                   ('sport'),
                                   ('business'),
                                   ('romantic'),
                                   ('streetwear'),
                                   ('classic'),
                                   ('bohemian'),
                                   ('minimalist');

-- Цвета
INSERT INTO color_types (name, hex) VALUES
                                        ('black',  '#000000'),
                                        ('white',  '#FFFFFF'),
                                        ('grey',   '#808080'),
                                        ('beige',  '#F5F5DC'),
                                        ('brown',  '#A52A2A'),
                                        ('red',    '#FF0000'),
                                        ('blue',   '#0000FF'),
                                        ('green',  '#008000'),
                                        ('yellow', '#FFFF00'),
                                        ('pink',   '#FFC0CB'),
                                        ('purple', '#800080'),
                                        ('orange', '#FFA500');

-- Музыкальные жанры
INSERT INTO music (name) VALUES
                             ('rock'),
                             ('jazz'),
                             ('hip-hop'),
                             ('classical'),
                             ('pop'),
                             ('indie'),
                             ('electronic'),
                             ('rnb');
