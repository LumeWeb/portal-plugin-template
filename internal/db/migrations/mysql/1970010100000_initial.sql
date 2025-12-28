-- +goose Up
-- Initial database schema for the template plugin
-- This migration establishes the base schema for storing items
--
-- Usage:
-- This migration runs automatically when the plugin is first installed
-- or when the database is initialized. It is idempotent and can be
-- run multiple times safely.
--
-- Tables:
-- items: Stores the basic item information with timestamps for tracking
--        creation, updates, and soft deletes

CREATE TABLE IF NOT EXISTS `items`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT,
    `created_at`  datetime(3)     DEFAULT NULL,
    `updated_at`  datetime(3)     DEFAULT NULL,
    `deleted_at`  datetime(3)     DEFAULT NULL,
    `name`        varchar(255)    NOT NULL,
    `description` longtext,
    PRIMARY KEY (`id`)
);

-- +goose Down
DROP TABLE IF EXISTS `items`;
