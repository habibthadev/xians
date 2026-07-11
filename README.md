# xians

Share files instantly — no sign-up, no fuss. Drag, drop, and share files at full quality with expiring links, password protection, and download limits.

Built by [habibthadev](https://habibthadev.tech).

## Features

- **Instant file sharing** — upload files and get a shareable link in seconds. No accounts, no sign-up.
- **Password protection** — optionally protect galleries with a password. Recipients enter the password to access files.
- **Expiring links** — choose from preset TTLs (24h, 7d, 30d, never) or set a download count limit.
- **Direct-to-cloud uploads** — files stream directly to Cloudflare R2 via presigned URLs, zero server buffering.
- **View vs download** — image/video previews use a streaming proxy (no download count), actual downloads use presigned URLs (counts decremented).
- **Gallery grid** — responsive file grid with hover preview overlay and fullscreen lightbox.
- **Bulk download** — download all files as a single ZIP archive.
- **OG image generation** — dynamic social preview images with Satori (JSX→SVG) + resvg (SVG→PNG), matching the app's Geist/Anton font stack and theme colors.
- **Animated UI** — GSAP-powered micro-animations (staggered entries, hover states, button clicks, lightbox transitions) with proper Svelte 5 `$effect`/`gsap.context()` lifecycle management.
- **Responsive** — mobile-first layout that works on any device.

## Architecture

```
xians/
├── apps/
│   ├── api/          # Go REST API (chi router, GORM, Cloudflare R2)
│   └── web/          # Svelte 5 + SvelteKit frontend (adapter-vercel)
├── packages/
│   └── shared-types/ # Shared TypeScript types (Zod schemas)
└── pnpm-workspace.yaml
```

## Tech Stack

### Frontend (`apps/web`)
- **Framework**: Svelte 5 (runes: `$state`, `$derived`, `$effect`, `$props`)
- **Meta-framework**: SvelteKit 2 (`adapter-vercel`)
- **Styling**: Tailwind CSS v4 with custom design tokens
- **Animation**: GSAP 3.15 (`gsap.context()` + `ctx.revert()` for Svelte 5 cleanup)
- **Fonts**: Geist (body), Anton (display), Poppins (brand UI)
- **Icons**: Iconify (`@iconify/svelte`)
- **Validation**: Zod
- **OG Images**: Satori (JSX→SVG) + `@resvg/resvg-js` (SVG→PNG)

### Backend (`apps/api`)
- **Language**: Go 1.26
- **Router**: chi (go-chi/chi/v5)
- **Database**: PostgreSQL + GORM
- **Storage**: Cloudflare R2 (S3-compatible)
- **Auth**: JWT (golang-jwt/jwt/v5) for gallery password access
- **CORS**: Custom middleware with origin allowlist

### Infrastructure
- **Frontend host**: Vercel (Node.js 22.x, serverless functions)
- **Backend host**: Fly.io (Docker, single region)
- **Database**: PostgreSQL (managed, e.g. Neon, Fly Postgres)
- **Storage**: Cloudflare R2 (S3-compatible, no egress fees)
- **Package manager**: pnpm workspaces

## Getting Started

### Prerequisites

- Node.js >= 20
- pnpm >= 9
- Go >= 1.26
- PostgreSQL instance
- Cloudflare R2 bucket

### Installation

```bash
# Clone the repository
git clone https://github.com/habibthadev/xians.git
cd xians

# Install dependencies
pnpm install

# Set up environment variables
cp apps/api/.env.example apps/api/.env
cp apps/web/.env.example apps/web/.env
```

### Environment Variables

**API** (`apps/api/.env`):

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SIGNING_SECRET` | At least 32 characters, used for gallery access JWTs |
| `R2_ACCOUNT_ID` | Cloudflare R2 account ID |
| `R2_ACCESS_KEY_ID` | Cloudflare R2 access key |
| `R2_SECRET_ACCESS_KEY` | Cloudflare R2 secret key |
| `R2_BUCKET_NAME` | R2 bucket name |
| `R2_PUBLIC_HOSTNAME` | R2 public bucket hostname (for presigned URLs) |
| `ALLOWED_ORIGIN` | CORS origin (e.g. `http://localhost:5173` for dev) |
| `PORT` | Server port (default: 8080) |

**Web** (`apps/web/.env`):

| Variable | Description | Scope |
|---|---|---|
| `PUBLIC_API_URL` | API URL exposed to browser | Client (public) |
| `API_URL` | API URL for SSR fetches | Server only |

### Development

```bash
# Start both API and web with hot-reload
pnpm dev

# Or start individually:
pnpm dev:web   # SvelteKit at localhost:5173
pnpm dev:api   # Go API at localhost:8080
```

### Build

```bash
pnpm build:web   # SvelteKit production build
pnpm build:api   # Go binary build
```

## Project Structure

### `apps/web/src/`

```
src/
├── routes/
│   ├── +layout.svelte        # Global layout, SEO meta tags, OG image
│   ├── +page.svelte          # Landing page (hero, features, how-it-works, FAQ)
│   ├── og/+server.ts         # Dynamic OG image generation (Satori + resvg)
│   ├── upload/+page.svelte   # Upload flow (drop → queue → settings → success)
│   ├── g/
│   │   └── [slug]/
│   │       ├── +page.server.ts  # SSR: gallery data fetch
│   │       └── +page.svelte     # Gallery: password gate → grid → lightbox
│   └── ...
├── lib/
│   ├── components/
│   │   ├── gallery/          # PasswordGate, GalleryGrid, MediaCard, Lightbox
│   │   ├── upload/           # Dropzone, FileQueue, UploadSettings, UploadSuccess
│   │   └── ui/               # Button, Input, Toast
│   └── api/
│       ├── env.ts            # API_URL from $env/dynamic/public
│       └── client.ts         # Fetch wrapper
└── app.css                   # Tailwind v4 theme tokens
```

### `apps/api/internal/`

```
internal/
├── cmd/
│   ├── server/main.go        # Entry point, routes, middleware
│   └── reaper/main.go        # Cron job for expired gallery cleanup
├── config/config.go          # Environment configuration
├── database/database.go      # GORM connection
├── handlers/
│   ├── batch.go              # Gallery CRUD handlers
│   ├── auth.go               # JWT cookie management, gallery access check
│   └── ...
├── middleware/
│   ├── cors.go               # CORS with origin allowlist
│   └── ...
├── models/
│   ├── batch.go              # GORM models (Batch, File)
│   └── repository.go         # Database operations
├── services/upload.go        # Business logic
└── storage/r2.go             # Cloudflare R2 operations
```

## API Endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/batches/init` | Initialize a new upload batch |
| `PUT` | `/api/batches/{id}/files/{fileId}` | Upload file (presigned R2 URL) |
| `POST` | `/api/batches/{id}/complete` | Finalize batch, create gallery |
| `GET` | `/api/g/{slug}` | Get gallery metadata + files |
| `POST` | `/api/g/{slug}/unlock` | Unlock password-protected gallery |
| `GET` | `/api/g/{slug}/files/{fileId}` | Download file (counts toward limit) |
| `GET` | `/api/g/{slug}/files/{fileId}/view` | View file (streaming proxy, no count) |
| `GET` | `/api/g/{slug}/zip` | Download all files as ZIP |
| `POST` | `/api/g/{slug}/report` | Report a gallery |
| `DELETE` | `/api/g/{slug}` | Delete gallery (requires owner token) |
| `GET` | `/api/health` | Health check |

## Deployment

### API → Fly.io

```bash
cd apps/api

# Launch with docker build
flyctl launch --no-deploy

# Set secrets (never in fly.toml)
flyctl secrets set DATABASE_URL="..." R2_ACCOUNT_ID="..." R2_ACCESS_KEY_ID="..." \
  R2_SECRET_ACCESS_KEY="..." R2_BUCKET_NAME="..." R2_PUBLIC_HOSTNAME="..." \
  JWT_SIGNING_SECRET="..."

# Deploy
flyctl deploy
```

Non-secret config like `ALLOWED_ORIGIN` and rate limits are in `fly.toml`.

### Web → Vercel

Set **Root Directory** to `apps/web` in Vercel project settings, or deploy with:

```bash
npx vercel --root apps/web
```

Set environment variables in Vercel dashboard:

| Variable | Value |
|---|---|
| `PUBLIC_API_URL` | `https://xians-api.fly.dev` |
| `API_URL` | `https://xians-api.fly.dev` |

## Attribution

Built by [habibthadev](https://habibthadev.tech).
