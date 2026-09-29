import 'package:flutter/foundation.dart'
    show kIsWeb, TargetPlatform, defaultTargetPlatform;

/// Runtime configuration.
///
/// Override at build/run time with:
///   flutter run --dart-define=API_BASE_URL=https://api.example.com
class Env {
  Env._();

  static const String _override = String.fromEnvironment('API_BASE_URL');

  /// Base URL of the Go backend. Defaults to the local dev server, picking
  /// the right host for the platform actually running the app:
  ///   - Chrome / macOS / iOS simulator → `localhost` reaches the host directly.
  ///   - Android emulator → `localhost` refers to the emulator itself, so
  ///     `10.0.2.2` is used to reach the host machine instead.
  /// A physical device on the same Wi-Fi still needs an explicit
  /// `--dart-define=API_BASE_URL=http://<your-lan-ip>:8080`.
  static String get apiBaseUrl {
    if (_override.isNotEmpty) return _override;
    if (!kIsWeb && defaultTargetPlatform == TargetPlatform.android) {
      return 'http://10.0.2.2:8080';
    }
    return 'http://localhost:8080';
  }

  /// API version prefix used by the backend router (`/api/v1`).
  static const String apiPrefix = '/api/v1';

  static String get apiUrl => '$apiBaseUrl$apiPrefix';
}
