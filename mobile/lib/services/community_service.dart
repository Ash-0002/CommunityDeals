import '../core/api_client.dart';
import '../models/community.dart';

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

  Future<List<Community>> discover({
    String? city,
    String? type,
    int page = 1,
    int limit = 20,
  }) async {
    final data = await _api.get('/communities', query: {
      if (city != null) 'city': city,
      if (type != null) 'type': type,
      'page': page,
      'limit': limit,
    });
    return _parseList(data);
  }

  Future<Community> byId(String id) async {
    final data = await _api.get('/communities/$id') as Map<String, dynamic>;
    return Community.fromJson(data);
  }

  Future<String> join(String id, {String? inviteCode}) async {
    final data = await _api.post('/communities/$id/join', body: {
      if (inviteCode != null) 'invite_code': inviteCode,
    }) as Map<String, dynamic>;
    return data['membership_status'] as String? ?? 'PENDING';
  }
}
