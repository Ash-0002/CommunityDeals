import 'package:flutter/material.dart';

import '../config/theme.dart';
import '../core/format.dart';
import '../models/campaign.dart';
import 'common.dart';

class CampaignCard extends StatelessWidget {
  const CampaignCard({super.key, required this.item, required this.onTap});

  final CampaignListItem item;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final muted = Theme.of(context).brightness == Brightness.dark
        ? AppColors.darkMuted
        : AppColors.lightMuted;
    final unlocked = item.minParticipants > 0 &&
        item.participantCount >= item.minParticipants;

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onTap,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (item.imageUrl != null && item.imageUrl!.isNotEmpty)
              AspectRatio(
                aspectRatio: 16 / 8,
                child: Stack(
                  fit: StackFit.expand,
                  children: [
                    Image.network(
                      item.imageUrl!,
                      fit: BoxFit.cover,
                      errorBuilder: (_, __, ___) =>
                          Container(color: AppColors.primary.withOpacity(0.08)),
                    ),
                    Positioned(
                      right: 10,
                      top: 10,
                      child: StatusChip(item.status),
                    ),
                    if (unlocked)
                      Positioned(
                        left: 10,
                        top: 10,
                        child: _Pill(
                          icon: Icons.check_circle_rounded,
                          label: 'Unlocked',
                        ),
                      ),
                  ],
                ),
              ),
            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          item.title,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                              fontSize: 15, fontWeight: FontWeight.w800),
                        ),
                      ),
                      if (item.imageUrl == null || item.imageUrl!.isEmpty) ...[
                        const SizedBox(width: 8),
                        StatusChip(item.status),
                      ],
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(item.serviceName,
                      style: TextStyle(fontSize: 12, color: muted)),
                  const SizedBox(height: 12),
                  GroupProgressBar(value: item.progress),
                  const SizedBox(height: 8),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        '${item.participantCount}/${item.minParticipants} joined',
                        style: TextStyle(
                            fontSize: 12,
                            fontWeight: FontWeight.w600,
                            color: muted),
                      ),
                      Text(
                        Money.fromPaise(item.currentPrice),
                        style: const TextStyle(
                            fontSize: 15,
                            fontWeight: FontWeight.w800,
                            color: AppColors.primary),
                      ),
                    ],
                  ),
                  if (Dates.daysLeft(item.endDate).isNotEmpty) ...[
                    const SizedBox(height: 6),
                    Text(Dates.daysLeft(item.endDate),
                        style: TextStyle(fontSize: 11, color: muted)),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _Pill extends StatelessWidget {
  const _Pill({required this.icon, required this.label});
  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.92),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: AppColors.primaryDark),
          const SizedBox(width: 4),
          Text(
            label,
            style: const TextStyle(
              fontSize: 10.5,
              fontWeight: FontWeight.w800,
              color: AppColors.primaryDark,
            ),
          ),
        ],
      ),
    );
  }
}
