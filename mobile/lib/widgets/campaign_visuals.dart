import 'dart:math' as math;

import 'package:flutter/material.dart';

import '../config/theme.dart';
import '../core/format.dart';
import '../models/participant.dart';

/// Full-bleed hero image for a campaign, with a category pill overlaid —
/// e.g. "GROUP BOOKING" on a soft gradient, matching the reference designs.
class CampaignHeroBanner extends StatelessWidget {
  const CampaignHeroBanner({
    super.key,
    required this.imageUrl,
    required this.label,
  });

  final String? imageUrl;
  final String label;

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(20),
      child: AspectRatio(
        aspectRatio: 16 / 10,
        child: Stack(
          fit: StackFit.expand,
          children: [
            Container(
              decoration: const BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                  colors: [Color(0xFFDCEBFF), Color(0xFFEFF6FF)],
                ),
              ),
            ),
            if (imageUrl != null && imageUrl!.isNotEmpty)
              Image.network(
                imageUrl!,
                fit: BoxFit.cover,
                loadingBuilder: (context, child, progress) =>
                    progress == null ? child : const SizedBox.shrink(),
                errorBuilder: (_, __, ___) => const SizedBox.shrink(),
              ),
            Positioned(
              left: 14,
              top: 14,
              child: Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.92),
                  borderRadius: BorderRadius.circular(999),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.groups_rounded,
                        size: 14, color: AppColors.primaryDark),
                    const SizedBox(width: 6),
                    Text(
                      label.toUpperCase(),
                      style: const TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.w800,
                        color: AppColors.primaryDark,
                        letterSpacing: 0.4,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Overlapping avatar stack + "N people joined" pill + an animated
/// "Group booking unlocked" confetti banner once the minimum is hit.
/// Mirrors the reference screenshots' social-proof strip.
class JoinedAvatarsRow extends StatelessWidget {
  const JoinedAvatarsRow({
    super.key,
    required this.participants,
    required this.totalJoined,
    required this.unlocked,
  });

  final List<Participant> participants;
  final int totalJoined;
  final bool unlocked;

  @override
  Widget build(BuildContext context) {
    const maxShown = 6;
    final shown = participants.take(maxShown).toList();
    final overflow = totalJoined - shown.length;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: AppColors.lightBorder),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                width: 40,
                height: 40,
                decoration: const BoxDecoration(
                  color: AppColors.primary,
                  shape: BoxShape.circle,
                ),
                child: const Icon(Icons.groups_rounded,
                    color: Colors.white, size: 20),
              ),
              const SizedBox(width: 12),
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: const Color(0xFFFDE68A),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text.rich(
                  TextSpan(
                    children: [
                      TextSpan(
                        text: '$totalJoined ',
                        style: const TextStyle(
                            fontWeight: FontWeight.w900, fontSize: 15),
                      ),
                      const TextSpan(
                        text: 'people joined',
                        style: TextStyle(
                            fontWeight: FontWeight.w700, fontSize: 13),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
          if (shown.isNotEmpty) ...[
            const SizedBox(height: 14),
            SizedBox(
              height: 40,
              child: Stack(
                children: [
                  for (var i = 0; i < shown.length; i++)
                    Positioned(
                      left: i * 26.0,
                      child: _Avatar(participant: shown[i]),
                    ),
                  if (overflow > 0)
                    Positioned(
                      left: shown.length * 26.0,
                      child: Container(
                        width: 36,
                        height: 36,
                        alignment: Alignment.center,
                        decoration: BoxDecoration(
                          color: AppColors.primary.withOpacity(0.12),
                          shape: BoxShape.circle,
                          border: Border.all(color: Colors.white, width: 2),
                        ),
                        child: Text(
                          '+$overflow',
                          style: const TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w800,
                            color: AppColors.primaryDark,
                          ),
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ],
          if (unlocked) ...[
            const SizedBox(height: 14),
            const _UnlockedBanner(),
          ],
        ],
      ),
    );
  }
}

class _Avatar extends StatelessWidget {
  const _Avatar({required this.participant});
  final Participant participant;

  @override
  Widget build(BuildContext context) {
    final url = participant.avatarUrl;
    return Container(
      width: 36,
      height: 36,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        border: Border.all(color: Colors.white, width: 2),
        color: AppColors.primary.withOpacity(0.15),
        image: (url != null && url.isNotEmpty)
            ? DecorationImage(image: NetworkImage(url), fit: BoxFit.cover)
            : null,
      ),
      alignment: Alignment.center,
      child: (url == null || url.isEmpty)
          ? Text(
              participant.initials,
              style: const TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.w800,
                color: AppColors.primaryDark,
              ),
            )
          : null,
    );
  }
}

/// Confetti-flanked "Group booking unlocked" pill, shown once the campaign
/// hits its minimum participant count.
class _UnlockedBanner extends StatefulWidget {
  const _UnlockedBanner();

  @override
  State<_UnlockedBanner> createState() => _UnlockedBannerState();
}

class _UnlockedBannerState extends State<_UnlockedBanner>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 900),
    )..forward();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _controller,
      builder: (context, child) {
        return SizedBox(
          height: 44,
          child: Stack(
            alignment: Alignment.center,
            clipBehavior: Clip.none,
            children: [
              CustomPaint(
                size: const Size(double.infinity, 44),
                painter: _ConfettiPainter(progress: _controller.value),
              ),
              child!,
            ],
          ),
        );
      },
      child: ScaleTransition(
        scale: CurvedAnimation(parent: _controller, curve: Curves.elasticOut),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
          decoration: BoxDecoration(
            color: const Color(0xFFDCFCE7),
            borderRadius: BorderRadius.circular(999),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 18,
                height: 18,
                decoration: const BoxDecoration(
                  color: AppColors.primary,
                  shape: BoxShape.circle,
                ),
                child: const Icon(Icons.check, size: 12, color: Colors.white),
              ),
              const SizedBox(width: 8),
              const Text(
                'Group booking unlocked',
                style: TextStyle(
                  fontSize: 12.5,
                  fontWeight: FontWeight.w800,
                  color: AppColors.primaryDark,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ConfettiPainter extends CustomPainter {
  _ConfettiPainter({required this.progress});
  final double progress;

  static const _colors = [
    Color(0xFFFBBF24),
    AppColors.primary,
    Color(0xFF60A5FA),
    Color(0xFFF472B6),
  ];

  @override
  void paint(Canvas canvas, Size size) {
    final rnd = math.Random(7); // stable layout across rebuilds
    final paint = Paint();
    const count = 14;
    for (var i = 0; i < count; i++) {
      final side = i.isEven ? -1 : 1;
      final baseX = size.width / 2 + side * (30 + rnd.nextDouble() * 70);
      final baseY = size.height / 2;
      final dy = -progress * (14 + rnd.nextDouble() * 18);
      final dx = side * progress * (6 + rnd.nextDouble() * 10);
      final opacity = (1 - progress).clamp(0.0, 1.0);
      paint.color = _colors[i % _colors.length].withOpacity(opacity);
      final rect = Rect.fromCenter(
        center: Offset(baseX + dx, baseY + dy),
        width: 5,
        height: 5,
      );
      canvas.save();
      canvas.translate(rect.center.dx, rect.center.dy);
      canvas.rotate(progress * math.pi * (i.isEven ? 2 : -2));
      canvas.translate(-rect.center.dx, -rect.center.dy);
      canvas.drawRect(rect, paint);
      canvas.restore();
    }
  }

  @override
  bool shouldRepaint(covariant _ConfettiPainter oldDelegate) =>
      oldDelegate.progress != progress;
}

/// Starburst "% OFF" sticker with a strikethrough original price, the
/// current price, and the amount saved — mirrors the reference design's
/// green discount badge.
class DiscountBadge extends StatelessWidget {
  const DiscountBadge({
    super.key,
    required this.originalPrice,
    required this.currentPrice,
    required this.percentOff,
  });

  final int originalPrice; // paise
  final int currentPrice; // paise
  final int percentOff;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: AppColors.lightBorder),
      ),
      child: Row(
        children: [
          if (percentOff > 0)
            SizedBox(
              width: 72,
              height: 72,
              child: CustomPaint(
                painter: _StarburstPainter(),
                child: Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        '$percentOff%',
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 17,
                          fontWeight: FontWeight.w900,
                          height: 1,
                        ),
                      ),
                      const Text(
                        'OFF',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 10,
                          fontWeight: FontWeight.w800,
                          letterSpacing: 0.5,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('Exclusive group price',
                    style: TextStyle(fontSize: 12, color: AppColors.lightMuted)),
                const SizedBox(height: 4),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    if (originalPrice > currentPrice) ...[
                      Text(
                        Money.fromPaise(originalPrice),
                        style: const TextStyle(
                          fontSize: 14,
                          color: AppColors.lightMuted,
                          decoration: TextDecoration.lineThrough,
                        ),
                      ),
                      const SizedBox(width: 8),
                    ],
                    Text(
                      Money.fromPaise(currentPrice),
                      style: const TextStyle(
                        fontSize: 24,
                        fontWeight: FontWeight.w900,
                        color: AppColors.primaryDark,
                      ),
                    ),
                  ],
                ),
                if (originalPrice > currentPrice) ...[
                  const SizedBox(height: 4),
                  Container(
                    padding: const EdgeInsets.symmetric(
                        horizontal: 8, vertical: 3),
                    decoration: BoxDecoration(
                      color: const Color(0xFFDCFCE7),
                      borderRadius: BorderRadius.circular(999),
                    ),
                    child: Text(
                      'You save ${Money.fromPaise(originalPrice - currentPrice)}',
                      style: const TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        color: AppColors.primaryDark,
                      ),
                    ),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _StarburstPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final outerR = size.width / 2;
    final innerR = outerR * 0.82;
    const points = 12;
    final path = Path();
    for (var i = 0; i < points * 2; i++) {
      final r = i.isEven ? outerR : innerR;
      final angle = (math.pi / points) * i - math.pi / 2;
      final p = Offset(
        center.dx + r * math.cos(angle),
        center.dy + r * math.sin(angle),
      );
      if (i == 0) {
        path.moveTo(p.dx, p.dy);
      } else {
        path.lineTo(p.dx, p.dy);
      }
    }
    path.close();

    final paint = Paint()
      ..shader = const LinearGradient(
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
        colors: [AppColors.primary, AppColors.primaryDark],
      ).createShader(Rect.fromCircle(center: center, radius: outerR));
    canvas.drawPath(path, paint);
  }

  @override
  bool shouldRepaint(covariant _StarburstPainter oldDelegate) => false;
}
