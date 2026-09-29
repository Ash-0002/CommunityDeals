import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../config/theme.dart';
import '../core/api_exception.dart';
import '../core/format.dart';
import '../models/community.dart';
import '../services/campaign_service.dart';
import '../services/community_service.dart';
import '../widgets/common.dart';

/// A mutable pricing tier row edited in the form — separate from the
/// immutable `PricingTier` model used to render read-only tiers.
class _TierRow {
  _TierRow({required this.minCount, this.maxCount = 0, required this.priceRupees});
  int minCount;
  int maxCount; // 0 = unlimited
  int priceRupees; // whole rupees, converted to paise on submit
}

class CreateCampaignScreen extends StatefulWidget {
  const CreateCampaignScreen({super.key});

  @override
  State<CreateCampaignScreen> createState() => _CreateCampaignScreenState();
}

class _CreateCampaignScreenState extends State<CreateCampaignScreen> {
  final _formKey = GlobalKey<FormState>();
  final _titleCtrl = TextEditingController();
  final _serviceCtrl = TextEditingController();
  final _descCtrl = TextEditingController();
  final _imageCtrl = TextEditingController();
  final _minCtrl = TextEditingController(text: '10');
  final _maxCtrl = TextEditingController();

  late Future<List<Community>> _communitiesFuture;
  String? _communityId;

  DateTime _serviceDate = DateTime.now().add(const Duration(days: 5));
  DateTime _endDate = DateTime.now().add(const Duration(days: 3));

  final List<_TierRow> _tiers = [
    _TierRow(minCount: 1, maxCount: 9, priceRupees: 999),
    _TierRow(minCount: 10, maxCount: 0, priceRupees: 499),
  ];

  bool _submitting = false;

  @override
  void initState() {
    super.initState();
    _communitiesFuture = context.read<CommunityService>().myCommunities();
  }

  @override
  void dispose() {
    _titleCtrl.dispose();
    _serviceCtrl.dispose();
    _descCtrl.dispose();
    _imageCtrl.dispose();
    _minCtrl.dispose();
    _maxCtrl.dispose();
    super.dispose();
  }

  Future<void> _pickDate({required bool isServiceDate}) async {
    final initial = isServiceDate ? _serviceDate : _endDate;
    final picked = await showDatePicker(
      context: context,
      initialDate: initial,
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 365)),
    );
    if (picked == null) return;
    setState(() {
      if (isServiceDate) {
        _serviceDate = picked;
      } else {
        _endDate = picked;
      }
    });
  }

  void _addTier() {
    setState(() {
      final last = _tiers.last;
      _tiers.last.maxCount = last.minCount + 9;
      _tiers.add(_TierRow(
        minCount: last.maxCount == 0 ? last.minCount + 10 : last.maxCount + 1,
        maxCount: 0,
        priceRupees: (last.priceRupees * 0.8).round(),
      ));
    });
  }

  void _removeTier(int index) {
    if (_tiers.length <= 1) return;
    setState(() => _tiers.removeAt(index));
  }

  String? _validateTiers() {
    for (var i = 1; i < _tiers.length; i++) {
      if (_tiers[i].minCount <= _tiers[i - 1].minCount) {
        return 'Tier ${i + 1}\'s "from" count must be greater than the tier above it';
      }
      if (_tiers[i].priceRupees > _tiers[i - 1].priceRupees) {
        return 'Tier ${i + 1}\'s price should be equal to or lower than the tier above it';
      }
    }
    if (_tiers.first.minCount != 1) {
      return 'The first tier must start from 1 person';
    }
    return null;
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    if (_communityId == null) {
      _snack('Choose a community first', error: true);
      return;
    }
    final tierError = _validateTiers();
    if (tierError != null) {
      _snack(tierError, error: true);
      return;
    }

    setState(() => _submitting = true);
    try {
      final campaign = await context.read<CampaignService>().create(
            NewCampaign(
              communityId: _communityId!,
              serviceName: _serviceCtrl.text.trim(),
              title: _titleCtrl.text.trim(),
              description: _descCtrl.text.trim(),
              imageUrl: _imageCtrl.text.trim(),
              minParticipants: int.parse(_minCtrl.text.trim()),
              maxParticipants: int.tryParse(_maxCtrl.text.trim()) ?? 0,
              serviceDate: _serviceDate,
              startDate: DateTime.now(),
              endDate: _endDate,
              pricingTiers: _tiers
                  .map((t) => PricingTierInput(
                        minCount: t.minCount,
                        maxCount: t.maxCount,
                        price: t.priceRupees * 100,
                      ))
                  .toList(),
            ),
          );
      if (!mounted) return;
      _snack('Deal created — share the link to get people joining!');
      context.pushReplacement('/campaigns/${campaign.id}');
    } on ApiException catch (e) {
      _snack(e.message, error: true);
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  void _snack(String msg, {bool error = false}) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(msg), backgroundColor: error ? AppColors.danger : null),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Start a group deal')),
      body: FutureBuilder<List<Community>>(
        future: _communitiesFuture,
        builder: (context, snap) {
          if (snap.connectionState == ConnectionState.waiting) {
            return const LoadingView();
          }
          if (snap.hasError) {
            return ErrorView(message: '${snap.error}');
          }
          final communities = snap.data ?? [];
          _communityId ??= communities.isNotEmpty ? communities.first.id : null;

          return Form(
            key: _formKey,
            child: ListView(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 120),
              children: [
                if (communities.isEmpty)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Deals are posted inside a society — join one first.',
                          style: TextStyle(color: AppColors.lightMuted),
                        ),
                        const SizedBox(height: 10),
                        OutlinedButton.icon(
                          onPressed: () => context.push('/communities'),
                          icon: const Icon(Icons.search, size: 16),
                          label: const Text('Find your society'),
                        ),
                      ],
                    ),
                  )
                else
                  DropdownButtonFormField<String>(
                    initialValue: _communityId,
                    decoration: const InputDecoration(labelText: 'Community'),
                    items: [
                      for (final c in communities)
                        DropdownMenuItem(value: c.id, child: Text(c.name)),
                    ],
                    onChanged: (v) => setState(() => _communityId = v),
                  ),
                const SizedBox(height: 14),
                TextFormField(
                  controller: _serviceCtrl,
                  decoration: const InputDecoration(labelText: 'Service (e.g. AC Service)'),
                  validator: (v) => (v == null || v.trim().isEmpty) ? 'Required' : null,
                ),
                const SizedBox(height: 14),
                TextFormField(
                  controller: _titleCtrl,
                  decoration: const InputDecoration(labelText: 'Deal title (e.g. AC Service This Sunday)'),
                  validator: (v) => (v == null || v.trim().isEmpty) ? 'Required' : null,
                ),
                const SizedBox(height: 14),
                TextFormField(
                  controller: _descCtrl,
                  maxLines: 3,
                  decoration: const InputDecoration(labelText: 'Description (optional)'),
                ),
                const SizedBox(height: 14),
                TextFormField(
                  controller: _imageCtrl,
                  decoration: const InputDecoration(
                    labelText: 'Image URL (optional)',
                    hintText: 'https://…',
                  ),
                ),
                const SizedBox(height: 20),
                Row(
                  children: [
                    Expanded(
                      child: TextFormField(
                        controller: _minCtrl,
                        keyboardType: TextInputType.number,
                        decoration: const InputDecoration(labelText: 'Min. participants'),
                        validator: (v) =>
                            (int.tryParse(v ?? '') ?? 0) < 1 ? 'Enter a number ≥ 1' : null,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: TextFormField(
                        controller: _maxCtrl,
                        keyboardType: TextInputType.number,
                        decoration: const InputDecoration(
                            labelText: 'Max (optional)', hintText: '0 = no limit'),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 20),
                Row(
                  children: [
                    Expanded(child: _DatePickerTile(
                      label: 'Service date',
                      date: _serviceDate,
                      onTap: () => _pickDate(isServiceDate: true),
                    )),
                    const SizedBox(width: 12),
                    Expanded(child: _DatePickerTile(
                      label: 'Joining closes',
                      date: _endDate,
                      onTap: () => _pickDate(isServiceDate: false),
                    )),
                  ],
                ),
                const SizedBox(height: 24),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text('Dynamic pricing',
                        style: TextStyle(fontSize: 15, fontWeight: FontWeight.w800)),
                    TextButton.icon(
                      onPressed: _addTier,
                      icon: const Icon(Icons.add, size: 18),
                      label: const Text('Add tier'),
                    ),
                  ],
                ),
                const Text(
                  'Price drops automatically as more people join. The first tier must start from 1.',
                  style: TextStyle(fontSize: 12, color: AppColors.lightMuted),
                ),
                const SizedBox(height: 10),
                for (var i = 0; i < _tiers.length; i++)
                  _TierEditor(
                    tier: _tiers[i],
                    index: i,
                    canRemove: _tiers.length > 1,
                    onChanged: () => setState(() {}),
                    onRemove: () => _removeTier(i),
                  ),
              ],
            ),
          );
        },
      ),
      bottomNavigationBar: SafeArea(
        minimum: const EdgeInsets.all(16),
        child: FilledButton(
          onPressed: _submitting ? null : _submit,
          child: _submitting
              ? const SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
              : const Text('Publish deal'),
        ),
      ),
    );
  }
}

class _DatePickerTile extends StatelessWidget {
  const _DatePickerTile({required this.label, required this.date, required this.onTap});
  final String label;
  final DateTime date;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          border: Border.all(color: AppColors.lightBorder),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(label, style: const TextStyle(fontSize: 11, color: AppColors.lightMuted)),
            const SizedBox(height: 4),
            Text(Dates.short(date.toIso8601String()),
                style: const TextStyle(fontWeight: FontWeight.w700)),
          ],
        ),
      ),
    );
  }
}

class _TierEditor extends StatelessWidget {
  const _TierEditor({
    required this.tier,
    required this.index,
    required this.canRemove,
    required this.onChanged,
    required this.onRemove,
  });

  final _TierRow tier;
  final int index;
  final bool canRemove;
  final VoidCallback onChanged;
  final VoidCallback onRemove;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.lightBorder),
      ),
      child: Row(
        children: [
          Expanded(
            child: _NumberField(
              label: 'From',
              value: tier.minCount,
              onChanged: (v) {
                tier.minCount = v;
                onChanged();
              },
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: _NumberField(
              label: 'To (0=∞)',
              value: tier.maxCount,
              onChanged: (v) {
                tier.maxCount = v;
                onChanged();
              },
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: _NumberField(
              label: '₹ price',
              value: tier.priceRupees,
              onChanged: (v) {
                tier.priceRupees = v;
                onChanged();
              },
            ),
          ),
          IconButton(
            onPressed: canRemove ? onRemove : null,
            icon: const Icon(Icons.delete_outline, size: 20),
            color: AppColors.danger,
          ),
        ],
      ),
    );
  }
}

class _NumberField extends StatelessWidget {
  const _NumberField({required this.label, required this.value, required this.onChanged});
  final String label;
  final int value;
  final ValueChanged<int> onChanged;

  @override
  Widget build(BuildContext context) {
    return TextFormField(
      initialValue: value.toString(),
      keyboardType: TextInputType.number,
      decoration: InputDecoration(labelText: label, isDense: true),
      onChanged: (v) => onChanged(int.tryParse(v) ?? 0),
    );
  }
}
