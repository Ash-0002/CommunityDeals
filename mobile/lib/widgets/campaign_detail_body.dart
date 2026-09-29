import 'package:flutter/material.dart';
import 'package:share_plus/share_plus.dart';

import '../config/theme.dart';
import '../core/format.dart';
import '../models/campaign.dart';
import '../models/participant.dart';
import 'campaign_visuals.dart';
import 'common.dart';

/// Shared scrollable content for both the authenticated campaign detail
/// screen and the public (slug) campaign page.
class CampaignDetailBody extends StatelessWidget {
  const CampaignDetailBody({
    super.key,
    required this.campaign,
    this.participants = const [],
  });

  final Campaign campaign;
  final List<Participant> participants;

  @override
  Widget build(BuildContext context) {
    final muted = Theme.of(context).brightness == Brightness.dark
        ? AppColors.darkMuted
        : AppColors.lightMuted;

    final firstTierPrice =
        campaign.pricingTiers.isNotEmpty ? campaign.pricingTiers.first.price : campaign.currentPrice;
    final percentOff = firstTierPrice > campaign.currentPrice
        ? (((firstTierPrice - campaign.currentPrice) / firstTierPrice) * 100).round()
        : 0;

    return ListView(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 120),
      children: [
        CampaignHeroBanner(imageUrl: campaign.imageUrl, label: 'Group booking'),
        const SizedBox(height: 16),

        Row(
          children: [
            Expanded(
              child: Text(
                campaign.title,
                style: const TextStyle(
                    fontSize: 22, fontWeight: FontWeight.w900),
              ),
            ),
            const SizedBox(width: 8),
            StatusChip(campaign.status),
          ],
        ),
        const SizedBox(height: 4),
        Row(
          children: [
            Expanded(
              child: Text(campaign.serviceName, style: TextStyle(color: muted)),
            ),
            IconButton(
              tooltip: 'Share',
              onPressed: () => Share.share(
                'Join "${campaign.title}" on CommunityDeals — ${campaign.shareUrl}',
                subject: campaign.title,
              ),
              icon: const Icon(Icons.ios_share_rounded, size: 20),
              color: AppColors.primaryDark,
            ),
          ],
        ),
        const SizedBox(height: 16),

        JoinedAvatarsRow(
          participants: participants,
          totalJoined: campaign.participantCount,
          unlocked: campaign.minimumReached,
        ),
        const SizedBox(height: 16),

        DiscountBadge(
          originalPrice: firstTierPrice,
          currentPrice: campaign.currentPrice,
          percentOff: percentOff,
        ),
        const SizedBox(height: 16),

        // Progress card
        Card(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text('Flats joined',
                        style: TextStyle(
                            color: muted,
                            fontSize: 12,
                            fontWeight: FontWeight.w600)),
                    Text('${campaign.participantCount} / ${campaign.minParticipants}',
                        style: const TextStyle(
                            fontWeight: FontWeight.w800, fontSize: 15)),
                  ],
                ),
                const SizedBox(height: 10),
                GroupProgressBar(value: campaign.progress, height: 12),
                const SizedBox(height: 8),
                if (campaign.nextTier != null)
                  Text(
                    '${campaign.participantsToNextTier} more to unlock '
                    '${Money.fromPaise(campaign.nextTier!.price)}',
                    style: const TextStyle(
                        color: AppColors.primaryDark,
                        fontSize: 12,
                        fontWeight: FontWeight.w700),
                  )
                else
                  const Text(
                    'Best group price unlocked 🎉',
                    style: TextStyle(
                        color: AppColors.primaryDark,
                        fontSize: 12,
                        fontWeight: FontWeight.w700),
                  ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 16),

        if (campaign.description.isNotEmpty) ...[
          const Text('About this deal',
              style: TextStyle(fontSize: 15, fontWeight: FontWeight.w800)),
          const SizedBox(height: 6),
          Text(campaign.description,
              style: TextStyle(color: muted, height: 1.5)),
          const SizedBox(height: 20),
        ],

        const Text('Pricing tiers',
            style: TextStyle(fontSize: 15, fontWeight: FontWeight.w800)),
        const SizedBox(height: 8),
        ...campaign.pricingTiers.map((t) {
          final active = campaign.participantCount >= t.minCount &&
              (t.maxCount == 0 || campaign.participantCount <= t.maxCount);
          return Container(
            margin: const EdgeInsets.only(bottom: 8),
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: active
                  ? AppColors.primary.withOpacity(0.08)
                  : Theme.of(context).colorScheme.surface,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: active ? AppColors.primary : AppColors.lightBorder,
              ),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    if (active) ...[
                      const Icon(Icons.check_circle_rounded,
                          size: 16, color: AppColors.primary),
                      const SizedBox(width: 6),
                    ],
                    Text(
                      t.maxCount == 0
                          ? '${t.minCount}+ people'
                          : '${t.minCount}–${t.maxCount} people',
                      style: const TextStyle(fontWeight: FontWeight.w600),
                    ),
                  ],
                ),
                Text(
                  Money.fromPaise(t.price),
                  style: TextStyle(
                    fontWeight: FontWeight.w800,
                    color: active ? AppColors.primaryDark : null,
                  ),
                ),
              ],
            ),
          );
        }),
        const SizedBox(height: 20),

        const Text('Booking details',
            style: TextStyle(fontSize: 15, fontWeight: FontWeight.w800)),
        const SizedBox(height: 4),
        _InfoRow(
            icon: Icons.calendar_today_rounded,
            label: 'Service date',
            value: Dates.short(campaign.serviceDate)),
        _InfoRow(
            icon: Icons.schedule_rounded,
            label: 'Deal closes',
            value: Dates.short(campaign.endDate)),
        _InfoRow(
            icon: Icons.hourglass_bottom_rounded,
            label: 'Days remaining',
            value: Dates.daysLeft(campaign.endDate)),
      ],
    );
  }
}

class _InfoRow extends StatelessWidget {
  const _InfoRow({required this.icon, required this.label, required this.value});
  final IconData icon;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        children: [
          Container(
            width: 32,
            height: 32,
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: AppColors.primary.withOpacity(0.1),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(icon, size: 16, color: AppColors.primaryDark),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(label, style: const TextStyle(color: AppColors.lightMuted)),
          ),
          Text(value, style: const TextStyle(fontWeight: FontWeight.w600)),
        ],
      ),
    );
  }
}
