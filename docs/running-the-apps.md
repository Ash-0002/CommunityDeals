# Running all three apps

CommunityDeals has three pieces that normally all run at once during
development:

| App      | Path       | Tech                 | Default URL              |
| -------- | ---------- | -------------------- | ------------------------- |
| Backend  | `backend/` | Go + Postgres + Redis | http://localhost:8080     |
| Web      | `web/`     | Next.js               | http://localhost:3000     |
| Mobile   | `mobile/`  | Flutter               | emulator/simulator/device |

Start them in this order — web and mobile both call the backend API.

## 1. Backend (Go API)

**Prerequisites:** Go, Docker (for Postgres + Redis).

```bash
cd backend

# first time only: copy env template and fill in secrets if needed
cp .env.example .env

# start Postgres + Redis (and optionally the API) in Docker
make docker-up

# run pending DB migrations
make migrate-up

# run the API server locally (hot-reloads via `go run`, not the Docker `api` container)
make run
```

The server listens on `http://localhost:8080` (see `APP_PORT` in `.env`).
`APP_BASE_URL` (defaults to `http://localhost:3000`) is used to build the
shareable `/c/:slug` campaign links — point it at your web app's real origin
in production.

To stop the Docker services: `make docker-down` (add `-v`/`make docker-reset`
to wipe the database volume).

Useful commands: `make test`, `make lint`, `make migrate-down`.

### Demo data

To skip manually creating a community/campaigns every time, seed the database:

```bash
make seed
```

Safe to re-run. It creates a community ("Green Valley Society"), 25 demo
users, and 3 campaigns in different states (early-stage, near the unlock
threshold, and already unlocked) so every screen has something to show.
Log in on web or mobile with:

- **Phone:** `9810000001` (Rahul Sharma — community admin, can create deals)
- **OTP:** `111111` (fixed in dev — `OTP_DEV_MODE=true`)
- Any of `9810000002` … `9810000025` works too, for a regular member view.

## 2. Web (Next.js)

**Prerequisites:** Node.js 18+.

```bash
cd web
npm install
npm run dev
```

Opens at `http://localhost:3000`. It talks to the backend at
`http://localhost:8080` — check `web/lib` for the base URL / API client if you
need to point it elsewhere.

Other commands: `npm run build`, `npm run lint`, `npm run type-check`.

## 3. Mobile (Flutter)

**Prerequisites:** Flutter SDK (`>=3.22.0`). See
[`flutter-primer.md`](./flutter-primer.md) if this is your first time with Flutter.

```bash
cd mobile

# first time only — platform folders (android/, ios/, web/...) aren't
# committed to git, so generate them once:
flutter create --project-name community_deals --platforms=android,ios,web .
flutter pub get
```

> If `flutter create` overwrites `pubspec.yaml`, restore it with
> `git checkout pubspec.yaml` and run `flutter pub get` again.

Then run the app — the backend host is picked automatically per platform
(Chrome/macOS/iOS simulator → `localhost:8080`, Android emulator → `10.0.2.2:8080`):

```bash
flutter run

# only needed for a physical device on the same Wi-Fi — use your machine's LAN IP
flutter run --dart-define=API_BASE_URL=http://192.168.1.20:8080
```

## Running all three together

Open three terminal tabs:

```bash
# Terminal 1
cd backend && make docker-up && make migrate-up && make seed && make run

# Terminal 2
cd web && npm run dev

# Terminal 3
cd mobile && flutter run
```

Backend must be reachable before web/mobile auth flows (OTP login) will work.
In development, `OTP_DEV_MODE=true` in `backend/.env` means any OTP request
returns a fixed code (`111111`) instead of sending a real SMS.
