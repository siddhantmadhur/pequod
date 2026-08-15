-- name: GetProfiles :many
SELECT id, username, password, type FROM profiles;

-- name: IsFinishedSetup :one 
SELECT count(*) FROM profiles;

-- name: CreateProfile :exec
INSERT INTO profiles (username, password, type) 
VALUES ( ?, ?, ? );

-- name: GetAdminUser :one
SELECT id, username, password, type FROM profiles
WHERE type = 0;

-- name: UpdateUser :exec
UPDATE profiles 
SET username = ?, password = ?
WHERE id = ?; 

-- name: GetUserFromUsername :one
SELECT id, username, password, type FROM profiles 
WHERE username = ?;

-- name: GetUserByID :one
SELECT id, username, password, type FROM profiles 
WHERE id = ?;

-- name: CreateSession :one
INSERT INTO sessions (id, user_id, created_at, access_token, refresh_token, refresh_expires_at, access_expires_at, device, device_name, client_name, client_version)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM sessions
WHERE id = ?;

-- name: CreateMediaLibrary :one
INSERT INTO media_library(created_at, name, description, device_path, media_type, owner_id) 
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetAllMediaLibraries :many
SELECT * FROM media_library;

-- name: GetMediaLibrary :one
SELECT * FROM media_library
WHERE id = ?;

-- name: AddNewContentFile :one
INSERT INTO content_library (
    media_library_id,
    created_at,
    file_path,
    name,
    media_title,
    description,
    cover_url,
    parent_id,
    external_provider,
    external_provider_id,
    media_type,
    classifier
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetContentFromPath :one 
SELECT * FROM content_library
WHERE file_path = ?;

-- name: GetContentFromExternalId :one 
SELECT * FROM content_library
WHERE external_provider_id = ?;

-- name: GetContentByID :one
SELECT * FROM content_library
WHERE id = ?;

-- name: GetAllContentFiles :many
SELECT * FROM content_library
WHERE media_library_id = ?;

-- name: GetAllShows :many
SELECT * FROM content_library
WHERE media_library_id = ? AND media_type = ?;

-- name: GetContentFromParentId :many
SELECT * FROM content_library
WHERE parent_id = ?;
