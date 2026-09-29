import '../core/api_client.dart';
import '../models/community.dart';

/// Outcome of a join request. The backend returns APPROVED immediately for
/// open communities, or PENDING when the community requires admin approval.
class JoinCommunityResult {
  const JoinCommunityResult({required this.status, required this.message});

  final String status;
  final String message;

  bool get isPending => status == 'PENDING';
}

class CommunityService {
  CommunityService({ApiClient? api}) : _api = api ?? ApiClient.instance;

  final ApiClient _api;

  List<Community> _parseList(dynamic data) {
    final raw = data is Map<String, dynamic>
        ? (data['communities'] as List<dynamic>? ?? const [])
        : (data as List<dynamic>? ?? const []);
    return raw
        .map((e) => Community.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<Community>> myCommunities() async {
    final data = await _api.get('/users/communities');
    return _parseList(data);
  }

  /// Location-based discovery — the backend filters on city / PIN code / type.
  Future<List<Community>> discover({
    String? city,
    String? pinCode,
    String? type,
    int page = 1,
    int limit = 20,
  }) async {
    final data = await _api.get('/communities', query: {
      if (city != null && city.isNotEmpty) 'city': city,
      if (pinCode != null && pinCode.isNotEmpty) 'pin_code': pinCode,
      if (type != null && type.isNotEmpty) 'type': type,
      'page': page,
      'limit': limit,
    });
    return _parseList(data);
  }

  Future<Community> byId(String id) async {
    final data = await _api.get('/communities/$id') as Map<String, dynamic>;
    return Community.fromJson(data);
  }

  Future<JoinCommunityResult> join(String id, {String? inviteCode}) async {
    final data = await _api.post('/communities/$id/join', body: {
      if (inviteCode != null && inviteCode.isNotEmpty) 'invite_code': inviteCode,
    }) as Map<String, dynamic>;
    return JoinCommunityResult(
      status: data['membership_status'] as String? ?? 'PENDING',
      message: data['message'] as String? ?? 'Join request sent',
    );
  }
}
