import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import 'config/theme.dart';
import 'router.dart';
import 'services/campaign_service.dart';
import 'services/community_service.dart';
import 'state/auth_controller.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const CommunityDealsApp());
}

class CommunityDealsApp extends StatefulWidget {
  const CommunityDealsApp({super.key});

  @override
  State<CommunityDealsApp> createState() => _CommunityDealsAppState();
}

class _CommunityDealsAppState extends State<CommunityDealsApp> {
  late final AuthController _auth;
  late final GoRouter router;

  @override
  void initState() {
    super.initState();
    _auth = AuthController();
    router = buildRouter(_auth);
    _auth.bootstrap();
  }

  @override
  void dispose() {
    _auth.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider.value(value: _auth),
        Provider(create: (_) => CampaignService()),
        Provider(create: (_) => CommunityService()),
      ],
      child: MaterialApp.router(
        title: 'CommunityDeals',
        debugShowCheckedModeBanner: false,
        theme: AppTheme.light(),
        darkTheme: AppTheme.dark(),
        themeMode: ThemeMode.system,
        routerConfig: router,
      ),
    );
  }
}
