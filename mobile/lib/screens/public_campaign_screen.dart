import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../models/campaign.dart';
import '../services/campaign_service.dart';
import '../state/auth_controller.dart';
import '../widgets/campaign_detail_body.dart';
import '../widgets/common.dart';

/// Unauthenticated campaign page opened from a shared WhatsApp link
/// (`communitydeals://c/<slug>` or a web `/c/<slug>` URL).
class PublicCampaignScreen extends StatefulWidget {
  const PublicCampaignScreen({super.key, required this.slug});

  final String slug;

  @override
  State<PublicCampaignScreen> createState() => _PublicCampaignScreenState();
}

class _PublicCampaignScreenState extends State<PublicCampaignScreen> {
  late Future<Campaign> _future;

  @override
  void initState() {
    super.initState();
    _future = context.read<CampaignService>().bySlug(widget.slug);
  }

  void _reload() {
    setState(() {
      _future = context.read<CampaignService>().bySlug(widget.slug);
    });
  }

  @override
  Widget build(BuildContext context) {
    final authed = context.watch<AuthController>().isAuthenticated;

    return Scaffold(
      appBar: AppBar(title: const Text('CommunityDeals')),
      body: FutureBuilder<Campaign>(
        future: _future,
        builder: (context, snap) {
          if (snap.connectionState == ConnectionState.waiting) {
            return const LoadingView();
          }
          if (snap.hasError) {
            return ErrorView(message: '${snap.error}', onRetry: _reload);
          }
          return CampaignDetailBody(campaign: snap.data!);
        },
      ),
      bottomNavigationBar: SafeArea(
        minimum: const EdgeInsets.all(16),
        child: FilledButton(
          onPressed: () {
            if (authed) {
              final c = _future;
              c.then((camp) {
                if (mounted) context.go('/campaigns/${camp.id}');
              });
            } else {
              context.go('/auth');
            }
          },
          child: Text(authed ? 'Open in app' : 'Log in to join'),
        ),
      ),
    );
  }
}
