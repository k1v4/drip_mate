ALTER TABLE users
    DROP CONSTRAINT IF EXISTS fk_users_access_level;

ALTER TABLE catalog
    DROP CONSTRAINT IF EXISTS fk_catalog_category;

ALTER TABLE catalog
    DROP CONSTRAINT IF EXISTS fk_catalog_season;

ALTER TABLE music_user
    DROP CONSTRAINT IF EXISTS fk_music_user_user;

ALTER TABLE music_user
    DROP CONSTRAINT IF EXISTS fk_music_user_music;

ALTER TABLE style_user
    DROP CONSTRAINT IF EXISTS fk_style_user_user;

ALTER TABLE style_user
    DROP CONSTRAINT IF EXISTS fk_style_user_style;

ALTER TABLE color_user
    DROP CONSTRAINT IF EXISTS fk_color_user_user;

ALTER TABLE color_user
    DROP CONSTRAINT IF EXISTS fk_color_user_color;

ALTER TABLE style_catalog
    DROP CONSTRAINT IF EXISTS fk_style_catalog_catalog;

ALTER TABLE style_catalog
    DROP CONSTRAINT IF EXISTS fk_style_catalog_style;

ALTER TABLE color_catalog
    DROP CONSTRAINT IF EXISTS fk_color_catalog_catalog;

ALTER TABLE color_catalog
    DROP CONSTRAINT IF EXISTS fk_color_catalog_color;

ALTER TABLE saved_outfits_name
    DROP CONSTRAINT IF EXISTS fk_saved_outfits_name_user;

ALTER TABLE saved_outfits
    DROP CONSTRAINT IF EXISTS fk_saved_outfits_outfit_name;

ALTER TABLE saved_outfits
    DROP CONSTRAINT IF EXISTS fk_saved_outfits_catalog;

ALTER TABLE recommendation_log
    DROP CONSTRAINT IF EXISTS fk_recommendation_log_user;

ALTER TABLE user_interactions
    DROP CONSTRAINT IF EXISTS fk_user_interactions_user;

ALTER TABLE user_interactions
    DROP CONSTRAINT IF EXISTS fk_user_interactions_recommendation_log;



DROP TABLE IF EXISTS user_interactions;
DROP TABLE IF EXISTS recommendation_log;
DROP TABLE IF EXISTS saved_outfits;
DROP TABLE IF EXISTS saved_outfits_name;
DROP TABLE IF EXISTS color_catalog;
DROP TABLE IF EXISTS style_catalog;
DROP TABLE IF EXISTS color_user;
DROP TABLE IF EXISTS style_user;
DROP TABLE IF EXISTS music_user;
DROP TABLE IF EXISTS catalog;
DROP TABLE IF EXISTS music;
DROP TABLE IF EXISTS color_types;
DROP TABLE IF EXISTS style_types;
DROP TABLE IF EXISTS category;
DROP TABLE IF EXISTS season;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS access_level;