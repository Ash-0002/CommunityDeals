import '../core/api_client.dart';
import '../models/campaign.dart';
import '../models/participant.dart';

/// One pricing bracket submitted when creating a campaign.
/// e.g. 10–24 people → ₹500 (price in paise).
class PricingTierInput {
  const PricingTierInput({
    required this.minCount,
    required this.maxCount,
    required this.price,
  });

  final int minCount;
  final int maxCount; // 0 = unlimited / last tier
  final int price; // paise

  Map<String, dynamic> toJson() => {
        'min_count': minCount,
        'max_count': maxCount,
        'price': price,
      };
}

class NewCampaign {
  const NewCampaign({
    required this.communityId,
    required this.serviceName,
    required this.title,
    required this.description,
    this.imageUrl,
    required this.minParticipants,
    required this.maxParticipants,
    required this.serviceDate,
    required this.startDate,
    required this.endDate,
    required this.pricingTiers,
  });

  final String communityId;
  final String serviceName;
  final String title;
  final String description;
  final String? imageUrl;
  final int minParticipants;
  final int maxParticipants;
  final DateTime serviceDate;
  final DateTime startDate;
  final DateTime endDate;
  final List<PricingTierInput> pricingTiers;

  Map<String, dynamic> toJson() => {
        'community_id': communityId,
        'service_name': serviceName,
        'title': title,
        'description': description,
        if (imageUrl != null && imageUrl!.isNotEmpty) 'image_url': imageUrl,
        'min_participants': minParticipants,
        'max_participants': maxParticipants,
        'service_date': serviceDate.toUtc().toIso8601String(),
        'start_date': startDate.toUtc().toIso8601String(),
        'end_date': endDate.toUtc().toIso8601String(),
        'pricing_tiers': pricingTiers.map((t) => t.toJson()).toList(),
      };
}

class CampaignPage {
  const CampaignPage({
    required this.campaigns,
    required this.total,
    required this.page,
    required this.hasMore,
  });

  final List<CampaignListItem> campaigns;
  final int total;
  final int page;
  final bool hasMore;
}

class JoinResult {
  const JoinResult({
    required this.participantCount,
    required this.priceLocked,
    required this.message,
  });

  final int participantCount;
  final int priceLocked;
  final String message;
}

class CampaignService {
  CampaignService({ApiClient? api}) : _api = api ?? ApiClient.instance;

  final ApiClient _api;

  Future<CampaignPage> list({
    String? communityId,
    String? status,
    int page = 1,
    int limit = 20,
  }) async {
    final data = await _api.get('/campaigns', query: {
      if (communityId != null) 'community_id': communityId,
      if (status != null) 'status': status,
      'page': page,
      'limit': limit,
    }) as Map<String, dynamic>;

    return CampaignPage(
      campaigns: (data['campaigns'] as List<dynamic>? ?? [])
          .map((e) => CampaignListItem.fromJson(e as Map<String, dynamic>))
          .toList(),
      total: (data['total'] as num?)?.toInt() ?? 0,
      page: (data['page'] as num?)?.toInt() ?? page,
      hasMore: data['has_more'] as bool? ?? false,
    );
  }

  Future<Campaign> byId(String id) async {
    final data = await _api.get('/campaigns/$id') as Map<String, dynamic>;
    return Campaign.fromJson(data);
  }

  /// Public, unauthenticated campaign lookup used for shared WhatsApp links.
  Future<Campaign> bySlug(String slug) async {
    final data = await _api.get('/c/$slug') as Map<String, dynamic>;
    return Campaign.fromJson(data);
  }

  Future<JoinResult> join(String id) async {
    final data =
        await _api.post('/campaigns/$id/join') as Map<String, dynamic>;
    return JoinResult(
      participantCount: (data['participant_count'] as num?)?.toInt() ?? 0,
      priceLocked: (data['price_locked'] as num?)?.toInt() ?? 0,
      message: data['message'] as String? ?? 'Joined',
    );
  }

  Future<void> leave(String id) => _api.post('/campaigns/$id/leave');

  Future<Campaign> create(NewCampaign input) async {
    final data =
        await _api.post('/campaigns', body: input.toJson()) as Map<String, dynamic>;
    return Campaign.fromJson(data);
  }

  Future<List<Participant>> participants(String id, {int limit = 50}) async {
    final data = await _api
        .get('/campaigns/$id/participants', query: {'limit': limit}) as Map<String, dynamic>;
    return (data['participants'] as List<dynamic>? ?? [])
        .map((e) => Participant.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}
