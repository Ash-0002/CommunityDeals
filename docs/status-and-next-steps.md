 # Status & next steps

A snapshot of what's built across the three apps, what's rough (design), and
what's clearly still missing. Update this as things change — it's meant to
stay current, not be a one-time writeup.

## What exists today

**Backend (Go)** — `backend/`
- Domains: users, communities, campaigns (`internal/domain`, `internal/dto`).
- Auth: phone + OTP flow, JWT access/refresh, dev-mode fixed OTP (`internal/service/auth_service.go`).
- Handlers/router for auth, campaigns, communities (`internal/handler`).
- Postgres migrations for users, communities, campaigns tables.
- Some test coverage (`campaign_repository_test.go`, `campaign_service_test.go`).
- No service layer for users beyond auth (no `user_service.go`) — profile
  reads/writes currently go straight through the repository from the handler.

**Web (Next.js)** — `web/`
- Pages: landing, login (OTP), dashboard, campaigns list + detail, public
  campaign page (`/c/[slug]`), profile, **create-campaign** (`/campaigns/new`).
- Shared UI: campaign card, campaign detail body, status chip, progress bar,
  theme toggle, sidebar layout, plus the same **visual redesign as mobile**
  (`components/campaign/campaign-visuals.tsx`) — hero image banner,
  overlapping avatar stack with an animated "unlocked" pill, and a starburst
  "% OFF" discount badge. Campaign detail also fetches real joined
  participants and has a working share button (native share sheet, falls
  back to clipboard).
- Dashboard now differs from the Campaigns catalog: it shows a "today's
  spotlight" deal (closest to unlocking, not yet joined) plus a stats row
  (deals joined, amount saved, communities) instead of repeating the list.
- Auth-aware middleware protecting `/dashboard`, `/campaigns`, `/profile`.
- Fixed two build-blocking bugs found this pass: the project's `.eslintrc.json`
  referenced `@typescript-eslint/no-unused-vars` without the plugin installed
  (failed `npm run build` for every file); and `/auth/login` used
  `useSearchParams()` without a Suspense boundary (failed static prerendering).
  `npm run build` is now clean.

**Mobile (Flutter)** — `mobile/`
- Same core flows as web: OTP auth, dashboard, campaigns list/detail, public
  campaign screen, profile.
- `dio`-based API client with 401 refresh handling, secure token storage,
  `go_router` with auth redirects.
- Platform folders (`android/`, `ios/`, `web/`) generated but not committed —
  see [`running-the-apps.md`](./running-the-apps.md).
- **Create-campaign flow** (`screens/create_campaign_screen.dart`) with an
  inline dynamic-pricing tier editor — add/remove tiers, price drops as the
  join count rises. Posts to `POST /campaigns`.
- **Join / cancel flow** — joining is unchanged; leaving now confirms first
  ("Cancel your spot?") since it forfeits the user's locked-in price.
  Backend `POST /campaigns/:id/leave` was already there.
- **Share campaign link** — a share icon on the detail screen opens the
  native share sheet (`share_plus`) with the campaign's `share_url`, which is
  now built from a configurable `APP_BASE_URL` instead of a hardcoded domain.
- **User settings** (`screens/settings_screen.dart`) — edit name/email
  (`PATCH /users/profile`), local notification toggles. Reachable from the
  Profile tab.
- **Visual redesign** matching the reference screenshots: a hero image
  banner, an overlapping avatar stack with a "N people joined" pill and a
  confetti "Group booking unlocked" animation once the minimum is hit, and a
  starburst "% OFF" discount badge with strikethrough pricing
  (`widgets/campaign_visuals.dart`).
- Campaign detail now shows real joined participants (name + avatar) via a
  new enriched `GET /campaigns/:id/participants` response.
- **Fixed a real login bug:** `lib/config/env.dart`'s default backend URL was
  hardcoded to `10.0.2.2:8080` (the Android-emulator-only alias for the host
  machine) regardless of platform. Running `flutter run` on Chrome, macOS, or
  the iOS simulator without an explicit `--dart-define` meant every API call —
  starting with "Send OTP" on the login screen — silently failed. It now
  picks `localhost:8080` unless it detects it's actually running on the
  Android emulator.
- Dashboard now differs from the Campaigns tab: a "today's spotlight" deal
  (closest to unlocking, not yet joined, rendered with the full hero/avatar/
  discount treatment) plus a stats row (deals joined, amount saved,
  communities), instead of repeating the same campaign list.

## Design — known rough edges

Mobile's campaign screens got a visual pass (hero images, avatar stacks,
discount badges — see above); web has not, so the two now look noticeably
different. Specific things worth tackling:

- **No shared design system across web and mobile.** Web has a Tailwind
  color/shadow scale (`web/tailwind.config.ts`); mobile has its own theme
  tokens (`mobile/lib/config/theme.dart`). They aren't derived from the same
  source, so colors/spacing/type will drift between platforms unless kept in
  sync by hand.
- **Inconsistent component polish.** Cards, status chips, and progress bars
  exist on both platforms but haven't been reviewed side-by-side for visual
  consistency (spacing, empty states, loading states).
- **No dark mode parity check.** Web has a theme toggle; unclear whether
  mobile's theme has been tested equally in dark mode.
- **No design references/mockups checked into the repo.** Decisions are
  currently being made ad hoc in code. Worth deciding whether to introduce a
  lightweight design source of truth (even a Figma file or a shared token
  file) before building more screens.

## Missing / next steps

Roughly in order of what likely unblocks the most:

1. **Community creation.** Campaign creation now exists on both web and
   mobile, but there's still no flow to *create a community* — the
   create-campaign screens assume the user already belongs to one.
2. **Web has no settings-equivalent notification toggles.** Mobile's settings
   screen has (local-only) notification toggles; web's `/profile` page only
   covers name/email + logout. Minor gap, low priority.
3. **User service layer (backend).** Profile get/update logic still lives
   outside a dedicated service — add `internal/service/user_service.go` to
   match the pattern used for campaigns/communities.
4. **Push/notification story.** The mobile settings screen has notification
   toggles, but they're local-only UI state — no backend notification
   service or WhatsApp/push integration behind them yet.
5. **Real SMS provider wiring.** `SMS_PROVIDER=twilio` is configured in
   `.env.example` but `TWILIO_*` values are blank — OTP only works in
   `OTP_DEV_MODE` right now.
6. **Mobile CI / build not set up.** No committed CI config for running
   `flutter analyze`/`flutter test` on PRs (backend and web don't appear to
   have CI either — worth deciding if/when to add).
7. **Deep linking not finished.** `communitydeals://c/<slug>` is documented
   in `mobile/README.md` but the intent-filter / associated-domain config
   depends on the (uncommitted) generated `android/`/`ios/` folders, so it
   needs to be redone by whoever runs `flutter create` next.
8. **No end-to-end test across all three apps** (e.g. sign up on web, see it
   reflected on mobile) — only backend has automated tests today.
9. **Two pre-existing broken test files.** `internal/service/campaign_service_test.go`
   and `internal/repository/campaign_repository_test.go` reference types/fields
   that no longer exist in the domain model (`repository.ErrCommunityNotFound`,
   `domain.ParticipantStatusActive`, `tier.Label`, etc.) — they fail to even
   compile (`go vet ./...` catches it). Not something introduced by the recent
   work; needs a pass to reconcile the tests with the current domain model.

## How to use this doc

When you finish something on this list, delete it (or move to a "done"
section) rather than leaving it stale. When you notice a new gap, add it here
instead of just remembering it.
