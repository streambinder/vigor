import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:provider/provider.dart';
import '../config/api_config.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../providers/auth_provider.dart';
import '../widgets/adaptive/adaptive.dart';
import '../models/injury.dart';
import '../models/gym.dart';
import '../models/user.dart';
import '../services/service_locator.dart';
import '../services/preferences_service.dart';
import '../widgets/gym_form_dialog.dart';
import '../utils/knowledge_labels.dart';
import 'profile_edit_screen.dart';
import 'settings_screen.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key});

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  int _avatarVersion = 0;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final locator = context.read<ServiceLocator>();
      if (locator.gymsNotifier.value == null) {
        locator.refreshGyms();
      }
    });
  }

  Future<void> _pickAndUploadAvatar() async {
    final picked = await ImagePicker().pickImage(source: ImageSource.gallery, maxWidth: 256, maxHeight: 256);
    if (picked == null) return;

    final bytes = await picked.readAsBytes();
    final response = await context.read<ServiceLocator>().userService.uploadAvatar(bytes, picked.name);
    if (!mounted) return;
    if (response.isSuccess) {
      setState(() => _avatarVersion++);
      AdaptiveNotification.show(context: context, message: 'Avatar updated');
    } else {
      AdaptiveNotification.showError(context: context, message: response.error ?? 'Upload failed');
    }
  }

  Future<void> _showGymDialog({Gym? gym}) async {
    final result = await showDialog<Map<String, dynamic>>(
      context: context,
      barrierDismissible: true,
      builder: (context) => GymFormDialog(gym: gym),
    );

    if (result != null) {
      final modifierVariants = result['modifier_variants'] as Map<String, List<double>>?;
      if (gym == null) {
        await _addGym(result['name'], result['equipment'], modifierVariants);
      } else {
        await _updateGym(gym.id, result['name'], result['equipment'], modifierVariants);
      }
    }
  }

  Future<void> _addGym(String name, List<String> equipment, Map<String, List<double>>? modifierVariants) async {
    final l10n = AppLocalizations.of(context);
    final response = await context.read<ServiceLocator>().gymService.createGym(name: name, equipment: equipment, modifierVariants: modifierVariants);
    if (mounted) {
      if (response.isSuccess) {
        AdaptiveNotification.show(context: context, message: l10n.gymAddedSuccessfully);
      } else {
        AdaptiveNotification.showError(context: context, message: l10n.failedToAddGym, rawError: response.error);
      }
    }
  }

  Future<void> _updateGym(String id, String name, List<String> equipment, Map<String, List<double>>? modifierVariants) async {
    final l10n = AppLocalizations.of(context);
    final response = await context.read<ServiceLocator>().gymService.updateGym(id: id, name: name, equipment: equipment, modifierVariants: modifierVariants);
    if (mounted) {
      if (response.isSuccess) {
        AdaptiveNotification.show(context: context, message: l10n.gymUpdatedSuccessfully);
      } else {
        AdaptiveNotification.showError(context: context, message: l10n.failedToUpdateGym, rawError: response.error);
      }
    }
  }

  Future<void> _deleteGym(Gym gym) async {
    final l10n = AppLocalizations.of(context);
    final gymService = context.read<ServiceLocator>().gymService;
    final prefsService = context.read<PreferencesService>();
    final shouldDelete = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: l10n.deleteGym,
      content: l10n.deleteGymConfirmation(gym.name),
      actions: [
        AdaptiveDialogAction(label: l10n.cancel, onPressed: () => Navigator.of(context).pop(false)),
        AdaptiveDialogAction(label: l10n.delete, isDestructive: true, onPressed: () => Navigator.of(context).pop(true)),
      ],
    );
    if (!mounted) return;

    if (shouldDelete == true) {
      final response = await gymService.deleteGym(gym.id);
      if (!mounted) return;
      if (response.isSuccess) {
        await prefsService.clearDefaultGymIfMatches(gym.id);
        if (!mounted) return;
        AdaptiveNotification.show(context: context, message: l10n.gymDeletedSuccessfully);
      } else {
        AdaptiveNotification.showError(context: context, message: l10n.failedToDeleteGym, rawError: response.error);
      }
    }
  }

  Future<void> _toggleDefaultGym(String id) async {
    final prefs = context.read<PreferencesService>();
    final current = prefs.defaultGymId;
    await prefs.setDefaultGymId(current == id ? null : id);
    setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    final authProvider = context.watch<AuthProvider>();
    final user = authProvider.currentUser;
    final l10n = AppLocalizations.of(context);

    final locator = context.read<ServiceLocator>();

    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(
        automaticallyImplyLeading: false,
        title: Text(l10n.profile),
        actions: [
          AdaptiveIconButton(
            icon: const Icon(Icons.refresh),
            tooltip: l10n.refresh,
            onPressed: () async {
              await authProvider.refreshUserData();
              await locator.refreshGyms();
              if (context.mounted) {
                AdaptiveNotification.show(context: context, message: l10n.userDataRefreshed, duration: const Duration(seconds: 2));
              }
            },
          ),
          AdaptiveIconButton(
            icon: const Icon(Icons.logout),
            tooltip: l10n.logout,
            onPressed: () => _showLogoutDialog(context, l10n),
          ),
        ],
      ),
      body: user == null
          ? const Center(child: AdaptiveLoadingIndicator())
          : RefreshIndicator(
              onRefresh: () async {
                await authProvider.refreshUserData();
                await locator.refreshGyms();
              },
              color: VigorColors.persimmon,
              child: ListView.builder(
                padding: VigorSpacing.paddingLg,
                itemCount: 5,
                itemBuilder: (context, index) {
                  switch (index) {
                    case 0:
                      return Padding(
                        padding: const EdgeInsets.only(bottom: VigorSpacing.xl),
                        child: _buildProfileHeader(user, l10n),
                      );
                    case 1:
                      return Padding(
                        padding: const EdgeInsets.only(bottom: VigorSpacing.xl),
                        child: _buildDataSections(user, l10n),
                      );
                    case 2:
                      return Padding(
                        padding: const EdgeInsets.only(bottom: VigorSpacing.xl),
                        child: _buildGymsSection(l10n),
                      );
                    case 3:
                      return Padding(
                        padding: const EdgeInsets.only(bottom: VigorSpacing.lg),
                        child: _buildOtherSection(l10n),
                      );
                    default:
                      return const SizedBox.shrink();
                  }
                },
              ),
            ),
    );
  }

  Widget _buildProfileHeader(User user, AppLocalizations l10n) {
    final now = DateTime.now();
    final birth = user.profile.birthdate;
    int age = now.year - birth.year;
    if (now.month < birth.month || (now.month == birth.month && now.day < birth.day)) {
      age--;
    }

    return Container(
      padding: VigorSpacing.paddingLg,
      decoration: BoxDecoration(
        color: VigorColors.surface(context),
        borderRadius: VigorRadius.radiusLg,
      ),
      child: Column(
        children: [
          Row(
            children: [
              // avatar with solid stone border (no gradient ring)
              GestureDetector(
                onTap: _pickAndUploadAvatar,
                child: Container(
                  padding: const EdgeInsets.all(2),
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    border: Border.all(color: VigorColors.stone, width: 2),
                  ),
                  child: CachedNetworkImage(
                    imageUrl: '${ApiConfig.avatarUrl(user.id)}?v=$_avatarVersion',
                    imageBuilder: (context, imageProvider) => CircleAvatar(
                      radius: 36,
                      backgroundImage: imageProvider,
                    ),
                    placeholder: (context, url) => CircleAvatar(
                      radius: 36,
                      backgroundColor: VigorColors.surface(context),
                      child: const AdaptiveLoadingIndicator(),
                    ),
                    errorWidget: (context, url, error) => CircleAvatar(
                      radius: 36,
                      backgroundColor: VigorColors.surface(context),
                      child: Icon(Icons.person, size: 40, color: VigorColors.stone),
                    ),
                  ),
                ),
              ),
              const SizedBox(width: VigorSpacing.md),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      '${user.profile.firstName} ${user.profile.lastName}',
                      style: VigorTypography.headline.copyWith(
                        fontSize: 20,
                        color: VigorColors.textPrimary(context),
                      ),
                    ),
                    const SizedBox(height: VigorSpacing.xs),
                    Text(
                      user.email,
                      style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context)),
                    ),
                    const SizedBox(height: VigorSpacing.sm),
                    Wrap(
                      spacing: VigorSpacing.xs,
                      runSpacing: VigorSpacing.xs,
                      children: [
                        _buildStatPill(Icons.cake, '$age'),
                        _buildStatPill(Icons.height, '${user.profile.height.toInt()} cm'),
                        _buildStatPill(Icons.monitor_weight, '${user.profile.weight.toInt()} kg'),
                        _buildStatPill(
                          user.profile.gender == 'male' ? Icons.male : Icons.female,
                          user.profile.gender == 'male' ? l10n.male : l10n.female,
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              // edit button: stone, not persimmon (not a primary CTA)
              GestureDetector(
                onTap: () async {
                  final result = await Navigator.of(context).push<bool>(
                    MaterialPageRoute(builder: (_) => ProfileEditScreen(profile: user.profile)),
                  );
                  if (result == true && mounted) {
                    context.read<AuthProvider>().refreshUserData();
                  }
                },
                child: Container(
                  padding: VigorSpacing.paddingSm,
                  decoration: BoxDecoration(
                    color: VigorColors.stone.withValues(alpha: 0.15),
                    borderRadius: VigorRadius.radiusFull,
                  ),
                  child: const Icon(Icons.edit, color: VigorColors.stone, size: 20),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildQuickStats(User user, AppLocalizations l10n) {
    final now = DateTime.now();
    final birth = user.profile.birthdate;
    int age = now.year - birth.year;
    // adjust if birthday hasn't occurred yet this year
    if (now.month < birth.month || (now.month == birth.month && now.day < birth.day)) {
      age--;
    }
    // kanso: all stat pills use stone for visual calm
    return Wrap(
      spacing: VigorSpacing.sm,
      runSpacing: VigorSpacing.sm,
      alignment: WrapAlignment.center,
      children: [
        _buildStatPill(Icons.cake, '$age'),
        _buildStatPill(Icons.height, '${user.profile.height.toInt()} cm'),
        _buildStatPill(Icons.monitor_weight, '${user.profile.weight.toInt()} kg'),
        _buildStatPill(
          user.profile.gender == 'male' ? Icons.male : Icons.female,
          user.profile.gender == 'male' ? l10n.male : l10n.female,
        ),
      ],
    );
  }

  Widget _buildStatPill(IconData icon, String value) {
    // using VigorTypography.data for numeric stats per identity.md
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
      decoration: BoxDecoration(
        color: VigorColors.stone.withValues(alpha: 0.1),
        borderRadius: VigorRadius.radiusFull,
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, color: VigorColors.stone, size: 14),
          const SizedBox(width: VigorSpacing.xs),
          Text(value, style: VigorTypography.data.copyWith(color: VigorColors.stone, fontSize: 12)),
        ],
      ),
    );
  }

  Widget _buildDataSections(User user, AppLocalizations l10n) {
    final goals = _getGoals(user.profile.data);
    final injuries = _getInjuries(user.profile.data);
    final limitations = _getLimitations(user.profile.data);
    final favExercises = _getFavoriteExercises(user.profile.data);
    final favEquipment = _getFavoriteEquipment(user.profile.data);

    return Column(
      children: [
        if (goals.isNotEmpty)
          _buildCollapsibleSection(
            icon: Icons.flag,
            title: l10n.goals,
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: goals.map((g) {
                  final desc = KnowledgeLabels.goalDescription(g, l10n);
                  return Padding(
                    padding: const EdgeInsets.symmetric(vertical: VigorSpacing.xs),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(KnowledgeLabels.goalLabel(g, l10n), style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
                        if (desc.isNotEmpty)
                          Text(desc, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context))),
                      ],
                    ),
                  );
                }).toList(),
              ),
            ],
          ),
        if (injuries.isNotEmpty) ...[
          const SizedBox(height: VigorSpacing.md),
          _buildCollapsibleSection(
            icon: Icons.healing,
            title: l10n.injuries,
            children: injuries.map((i) => _buildListItem(i.description, subtitle: l10n.yearLabel(i.year))).toList(),
          ),
        ],
        if (limitations.isNotEmpty) ...[
          const SizedBox(height: VigorSpacing.md),
          _buildCollapsibleSection(
            icon: Icons.warning_amber,
            title: l10n.limitations,
            children: [_buildChipWrap(limitations)],
          ),
        ],
        if (favExercises.isNotEmpty || favEquipment.isNotEmpty) ...[
          const SizedBox(height: VigorSpacing.md),
          _buildCollapsibleSection(
            icon: Icons.favorite,
            title: l10n.favorites,
            children: [
              if (favExercises.isNotEmpty) ...[
                Align(
                  alignment: Alignment.centerLeft,
                  child: Padding(
                    padding: const EdgeInsets.only(bottom: VigorSpacing.sm),
                    child: Text(l10n.exercises, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context))),
                  ),
                ),
                _buildChipWrap(favExercises),
              ],
              if (favEquipment.isNotEmpty) ...[
                const SizedBox(height: VigorSpacing.md),
                Align(
                  alignment: Alignment.centerLeft,
                  child: Padding(
                    padding: const EdgeInsets.only(bottom: VigorSpacing.sm),
                    child: Text(l10n.equipment, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context))),
                  ),
                ),
                _buildChipWrap(favEquipment),
              ],
            ],
          ),
        ],
      ],
    );
  }

  Widget _buildCollapsibleSection({
    required IconData icon,
    required String title,
    required List<Widget> children,
  }) {
    // seijaku: calm interface, stone for all section icons
    return Container(
      decoration: BoxDecoration(
        color: VigorColors.surface(context),
        borderRadius: VigorRadius.radiusMd,
      ),
      child: Theme(
        data: Theme.of(context).copyWith(dividerColor: Colors.transparent),
        child: ExpansionTile(
          leading: Container(
            padding: VigorSpacing.paddingSm,
            decoration: BoxDecoration(color: VigorColors.stone.withValues(alpha: 0.1), borderRadius: VigorRadius.radiusSm),
            child: Icon(icon, color: VigorColors.stone, size: 20),
          ),
          title: Text(title, style: VigorTypography.headline.copyWith(fontSize: 16, color: VigorColors.textPrimary(context))),
          childrenPadding: const EdgeInsets.only(left: VigorSpacing.md, right: VigorSpacing.md, bottom: VigorSpacing.md),
          children: children,
        ),
      ),
    );
  }

  Widget _buildListItem(String text, {String? subtitle}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: VigorSpacing.xs),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            margin: const EdgeInsets.only(top: 6, right: VigorSpacing.sm),
            width: 6,
            height: 6,
            decoration: const BoxDecoration(color: VigorColors.stone, shape: BoxShape.circle),
          ),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(text, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
                if (subtitle != null)
                  Text(subtitle, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context))),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildChipWrap(List<String> items) {
    // kanso: stone for all chips, no color explosion
    return Align(
      alignment: Alignment.centerLeft,
      child: Wrap(
        spacing: VigorSpacing.xs,
        runSpacing: VigorSpacing.xs,
        children: items.map((item) => Container(
          padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
          decoration: BoxDecoration(
            color: VigorColors.stone.withValues(alpha: 0.1),
            borderRadius: VigorRadius.radiusFull,
          ),
          child: Text(item, style: VigorTypography.caption.copyWith(color: VigorColors.stone)),
        )).toList(),
      ),
    );
  }

  Widget _buildGymsSection(AppLocalizations l10n) {
    final locator = context.read<ServiceLocator>();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // section header
        Row(
          children: [
            // stone icon, no gradient
            Icon(Icons.fitness_center, color: VigorColors.stone, size: 24),
            const SizedBox(width: VigorSpacing.sm),
            Expanded(
              child: Text(l10n.myGyms, style: VigorTypography.headline.copyWith(fontSize: 18, color: VigorColors.textPrimary(context))),
            ),
            // add gym button: persimmon is ok here, it's a primary CTA. solid color, no gradient (seijaku)
            GestureDetector(
              onTap: () => _showGymDialog(),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.sm),
                decoration: const BoxDecoration(
                  color: VigorColors.persimmon,
                  borderRadius: VigorRadius.radiusFull,
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.add, color: Colors.white, size: 16),
                    const SizedBox(width: VigorSpacing.xs),
                    Text(l10n.addGym, style: VigorTypography.label.copyWith(color: Colors.white)),
                  ],
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: VigorSpacing.md),
        // gyms list
        ValueListenableBuilder<List<Gym>?>(
          valueListenable: locator.gymsNotifier,
          builder: (context, gyms, _) {
            if (gyms == null) {
              return const Center(child: AdaptiveLoadingIndicator());
            }
            if (gyms.isEmpty) {
              return SizedBox(
                width: double.infinity,
                child: Container(
                  padding: VigorSpacing.paddingLg,
                  decoration: BoxDecoration(
                    color: VigorColors.surface(context),
                    borderRadius: VigorRadius.radiusMd,
                  ),
                  child: Column(
                    children: [
                      Icon(Icons.fitness_center, size: 48, color: VigorColors.stone.withValues(alpha: 0.5)),
                      const SizedBox(height: VigorSpacing.sm),
                      Text(l10n.noGymsAddedYet, style: VigorTypography.body.copyWith(color: VigorColors.textSecondary(context))),
                    ],
                  ),
                ),
              );
            }
            return Column(children: gyms.map((gym) => _buildGymCard(gym, l10n)).toList());
          },
        ),
      ],
    );
  }

  Widget _buildGymCard(Gym gym, AppLocalizations l10n) {
    final isDefault = context.read<PreferencesService>().defaultGymId == gym.id;
    return Container(
      margin: const EdgeInsets.only(bottom: VigorSpacing.md),
      decoration: BoxDecoration(
        color: VigorColors.surface(context),
        borderRadius: VigorRadius.radiusMd,
        // kintsugi: gold border for default gym only
        border: isDefault ? Border.all(color: VigorColors.gold, width: 1) : null,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: VigorSpacing.paddingMd,
            child: Row(
              children: [
                Container(
                  padding: VigorSpacing.paddingSm,
                  decoration: BoxDecoration(
                    color: VigorColors.stone.withValues(alpha: 0.1),
                    borderRadius: VigorRadius.radiusSm,
                  ),
                  child: const Icon(Icons.fitness_center, color: VigorColors.stone, size: 20),
                ),
                const SizedBox(width: VigorSpacing.md),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(gym.name, style: VigorTypography.headline.copyWith(fontSize: 16, color: VigorColors.textPrimary(context))),
                    ],
                  ),
                ),
                // action buttons: stone for inactive star, gold for active (kintsugi)
                _buildGymAction(isDefault ? Icons.star : Icons.star_border, isDefault ? VigorColors.gold : VigorColors.stone, () => _toggleDefaultGym(gym.id)),
                _buildGymAction(Icons.edit, VigorColors.stone, () => _showGymDialog(gym: gym)),
                _buildGymAction(Icons.delete, VigorColors.error, () => _deleteGym(gym)),
              ],
            ),
          ),
          if (gym.equipment.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(left: VigorSpacing.md, right: VigorSpacing.md, bottom: VigorSpacing.md),
              child: Wrap(
                spacing: VigorSpacing.xs,
                runSpacing: VigorSpacing.xs,
                children: gym.equipment.map((eq) => Container(
                  padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
                  decoration: BoxDecoration(
                    color: VigorColors.stone.withValues(alpha: 0.1),
                    borderRadius: VigorRadius.radiusFull,
                  ),
                  child: Text(eq, style: VigorTypography.caption.copyWith(color: VigorColors.stone)),
                )).toList(),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildGymAction(IconData icon, Color color, VoidCallback onTap) {
    return GestureDetector(
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.all(VigorSpacing.sm),
        child: Icon(icon, size: 20, color: color),
      ),
    );
  }

  Widget _buildSettingsButton(AppLocalizations l10n) {
    return GestureDetector(
      onTap: () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => const SettingsScreen())),
      child: Container(
        padding: VigorSpacing.paddingMd,
        decoration: BoxDecoration(
          color: VigorColors.surface(context),
          borderRadius: VigorRadius.radiusMd,
        ),
        child: Row(
          children: [
            Container(
              padding: VigorSpacing.paddingSm,
              decoration: BoxDecoration(color: VigorColors.stone.withValues(alpha: 0.1), borderRadius: VigorRadius.radiusSm),
              child: const Icon(Icons.settings, color: VigorColors.stone, size: 20),
            ),
            const SizedBox(width: VigorSpacing.md),
            Expanded(child: Text(l10n.settings, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context), fontWeight: FontWeight.w500))),
            const Icon(Icons.chevron_right, color: VigorColors.stone, size: 20),
          ],
        ),
      ),
    );
  }

  Widget _buildOtherSection(AppLocalizations l10n) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // section header
        Row(
          children: [
            Icon(Icons.more_horiz, color: VigorColors.stone, size: 24),
            const SizedBox(width: VigorSpacing.sm),
            Text(l10n.other, style: VigorTypography.headline.copyWith(fontSize: 18, color: VigorColors.textPrimary(context))),
          ],
        ),
        const SizedBox(height: VigorSpacing.md),
        _buildSettingsButton(l10n),
      ],
    );
  }

  Future<void> _showLogoutDialog(BuildContext context, AppLocalizations l10n) async {
    final shouldLogout = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: l10n.logout,
      content: l10n.logoutConfirmation,
      actions: [
        AdaptiveDialogAction(label: l10n.cancel, onPressed: () => Navigator.of(context).pop(false)),
        AdaptiveDialogAction(label: l10n.logout, isDestructive: true, onPressed: () => Navigator.of(context).pop(true)),
      ],
    );
    if (shouldLogout == true && context.mounted) {
      context.read<ServiceLocator>().clearServices();
      await context.read<AuthProvider>().logout();
    }
  }

  String _capitalizeFirst(String text) => text.isEmpty ? text : text[0].toUpperCase() + text.substring(1);

  List<String> _getGoals(Map<String, dynamic> data) {
    try {
      if (data['goals'] != null) return (data['goals'] as List).cast<String>();
    } catch (_) {}
    return [];
  }

  List<Injury> _getInjuries(Map<String, dynamic> data) {
    try {
      if (data['injuries'] != null) return (data['injuries'] as List).map((i) => Injury.fromJson(i)).toList();
    } catch (_) {}
    return [];
  }

  List<String> _getLimitations(Map<String, dynamic> data) {
    try {
      if (data['limitations'] != null) return (data['limitations'] as List).cast<String>();
    } catch (_) {}
    return [];
  }

  List<String> _getFavoriteExercises(Map<String, dynamic> data) {
    try {
      if (data['preferences'] != null) {
        final prefs = data['preferences'] as Map<String, dynamic>;
        if (prefs['exercises'] != null) return (prefs['exercises'] as List).cast<String>();
      }
    } catch (_) {}
    return [];
  }

  List<String> _getFavoriteEquipment(Map<String, dynamic> data) {
    try {
      if (data['preferences'] != null) {
        final prefs = data['preferences'] as Map<String, dynamic>;
        if (prefs['equipment'] != null) return (prefs['equipment'] as List).cast<String>();
      }
    } catch (_) {}
    return [];
  }
}
