import 'dart:ui';
import 'package:flutter/material.dart';
import '../../services/app_logger.dart';
import '../../utils/platform_helper.dart';
import '../../theme/liquid_glass_theme.dart';

/// Platform-adaptive notification that always appears above dialogs/modals.
/// Uses overlay with rootOverlay: true to ensure proper z-ordering.
class AdaptiveNotification {
  static void show({
    required BuildContext context,
    required String message,
    Duration duration = const Duration(seconds: 3),
    bool isError = false,
  }) {
    final overlay = Overlay.of(context, rootOverlay: true);
    late OverlayEntry overlayEntry;

    overlayEntry = OverlayEntry(
      builder: (ctx) => _NotificationWidget(
        message: message,
        isError: isError,
        useLiquidGlass: PlatformHelper.useLiquidGlass,
        onDismiss: () => overlayEntry.remove(),
      ),
    );

    overlay.insert(overlayEntry);
    Future.delayed(duration, () {
      if (overlayEntry.mounted) {
        overlayEntry.remove();
      }
    });
  }

  /// shows an error notification with optional raw error logging
  /// [message] is the user-friendly message to display
  /// [rawError] is the technical error to log to console (not shown to user)
  static void showError({
    required BuildContext context,
    required String message,
    String? rawError,
  }) {
    if (rawError != null && rawError.isNotEmpty) {
      AppLogger.error('[Error] $rawError');
    }
    show(context: context, message: message, isError: true);
  }
}

class _NotificationWidget extends StatelessWidget {
  final String message;
  final bool isError;
  final bool useLiquidGlass;
  final VoidCallback onDismiss;

  const _NotificationWidget({
    required this.message,
    required this.isError,
    required this.useLiquidGlass,
    required this.onDismiss,
  });

  @override
  Widget build(BuildContext context) {
    return Positioned(
      top: MediaQuery.of(context).padding.top + 16,
      left: 16,
      right: 16,
      child: Material(
        color: Colors.transparent,
        child: useLiquidGlass ? _buildLiquidGlass(context) : _buildMaterial(context),
      ),
    );
  }

  Widget _buildLiquidGlass(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(12),
      child: BackdropFilter(
        filter: ImageFilter.blur(sigmaX: 10, sigmaY: 10),
        child: Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: isError
                ? LiquidGlassTheme.errorColor.withOpacity(0.9)
                : Colors.white.withOpacity(0.9),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: Colors.white.withOpacity(0.3),
              width: 1,
            ),
            boxShadow: LiquidGlassTheme.softShadow,
          ),
          child: _buildContent(
            iconColor: isError ? Colors.white : LiquidGlassTheme.successColor,
            textColor: isError ? Colors.white : Colors.black87,
          ),
        ),
      ),
    );
  }

  Widget _buildMaterial(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isError ? colorScheme.error : colorScheme.inverseSurface,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.2),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: _buildContent(
        iconColor: isError ? colorScheme.onError : colorScheme.onInverseSurface,
        textColor: isError ? colorScheme.onError : colorScheme.onInverseSurface,
      ),
    );
  }

  Widget _buildContent({required Color iconColor, required Color textColor}) {
    return Row(
      children: [
        Icon(
          isError ? Icons.error_outline : Icons.check_circle_outline,
          color: iconColor,
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Text(
            message,
            style: TextStyle(
              color: textColor,
              fontSize: 15,
              fontWeight: FontWeight.w500,
            ),
          ),
        ),
      ],
    );
  }
}
