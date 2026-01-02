import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../services/user_service.dart';
import '../services/secure_storage_service.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';
import 'adaptive/adaptive.dart';

/// shows a searchable user select dialog and returns the selected user id
Future<UserInfo?> showUserSelectDialog({
  required BuildContext context,
  String title = 'Select User',
  String? excludeUserId,
}) async {
  return showDialog<UserInfo>(
    context: context,
    builder: (ctx) => _UserSelectDialog(
      title: title,
      excludeUserId: excludeUserId,
    ),
  );
}

class _UserSelectDialog extends StatefulWidget {
  final String title;
  final String? excludeUserId;

  const _UserSelectDialog({required this.title, this.excludeUserId});

  @override
  State<_UserSelectDialog> createState() => _UserSelectDialogState();
}

class _UserSelectDialogState extends State<_UserSelectDialog> {
  final _searchController = TextEditingController();
  List<UserInfo> _users = [];
  List<UserInfo> _filteredUsers = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadUsers();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadUsers() async {
    final storage = context.read<SecureStorageService>();
    final userService = UserService(storageService: storage);
    final response = await userService.getUsers();

    if (mounted) {
      setState(() {
        _isLoading = false;
        if (response.isSuccess && response.data != null) {
          _users = response.data!;
          if (widget.excludeUserId != null) {
            _users = _users.where((u) => u.id != widget.excludeUserId).toList();
          }
          _filteredUsers = _users;
        } else {
          _error = response.error ?? 'Failed to load users';
        }
      });
    }
  }

  void _filterUsers(String query) {
    setState(() {
      if (query.isEmpty) {
        _filteredUsers = _users;
      } else {
        final lower = query.toLowerCase();
        _filteredUsers = _users.where((u) => u.displayName.toLowerCase().contains(lower)).toList();
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Dialog(
      backgroundColor: PlatformHelper.useLiquidGlass ? Colors.transparent : null,
      child: Container(
        constraints: const BoxConstraints(maxWidth: 400, maxHeight: 500),
        decoration: PlatformHelper.useLiquidGlass
            ? LiquidGlassTheme.glassDecoration(
                borderRadius: 16,
                isDark: isDark,
                border: Border.all(
                  color: isDark ? Colors.white.withValues(alpha: 0.2) : Colors.black.withValues(alpha: 0.1),
                  width: 1.5,
                ),
              )
            : BoxDecoration(
                color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
                borderRadius: VigorRadius.radiusLg,
              ),
        child: ClipRRect(
          borderRadius: VigorRadius.radiusLg,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Padding(
                padding: VigorSpacing.paddingMd,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      widget.title,
                      style: VigorTypography.headline.copyWith(
                        fontSize: 20,
                        color: VigorColors.textPrimary(context),
                      ),
                    ),
                    SizedBox(height: VigorSpacing.sm),
                    AdaptiveTextField(
                      controller: _searchController,
                      placeholder: l10n.searchByName,
                      prefix: const Icon(Icons.search, size: 20),
                      onChanged: _filterUsers,
                    ),
                  ],
                ),
              ),
              const Divider(height: 1),
              Flexible(child: _buildContent(l10n)),
              const Divider(height: 1),
              Padding(
                padding: VigorSpacing.paddingMd,
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    AdaptiveTextButton(
                      onPressed: () => Navigator.of(context).pop(null),
                      child: Text(l10n.cancel),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildContent(AppLocalizations l10n) {
    if (_isLoading) {
      return Padding(
        padding: VigorSpacing.paddingXl,
        child: const AdaptiveLoadingIndicator(),
      );
    }

    if (_error != null) {
      return Padding(
        padding: VigorSpacing.paddingXl,
        child: Text(
          _error!,
          style: TextStyle(color: VigorColors.error),
        ),
      );
    }

    if (_filteredUsers.isEmpty) {
      return Padding(
        padding: VigorSpacing.paddingXl,
        child: Text(
          _users.isEmpty ? l10n.noUsersAvailable : l10n.noMatchingUsers,
          style: VigorTypography.body.copyWith(
            color: VigorColors.textMuted(context),
          ),
        ),
      );
    }

    return ListView.builder(
      shrinkWrap: true,
      itemCount: _filteredUsers.length,
      itemBuilder: (ctx, index) {
        final user = _filteredUsers[index];
        return ListTile(
          leading: CircleAvatar(
            backgroundColor: VigorColors.orange.withValues(alpha: 0.2),
            child: Text(
              user.firstName.isNotEmpty ? user.firstName[0].toUpperCase() : '?',
              style: TextStyle(color: VigorColors.orange),
            ),
          ),
          title: Text(
            user.displayName,
            style: VigorTypography.body.copyWith(
              color: VigorColors.textPrimary(context),
            ),
          ),
          onTap: () => Navigator.of(context).pop(user),
        );
      },
    );
  }
}
