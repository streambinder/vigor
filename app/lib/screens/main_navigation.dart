import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../config/api_config.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/gym.dart';
import '../providers/auth_provider.dart';
import '../services/service_locator.dart';
import '../utils/platform_helper.dart';
import '../widgets/navigation/liquid_glass_nav_bar.dart';
import '../widgets/training_generation_modal.dart';
import 'home_page.dart';
import 'activity_screen.dart';
import 'profile_screen.dart';
import 'training_details_screen.dart';
import 'flow_details_screen.dart';

class MainNavigation extends StatefulWidget {
  const MainNavigation({super.key});

  static final GlobalKey<MainNavigationState> navKey = GlobalKey();

  /// Navigate to a specific tab by index
  static void navigateToTab(int index) {
    navKey.currentState?._onTabTapped(index);
  }

  @override
  State<MainNavigation> createState() => MainNavigationState();
}

class MainNavigationState extends State<MainNavigation> {
  int _currentIndex = 0;

  // use IndexedStack to preserve screen state across tab switches
  static const List<Widget> _screens = [
    HomePage(),
    ActivityScreen(),
    ProfileScreen(),
  ];

  void _onTabTapped(int index) {
    setState(() {
      _currentIndex = index;
    });
  }

  void _showTrainingGenerationModal() {
    final locator = context.read<ServiceLocator>();
    final gyms = locator.gymsNotifier.value ?? [];
    showDialog(
      context: context,
      barrierDismissible: true,
      builder: (dialogContext) => ValueListenableBuilder<List<Gym>?>(
        valueListenable: locator.gymsNotifier,
        builder: (context, currentGyms, _) => TrainingGenerationModal(
          gyms: currentGyms ?? gyms,
          onSuccess: (training) {
            Navigator.of(dialogContext).push(
              MaterialPageRoute(
                builder: (context) => TrainingDetailsScreen(training: training),
              ),
            );
          },
          onFlowSuccess: (flowSession) {
            Navigator.of(dialogContext).push(
              MaterialPageRoute(
                builder: (context) => FlowDetailsScreen(flowSession: flowSession),
              ),
            );
          },
        ),
      ),
    );
  }

  /// builds a small circular avatar for the profile tab icon,
  /// falls back to person icon if user has no avatar.
  /// wraps with a sync spinner when health data is syncing.
  Widget _buildProfileTabAvatar(String userId, bool isSelected) {
    final borderColor = isSelected ? VigorColors.indigo : VigorColors.stone;
    final avatar = CachedNetworkImage(
      imageUrl: ApiConfig.avatarUrl(userId),
      imageBuilder: (context, imageProvider) => Container(
        width: 24,
        height: 24,
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          image: DecorationImage(image: imageProvider, fit: BoxFit.cover),
          border: Border.all(color: borderColor, width: 1.5),
        ),
      ),
      placeholder: (context, url) => Icon(
        Icons.person_rounded,
        color: isSelected ? VigorColors.indigo : VigorColors.stone,
        size: 24,
      ),
      errorWidget: (context, url, error) => Icon(
        Icons.person_rounded,
        color: isSelected ? VigorColors.indigo : VigorColors.stone,
        size: 24,
      ),
    );

    final healthService = context.read<ServiceLocator>().healthDataService;
    if (healthService == null) return avatar;

    return ValueListenableBuilder<bool>(
      valueListenable: healthService.syncing,
      builder: (context, isSyncing, _) => SizedBox(
        width: 28,
        height: 28,
        child: Stack(
          alignment: Alignment.center,
          children: [
            avatar,
            if (isSyncing)
              SizedBox(
                width: 28,
                height: 28,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  color: VigorColors.indigoAdaptive(context),
                ),
              ),
          ],
        ),
      ),
    );
  }

  // animated FAB: collapsed on home, expanded on activity, hidden on profile
  Widget _buildFAB(AppLocalizations l10n) {
    final isVisible = _currentIndex != 2;
    final isExpanded = _currentIndex == 1;

    final label = Text(
      l10n.generateSession,
      style: VigorTypography.label.copyWith(
        color: Colors.white,
        fontWeight: FontWeight.w600,
      ),
      maxLines: 1,
      overflow: TextOverflow.clip,
    );

    return AnimatedScale(
      scale: isVisible ? 1.0 : 0.0,
      duration: const Duration(milliseconds: 120),
      curve: Curves.easeInOut,
      child: Material(
        color: VigorColors.persimmon,
        borderRadius: BorderRadius.circular(16),
        elevation: 4,
        child: InkWell(
          onTap: isVisible ? _showTrainingGenerationModal : null,
          borderRadius: BorderRadius.circular(16),
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 150),
            curve: Curves.easeInOut,
            padding: isExpanded
                ? const EdgeInsets.symmetric(horizontal: 20, vertical: 14)
                : const EdgeInsets.all(14),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(Icons.bolt, color: Colors.white, size: 22),
                AnimatedSize(
                  duration: const Duration(milliseconds: 150),
                  curve: Curves.easeInOut,
                  alignment: Alignment.centerLeft,
                  child: isExpanded
                      ? Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [const SizedBox(width: 8), label],
                        )
                      : const SizedBox.shrink(),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final fab = _buildFAB(l10n);
    final userId = context.watch<AuthProvider>().currentUser?.id;

    return PopScope(
      canPop: _currentIndex == 0,
      onPopInvokedWithResult: (didPop, _) {
        if (!didPop) _onTabTapped(0);
      },
      child: _buildScaffold(context, l10n, fab, userId),
    );
  }

  Widget _buildScaffold(
    BuildContext context,
    AppLocalizations l10n,
    Widget fab,
    String? userId,
  ) {
    if (PlatformHelper.useLiquidGlass) {
      // iOS-style with Liquid Glass navigation
      return Scaffold(
        extendBody: true,
        body: Stack(
          children: [
            // IndexedStack preserves state across tab switches
            Padding(
              padding: const EdgeInsets.only(bottom: 100),
              child: IndexedStack(
                index: _currentIndex,
                children: _screens,
              ),
            ),
            // FAB above the nav bar
            Positioned(
              right: 16,
              bottom: MediaQuery.of(context).padding.bottom + 90,
              child: fab,
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
                    customIcon: userId != null
                        ? _buildProfileTabAvatar(userId, _currentIndex == 2)
                        : null,
                  ),
                ],
              ),
            ),
          ],
        ),
      );
    } else {
      // Material Design navigation with IndexedStack
      final profileIcon = userId != null
          ? _buildProfileTabAvatar(userId, false)
          : const Icon(Icons.person_outline);
      final profileSelectedIcon = userId != null
          ? _buildProfileTabAvatar(userId, true)
          : const Icon(Icons.person_rounded);

      return Scaffold(
        body: IndexedStack(
          index: _currentIndex,
          children: _screens,
        ),
        floatingActionButton: fab,
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
              icon: profileIcon,
              selectedIcon: profileSelectedIcon,
              label: l10n.navProfile,
            ),
          ],
        ),
      );
    }
  }
}
