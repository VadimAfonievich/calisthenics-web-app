# Exercise demo media

Exercises have separate cover and demo semantics. `exercises.demo_media_id` references `media_assets`; APIs expose only public URL/type/MIME/poster metadata and use joins rather than per-row lookups. Missing or failed media renders a controlled placeholder. Catalog cards do not render videos, so listing 247 exercises creates no video requests.

Supported demo assets are H.264 MP4 (`video/mp4`), WebM (`video/webm`), GIF (`image/gif`), and JPEG/PNG/WebP static fallbacks. Attachment is limited to 5 MB; videos must declare a duration of 1–6 seconds. Recommended production encoding is 720×720 or portrait-safe, 24–30 fps, muted/no audio, 1–4 seconds, and loop-safe start/end.

`ExerciseDemoMedia` uses muted autoplay, loop and plays-inline for video. It displays a poster/static preview for `prefers-reduced-motion`, falls back cleanly after load errors, and never renders demos in compact catalog cards. Autoplay failure does not affect workout timers, voice coaching or action buttons.

Coach Studio selects an existing ready media asset. Coaches may attach assets to their own exercises; admin/super-admin may manage system exercises. The existing runtime has a Storage interface with `Upload`, `Delete`, and `PublicURL`, but the configured provider is currently unavailable. The MVP therefore supports already prepared external/data assets and shows a controlled message; it does not transcode or run FFmpeg. Object keys should use `exercises/<standard_key>/demo.<ext>` or `exercises/<exercise_uuid>/<asset>`, never a user filename alone.

Bulk assignment uses an ID-only manifest, never scraping or arbitrary remote URLs:

```sh
go run ./cmd/exercise-demo-media --file ../docs/exercise-demo-media.example.json --validate-only
go run ./cmd/exercise-demo-media --file ../docs/exercise-demo-media.json --dry-run
go run ./cmd/exercise-demo-media --file ../docs/exercise-demo-media.json --confirm
```

Only owned/original, licensed, purpose-generated, or license-compliant stock/CC media may be used. Do not copy media from fitness applications. Replacing a relation does not delete the old asset. Media Library reference checks include `demo_media_id`, preventing deletion while used. External object deletion is outside this foundation.
