import 'package:flutter/material.dart';

import '../config/theme.dart';

/// Full-screen centered spinner.
class LoadingView extends StatelessWidget {
  const LoadingView({super.key});
  @override
  Widget build(BuildContext context) =>
      const Center(child: CircularProgressIndicator());
}

/// Full-screen error state with a retry button.
class ErrorView extends StatelessWidget {
  const ErrorView({super.key, required this.message, this.onRetry});

  final String message;
  final VoidCallback? onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.cloud_off, size: 40, color: AppColors.lightMuted),
            const SizedBox(height: 12),
            Text(
              message,
              textAlign: TextAlign.center,
              style: const TextStyle(fontWeight: FontWeight.w600),
            ),
            if (onRetry != null) ...[
              const SizedBox(height: 16),
              OutlinedButton(onPressed: onRetry, child: const Text('Retry')),
            ],
          ],
        ),
      ),
    );
  }
}

/// Green progress bar toward the campaign minimum.
class GroupProgressBar extends StatelessWidget {
  const GroupProgressBar({
    super.key,
    required this.value,
    this.height = 10,
  });

  final double value;
  final double height;

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(999),
      child: LinearProgressIndicator(
        value: value.clamp(0, 1),
        minHeight: height,
        backgroundColor:
            Theme.of(context).colorScheme.primary.withOpacity(0.12),
        valueColor: const AlwaysStoppedAnimation(AppColors.primary),
      ),
    );
  }
}

/// Small colored pill for campaign status.
class StatusChip extends StatelessWidget {
  const StatusChip(this.status, {super.key});
  final String status;

  @override
  Widget build(BuildContext context) {
    final (bg, fg, label) = switch (status) {
      'PUBLISHED' => (const Color(0xFFDBEAFE), const Color(0xFF1D4ED8), 'Open'),
      'MINIMUM_REACHED' => (
          const Color(0xFFDCFCE7),
          AppColors.primaryDark,
          'Min reached'
        ),
      'CONFIRMED' => (
          const Color(0xFFDCFCE7),
          AppColors.primaryDark,
          'Confirmed'
        ),
      'COMPLETED' => (const Color(0xFFE2E8F0), Color(0xFF334155), 'Completed'),
      'CANCELLED' => (const Color(0xFFFEE2E2), AppColors.danger, 'Cancelled'),
      _ => (const Color(0xFFE2E8F0), Color(0xFF334155), status),
    };
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: TextStyle(fontSize: 11, fontWeight: FontWeight.w700, color: fg),
      ),
    );
  }
}
