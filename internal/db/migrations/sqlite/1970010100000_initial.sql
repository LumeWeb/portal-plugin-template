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
-- SQLite version of the schema

CREATE TABLE IF NOT EXISTS `items`
(
    `id`         integer PRIMARY KEY AUTOINCREMENT,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime,
    `name`       text    NOT NULL,
    `description` text
);

-- +goose Down
DROP TABLE IF EXISTS `items`;
