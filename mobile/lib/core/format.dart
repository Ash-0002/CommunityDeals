import 'package:intl/intl.dart';

/// Formatting helpers. Prices from the backend are integers in **paise**.
class Money {
  Money._();

  static final NumberFormat _inr = NumberFormat.currency(
    locale: 'en_IN',
    symbol: '₹',
    decimalDigits: 0,
  );

  /// `45000` (paise) → `₹450`
  static String fromPaise(int paise) => _inr.format(paise / 100);
}

class Dates {
  Dates._();

  static String short(String? iso) {
    if (iso == null || iso.isEmpty) return '—';
    final d = DateTime.tryParse(iso);
    if (d == null) return '—';
    return DateFormat('d MMM yyyy').format(d.toLocal());
  }

  static String daysLeft(String? iso) {
    if (iso == null || iso.isEmpty) return '';
    final d = DateTime.tryParse(iso);
    if (d == null) return '';
    final diff = d.toLocal().difference(DateTime.now()).inDays;
    if (diff < 0) return 'Ended';
    if (diff == 0) return 'Ends today';
    return '$diff ${diff == 1 ? 'day' : 'days'} left';
  }
}
