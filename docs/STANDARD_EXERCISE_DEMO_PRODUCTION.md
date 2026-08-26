# Standard exercise demo production pipeline

## Release gate

Production demo work starts only after a fresh custom-format PostgreSQL backup passes `pg_restore --list`, migration 14 is clean, and the API/web revision is healthy. Never put credentials, database URLs, or licensed source files in Git. The first acceptance batch is limited to 16 exercises; the 247-item library is explicitly out of scope until that batch is approved.

## Visual and encoding specification

- One clip shows one exercise, not a tutorial. The movement must be understandable within 1–2 seconds; technique details remain in Exercise text.
- Preferred duration is 1–4 seconds; the hard attachment limit is 6 seconds. Use 24–30 fps, no audio, no music and no spoken track.
- Primary delivery is H.264 MP4. WebM is optional; GIF or a JPEG/PNG/WebP poster is a fallback. Maximum file size is 5 MB.
- Prefer 720×720 (1:1) or 720×900 (4:5), framed safely for a narrow mobile viewport.
- Use one athlete in neutral sportswear, a clean high-contrast non-distracting background, a static camera and the full range of motion in frame.
- Do not add embedded text, watermarks, third-party logos, flashing edits or camera motion. Start and end poses should join cleanly when possible.

Camera angles:

| Movement | Preferred view |
| --- | --- |
| Push exercises | Side or front three-quarter |
| Pull-up | Side/front three-quarter with the full bar visible |
| Squat and lunge | Side/front three-quarter |
| Handstand and planche | Side |
| Front lever | Side |
| Muscle-up | Side/front three-quarter |
| Joint rotations | Front or slight three-quarter so the joint path is obvious |

Only original, commissioned, correctly licensed, or commercially permitted generated media may be used. Do not download clips from Leap, YouTube, Instagram, TikTok, other fitness applications, or unrelated sites.

## Object storage

The application keeps binary data outside PostgreSQL and Render's filesystem. `storage.Provider` remains the boundary; `storage.S3` implements it through an S3-compatible API. Configure these secret/runtime values in Render, never in source control:

- `OBJECT_STORAGE_PROVIDER=s3`
- `OBJECT_STORAGE_ENDPOINT` — HTTP(S) S3 API origin without a path
- `OBJECT_STORAGE_REGION`
- `OBJECT_STORAGE_BUCKET`
- `OBJECT_STORAGE_ACCESS_KEY`
- `OBJECT_STORAGE_SECRET_KEY`
- `OBJECT_STORAGE_PUBLIC_URL` — stable HTTPS CDN/public bucket base URL

No provider is selected by this change. Bucket CORS, public-read/CDN policy and lifecycle policy must be configured in the chosen provider. Public demos use stable URLs and `Cache-Control: public, max-age=31536000, immutable`; do not use short-lived signed URLs.

Stable object identities are:

- system: `exercises/standard/<standard_key>/demo.<ext>`
- future coach uploads: `exercises/coaches/<owner-id>/<exercise-id>/<asset>`

## Test batch

[`exercise-demo-media-test-batch.json`](exercise-demo-media-test-batch.json) contains the verified catalog keys for this 16-item batch. Its `media_asset_id` values deliberately remain `null` until real licensed assets are uploaded. `--validate-only` accepts this planning state, while `--dry-run` and `--confirm` reject pending IDs, preventing fake production attachments.

After an approved local file is available, validate without writing:

```sh
cd backend
go run ./cmd/exercise-demo-upload --standard-key push-up-standard --file /approved/demo.mp4 --duration-seconds 4 --dry-run
```

Then, only with production DB and object-storage variables supplied through the operator's secret environment:

```sh
go run ./cmd/exercise-demo-upload --standard-key push-up-standard --file /approved/demo.mp4 --duration-seconds 4 --confirm
```

The helper checks a real regular file, content/extension MIME agreement, the 5 MB limit, a trusted 1–6 second video duration, and a system-owned standard key. It uploads to the stable key, creates or updates a system-owned `media_assets` row, and attaches it in one DB transaction. It cleans up a newly created object when the DB transaction fails. An overwrite of an already-existing object cannot be rolled back atomically with PostgreSQL; use bucket versioning and retain source files for recovery.

After all approved assets have media IDs, replace the `null` values and run:

```sh
go run ./cmd/exercise-demo-media --file ../docs/exercise-demo-media-test-batch.json --validate-only
go run ./cmd/exercise-demo-media --file ../docs/exercise-demo-media-test-batch.json --dry-run
# confirm only after missing=0 and conflicts=0
go run ./cmd/exercise-demo-media --file ../docs/exercise-demo-media-test-batch.json --confirm
```

The mapping command reports `attach`, `replace`, `unchanged`, `missing`, and `conflicts`; confirm is blocked by missing exercises, invalid media, or conflicts.

## Acceptance checklist

Use 1–3 licensed assets before expanding the batch.

- Android Telegram WebView and iOS Telegram WebView: detail video is muted, loops inline, has no controls and does not enter fullscreen.
- REP workout: demo loops while “Готово” remains usable. TIMED workout: prepare countdown and timer start normally. Voice Coach output and session saves are unchanged.
- Reject autoplay or simulate a media error: the workout continues and a poster/controlled fallback appears.
- Enable reduced motion: no video autoplay; poster/static preview appears.
- Simulate a slow network: countdown, set controls and session persistence continue independently of media loading.
- Open the 247-item exercise catalog and inspect Network: zero MP4/WebM requests are expected because catalog cards do not mount video elements. Search and all filters remain responsive.
- Coach-owned exercise: upload/select, save, reopen, replace and remove. Replacing/removing a relation must not delete the media asset.
- Coach cannot edit a system exercise or delete system media referenced by it; admin/super-admin performs system attachment.

Record device/OS/Telegram versions, asset keys, request counts and failures in the release report. Do not call the batch accepted based only on desktop automated tests.
