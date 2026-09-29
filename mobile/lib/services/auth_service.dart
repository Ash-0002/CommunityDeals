import '../core/api_client.dart';
import '../core/token_store.dart';
import '../models/user.dart';

class OtpChallenge {
  const OtpChallenge({required this.phone, required this.expiresInSeconds});
  final String phone;
  final int expiresInSeconds;
}

class AuthService {
  AuthService({ApiClient? api}) : _api = api ?? ApiClient.instance;

  final ApiClient _api;

  Future<OtpChallenge> sendOtp(String phone) async {
    final data = await _api.post('/auth/send-otp', body: {'phone': phone})
        as Map<String, dynamic>;
    return OtpChallenge(
      phone: data['phone'] as String? ?? phone,
      expiresInSeconds: (data['expires_in_seconds'] as num?)?.toInt() ?? 300,
    );
  }

  /// Verifies the OTP, persists the token pair, and returns the user.
  Future<User> verifyOtp({required String phone, required String otp}) async {
    final data = await _api
            .post('/auth/verify-otp', body: {'phone': phone, 'otp': otp})
        as Map<String, dynamic>;

    await TokenStore.instance.save(
      accessToken: data['access_token'] as String,
      refreshToken: data['refresh_token'] as String,
    );
    return User.fromJson(data['user'] as Map<String, dynamic>);
  }

  Future<User> me() async {
    final data = await _api.get('/users/profile') as Map<String, dynamic>;
    return User.fromJson(data);
  }

  Future<User> updateProfile({String? name, String? email}) async {
    final body = <String, dynamic>{};
    if (name != null) body['name'] = name;
    if (email != null) body['email'] = email;
    final data =
        await _api.patch('/users/profile', body: body) as Map<String, dynamic>;
    return User.fromJson(data);
  }

  Future<void> logout() async {
    try {
      await _api.post('/auth/logout');
    } catch (_) {
      // best effort — clear locally regardless
    }
    await TokenStore.instance.clear();
  }
}
