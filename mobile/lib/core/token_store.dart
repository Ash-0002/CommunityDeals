import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// Persists the JWT access / refresh token pair in the platform keystore
/// (Keychain on iOS, EncryptedSharedPreferences on Android).
class TokenStore {
  TokenStore._();
  static final TokenStore instance = TokenStore._();

  static const _accessKey = 'cd.access_token';
  static const _refreshKey = 'cd.refresh_token';

  final FlutterSecureStorage _storage = const FlutterSecureStorage(
    aOptions: AndroidOptions(encryptedSharedPreferences: true),
  );

  String? _accessToken;
  String? get accessToken => _accessToken;

  Future<void> load() async {
    _accessToken = await _storage.read(key: _accessKey);
  }

  Future<String?> readRefreshToken() => _storage.read(key: _refreshKey);

  Future<void> save({
    required String accessToken,
    required String refreshToken,
  }) async {
    _accessToken = accessToken;
    await _storage.write(key: _accessKey, value: accessToken);
    await _storage.write(key: _refreshKey, value: refreshToken);
  }

  Future<void> clear() async {
    _accessToken = null;
    await _storage.delete(key: _accessKey);
    await _storage.delete(key: _refreshKey);
  }

  bool get hasToken => _accessToken != null && _accessToken!.isNotEmpty;
}
