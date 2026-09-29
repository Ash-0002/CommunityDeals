import 'pricing_tier.dart';

/// Lightweight campaign row used by list endpoints (`GET /campaigns`).
class CampaignListItem {
  const CampaignListItem({
    required this.id,
    required this.title,
    required this.serviceName,
    required this.status,
    required this.participantCount,
    required this.minParticipants,
    required this.currentPrice,
    this.imageUrl,
    this.endDate,
    this.serviceDate,
    this.slug,
    this.isJoined = false,
    this.firstTierPrice = 0,
  });

  final String id;
  final String title;
  final String serviceName;
  final String status;
  final int participantCount;
  final int minParticipants;
  final int currentPrice; // paise
  final String? imageUrl;
  final String? endDate;
  final String? serviceDate;
  final String? slug;
  final bool isJoined;
  final int firstTierPrice; // paise — un-discounted tier 1 price

  int get amountSaved => (firstTierPrice - currentPrice).clamp(0, firstTierPrice);

  double get progress =>
      minParticipants == 0 ? 0 : (participantCount / minParticipants).clamp(0, 1);

  factory CampaignListItem.fromJson(Map<String, dynamic> json) =>
      CampaignListItem(
        id: json['id'] as String? ?? '',
        title: json['title'] as String? ?? '',
        serviceName: json['service_name'] as String? ?? '',
        status: json['status'] as String? ?? 'DRAFT',
        participantCount: (json['participant_count'] as num?)?.toInt() ?? 0,
        minParticipants: (json['min_participants'] as num?)?.toInt() ?? 0,
        currentPrice: (json['current_price'] as num?)?.toInt() ?? 0,
        imageUrl: json['image_url'] as String?,
        endDate: json['end_date'] as String?,
        serviceDate: json['service_date'] as String?,
        slug: json['slug'] as String?,
        isJoined: json['is_joined'] as bool? ?? false,
        firstTierPrice: (json['first_tier_price'] as num?)?.toInt() ?? 0,
      );
}

/// Full campaign detail (`GET /campaigns/:id` and `GET /c/:slug`).
class Campaign {
  const Campaign({
    required this.id,
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
    required this.status,
    required this.participantCount,
    required this.slug,
    required this.shareUrl,
    required this.pricingTiers,
    required this.currentPrice,
    this.nextTier,
    required this.participantsToNextTier,
    required this.isJoined,
  });

  final String id;
  final String communityId;
  final String serviceName;
  final String title;
  final String description;
  final String? imageUrl;
  final int minParticipants;
  final int maxParticipants;
  final String serviceDate;
  final String startDate;
  final String endDate;
  final String status;
  final int participantCount;
  final String slug;
  final String shareUrl;
  final List<PricingTier> pricingTiers;
  final int currentPrice; // paise
  final PricingTier? nextTier;
  final int participantsToNextTier;
  final bool isJoined;

  double get progress => minParticipants == 0
      ? 0
      : (participantCount / minParticipants).clamp(0, 1);

  bool get minimumReached => participantCount >= minParticipants;

  factory Campaign.fromJson(Map<String, dynamic> json) => Campaign(
        id: json['id'] as String? ?? '',
        communityId: json['community_id'] as String? ?? '',
        serviceName: json['service_name'] as String? ?? '',
        title: json['title'] as String? ?? '',
        description: json['description'] as String? ?? '',
        imageUrl: json['image_url'] as String?,
        minParticipants: (json['min_participants'] as num?)?.toInt() ?? 0,
        maxParticipants: (json['max_participants'] as num?)?.toInt() ?? 0,
        serviceDate: json['service_date'] as String? ?? '',
        startDate: json['start_date'] as String? ?? '',
        endDate: json['end_date'] as String? ?? '',
        status: json['status'] as String? ?? 'DRAFT',
        participantCount: (json['participant_count'] as num?)?.toInt() ?? 0,
        slug: json['slug'] as String? ?? '',
        shareUrl: json['share_url'] as String? ?? '',
        pricingTiers: (json['pricing_tiers'] as List<dynamic>? ?? [])
            .map((e) => PricingTier.fromJson(e as Map<String, dynamic>))
            .toList(),
        currentPrice: (json['current_price'] as num?)?.toInt() ?? 0,
        nextTier: json['next_tier'] == null
            ? null
            : PricingTier.fromJson(json['next_tier'] as Map<String, dynamic>),
        participantsToNextTier:
            (json['participants_to_next_tier'] as num?)?.toInt() ?? 0,
        isJoined: json['is_joined'] as bool? ?? false,
      );
}
