import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../config/theme.dart';
import '../core/api_exception.dart';
import '../models/campaign.dart';
import '../models/participant.dart';
import '../services/campaign_service.dart';
import '../widgets/campaign_detail_body.dart';
import '../widgets/common.dart';

class CampaignDetailScreen extends StatefulWidget {
  const CampaignDetailScreen({super.key, required this.id});

  final String id;

  @override
  State<CampaignDetailScreen> createState() => _CampaignDetailScreenState();
}

class _DetailData {
  _DetailData({required this.campaign, required this.participants});
  final Campaign campaign;
  final List<Participant> participants;
}

class _CampaignDetailScreenState extends State<CampaignDetailScreen> {
  late Future<_DetailData> _future;
  bool _joining = false;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<_DetailData> _load() async {
    final service = context.read<CampaignService>();
    final results = await Future.wait([
      service.byId(widget.id),
      service.participants(widget.id),
    ]);
    return _DetailData(
      campaign: results[0] as Campaign,
      participants: results[1] as List<Participant>,
    );
  }

  void _reload() {
    if (!mounted) return;
    setState(() => _future = _load());
  }

  Future<void> _toggleJoin(Campaign campaign) async {
    final service = context.read<CampaignService>();

    if (campaign.isJoined) {
      final confirm = await showDialog<bool>(
        context: context,
        builder: (ctx) => AlertDialog(
          title: const Text('Cancel your spot?'),
          content: Text(
              'You\'ll leave "${campaign.title}" and lose your locked-in price. You can rejoin later if the deal is still open.'),
          actions: [
            TextButton(
                onPressed: () => Navigator.pop(ctx, false),
                child: const Text('Keep my spot')),
            TextButton(
              onPressed: () => Navigator.pop(ctx, true),
              child: const Text('Cancel spot',
                  style: TextStyle(color: AppColors.danger)),
            ),
          ],
        ),
      );
      if (confirm != true) return;
    }

    setState(() => _joining = true);
    try {
      if (campaign.isJoined) {
        await service.leave(campaign.id);
        _snack('You left this deal');
      } else {
        final r = await service.join(campaign.id);
        _snack(r.message);
      }
      _reload();
    } on ApiException catch (e) {
      _snack(e.message, error: true);
    } finally {
      if (mounted) setState(() => _joining = false);
    }
  }

  void _snack(String msg, {bool error = false}) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg),
        backgroundColor: error ? AppColors.danger : null,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Deal details')),
      body: FutureBuilder<_DetailData>(
        future: _future,
        builder: (context, snap) {
          if (snap.connectionState == ConnectionState.waiting) {
            return const LoadingView();
          }
          if (snap.hasError) {
            return ErrorView(message: '${snap.error}', onRetry: _reload);
          }
          return CampaignDetailBody(
            campaign: snap.data!.campaign,
            participants: snap.data!.participants,
          );
        },
      ),
      bottomNavigationBar: FutureBuilder<_DetailData>(
        future: _future,
        builder: (context, snap) {
          if (!snap.hasData) return const SizedBox.shrink();
          final c = snap.data!.campaign;
          final closed = c.status == 'CANCELLED' || c.status == 'COMPLETED';
          return SafeArea(
            minimum: const EdgeInsets.all(16),
            child: FilledButton(
              style: c.isJoined
                  ? FilledButton.styleFrom(
                      backgroundColor: Colors.transparent,
                      foregroundColor: AppColors.danger,
                      side: const BorderSide(color: AppColors.danger),
                    )
                  : null,
              onPressed: (_joining || closed) ? null : () => _toggleJoin(c),
              child: _joining
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(
                          strokeWidth: 2, color: Colors.white))
                  : Text(closed
                      ? 'Deal closed'
                      : c.isJoined
                          ? 'Cancel my spot'
                          : 'Join this deal'),
            ),
          );
        },
      ),
    );
  }
}
