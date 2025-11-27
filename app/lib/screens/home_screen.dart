import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/auth_provider.dart';
import '../widgets/adaptive/adaptive.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final authProvider = context.watch<AuthProvider>();
    final user = authProvider.currentUser;

    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(
        title: const Text('Vigor'),
        actions: [
          AdaptiveIconButton(
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
            onPressed: () async {
              await authProvider.refreshUserData();
              if (context.mounted) {
                AdaptiveNotification.show(
                  context: context,
                  message: 'User data refreshed',
                  duration: const Duration(seconds: 2),
                );
              }
            },
          ),
          AdaptiveIconButton(
            icon: const Icon(Icons.logout),
            tooltip: 'Logout',
            onPressed: () async {
              final shouldLogout = await AdaptiveAlertDialog.show<bool>(
                context: context,
                title: 'Logout',
                content: 'Are you sure you want to logout?',
                actions: [
                  AdaptiveDialogAction(
                    label: 'Cancel',
                    onPressed: () => Navigator.of(context).pop(false),
                  ),
                  AdaptiveDialogAction(
                    label: 'Logout',
                    isDestructive: true,
                    onPressed: () => Navigator.of(context).pop(true),
                  ),
                ],
              );

              if (shouldLogout == true && context.mounted) {
                await context.read<AuthProvider>().logout();
              }
            },
          ),
        ],
      ),
      body: user == null
          ? const Center(child: AdaptiveLoadingIndicator())
          : RefreshIndicator(
              onRefresh: () async {
                await authProvider.refreshUserData();
              },
              child: SingleChildScrollView(
                physics: const AlwaysScrollableScrollPhysics(),
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Welcome section
                    AdaptiveCard(
                      child: Row(
                        children: [
                          CircleAvatar(
                            radius: 30,
                            backgroundColor: PlatformHelper.useLiquidGlass
                                ? LiquidGlassTheme.primaryColor.withOpacity(0.2)
                                : Theme.of(context).colorScheme.primaryContainer,
                            child: Icon(
                              Icons.person,
                              size: 35,
                              color: PlatformHelper.useLiquidGlass
                                  ? LiquidGlassTheme.primaryColor
                                  : Theme.of(context).colorScheme.primary,
                            ),
                          ),
                          const SizedBox(width: 16),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  'Welcome back!',
                                  style: PlatformHelper.useLiquidGlass
                                      ? LiquidGlassTheme.headlineStyle
                                      : const TextStyle(
                                          fontSize: 20,
                                          fontWeight: FontWeight.bold,
                                        ),
                                ),
                                const SizedBox(height: 4),
                                Text(
                                  user.email,
                                  style: PlatformHelper.useLiquidGlass
                                      ? LiquidGlassTheme.captionStyle
                                      : const TextStyle(
                                          fontSize: 14,
                                          color: Colors.grey,
                                        ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 24),

                    // User ID
                    Text(
                      'Account Information',
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.headlineStyle
                          : const TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                            ),
                    ),
                    const SizedBox(height: 8),
                    AdaptiveCard(
                      child: AdaptiveListTile(
                        leading: const Icon(Icons.fingerprint),
                        title: const Text('User ID'),
                        subtitle: Text(user.id),
                      ),
                    ),
                    const SizedBox(height: 8),
                    AdaptiveCard(
                      child: AdaptiveListTile(
                        leading: const Icon(Icons.email),
                        title: const Text('Email'),
                        subtitle: Text(user.email),
                      ),
                    ),

                    // Profile section
                    const SizedBox(height: 24),
                    Text(
                      'Profile',
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.headlineStyle
                          : const TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                            ),
                    ),
                    const SizedBox(height: 8),
                    AdaptiveCard(
                      child: AdaptiveListTile(
                        leading: const Icon(Icons.cake),
                        title: const Text('Birthdate'),
                        subtitle: Text(
                          _formatDate(user.profile.birthdate),
                        ),
                      ),
                    ),
                    const SizedBox(height: 8),
                    AdaptiveCard(
                      child: AdaptiveListTile(
                        leading: const Icon(Icons.language),
                        title: const Text('Language'),
                        subtitle: Text(user.profile.language),
                      ),
                    ),
                    const SizedBox(height: 8),
                    AdaptiveCard(
                      child: AdaptiveListTile(
                        leading: const Icon(Icons.height),
                        title: const Text('Height'),
                        subtitle: Text('${user.profile.height} cm'),
                      ),
                    ),
                    const SizedBox(height: 8),
                    AdaptiveCard(
                      child: AdaptiveListTile(
                        leading: const Icon(Icons.monitor_weight),
                        title: const Text('Weight'),
                        subtitle: Text('${user.profile.weight} kg'),
                      ),
                    ),

                    const SizedBox(height: 24),

                    // Quick actions
                    Text(
                      'Quick Actions',
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.headlineStyle
                          : const TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                            ),
                    ),
                    const SizedBox(height: 8),
                    AdaptiveCard(
                      child: Column(
                        children: [
                          AdaptiveListTile(
                            leading: const Icon(Icons.edit),
                            title: const Text('Edit Profile'),
                            onTap: () {
                              // TODO: Navigate to profile edit screen
                              AdaptiveNotification.show(
                                context: context,
                                message: 'Profile editing coming soon!',
                              );
                            },
                          ),
                          if (!PlatformHelper.useLiquidGlass)
                            const Divider(height: 1),
                          AdaptiveListTile(
                            leading: const Icon(Icons.fitness_center),
                            title: const Text('Start Training'),
                            onTap: () {
                              // TODO: Navigate to training screen
                              AdaptiveNotification.show(
                                context: context,
                                message: 'Training feature coming soon!',
                              );
                            },
                          ),
                          if (!PlatformHelper.useLiquidGlass)
                            const Divider(height: 1),
                          AdaptiveListTile(
                            leading: const Icon(Icons.settings),
                            title: const Text('Settings'),
                            onTap: () {
                              // TODO: Navigate to settings screen
                              AdaptiveNotification.show(
                                context: context,
                                message: 'Settings coming soon!',
                              );
                            },
                          ),
                        ],
                      ),
                    ),

                    const SizedBox(height: 24),

                    // Danger zone
                    Text(
                      'Danger Zone',
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.headlineStyle.copyWith(
                              color: LiquidGlassTheme.errorColor,
                            )
                          : const TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                              color: Colors.red,
                            ),
                    ),
                    const SizedBox(height: 8),
                    AdaptiveCard(
                      child: AdaptiveListTile(
                        leading: Icon(
                          Icons.delete_forever,
                          color: PlatformHelper.useLiquidGlass
                              ? LiquidGlassTheme.errorColor
                              : Colors.red,
                        ),
                        title: Text(
                          'Delete Account',
                          style: TextStyle(
                            color: PlatformHelper.useLiquidGlass
                                ? LiquidGlassTheme.errorColor
                                : Colors.red,
                          ),
                        ),
                        onTap: () async {
                          final shouldDelete =
                              await AdaptiveAlertDialog.show<bool>(
                            context: context,
                            title: 'Delete Account',
                            content:
                                'Are you sure you want to delete your account? This action cannot be undone.',
                            actions: [
                              AdaptiveDialogAction(
                                label: 'Cancel',
                                onPressed: () => Navigator.of(context).pop(false),
                              ),
                              AdaptiveDialogAction(
                                label: 'Delete',
                                isDestructive: true,
                                onPressed: () => Navigator.of(context).pop(true),
                              ),
                            ],
                          );

                          if (shouldDelete == true && context.mounted) {
                            final success = await authProvider.deleteAccount();
                            if (context.mounted) {
                              if (success) {
                                AdaptiveNotification.show(
                                  context: context,
                                  message: 'Account deleted successfully',
                                );
                              } else {
                                AdaptiveNotification.showError(
                                  context: context,
                                  message: authProvider.errorMessage ??
                                      'Failed to delete account',
                                );
                              }
                            }
                          }
                        },
                      ),
                    ),
                  ],
                ),
              ),
            ),
    );
  }

  String _formatDate(DateTime date) {
    return '${date.day.toString().padLeft(2, '0')}/${date.month.toString().padLeft(2, '0')}/${date.year}';
  }
}
