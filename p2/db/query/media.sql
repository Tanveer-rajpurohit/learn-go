-- name: CreateMediaAsset :one
INSERT INTO media_assets (user_id, target_size, raw_url, status)
VALUES ($1, $2, $3, 'pending')
RETURNING *;

-- name: GetMediaAsset :one
SELECT * FROM media_assets WHERE id = $1;

-- name: UpdateMediaStatus :exec
UPDATE media_assets SET status = $2, updated_at = NOW() WHERE id = $1;

-- name: UpdateMediaVariants :one
UPDATE media_assets
SET
    variant_sd = COALESCE(sqlc.narg('variant_sd'), variant_sd),
    variant_hd = COALESCE(sqlc.narg('variant_hd'), variant_hd),
    variant_raw = COALESCE(sqlc.narg('variant_raw'), variant_raw),
    status = 'completed',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateMediaFailed :exec
UPDATE media_assets
SET
    status = 'failed',
    error_message = $2,
    retry_count = retry_count + 1,
    updated_at = NOW()
WHERE id = $1;