import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:community_deals/config/theme.dart';
import 'package:community_deals/screens/splash_screen.dart';

void main() {
  testWidgets('splash screen renders the brand name', (tester) async {
    await tester.pumpWidget(
      MaterialApp(theme: AppTheme.light(), home: const SplashScreen()),
    );
    expect(find.text('CommunityDeals'), findsOneWidget);
  });
}
