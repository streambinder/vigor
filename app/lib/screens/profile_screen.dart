import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/auth_provider.dart';
import '../widgets/adaptive/adaptive.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';
import '../models/goal.dart';
import '../models/injury.dart';
import '../models/gym.dart';
import '../services/gym_service.dart';
import '../services/secure_storage_service.dart';
import '../services/preferences_service.dart';
import '../widgets/gym_form_dialog.dart';
import 'profile_completion_modal.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key});

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  GymService? _gymService;
  PreferencesService? _prefsService;
  List<Gym>? _gyms;
  bool _isLoadingGyms = false;

  @override
  void initState() {
    super.initState();
    // Get the storage service from provider context
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final storage = context.read<SecureStorageService>();
      _gymService = GymService(storageService: storage);
      _prefsService = context.read<PreferencesService>();
      _loadGyms();
    });
  }

  Future<void> _loadGyms() async {
    if (_gymService == null) return;

    setState(() {
      _isLoadingGyms = true;
    });

    final response = await _gymService!.getGyms();
    if (response.isSuccess && mounted) {
      setState(() {
        _gyms = response.data;
        _isLoadingGyms = false;
      });
    } else if (mounted) {
      setState(() {
        _isLoadingGyms = false;
      });
      AdaptiveNotification.showError(
        context: context,
        message: response.error ?? 'Failed to load gyms',
      );
    }
  }

  Future<void> _showGymDialog({Gym? gym}) async {
    final result = await showDialog<Map<String, dynamic>>(
      context: context,
      builder: (context) => GymFormDialog(gym: gym),
    );

    if (result != null) {
      if (gym == null) {
        await _addGym(result['name'], result['equipment']);
      } else {
        await _updateGym(gym.name, result['name'], result['equipment']);
      }
    }
  }

  Future<void> _addGym(String name, List<String> equipment) async {
    if (_gymService == null) return;

    final response = await _gymService!.createGym(
      name: name,
      equipment: equipment,
    );

    if (response.isSuccess && mounted) {
      AdaptiveNotification.show(
        context: context,
        message: 'Gym added successfully',
      );
      await _loadGyms();
    } else if (mounted) {
      AdaptiveNotification.showError(
        context: context,
        message: response.error ?? 'Failed to add gym',
      );
    }
  }

  Future<void> _updateGym(String currentName, String newName, List<String> equipment) async {
    if (_gymService == null) return;

    final response = await _gymService!.updateGym(
      currentName: currentName,
      newName: newName != currentName ? newName : null,
      equipment: equipment,
    );

    if (response.isSuccess && mounted) {
      AdaptiveNotification.show(
        context: context,
        message: 'Gym updated successfully',
      );
      await _loadGyms();
    } else if (mounted) {
      AdaptiveNotification.showError(
        context: context,
        message: response.error ?? 'Failed to update gym',
      );
    }
  }

  Future<void> _deleteGym(String name) async {
    if (_gymService == null) return;
    final shouldDelete = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: 'Delete Gym',
      content: 'Are you sure you want to delete "$name"?',
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

    if (shouldDelete == true) {
      final response = await _gymService!.deleteGym(name);

      if (response.isSuccess && mounted) {
        await _prefsService?.clearDefaultGymIfMatches(name);
        AdaptiveNotification.show(
          context: context,
          message: 'Gym deleted successfully',
        );
        await _loadGyms();
      } else if (mounted) {
        AdaptiveNotification.showError(
          context: context,
          message: response.error ?? 'Failed to delete gym',
        );
      }
    }
  }

  Future<void> _toggleDefaultGym(String name) async {
    final current = _prefsService?.defaultGymName;
    final newDefault = current == name ? null : name;
    await _prefsService?.setDefaultGymName(newDefault);
    setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    final authProvider = context.watch<AuthProvider>();
    final user = authProvider.currentUser;

    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(
        title: const Text('Profile'),
        actions: [
          AdaptiveIconButton(
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
            onPressed: () async {
              await authProvider.refreshUserData();
              await _loadGyms();
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
                await _loadGyms();
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
                        leading: const Icon(Icons.person),
                        title: const Text('Gender'),
                        subtitle: Text(_capitalizeFirst(user.profile.gender)),
                      ),
                    ),
                    const SizedBox(height: 8),
                    AdaptiveCard(
                      child: AdaptiveListTile(
                        leading: const Icon(Icons.language),
                        title: const Text('Language'),
                        subtitle: Text(_capitalizeFirst(user.profile.language)),
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

                    // Goals section
                    if (_getGoals(user.profile.data).isNotEmpty) ...[
                      const SizedBox(height: 8),
                      AdaptiveCard(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Padding(
                              padding: const EdgeInsets.all(16.0),
                              child: Row(
                                children: [
                                  const Icon(Icons.flag),
                                  const SizedBox(width: 16),
                                  Text(
                                    'Goals',
                                    style: PlatformHelper.useLiquidGlass
                                        ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 16)
                                        : const TextStyle(
                                            fontSize: 16,
                                            fontWeight: FontWeight.bold,
                                          ),
                                  ),
                                ],
                              ),
                            ),
                            ..._getGoals(user.profile.data).map((goal) => Padding(
                                  padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
                                  child: Row(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      const Text('• ', style: TextStyle(fontSize: 18)),
                                      Expanded(
                                        child: Column(
                                          crossAxisAlignment: CrossAxisAlignment.start,
                                          children: [
                                            Text(goal.description),
                                            const SizedBox(height: 4),
                                            Text(
                                              'Started: ${_formatDate(goal.startDate)}',
                                              style: const TextStyle(
                                                fontSize: 12,
                                                color: Colors.grey,
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                    ],
                                  ),
                                )),
                            const SizedBox(height: 8),
                          ],
                        ),
                      ),
                    ],

                    // Injuries section
                    if (_getInjuries(user.profile.data).isNotEmpty) ...[
                      const SizedBox(height: 8),
                      AdaptiveCard(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Padding(
                              padding: const EdgeInsets.all(16.0),
                              child: Row(
                                children: [
                                  const Icon(Icons.healing),
                                  const SizedBox(width: 16),
                                  Text(
                                    'Injuries',
                                    style: PlatformHelper.useLiquidGlass
                                        ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 16)
                                        : const TextStyle(
                                            fontSize: 16,
                                            fontWeight: FontWeight.bold,
                                          ),
                                  ),
                                ],
                              ),
                            ),
                            ..._getInjuries(user.profile.data).map((injury) => Padding(
                                  padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
                                  child: Row(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      const Text('• ', style: TextStyle(fontSize: 18)),
                                      Expanded(
                                        child: Column(
                                          crossAxisAlignment: CrossAxisAlignment.start,
                                          children: [
                                            Text(injury.description),
                                            const SizedBox(height: 4),
                                            Text(
                                              'Year: ${injury.year}',
                                              style: const TextStyle(
                                                fontSize: 12,
                                                color: Colors.grey,
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                    ],
                                  ),
                                )),
                            const SizedBox(height: 8),
                          ],
                        ),
                      ),
                    ],

                    // Limitations section
                    if (_getLimitations(user.profile.data).isNotEmpty) ...[
                      const SizedBox(height: 8),
                      AdaptiveCard(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Padding(
                              padding: const EdgeInsets.all(16.0),
                              child: Row(
                                children: [
                                  const Icon(Icons.warning_amber),
                                  const SizedBox(width: 16),
                                  Text(
                                    'Limitations',
                                    style: PlatformHelper.useLiquidGlass
                                        ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 16)
                                        : const TextStyle(
                                            fontSize: 16,
                                            fontWeight: FontWeight.bold,
                                          ),
                                  ),
                                ],
                              ),
                            ),
                            ..._getLimitations(user.profile.data).map((limitation) => Padding(
                                  padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
                                  child: Row(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      const Text('• ', style: TextStyle(fontSize: 18)),
                                      Expanded(child: Text(limitation)),
                                    ],
                                  ),
                                )),
                            const SizedBox(height: 8),
                          ],
                        ),
                      ),
                    ],

                    const SizedBox(height: 24),

                    // Gyms section
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          'My Gyms',
                          style: PlatformHelper.useLiquidGlass
                              ? LiquidGlassTheme.headlineStyle
                              : const TextStyle(
                                  fontSize: 18,
                                  fontWeight: FontWeight.bold,
                                ),
                        ),
                        AdaptiveIconButton(
                          icon: const Icon(Icons.add),
                          tooltip: 'Add Gym',
                          onPressed: () => _showGymDialog(),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    if (_isLoadingGyms)
                      const Center(child: AdaptiveLoadingIndicator())
                    else if (_gyms == null || _gyms!.isEmpty)
                      AdaptiveCard(
                        child: Padding(
                          padding: const EdgeInsets.all(16.0),
                          child: Center(
                            child: Column(
                              children: [
                                Icon(
                                  Icons.fitness_center,
                                  size: 48,
                                  color: Colors.grey[400],
                                ),
                                const SizedBox(height: 8),
                                Text(
                                  'No gyms added yet',
                                  style: TextStyle(
                                    color: Colors.grey[600],
                                    fontStyle: FontStyle.italic,
                                  ),
                                ),
                                const SizedBox(height: 8),
                                TextButton.icon(
                                  onPressed: () => _showGymDialog(),
                                  icon: const Icon(Icons.add),
                                  label: const Text('Add Your First Gym'),
                                ),
                              ],
                            ),
                          ),
                        ),
                      )
                    else
                      ...(_gyms!.map((gym) {
                        final isDefault = _prefsService?.defaultGymName == gym.name;
                        return Padding(
                            padding: const EdgeInsets.only(bottom: 8.0),
                            child: AdaptiveCard(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Padding(
                                    padding: const EdgeInsets.all(16.0),
                                    child: Row(
                                      children: [
                                        const Icon(Icons.fitness_center),
                                        const SizedBox(width: 16),
                                        Expanded(
                                          child: Text(
                                            gym.name,
                                            style: PlatformHelper.useLiquidGlass
                                                ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 16)
                                                : const TextStyle(
                                                    fontSize: 16,
                                                    fontWeight: FontWeight.bold,
                                                  ),
                                          ),
                                        ),
                                        AdaptiveIconButton(
                                          icon: Icon(
                                            isDefault ? Icons.star : Icons.star_border,
                                            size: 20,
                                            color: isDefault ? Colors.amber : null,
                                          ),
                                          tooltip: isDefault ? 'Remove Default' : 'Set as Default',
                                          onPressed: () => _toggleDefaultGym(gym.name),
                                        ),
                                        AdaptiveIconButton(
                                          icon: const Icon(Icons.edit, size: 20),
                                          tooltip: 'Edit',
                                          onPressed: () => _showGymDialog(gym: gym),
                                        ),
                                        AdaptiveIconButton(
                                          icon: const Icon(Icons.delete, size: 20),
                                          tooltip: 'Delete',
                                          onPressed: () => _deleteGym(gym.name),
                                        ),
                                      ],
                                    ),
                                  ),
                                  if (gym.equipment.isNotEmpty)
                                    Padding(
                                      padding: const EdgeInsets.only(
                                        left: 16.0,
                                        right: 16.0,
                                        bottom: 16.0,
                                      ),
                                      child: Wrap(
                                        spacing: 8,
                                        runSpacing: 8,
                                        children: gym.equipment.map((equipment) {
                                          return Chip(
                                            label: Text(
                                              equipment,
                                              style: const TextStyle(fontSize: 12),
                                            ),
                                            padding: const EdgeInsets.symmetric(
                                              horizontal: 8,
                                              vertical: 4,
                                            ),
                                            backgroundColor: PlatformHelper.useLiquidGlass
                                                ? LiquidGlassTheme.primaryColor.withOpacity(0.2)
                                                : Theme.of(context).colorScheme.secondaryContainer,
                                          );
                                        }).toList(),
                                      ),
                                    ),
                                ],
                              ),
                            ),
                          );
                      })),

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
                              showDialog(
                                context: context,
                                barrierDismissible: true,
                                builder: (context) => ProfileCompletionModal(
                                  profile: user.profile,
                                  missingFields: const {},
                                ),
                              );
                            },
                          ),
                          if (!PlatformHelper.useLiquidGlass)
                            const Divider(height: 1),
                          AdaptiveListTile(
                            leading: Icon(
                              Icons.settings,
                              color: Colors.grey.withOpacity(0.5),
                            ),
                            title: Text(
                              'Settings',
                              style: TextStyle(color: Colors.grey.withOpacity(0.5)),
                            ),
                            onTap: null,
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

  String _capitalizeFirst(String text) {
    if (text.isEmpty) return text;
    return text[0].toUpperCase() + text.substring(1);
  }

  List<Goal> _getGoals(Map<String, dynamic> data) {
    try {
      if (data['goals'] != null) {
        return (data['goals'] as List).map((g) => Goal.fromJson(g)).toList();
      }
    } catch (e) {
      // Ignore parsing errors
    }
    return [];
  }

  List<Injury> _getInjuries(Map<String, dynamic> data) {
    try {
      if (data['injuries'] != null) {
        return (data['injuries'] as List).map((i) => Injury.fromJson(i)).toList();
      }
    } catch (e) {
      // Ignore parsing errors
    }
    return [];
  }

  List<String> _getLimitations(Map<String, dynamic> data) {
    try {
      if (data['limitations'] != null) {
        return (data['limitations'] as List).cast<String>();
      }
    } catch (e) {
      // Ignore parsing errors
    }
    return [];
  }
}
