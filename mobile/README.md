# CommunityDeals — Mobile (Flutter)

Flutter client for the CommunityDeals group-buying / service-booking platform.
Replaces the earlier Expo / React Native app.

## Stack

| Concern        | Choice                          |
| -------------- | ------------------------------- |
| Framework      | Flutter (Material 3)            |
| Routing        | `go_router`                     |
| State          | `provider` + `ChangeNotifier`   |
| HTTP           | `dio` (envelope unwrap + 401 refresh) |
| Token storage  | `flutter_secure_storage` (Keychain / EncryptedSharedPreferences) |
| Formatting     | `intl` (INR currency, dates)    |

## First-time setup

The `android/`, `ios/`, `web/`, `linux/`, `macos/`, `windows/` platform folders
are **not** committed. Generate them once (this keeps `lib/`, `pubspec.yaml`,
`test/` untouched):

```bash
cd mobile
flutter create --project-name community_deals --platforms=android,ios,web .
flutter pub get
```

If `flutter create` rewrites `pubspec.yaml`, restore it with `git checkout pubspec.yaml`
and re-run `flutter pub get`.

## Run

The backend host is picked automatically per platform (see `lib/config/env.dart`):
Chrome/macOS/iOS simulator use `localhost:8080`, the Android emulator uses
`10.0.2.2:8080` (its alias for the host machine). Just run:

```bash
flutter run
```

Only override it for a physical device on the same Wi-Fi (`localhost` and
`10.0.2.2` don't reach your laptop from a real phone):

```bash
flutter run --dart-define=API_BASE_URL=http://192.168.1.20:8080
```

## Layout

```
lib/
  config/      env (API base URL), theme tokens
  core/        api_client, token_store, api_exception, format helpers
  models/      user, community, pricing_tier, campaign
  services/    auth / campaign / community — one method per backend endpoint
  state/       auth_controller (session source of truth, drives router redirects)
  screens/     splash, auth/{phone,otp}, home_shell, dashboard,
               campaigns, campaign_detail, public_campaign, profile
  widgets/     common (loading/error/progress/status), campaign_card,
               campaign_detail_body
  router.dart  go_router config + auth redirect
  main.dart    providers + MaterialApp.router
```

## Backend endpoints used

- `POST /api/v1/auth/send-otp`, `POST /api/v1/auth/verify-otp`,
  `POST /api/v1/auth/refresh-token`, `POST /api/v1/auth/logout`
- `GET/PATCH /api/v1/users/profile`, `GET /api/v1/users/communities`
- `GET /api/v1/campaigns`, `GET /api/v1/campaigns/:id`,
  `POST /api/v1/campaigns/:id/{join,leave}`
- `GET /c/:slug` (public, unauthenticated)
- `GET /api/v1/communities`, `POST /api/v1/communities/:id/join`

## Deep links

Scheme `communitydeals://` — `communitydeals://c/<slug>` opens the public
campaign page. Configure the intent-filter / associated domain after running
`flutter create` (see `android/app/src/main/AndroidManifest.xml` and
`ios/Runner/Info.plist`).
