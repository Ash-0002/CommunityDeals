class PricingTier {
  const PricingTier({
    required this.minCount,
    required this.maxCount,
    required this.price,
    this.tierOrder = 0,
  });

  final int minCount;
  final int maxCount; // 0 = unlimited / last tier
  final int price; // paise
  final int tierOrder;

  factory PricingTier.fromJson(Map<String, dynamic> json) => PricingTier(
        minCount: (json['min_count'] as num?)?.toInt() ?? 0,
        maxCount: (json['max_count'] as num?)?.toInt() ?? 0,
        price: (json['price'] as num?)?.toInt() ?? 0,
        tierOrder: (json['tier_order'] as num?)?.toInt() ?? 0,
      );
}
