/// Normalised error raised by [ApiClient] for any non-2xx response or
/// transport failure. `code` mirrors the backend's `error.code` field.
class ApiException implements Exception {
  ApiException(this.message, {this.code, this.statusCode});

  final String message;
  final String? code;
  final int? statusCode;

  bool get isUnauthorized => statusCode == 401;

  @override
  String toString() => 'ApiException($statusCode $code): $message';
}
