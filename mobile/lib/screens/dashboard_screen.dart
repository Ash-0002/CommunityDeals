import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../config/theme.dart';
import '../core/format.dart';
import '../models/campaign.dart';
import '../models/community.dart';
import '../services/campaign_service.dart';
import '../services/community_service.dart';
import '../state/auth_controller.dart';
import '../widgets/campaign_card.dart';
import '../widgets/campaign_visuals.dart';
import '../widgets/common.dart';

class DashboardScreen extends StatefulWidget {
  const DashboardScreen({super.key});

  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen> {
  late Future<_DashboardData> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<_DashboardData> _load() async {
    final campaigns = context.read<CampaignService>();
    final communities = context.read<CommunityService>();
    final results = await Future.wait([
      communities.myCommunities(),
      campaigns.list(status: 'PUBLISHED', limit: 20),
      campaigns.list(status: 'MINIMUM_REACHED', limit: 20),
    ]);
    final active = [
      ...(results[1] as CampaignPage).campaigns,
      ...(results[2] as CampaignPage).campaigns,
    ];
    return _DashboardData(
      communities: results[0] as List<Community>,
      campaigns: active,
    );
  }

  Future<void> _refresh() async {
    final data = _load();
    setState(() => _future = data);
    await data;
  }

  @override
  Widget build(BuildContext context) {
    final user = context.watch<AuthController>().user;

    return Scaffold(
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Hello 👋',
                style: TextStyle(fontSize: 12, color: AppColors.lightMuted)),
            Text(user?.name.isNotEmpty == true ? user!.name : 'there',
                style: const TextStyle(
                    fontSize: 18, fontWeight: FontWeight.w800)),
          ],
        ),
      ),
      body: RefreshIndicator(
        onRefresh: _refresh,
        child: FutureBuilder<_DashboardData>(
          future: _future,
          builder: (context, snap) {
            if (snap.connectionState == ConnectionState.waiting) {
              return const LoadingView();
            }
            if (snap.hasError) {
              return ListView(children: [
                SizedBox(
                  height: MediaQuery.of(context).size.height * 0.7,
                  child: ErrorView(
                    message: '${snap.error}',
                    onRetry: _refresh,
                  ),
                ),
              ]);
            }
            final data = snap.data!;
            final spotlight = _pickSpotlight(data.campaigns);
            final rest = data.campaigns.where((c) => c.id != spotlight?.id).toList();

            return ListView(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 32),
              children: [
                _CommunitiesStrip(communities: data.communities),
                const SizedBox(height: 20),
                _StatsRow(campaigns: data.campaigns, communityCount: data.communities.length),
                const SizedBox(height: 24),
                if (spotlight != null) ...[
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text('Today\'s spotlight',
                          style: TextStyle(fontSize: 16, fontWeight: FontWeight.w800)),
                      TextButton(
                        onPressed: () => context.go('/campaigns'),
                        child: const Text('See all'),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  _SpotlightCard(
                    item: spotlight,
                    onTap: () => context.push('/campaigns/${spotlight.id}'),
                  ),
                  const SizedBox(height: 24),
                ],
                if (rest.isNotEmpty) ...[
                  const Text('More active deals',
                      style: TextStyle(fontSize: 16, fontWeight: FontWeight.w800)),
                  const SizedBox(height: 4),
                  ...rest.take(3).map(
                        (c) => CampaignCard(
                          item: c,
                          onTap: () => context.push('/campaigns/${c.id}'),
                        ),
                      ),
                ] else if (spotlight == null)
                  const Padding(
                    padding: EdgeInsets.symmetric(vertical: 32),
                    child: Center(
                      child: Text('No active deals right now',
                          style: TextStyle(color: AppColors.lightMuted)),
                    ),
                  ),
              ],
            );
          },
        ),
      ),
    );
  }

  /// Picks the deal most worth surfacing: prefer one the user hasn't joined
  /// yet and is closest to unlocking its next price tier; fall back to the
  /// first active campaign.
  CampaignListItem? _pickSpotlight(List<CampaignListItem> campaigns) {
    if (campaigns.isEmpty) return null;
    final candidates = campaigns.where((c) => !c.isJoined).toList();
    final pool = candidates.isNotEmpty ? candidates : campaigns;
    pool.sort((a, b) {
      final aLeft = (a.minParticipants - a.participantCount).clamp(0, 1 << 30);
      final bLeft = (b.minParticipants - b.participantCount).clamp(0, 1 << 30);
      return aLeft.compareTo(bLeft);
    });
    return pool.first;
  }
}

class _SpotlightCard extends StatelessWidget {
  const _SpotlightCard({required this.item, required this.onTap});
  final CampaignListItem item;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final unlocked = item.minParticipants > 0 && item.participantCount >= item.minParticipants;
    final percentOff = item.firstTierPrice > item.currentPrice
        ? (((item.firstTierPrice - item.currentPrice) / item.firstTierPrice) * 100).round()
        : 0;

    return InkWell(
      borderRadius: BorderRadius.circular(20),
      onTap: onTap,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          CampaignHeroBanner(imageUrl: item.imageUrl, label: 'Society announcement'),
          const SizedBox(height: 12),
          Text(item.title,
              style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w900)),
          const SizedBox(height: 2),
          Text(item.serviceName, style: const TextStyle(color: AppColors.lightMuted)),
          const SizedBox(height: 12),
          JoinedAvatarsRow(
            participants: const [],
            totalJoined: item.participantCount,
            unlocked: unlocked,
          ),
          const SizedBox(height: 12),
          DiscountBadge(
            originalPrice: item.firstTierPrice,
            currentPrice: item.currentPrice,
            percentOff: percentOff,
          ),
          const SizedBox(height: 12),
          GroupProgressBar(value: item.progress, height: 10),
          const SizedBox(height: 6),
          Text(
            '${item.participantCount} of ${item.minParticipants} joined'
            '${Dates.daysLeft(item.endDate).isNotEmpty ? ' · ${Dates.daysLeft(item.endDate)}' : ''}',
            style: const TextStyle(
                fontSize: 12, fontWeight: FontWeight.w600, color: AppColors.lightMuted),
          ),
        ],
      ),
    );
  }
}

class _StatsRow extends StatelessWidget {
  const _StatsRow({required this.campaigns, required this.communityCount});
  final List<CampaignListItem> campaigns;
  final int communityCount;

  @override
  Widget build(BuildContext context) {
    final joined = campaigns.where((c) => c.isJoined).toList();
    final saved = joined.fold<int>(0, (sum, c) => sum + c.amountSaved);

    return Row(
      children: [
        Expanded(child: _StatTile(label: 'Deals joined', value: '${joined.length}')),
        const SizedBox(width: 10),
        Expanded(child: _StatTile(label: 'You\'ve saved', value: Money.fromPaise(saved))),
        const SizedBox(width: 10),
        Expanded(child: _StatTile(label: 'Communities', value: '$communityCount')),
      ],
    );
  }
}

class _StatTile extends StatelessWidget {
  const _StatTile({required this.label, required this.value});
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 14, horizontal: 10),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: AppColors.lightBorder),
      ),
      child: Column(
        children: [
          Text(value,
              style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w900, color: AppColors.primaryDark)),
          const SizedBox(height: 2),
          Text(label,
              textAlign: TextAlign.center,
              style: const TextStyle(fontSize: 10.5, color: AppColors.lightMuted, fontWeight: FontWeight.w600)),
        ],
      ),
    );
  }
}

class _CommunitiesStrip extends StatelessWidget {
  const _CommunitiesStrip({required this.communities});
  final List<Community> communities;

  @override
  Widget build(BuildContext context) {
    if (communities.isEmpty) {
      return Card(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              const Icon(Icons.groups_outlined, color: AppColors.primary),
              const SizedBox(width: 12),
              const Expanded(
                child: Text('You have not joined a community yet.',
                    style: TextStyle(fontWeight: FontWeight.w600)),
              ),
            ],
          ),
        ),
      );
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('My communities',
            style: TextStyle(fontSize: 16, fontWeight: FontWeight.w800)),
        const SizedBox(height: 12),
        SizedBox(
          height: 92,
          child: ListView.separated(
            scrollDirection: Axis.horizontal,
            itemCount: communities.length,
            separatorBuilder: (_, __) => const SizedBox(width: 12),
            itemBuilder: (context, i) {
              final c = communities[i];
              return Container(
                width: 180,
                padding: const EdgeInsets.all(14),
                decoration: BoxDecoration(
                  color: Theme.of(context).colorScheme.surface,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: AppColors.lightBorder),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(c.name,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                            fontWeight: FontWeight.w800, fontSize: 14)),
                    const SizedBox(height: 4),
                    Text('${c.memberCount} members',
                        style: const TextStyle(
                            fontSize: 12, color: AppColors.lightMuted)),
                  ],
                ),
              );
            },
          ),
        ),
      ],
    );
  }
}

class _DashboardData {
  _DashboardData({required this.communities, required this.campaigns});
  final List<Community> communities;
  final List<CampaignListItem> campaigns;
}
