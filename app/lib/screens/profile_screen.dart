import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../providers/auth_provider.dart';
import '../widgets/adaptive/adaptive.dart';
import '../models/goal.dart';
import '../models/injury.dart';
import '../models/gym.dart';
import '../services/service_locator.dart';
import '../services/preferences_service.dart';
import '../widgets/gym_form_dialog.dart';
import 'profile_completion_modal.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key});

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  List<Gym>? _gyms;
  bool _isLoadingGyms = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _loadGyms());
  }

  Future<void> _loadGyms() async {
    setState(() => _isLoadingGyms = true);

    final response = await context.read<ServiceLocator>().gymService.getGyms();
    if (response.isSuccess && mounted) {
      setState(() {
        _gyms = response.data;
        _isLoadingGyms = false;
      });
    } else if (mounted) {
      setState(() => _isLoadingGyms = false);
      AdaptiveNotification.showError(
        context: context,
        message: AppLocalizations.of(context).failedToLoadGyms,
        rawError: response.error,
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
        await _updateGym(gym.id, result['name'], result['equipment']);
      }
    }
  }

  Future<void> _addGym(String name, List<String> equipment) async {
    final l10n = AppLocalizations.of(context);

    final response = await context.read<ServiceLocator>().gymService.createGym(name: name, equipment: equipment);
    if (response.isSuccess && mounted) {
      AdaptiveNotification.show(context: context, message: l10n.gymAddedSuccessfully);
      await _loadGyms();
    } else if (mounted) {
      AdaptiveNotification.showError(context: context, message: l10n.failedToAddGym, rawError: response.error);
    }
  }

  Future<void> _updateGym(String id, String name, List<String> equipment) async {
    final l10n = AppLocalizations.of(context);

    final response = await context.read<ServiceLocator>().gymService.updateGym(id: id, name: name, equipment: equipment);
    if (response.isSuccess && mounted) {
      AdaptiveNotification.show(context: context, message: l10n.gymUpdatedSuccessfully);
      await _loadGyms();
    } else if (mounted) {
      AdaptiveNotification.showError(context: context, message: l10n.failedToUpdateGym, rawError: response.error);
    }
  }

  Future<void> _deleteGym(Gym gym) async {
    final l10n = AppLocalizations.of(context);
    final shouldDelete = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: l10n.deleteGym,
      content: l10n.deleteGymConfirmation(gym.name),
      actions: [
        AdaptiveDialogAction(label: l10n.cancel, onPressed: () => Navigator.of(context).pop(false)),
        AdaptiveDialogAction(label: l10n.delete, isDestructive: true, onPressed: () => Navigator.of(context).pop(true)),
      ],
    );

    if (shouldDelete == true) {
      final response = await context.read<ServiceLocator>().gymService.deleteGym(gym.id);
      if (response.isSuccess && mounted) {
        await context.read<PreferencesService>().clearDefaultGymIfMatches(gym.id);
        AdaptiveNotification.show(context: context, message: l10n.gymDeletedSuccessfully);
        await _loadGyms();
      } else if (mounted) {
        AdaptiveNotification.showError(context: context, message: l10n.failedToDeleteGym, rawError: response.error);
      }
    }
  }

  Future<void> _toggleDefaultGym(String id) async {
    final current = _prefsService?.defaultGymId;
    await _prefsService?.setDefaultGymId(current == id ? null : id);
    setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    final authProvider = context.watch<AuthProvider>();
    final user = authProvider.currentUser;
    final l10n = AppLocalizations.of(context);

    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(
        title: Text(l10n.profile),
        actions: [
          AdaptiveIconButton(
            icon: const Icon(Icons.refresh),
            tooltip: l10n.refresh,
            onPressed: () async {
              await authProvider.refreshUserData();
              await _loadGyms();
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
                await _loadGyms();
              },
              color: VigorColors.orange,
              child: ListView(
                padding: VigorSpacing.paddingLg,
                children: [
                  _buildProfileHeader(user, l10n),
                  SizedBox(height: VigorSpacing.lg),
                  _buildQuickStats(user, l10n),
                  SizedBox(height: VigorSpacing.xl),
                  _buildDataSections(user, l10n),
                  SizedBox(height: VigorSpacing.xl),
                  _buildGymsSection(l10n),
                  SizedBox(height: VigorSpacing.xl),
                  _buildDangerZone(l10n, authProvider),
                  SizedBox(height: VigorSpacing.xxl),
                ],
              ),
            ),
    );
  }

  Widget _buildProfileHeader(user, AppLocalizations l10n) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Container(
      padding: VigorSpacing.paddingLg,
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            VigorColors.orange.withValues(alpha: 0.15),
            VigorColors.electricBlue.withValues(alpha: 0.15),
          ],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: VigorRadius.radiusLg,
        border: Border.all(
          color: isDark ? Colors.white.withValues(alpha: 0.1) : Colors.black.withValues(alpha: 0.05),
        ),
      ),
      child: Row(
        children: [
          // avatar with gradient ring
          Container(
            padding: const EdgeInsets.all(3),
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              gradient: LinearGradient(colors: [VigorColors.orange, VigorColors.electricBlue]),
            ),
            child: CircleAvatar(
              radius: 36,
              backgroundColor: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
              child: ShaderMask(
                shaderCallback: (bounds) => LinearGradient(
                  colors: [VigorColors.orange, VigorColors.electricBlue],
                ).createShader(bounds),
                child: const Icon(Icons.person, size: 40, color: Colors.white),
              ),
            ),
          ),
          SizedBox(width: VigorSpacing.md),
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
                SizedBox(height: VigorSpacing.xs),
                Text(
                  user.email,
                  style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context)),
                ),
              ],
            ),
          ),
          // edit button
          GestureDetector(
            onTap: () => showDialog(
              context: context,
              barrierDismissible: true,
              builder: (context) => ProfileCompletionModal(profile: user.profile, missingFields: const {}),
            ),
            child: Container(
              padding: VigorSpacing.paddingSm,
              decoration: BoxDecoration(
                color: VigorColors.orange.withValues(alpha: 0.15),
                borderRadius: VigorRadius.radiusFull,
              ),
              child: Icon(Icons.edit, color: VigorColors.orange, size: 20),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildQuickStats(user, AppLocalizations l10n) {
    final age = DateTime.now().year - user.profile.birthdate.year;
    return Wrap(
      spacing: VigorSpacing.sm,
      runSpacing: VigorSpacing.sm,
      alignment: WrapAlignment.center,
      children: [
        _buildStatPill(Icons.cake, '$age', VigorColors.orange),
        _buildStatPill(Icons.height, '${user.profile.height.toInt()} cm', VigorColors.electricBlue),
        _buildStatPill(Icons.monitor_weight, '${user.profile.weight.toInt()} kg', VigorColors.success),
        _buildStatPill(
          user.profile.gender == 'male' ? Icons.male : Icons.female,
          _capitalizeFirst(user.profile.gender),
          VigorColors.warning,
        ),
      ],
    );
  }

  Widget _buildStatPill(IconData icon, String value, Color color) {
    return Container(
      padding: EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.sm),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: VigorRadius.radiusFull,
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, color: color, size: 16),
          SizedBox(width: VigorSpacing.xs),
          Text(value, style: VigorTypography.label.copyWith(color: color, fontWeight: FontWeight.w600)),
        ],
      ),
    );
  }

  Widget _buildDataSections(user, AppLocalizations l10n) {
    final goals = _getGoals(user.profile.data);
    final injuries = _getInjuries(user.profile.data);
    final limitations = _getLimitations(user.profile.data);
    final favExercises = _getFavoriteExercises(user.profile.data);
    final favEquipment = _getFavoriteEquipment(user.profile.data);
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Column(
      children: [
        if (goals.isNotEmpty)
          _buildCollapsibleSection(
            icon: Icons.flag,
            title: l10n.goals,
            color: VigorColors.success,
            isDark: isDark,
            children: goals.map((g) => _buildListItem(g.description, subtitle: l10n.startedDate(_formatDate(g.startDate)))).toList(),
          ),
        if (injuries.isNotEmpty) ...[
          SizedBox(height: VigorSpacing.md),
          _buildCollapsibleSection(
            icon: Icons.healing,
            title: l10n.injuries,
            color: VigorColors.warning,
            isDark: isDark,
            children: injuries.map((i) => _buildListItem(i.description, subtitle: l10n.yearLabel(i.year))).toList(),
          ),
        ],
        if (limitations.isNotEmpty) ...[
          SizedBox(height: VigorSpacing.md),
          _buildCollapsibleSection(
            icon: Icons.warning_amber,
            title: l10n.limitations,
            color: VigorColors.error,
            isDark: isDark,
            children: limitations.map((lim) => _buildListItem(lim)).toList(),
          ),
        ],
        if (favExercises.isNotEmpty || favEquipment.isNotEmpty) ...[
          SizedBox(height: VigorSpacing.md),
          _buildCollapsibleSection(
            icon: Icons.favorite,
            title: l10n.favorites,
            color: VigorColors.electricBlue,
            isDark: isDark,
            children: [
              if (favExercises.isNotEmpty) ...[
                Padding(
                  padding: EdgeInsets.only(left: VigorSpacing.md, bottom: VigorSpacing.xs),
                  child: Text(l10n.exercises, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context))),
                ),
                ...favExercises.map((e) => _buildListItem(e)),
              ],
              if (favEquipment.isNotEmpty) ...[
                SizedBox(height: VigorSpacing.sm),
                Padding(
                  padding: EdgeInsets.only(left: VigorSpacing.md, bottom: VigorSpacing.xs),
                  child: Text(l10n.equipment, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context))),
                ),
                ...favEquipment.map((e) => _buildListItem(e)),
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
    required Color color,
    required bool isDark,
    required List<Widget> children,
  }) {
    return Container(
      decoration: BoxDecoration(
        color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
        borderRadius: VigorRadius.radiusMd,
        border: Border.all(color: color.withValues(alpha: 0.2)),
      ),
      child: Theme(
        data: Theme.of(context).copyWith(dividerColor: Colors.transparent),
        child: ExpansionTile(
          leading: Container(
            padding: VigorSpacing.paddingSm,
            decoration: BoxDecoration(color: color.withValues(alpha: 0.15), borderRadius: VigorRadius.radiusSm),
            child: Icon(icon, color: color, size: 20),
          ),
          title: Text(title, style: VigorTypography.headline.copyWith(fontSize: 16, color: VigorColors.textPrimary(context))),
          childrenPadding: EdgeInsets.only(left: VigorSpacing.md, right: VigorSpacing.md, bottom: VigorSpacing.md),
          children: children,
        ),
      ),
    );
  }

  Widget _buildListItem(String text, {String? subtitle}) {
    return Padding(
      padding: EdgeInsets.symmetric(vertical: VigorSpacing.xs),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            margin: EdgeInsets.only(top: 6, right: VigorSpacing.sm),
            width: 6,
            height: 6,
            decoration: BoxDecoration(color: VigorColors.orange, shape: BoxShape.circle),
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

  Widget _buildGymsSection(AppLocalizations l10n) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // section header
        Row(
          children: [
            ShaderMask(
              shaderCallback: (bounds) => LinearGradient(colors: [VigorColors.orange, VigorColors.electricBlue]).createShader(bounds),
              child: const Icon(Icons.fitness_center, color: Colors.white, size: 24),
            ),
            SizedBox(width: VigorSpacing.sm),
            Expanded(
              child: Text(l10n.myGyms, style: VigorTypography.headline.copyWith(fontSize: 18, color: VigorColors.textPrimary(context))),
            ),
            // add gym button
            GestureDetector(
              onTap: () => _showGymDialog(),
              child: Container(
                padding: EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.sm),
                decoration: BoxDecoration(
                  gradient: LinearGradient(colors: [VigorColors.orange, VigorColors.electricBlue]),
                  borderRadius: VigorRadius.radiusFull,
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.add, color: Colors.white, size: 16),
                    SizedBox(width: VigorSpacing.xs),
                    Text(l10n.addGym, style: VigorTypography.label.copyWith(color: Colors.white)),
                  ],
                ),
              ),
            ),
          ],
        ),
        SizedBox(height: VigorSpacing.md),
        // gyms list
        if (_isLoadingGyms)
          const Center(child: AdaptiveLoadingIndicator())
        else if (_gyms == null || _gyms!.isEmpty)
          Container(
            padding: VigorSpacing.paddingLg,
            decoration: BoxDecoration(
              color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
              borderRadius: VigorRadius.radiusMd,
              border: Border.all(color: isDark ? Colors.white.withValues(alpha: 0.1) : Colors.black.withValues(alpha: 0.05)),
            ),
            child: Column(
              children: [
                ShaderMask(
                  shaderCallback: (bounds) => LinearGradient(colors: [VigorColors.orange.withValues(alpha: 0.5), VigorColors.electricBlue.withValues(alpha: 0.5)]).createShader(bounds),
                  child: const Icon(Icons.fitness_center, size: 48, color: Colors.white),
                ),
                SizedBox(height: VigorSpacing.sm),
                Text(l10n.noGymsAddedYet, style: VigorTypography.body.copyWith(color: VigorColors.textSecondary(context))),
              ],
            ),
          )
        else
          ...(_gyms!.map((gym) => _buildGymCard(gym, l10n, isDark))),
      ],
    );
  }

  Widget _buildGymCard(Gym gym, AppLocalizations l10n, bool isDark) {
    final isDefault = _prefsService?.defaultGymId == gym.id;
    return Container(
      margin: EdgeInsets.only(bottom: VigorSpacing.md),
      decoration: BoxDecoration(
        color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
        borderRadius: VigorRadius.radiusMd,
        border: Border.all(
          color: isDefault ? VigorColors.orange.withValues(alpha: 0.5) : (isDark ? Colors.white.withValues(alpha: 0.1) : Colors.black.withValues(alpha: 0.05)),
          width: isDefault ? 2 : 1,
        ),
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
                    color: VigorColors.orange.withValues(alpha: 0.15),
                    borderRadius: VigorRadius.radiusSm,
                  ),
                  child: Icon(Icons.fitness_center, color: VigorColors.orange, size: 20),
                ),
                SizedBox(width: VigorSpacing.md),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(gym.name, style: VigorTypography.headline.copyWith(fontSize: 16, color: VigorColors.textPrimary(context))),
                      if (isDefault)
                        Container(
                          margin: EdgeInsets.only(top: VigorSpacing.xs),
                          padding: EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: 2),
                          decoration: BoxDecoration(color: VigorColors.orange, borderRadius: VigorRadius.radiusFull),
                          child: Text('Default', style: VigorTypography.caption.copyWith(color: Colors.white, fontWeight: FontWeight.w600)),
                        ),
                    ],
                  ),
                ),
                // action buttons
                _buildGymAction(isDefault ? Icons.star : Icons.star_border, isDefault ? Colors.amber : VigorColors.textSecondary(context), () => _toggleDefaultGym(gym.id)),
                _buildGymAction(Icons.edit, VigorColors.electricBlue, () => _showGymDialog(gym: gym)),
                _buildGymAction(Icons.delete, VigorColors.error, () => _deleteGym(gym)),
              ],
            ),
          ),
          if (gym.equipment.isNotEmpty)
            Padding(
              padding: EdgeInsets.only(left: VigorSpacing.md, right: VigorSpacing.md, bottom: VigorSpacing.md),
              child: Wrap(
                spacing: VigorSpacing.xs,
                runSpacing: VigorSpacing.xs,
                children: gym.equipment.map((eq) => Container(
                  padding: EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
                  decoration: BoxDecoration(
                    color: VigorColors.electricBlue.withValues(alpha: 0.15),
                    borderRadius: VigorRadius.radiusFull,
                  ),
                  child: Text(eq, style: VigorTypography.caption.copyWith(color: VigorColors.electricBlue)),
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
        padding: EdgeInsets.all(VigorSpacing.sm),
        child: Icon(icon, size: 20, color: color),
      ),
    );
  }

  Widget _buildDangerZone(AppLocalizations l10n, AuthProvider authProvider) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(Icons.warning, color: VigorColors.error, size: 20),
            SizedBox(width: VigorSpacing.sm),
            Text(l10n.dangerZone, style: VigorTypography.headline.copyWith(fontSize: 16, color: VigorColors.error)),
          ],
        ),
        SizedBox(height: VigorSpacing.sm),
        Container(
          decoration: BoxDecoration(
            color: VigorColors.error.withValues(alpha: 0.1),
            borderRadius: VigorRadius.radiusMd,
            border: Border.all(color: VigorColors.error.withValues(alpha: 0.3)),
          ),
          child: ListTile(
            leading: Icon(Icons.delete_forever, color: VigorColors.error),
            title: Text(l10n.deleteAccount, style: VigorTypography.body.copyWith(color: VigorColors.error, fontWeight: FontWeight.w500)),
            trailing: Icon(Icons.chevron_right, color: VigorColors.error),
            onTap: () => _showDeleteAccountDialog(l10n, authProvider),
          ),
        ),
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
      await context.read<AuthProvider>().logout();
    }
  }

  Future<void> _showDeleteAccountDialog(AppLocalizations l10n, AuthProvider authProvider) async {
    final shouldDelete = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: l10n.deleteAccount,
      content: l10n.deleteAccountConfirmation,
      actions: [
        AdaptiveDialogAction(label: l10n.cancel, onPressed: () => Navigator.of(context).pop(false)),
        AdaptiveDialogAction(label: l10n.delete, isDestructive: true, onPressed: () => Navigator.of(context).pop(true)),
      ],
    );

    if (shouldDelete == true && context.mounted) {
      final success = await authProvider.deleteAccount();
      if (context.mounted) {
        if (success) {
          AdaptiveNotification.show(context: context, message: l10n.accountDeletedSuccessfully);
        } else {
          AdaptiveNotification.showError(context: context, message: l10n.failedToDeleteAccount, rawError: authProvider.errorMessage);
        }
      }
    }
  }

  String _formatDate(DateTime date) => '${date.day.toString().padLeft(2, '0')}/${date.month.toString().padLeft(2, '0')}/${date.year}';

  String _capitalizeFirst(String text) => text.isEmpty ? text : text[0].toUpperCase() + text.substring(1);

  List<Goal> _getGoals(Map<String, dynamic> data) {
    try {
      if (data['goals'] != null) return (data['goals'] as List).map((g) => Goal.fromJson(g)).toList();
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
