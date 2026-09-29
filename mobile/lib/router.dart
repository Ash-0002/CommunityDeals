import 'package:go_router/go_router.dart';

import 'state/auth_controller.dart';
import 'screens/splash_screen.dart';
import 'screens/auth/phone_screen.dart';
import 'screens/auth/otp_screen.dart';
import 'screens/home_shell.dart';
import 'screens/dashboard_screen.dart';
import 'screens/campaigns_screen.dart';
import 'screens/campaign_detail_screen.dart';
import 'screens/communities_screen.dart';
import 'screens/create_campaign_screen.dart';
import 'screens/public_campaign_screen.dart';
import 'screens/profile_screen.dart';
import 'screens/settings_screen.dart';

GoRouter buildRouter(AuthController auth) {
  return GoRouter(
    initialLocation: '/splash',
    refreshListenable: auth,
    redirect: (context, state) {
      final loc = state.matchedLocation;
      final isPublic = loc.startsWith('/c/');
      final isAuthFlow = loc == '/auth' || loc == '/auth/otp';
      final isSplash = loc == '/splash';

      switch (auth.status) {
        case AuthStatus.unknown:
          return isSplash ? null : '/splash';
        case AuthStatus.unauthenticated:
          if (isPublic || isAuthFlow) return null;
          return '/auth';
        case AuthStatus.authenticated:
          if (isSplash || isAuthFlow) return '/';
          return null;
      }
      return null;
    },
    routes: [
      GoRoute(path: '/splash', builder: (_, __) => const SplashScreen()),
      GoRoute(path: '/auth', builder: (_, __) => const PhoneScreen()),
      GoRoute(
        path: '/auth/otp',
        builder: (_, state) =>
            OtpScreen(phone: (state.extra as String?) ?? ''),
      ),
      GoRoute(
        path: '/c/:slug',
        builder: (_, state) =>
            PublicCampaignScreen(slug: state.pathParameters['slug']!),
      ),
      GoRoute(
        path: '/campaigns/new',
        builder: (_, __) => const CreateCampaignScreen(),
      ),
      GoRoute(
        path: '/campaigns/:id',
        builder: (_, state) =>
            CampaignDetailScreen(id: state.pathParameters['id']!),
      ),
      GoRoute(
        path: '/settings',
        builder: (_, __) => const SettingsScreen(),
      ),
      GoRoute(
        path: '/communities',
        builder: (_, __) => const CommunitiesScreen(),
      ),
      ShellRoute(
        builder: (_, __, child) => HomeShell(child: child),
        routes: [
          GoRoute(path: '/', builder: (_, __) => const DashboardScreen()),
          GoRoute(
              path: '/campaigns', builder: (_, __) => const CampaignsScreen()),
          GoRoute(
              path: '/profile', builder: (_, __) => const ProfileScreen()),
        ],
      ),
    ],
  );
}
