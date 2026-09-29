import 'package:dio/dio.dart';

import '../config/env.dart';
import 'api_exception.dart';
import 'token_store.dart';

/// Thin wrapper around Dio that:
///  * injects the `Authorization: Bearer` header,
///  * unwraps the backend's `{ success, message, data, error }` envelope,
///  * transparently refreshes the access token once on a 401.
class ApiClient {
  ApiClient._() {
    _dio = Dio(
      BaseOptions(
        baseUrl: Env.apiUrl,
        connectTimeout: const Duration(seconds: 15),
        receiveTimeout: const Duration(seconds: 20),
        contentType: 'application/json',
      ),
    );

    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          final token = TokenStore.instance.accessToken;
          if (token != null && token.isNotEmpty) {
            options.headers['Authorization'] = 'Bearer $token';
          }
          handler.next(options);
        },
        onError: (error, handler) async {
          final response = error.response;
          final isAuthCall =
              error.requestOptions.path.contains('/auth/');
          if (response?.statusCode == 401 && !isAuthCall && !_isRefreshing) {
            final refreshed = await _refreshToken();
            if (refreshed) {
              try {
                final retry = await _retry(error.requestOptions);
                return handler.resolve(retry);
              } catch (_) {
                // fall through to the original error
              }
            }
          }
          handler.next(error);
        },
      ),
    );
  }

  static final ApiClient instance = ApiClient._();

  late final Dio _dio;
  bool _isRefreshing = false;

  /// Callback fired when the session can no longer be refreshed.
  void Function()? onSessionExpired;

  // ── Public verbs ──────────────────────────────────────────────────────────

  Future<dynamic> get(String path, {Map<String, dynamic>? query}) =>
      _send(() => _dio.get(path, queryParameters: query));

  Future<dynamic> post(String path, {Object? body}) =>
      _send(() => _dio.post(path, data: body));

  Future<dynamic> patch(String path, {Object? body}) =>
      _send(() => _dio.patch(path, data: body));

  // ── Internals ─────────────────────────────────────────────────────────────

  Future<dynamic> _send(Future<Response<dynamic>> Function() run) async {
    try {
      final res = await run();
      return _unwrap(res.data);
    } on DioException catch (e) {
      throw _toApiException(e);
    }
  }

  /// Returns the `data` field of the envelope, or the raw body if the
  /// response is not enveloped.
  dynamic _unwrap(dynamic body) {
    if (body is Map<String, dynamic> && body.containsKey('success')) {
      if (body['success'] == true) return body['data'];
      final err = body['error'];
      throw ApiException(
        (err is Map ? err['message'] : null)?.toString() ??
            body['message']?.toString() ??
            'Request failed',
        code: err is Map ? err['code']?.toString() : null,
      );
    }
    return body;
  }

  ApiException _toApiException(DioException e) {
    final data = e.response?.data;
    String message = 'Network error. Please try again.';
    String? code;
    if (data is Map) {
      final err = data['error'];
      message = (err is Map ? err['message'] : null)?.toString() ??
          data['message']?.toString() ??
          message;
      code = err is Map ? err['code']?.toString() : null;
    } else if (e.type == DioExceptionType.connectionError ||
        e.type == DioExceptionType.connectionTimeout) {
      message = 'Cannot reach the server. Check your connection.';
    }
    return ApiException(
      message,
      code: code,
      statusCode: e.response?.statusCode,
    );
  }

  Future<bool> _refreshToken() async {
    _isRefreshing = true;
    try {
      final refresh = await TokenStore.instance.readRefreshToken();
      if (refresh == null || refresh.isEmpty) {
        _expire();
        return false;
      }
      final res = await _dio.post(
        '/auth/refresh-token',
        data: {'refresh_token': refresh},
      );
      final data = _unwrap(res.data) as Map<String, dynamic>;
      await TokenStore.instance.save(
        accessToken: data['access_token'] as String,
        refreshToken: (data['refresh_token'] ?? refresh) as String,
      );
      return true;
    } catch (_) {
      _expire();
      return false;
    } finally {
      _isRefreshing = false;
    }
  }

  void _expire() {
    TokenStore.instance.clear();
    onSessionExpired?.call();
  }

  Future<Response<dynamic>> _retry(RequestOptions options) {
    return _dio.request(
      options.path,
      data: options.data,
      queryParameters: options.queryParameters,
      options: Options(method: options.method, headers: options.headers),
    );
  }
}
