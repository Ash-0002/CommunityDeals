import 'package:flutter/foundation.dart';

import '../core/api_client.dart';
import '../core/token_store.dart';
import '../models/user.dart';
import '../services/auth_service.dart';

enum AuthStatus { unknown, authenticated, unauthenticated }

/// Single source of truth for the session. Drives router redirects.
class AuthController extends ChangeNotifier {
  AuthController({AuthService? service})
      : _service = service ?? AuthService() {
    ApiClient.instance.onSessionExpired = _onSessionExpired;
  }

  final AuthService _service;

  AuthStatus _status = AuthStatus.unknown;
  AuthStatus get status => _status;

  User? _user;
  User? get user => _user;

  bool get isAuthenticated => _status == AuthStatus.authenticated;

  /// Called once on startup: restore token, then confirm it with the backend.
  Future<void> bootstrap() async {
    await TokenStore.instance.load();
    if (!TokenStore.instance.hasToken) {
      _set(AuthStatus.unauthenticated, null);
      return;
    }
    try {
      final me = await _service.me();
      _set(AuthStatus.authenticated, me);
    } catch (_) {
      await TokenStore.instance.clear();
      _set(AuthStatus.unauthenticated, null);
    }
  }

  Future<OtpChallenge> sendOtp(String phone) => _service.sendOtp(phone);

  Future<void> verifyOtp({required String phone, required String otp}) async {
    final user = await _service.verifyOtp(phone: phone, otp: otp);
    _set(AuthStatus.authenticated, user);
  }

  Future<void> refreshProfile() async {
    try {
      _set(AuthStatus.authenticated, await _service.me());
    } catch (_) {/* keep current */}
  }

  Future<void> updateProfile({String? name, String? email}) async {
    final updated = await _service.updateProfile(name: name, email: email);
    _set(AuthStatus.authenticated, updated);
  }

  void setUser(User user) => _set(AuthStatus.authenticated, user);

  Future<void> logout() async {
    await _service.logout();
    _set(AuthStatus.unauthenticated, null);
  }

  void _onSessionExpired() {
    _set(AuthStatus.unauthenticated, null);
  }

  void _set(AuthStatus status, User? user) {
    _status = status;
    _user = user;
    notifyListeners();
  }
}
