import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../config/theme.dart';
import '../core/api_exception.dart';
import '../models/community.dart';
import '../services/community_service.dart';
import '../widgets/common.dart';

/// Find and join a society. Without this a brand-new user has no way into a
/// community, and every other screen dead-ends ("join a community first").
/// Discovery is location-based — the backend filters on city / PIN code.
class CommunitiesScreen extends StatefulWidget {
  const CommunitiesScreen({super.key});

  @override
  State<CommunitiesScreen> createState() => _CommunitiesScreenState();
}

class _CommunitiesScreenState extends State<CommunitiesScreen> {
  final _searchCtrl = TextEditingController();

  late Future<_CommunitiesData> _future;
  String _query = '';
  String? _joiningId;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  @override
  void dispose() {
    _searchCtrl.dispose();
    super.dispose();
  }

  Future<_CommunitiesData> _load() async {
    final service = context.read<CommunityService>();
    // A 6-digit numeric query is a PIN code; anything else is a city name.
    final isPin = RegExp(r'^\d{4,6}$').hasMatch(_query.trim());
    final results = await Future.wait([
      service.myCommunities(),
      service.discover(
        city: isPin ? null : _query.trim(),
        pinCode: isPin ? _query.trim() : null,
        limit: 50,
      ),
    ]);
    return _CommunitiesData(
      mine: results[0],
      discovered: results[1],
    );
  }

  Future<void> _refresh() async {
    final f = _load();
    setState(() => _future = f);
    await f;
  }

  void _search(String value) {
    setState(() {
      _query = value;
      _future = _load();
    });
  }

  Future<void> _join(Community community) async {
    final service = context.read<CommunityService>();

    if (community.requiresApproval) {
      final proceed = await showDialog<bool>(
        context: context,
        builder: (ctx) => AlertDialog(
          title: const Text('Request to join?'),
          content: Text(
              '${community.name} approves members manually. Your request will '
              'be sent to the society admin.'),
          actions: [
            TextButton(
                onPressed: () => Navigator.pop(ctx, false),
                child: const Text('Cancel')),
            TextButton(
                onPressed: () => Navigator.pop(ctx, true),
                child: const Text('Send request')),
          ],
        ),
      );
      if (proceed != true) return;
    }

    setState(() => _joiningId = community.id);
    try {
      final result = await service.join(community.id);
      _snack(result.message);
      await _refresh();
    } on ApiException catch (e) {
      // A community using invite-code access will reject an empty code —
      // give the user a chance to enter one and retry.
      if (e.code == 'INVALID_INVITE_CODE') {
        final code = await _askInviteCode(community);
        if (code != null && code.isNotEmpty) {
          try {
            final result = await service.join(community.id, inviteCode: code);
            _snack(result.message);
            await _refresh();
          } on ApiException catch (e2) {
            _snack(e2.message, error: true);
          }
        }
      } else {
        _snack(e.message, error: true);
      }
    } finally {
      if (mounted) setState(() => _joiningId = null);
    }
  }

  Future<String?> _askInviteCode(Community community) {
    final ctrl = TextEditingController();
    return showDialog<String>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Invite code needed'),
        content: TextField(
          controller: ctrl,
          autofocus: true,
          textCapitalization: TextCapitalization.characters,
          decoration: InputDecoration(
            labelText: 'Invite code',
            hintText: 'Ask your ${community.name} admin',
          ),
        ),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx), child: const Text('Cancel')),
          TextButton(
              onPressed: () => Navigator.pop(ctx, ctrl.text.trim()),
              child: const Text('Join')),
        ],
      ),
    );
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
      appBar: AppBar(title: const Text('Find your society')),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
            child: TextField(
              controller: _searchCtrl,
              textInputAction: TextInputAction.search,
              onSubmitted: _search,
              decoration: InputDecoration(
                prefixIcon: const Icon(Icons.search),
                hintText: 'Search by city or PIN code',
                suffixIcon: _query.isEmpty
                    ? null
                    : IconButton(
                        icon: const Icon(Icons.clear),
                        onPressed: () {
                          _searchCtrl.clear();
                          _search('');
                        },
                      ),
              ),
            ),
          ),
          Expanded(
            child: RefreshIndicator(
              onRefresh: _refresh,
              child: FutureBuilder<_CommunitiesData>(
                future: _future,
                builder: (context, snap) {
                  if (snap.connectionState == ConnectionState.waiting) {
                    return const LoadingView();
                  }
                  if (snap.hasError) {
                    return ErrorView(
                        message: '${snap.error}', onRetry: _refresh);
                  }
                  final data = snap.data!;
                  final mineIds = data.mine.map((c) => c.id).toSet();
                  final others = data.discovered
                      .where((c) => !mineIds.contains(c.id))
                      .toList();

                  return ListView(
                    padding: const EdgeInsets.fromLTRB(16, 8, 16, 32),
                    children: [
                      if (data.mine.isNotEmpty) ...[
                        const _SectionTitle('My societies'),
                        ...data.mine.map((c) => _CommunityTile(
                              community: c,
                              joined: true,
                              busy: false,
                              onJoin: null,
                            )),
                        const SizedBox(height: 20),
                      ],
                      _SectionTitle(
                        _query.isEmpty
                            ? 'Societies you can join'
                            : 'Results for "$_query"',
                      ),
                      if (others.isEmpty)
                        Padding(
                          padding: const EdgeInsets.symmetric(vertical: 28),
                          child: Center(
                            child: Text(
                              _query.isEmpty
                                  ? 'No other societies listed yet.'
                                  : 'Nothing found for "$_query". Try a city '
                                      'name or a 6-digit PIN code.',
                              textAlign: TextAlign.center,
                              style:
                                  const TextStyle(color: AppColors.lightMuted),
                            ),
                          ),
                        )
                      else
                        ...others.map((c) => _CommunityTile(
                              community: c,
                              joined: false,
                              busy: _joiningId == c.id,
                              onJoin: () => _join(c),
                            )),
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

class _SectionTitle extends StatelessWidget {
  const _SectionTitle(this.text);
  final String text;

  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.only(bottom: 10),
        child: Text(text,
            style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w800)),
      );
}

class _CommunityTile extends StatelessWidget {
  const _CommunityTile({
    required this.community,
    required this.joined,
    required this.busy,
    required this.onJoin,
  });

  final Community community;
  final bool joined;
  final bool busy;
  final VoidCallback? onJoin;

  /// "Baner Road, Pune, Maharashtra 411045" from whichever parts exist.
  String get _location {
    final parts = <String>[
      if (community.address?.isNotEmpty == true) community.address!,
      if (community.city?.isNotEmpty == true) community.city!,
      if (community.state?.isNotEmpty == true) community.state!,
    ];
    final line = parts.join(', ');
    final pin = community.pinCode;
    if (pin != null && pin.isNotEmpty) {
      return line.isEmpty ? pin : '$line $pin';
    }
    return line;
  }

  @override
  Widget build(BuildContext context) {
    final location = _location;

    return Card(
      margin: const EdgeInsets.only(bottom: 10),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  child: Text(
                    community.name,
                    style: const TextStyle(
                        fontSize: 15, fontWeight: FontWeight.w800),
                  ),
                ),
                if (joined)
                  Container(
                    padding: const EdgeInsets.symmetric(
                        horizontal: 10, vertical: 4),
                    decoration: BoxDecoration(
                      color: const Color(0xFFDCFCE7),
                      borderRadius: BorderRadius.circular(999),
                    ),
                    child: const Text('Joined',
                        style: TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w700,
                            color: AppColors.primaryDark)),
                  ),
              ],
            ),
            if (location.isNotEmpty) ...[
              const SizedBox(height: 6),
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Icon(Icons.location_on_outlined,
                      size: 14, color: AppColors.lightMuted),
                  const SizedBox(width: 4),
                  Expanded(
                    child: Text(location,
                        style: const TextStyle(
                            fontSize: 12, color: AppColors.lightMuted)),
                  ),
                ],
              ),
            ],
            const SizedBox(height: 8),
            Row(
              children: [
                const Icon(Icons.groups_outlined,
                    size: 14, color: AppColors.lightMuted),
                const SizedBox(width: 4),
                Text('${community.memberCount} members',
                    style: const TextStyle(
                        fontSize: 12, color: AppColors.lightMuted)),
                if (community.requiresApproval) ...[
                  const SizedBox(width: 12),
                  const Icon(Icons.lock_outline,
                      size: 13, color: AppColors.lightMuted),
                  const SizedBox(width: 4),
                  const Text('Approval needed',
                      style: TextStyle(
                          fontSize: 12, color: AppColors.lightMuted)),
                ],
              ],
            ),
            if (!joined) ...[
              const SizedBox(height: 14),
              SizedBox(
                width: double.infinity,
                child: FilledButton(
                  style: FilledButton.styleFrom(
                    minimumSize: const Size.fromHeight(44),
                  ),
                  onPressed: busy ? null : onJoin,
                  child: busy
                      ? const SizedBox(
                          width: 18,
                          height: 18,
                          child: CircularProgressIndicator(
                              strokeWidth: 2, color: Colors.white))
                      : Text(community.requiresApproval
                          ? 'Request to join'
                          : 'Join society'),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _CommunitiesData {
  _CommunitiesData({required this.mine, required this.discovered});
  final List<Community> mine;
  final List<Community> discovered;
}
