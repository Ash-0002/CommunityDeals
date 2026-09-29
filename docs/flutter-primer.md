# Flutter primer (for this codebase)

You've worked with the web (Next.js) and backend (Go) sides already. This is
a quick map of Flutter/Dart concepts to what you already know, plus where
things live in `mobile/`.

## Mental model

- **Dart** is the language (similar territory to TypeScript — typed, null-safe,
  async/await). **Flutter** is the UI framework, roughly playing the role
  React plays for the web app.
- There's no HTML/CSS. Everything — layout, styling, text — is a Dart object
  called a **Widget**. You build UI by nesting widgets, the same way you'd
  nest JSX elements.
- Instead of a browser, Flutter compiles to a native app (or a Chrome-hosted
  web build) and manages its own rendering — no DOM.

## Widgets, in React terms

| React concept              | Flutter equivalent                                   |
| --------------------------- | ------------------------------------------------------ |
| Function component          | `StatelessWidget` (no internal state) or `StatefulWidget` (has state) |
| `useState`                  | `setState(() { ... })` inside a `StatefulWidget`, or a `provider` `ChangeNotifier` (used in this app — see below) |
| `useEffect`                 | `initState()` / `dispose()` lifecycle methods           |
| props                       | constructor parameters (usually `final` fields)         |
| CSS                         | widget properties like `padding:`, `color:`, `TextStyle` |
| React Router                | `go_router` (this app uses it — see `lib/router.dart`)   |
| `fetch`/`axios`             | `dio` (this app's HTTP client — see `lib/core/api_client.dart`) |
| Context API / Redux/Zustand | `provider` package + `ChangeNotifier` (this app's state approach — see `lib/state/auth_controller.dart`) |

## State management used here: `provider`

The app uses `provider` with `ChangeNotifier` classes (see
`lib/state/auth_controller.dart`). Pattern:

1. A class extends `ChangeNotifier` and holds some state (e.g. logged-in user).
2. It calls `notifyListeners()` whenever that state changes.
3. Widgets read it with `context.watch<AuthController>()` (rebuilds on change)
   or `context.read<AuthController>()` (one-off read, e.g. inside a button's
   `onPressed`).

This is conceptually similar to a small Redux/Zustand store, just built into
one class per concern.

## Project layout (`mobile/lib/`)

```
config/      env (API base URL), theme tokens        — like web/lib/config or .env
core/        api_client, token_store, api_exception   — like web/lib/api client + axios interceptors
models/      user, community, pricing_tier, campaign  — like web/lib/types.ts
services/    one class per backend resource (auth, campaign, community) — like web/lib/api/*.ts
state/       auth_controller (ChangeNotifier)         — like a React context/store
screens/     one file per full page                   — like web/app/**/page.tsx
widgets/     reusable pieces (cards, buttons, states)  — like web/components/
router.dart  all routes + auth redirect logic          — like web/middleware.ts + route tree
main.dart    app entry point, wires providers + router — like web/app/layout.tsx
```

## Commands you'll actually use

```bash
flutter pub get        # install dependencies (like npm install)
flutter run             # run on a connected device/emulator/simulator (like npm run dev)
flutter run -d chrome    # run in a Chrome tab instead of a device
flutter test             # run tests (like npm test)
flutter analyze          # static analysis / lint (like npm run lint)
flutter devices          # list available emulators/simulators/devices
r                        # (while `flutter run` is active) hot reload — like Fast Refresh
R                        # (while `flutter run` is active) hot restart (full reset of state)
```

## Emulator/simulator setup (one-time, OS-specific)

- **Android:** install Android Studio, open its Device Manager, create a
  virtual device, then `flutter emulators --launch <name>` or launch it from
  Android Studio before `flutter run`.
- **iOS (Mac only):** install Xcode, then `open -a Simulator` before
  `flutter run`.
- You can skip both and just run `flutter run -d chrome` to test in a browser
  while getting oriented, though some native-only features (secure storage,
  deep links) behave differently there.

## Gotchas specific to this repo

- `android/`, `ios/`, `web/`, `macos/` platform folders are **not** committed
  — run `flutter create --project-name community_deals --platforms=android,ios,web .`
  once after cloning (see [`running-the-apps.md`](./running-the-apps.md)).
- On the **Android emulator**, `localhost` refers to the emulator itself, not
  your host machine — that's why `lib/config/env.dart` defaults to
  `10.0.2.2:8080` to reach a backend running on your laptop.
- Override the backend URL per-run with
  `flutter run --dart-define=API_BASE_URL=http://...` rather than editing code.

## Where to learn more

- [flutter.dev/docs](https://flutter.dev/docs) — official docs, has a
  "Flutter for React Native devs" style guide.
- `dart.dev/language` — language tour if Dart syntax itself is unfamiliar.
