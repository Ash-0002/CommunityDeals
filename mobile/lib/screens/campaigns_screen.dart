import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../models/campaign.dart';
import '../services/campaign_service.dart';
import '../widgets/campaign_card.dart';
import '../widgets/common.dart';

class CampaignsScreen extends StatefulWidget {
  const CampaignsScreen({super.key});

  @override
  State<CampaignsScreen> createState() => _CampaignsScreenState();
}

class _CampaignsScreenState extends State<CampaignsScreen> {
  static const _filters = <String, String?>{
    'All': null,
    'Open': 'PUBLISHED',
    'Min reached': 'MINIMUM_REACHED',
    'Confirmed': 'CONFIRMED',
  };

  String _active = 'All';
  late Future<List<CampaignListItem>> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<List<CampaignListItem>> _load() async {
    final page = await context
        .read<CampaignService>()
        .list(status: _filters[_active], limit: 50);
    return page.campaigns;
  }

  void _select(String filter) {
    setState(() {
      _active = filter;
      _future = _load();
    });
  }

  Future<void> _refresh() async {
    final f = _load();
    setState(() => _future = f);
    await f;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Deals')),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => context.push('/campaigns/new'),
        icon: const Icon(Icons.add),
        label: const Text('Start a deal'),
      ),
      body: Column(
        children: [
          SizedBox(
            height: 52,
            child: ListView(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.symmetric(horizontal: 12),
              children: [
                for (final name in _filters.keys)
                  Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: ChoiceChip(
                      label: Text(name),
                      selected: _active == name,
                      onSelected: (_) => _select(name),
                    ),
                  ),
              ],
            ),
          ),
          Expanded(
            child: RefreshIndicator(
              onRefresh: _refresh,
              child: FutureBuilder<List<CampaignListItem>>(
                future: _future,
                builder: (context, snap) {
                  if (snap.connectionState == ConnectionState.waiting) {
                    return const LoadingView();
                  }
                  if (snap.hasError) {
                    return ErrorView(
                        message: '${snap.error}', onRetry: _refresh);
                  }
                  final list = snap.data!;
                  if (list.isEmpty) {
                    return ListView(children: const [
                      SizedBox(height: 120),
                      Center(child: Text('Nothing here yet')),
                    ]);
                  }
                  return ListView(
                    padding: const EdgeInsets.fromLTRB(16, 12, 16, 32),
                    children: [
                      for (final c in list)
                        CampaignCard(
                          item: c,
                          onTap: () => context.push('/campaigns/${c.id}'),
                        ),
                    ],
                  );
                },
              ),
            ),
          ),
        ],
      ),
    );
  }
}
