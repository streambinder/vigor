import 'dart:ui';
import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../services/app_logger.dart';
import '../../utils/platform_helper.dart';

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
      top: MediaQuery.of(context).padding.top + VigorSpacing.md,
      left: VigorSpacing.md,
      right: VigorSpacing.md,
      child: Material(
        color: Colors.transparent,
        child: useLiquidGlass ? _buildLiquidGlass(context) : _buildMaterial(context),
      ),
    );
  }

  Widget _buildLiquidGlass(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return ClipRRect(
      borderRadius: VigorRadius.radiusMd,
      child: BackdropFilter(
        filter: ImageFilter.blur(sigmaX: 10, sigmaY: 10),
        child: Container(
          padding: VigorSpacing.paddingMd,
          decoration: BoxDecoration(
            color: isError
                ? VigorColors.error.withValues(alpha: 0.9)
                : isDark
                    ? VigorColors.darkSurface.withValues(alpha: 0.9)
                    : Colors.white.withValues(alpha: 0.9),
            borderRadius: VigorRadius.radiusMd,
            border: Border.all(
              color: isDark
                  ? Colors.white.withValues(alpha: 0.1)
                  : Colors.white.withValues(alpha: 0.3),
              width: 1,
            ),
            boxShadow: VigorShadows.elevation2(context),
          ),
          child: _buildContent(
            iconColor: isError ? Colors.white : VigorColors.success,
            textColor: isError
                ? Colors.white
                : VigorColors.textPrimary(context),
          ),
        ),
      ),
    );
  }

  Widget _buildMaterial(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Container(
      padding: VigorSpacing.paddingMd,
      decoration: BoxDecoration(
        color: isError ? VigorColors.error : colorScheme.inverseSurface,
        borderRadius: VigorRadius.radiusMd,
        boxShadow: VigorShadows.elevation2(context),
      ),
      child: _buildContent(
        iconColor: isError ? Colors.white : colorScheme.onInverseSurface,
        textColor: isError ? Colors.white : colorScheme.onInverseSurface,
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
        const SizedBox(width: VigorSpacing.sm),
        Expanded(
          child: Text(
            message,
            style: VigorTypography.body.copyWith(
              color: textColor,
              fontWeight: FontWeight.w500,
            ),
          ),
        ),
      ],
    );
  }
}
