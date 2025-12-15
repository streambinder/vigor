import 'package:flutter/material.dart';
import '../generated/app_localizations.dart';
import '../utils/platform_helper.dart';
import '../widgets/navigation/liquid_glass_nav_bar.dart';
import 'home_page.dart';
import 'activity_screen.dart';
import 'profile_screen.dart';

class MainNavigation extends StatefulWidget {
  const MainNavigation({super.key});

  static final GlobalKey<_MainNavigationState> navKey = GlobalKey();

  /// Navigate to a specific tab by index
  static void navigateToTab(int index) {
    navKey.currentState?._onTabTapped(index);
  }

  @override
  State<MainNavigation> createState() => _MainNavigationState();
}

class _MainNavigationState extends State<MainNavigation> {
  int _currentIndex = 0;

  final List<Widget> _screens = const [
    HomePage(),
    ActivityScreen(),
    ProfileScreen(),
  ];

  void _onTabTapped(int index) {
    setState(() {
      _currentIndex = index;
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    if (PlatformHelper.useLiquidGlass) {
      // iOS-style with Liquid Glass navigation
      return Scaffold(
        extendBody: true,
        body: Stack(
          children: [
            // Main content with padding for the navigation bar
            Padding(
              padding: const EdgeInsets.only(bottom: 100),
              child: _screens[_currentIndex],
            ),
            // Liquid Glass navigation bar
            Positioned(
              left: 16,
              right: 16,
              bottom: MediaQuery.of(context).padding.bottom + 16,
              child: LiquidGlassNavBar(
                currentIndex: _currentIndex,
                onTap: _onTabTapped,
                items: [
                  LiquidGlassNavItem(
                    icon: Icons.home_rounded,
                    label: l10n.navHome,
                  ),
                  LiquidGlassNavItem(
                    icon: Icons.show_chart_rounded,
                    label: l10n.navActivity,
                  ),
                  LiquidGlassNavItem(
                    icon: Icons.person_rounded,
                    label: l10n.navProfile,
                  ),
                ],
              ),
            ),
          ],
        ),
      );
    } else {
      // Material Design navigation
      return Scaffold(
        body: _screens[_currentIndex],
        bottomNavigationBar: NavigationBar(
          selectedIndex: _currentIndex,
          onDestinationSelected: _onTabTapped,
          destinations: [
            NavigationDestination(
              icon: const Icon(Icons.home_outlined),
              selectedIcon: const Icon(Icons.home_rounded),
              label: l10n.navHome,
            ),
            NavigationDestination(
              icon: const Icon(Icons.show_chart_outlined),
              selectedIcon: const Icon(Icons.show_chart_rounded),
              label: l10n.navActivity,
            ),
            NavigationDestination(
              icon: const Icon(Icons.person_outline),
              selectedIcon: const Icon(Icons.person_rounded),
              label: l10n.navProfile,
            ),
          ],
        ),
      );
    }
  }
}
