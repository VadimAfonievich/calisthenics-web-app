# Media storage

Binary data is never stored in PostgreSQL or Render's ephemeral filesystem. `media_assets` stores ownership, type, provider/key, public URL, MIME type, dimensions, duration, and size. Published content references media by UUID. Deletion is rejected while a cover reference exists.

The domain-facing `storage.Provider` supports `Upload`, `Delete`, and `PublicURL`; production is intended for an S3-compatible provider such as Cloudflare R2, AWS S3, Backblaze B2, or MinIO. Until credentials/provider wiring is configured, `/coach/media/upload` returns controlled `OBJECT_STORAGE_UNAVAILABLE`, while HTTPS external image/video URLs remain usable. External YouTube, Rutube, VK Video, and direct URLs are referenced, never downloaded.

Configuration:

- `OBJECT_STORAGE_PROVIDER`
- `OBJECT_STORAGE_ENDPOINT`
- `OBJECT_STORAGE_REGION`
- `OBJECT_STORAGE_BUCKET`
- `OBJECT_STORAGE_ACCESS_KEY`
- `OBJECT_STORAGE_SECRET_KEY`
- `OBJECT_STORAGE_PUBLIC_URL`

The intended limits are 10 MB for JPEG/PNG/WebP and 500 MB for MP4/WebM. Adaptive streaming and transcoding are outside this phase.
