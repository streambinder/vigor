import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
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
    return Dialog(
      backgroundColor: PlatformHelper.useLiquidGlass ? Colors.transparent : null,
      child: Container(
        constraints: const BoxConstraints(maxWidth: 400, maxHeight: 500),
        decoration: PlatformHelper.useLiquidGlass
            ? LiquidGlassTheme.glassDecoration(
                borderRadius: 16,
                border: Border.all(color: Colors.white.withOpacity(0.2), width: 1.5),
              )
            : BoxDecoration(
                color: Theme.of(context).colorScheme.surface,
                borderRadius: BorderRadius.circular(16),
              ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(16),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      widget.title,
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 20)
                          : Theme.of(context).textTheme.titleLarge,
                    ),
                    const SizedBox(height: 12),
                    AdaptiveTextField(
                      controller: _searchController,
                      placeholder: 'Search by name',
                      prefix: const Icon(Icons.search, size: 20),
                      onChanged: _filterUsers,
                    ),
                  ],
                ),
              ),
              const Divider(height: 1),
              Flexible(child: _buildContent()),
              const Divider(height: 1),
              Padding(
                padding: const EdgeInsets.all(12),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    AdaptiveTextButton(
                      onPressed: () => Navigator.of(context).pop(null),
                      child: const Text('Cancel'),
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

  Widget _buildContent() {
    if (_isLoading) {
      return const Padding(
        padding: EdgeInsets.all(32),
        child: AdaptiveLoadingIndicator(),
      );
    }

    if (_error != null) {
      return Padding(
        padding: const EdgeInsets.all(32),
        child: Text(
          _error!,
          style: TextStyle(color: Theme.of(context).colorScheme.error),
        ),
      );
    }

    if (_filteredUsers.isEmpty) {
      return Padding(
        padding: const EdgeInsets.all(32),
        child: Text(
          _users.isEmpty ? 'No users available' : 'No matching users',
          style: PlatformHelper.useLiquidGlass
              ? LiquidGlassTheme.captionStyle
              : Theme.of(context).textTheme.bodyMedium?.copyWith(color: Colors.grey),
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
            backgroundColor: PlatformHelper.useLiquidGlass
                ? LiquidGlassTheme.primaryColor.withOpacity(0.2)
                : Theme.of(context).colorScheme.primaryContainer,
            child: Text(
              user.firstName.isNotEmpty ? user.firstName[0].toUpperCase() : '?',
              style: TextStyle(
                color: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.primaryColor
                    : Theme.of(context).colorScheme.onPrimaryContainer,
              ),
            ),
          ),
          title: Text(
            user.displayName,
            style: PlatformHelper.useLiquidGlass ? LiquidGlassTheme.bodyStyle : null,
          ),
          onTap: () => Navigator.of(context).pop(user),
        );
      },
    );
  }
}
