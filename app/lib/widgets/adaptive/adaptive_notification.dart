import 'dart:ui';
import 'package:flutter/material.dart';
import '../../utils/platform_helper.dart';
import '../../theme/liquid_glass_theme.dart';

/// Platform-adaptive notification/snackbar
class AdaptiveNotification {
  static void show({
    required BuildContext context,
    required String message,
    Duration duration = const Duration(seconds: 3),
    bool isError = false,
  }) {
    if (PlatformHelper.useLiquidGlass) {
      _showLiquidGlassNotification(
        context: context,
        message: message,
        duration: duration,
        isError: isError,
      );
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(message),
          duration: duration,
          backgroundColor: isError
              ? Theme.of(context).colorScheme.error
              : null,
        ),
      );
    }
  }

  static void _showLiquidGlassNotification({
    required BuildContext context,
    required String message,
    required Duration duration,
    required bool isError,
  }) {
    final overlay = Overlay.of(context);
    late OverlayEntry overlayEntry;

    overlayEntry = OverlayEntry(
      builder: (context) => Positioned(
        top: MediaQuery.of(context).padding.top + 16,
        left: 16,
        right: 16,
        child: Material(
          color: Colors.transparent,
          child: ClipRRect(
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
                child: Row(
                  children: [
                    if (isError)
                      const Icon(
                        Icons.error_outline,
                        color: Colors.white,
                      )
                    else
                      Icon(
                        Icons.check_circle_outline,
                        color: LiquidGlassTheme.successColor,
                      ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        message,
                        style: TextStyle(
                          color: isError ? Colors.white : Colors.black87,
                          fontSize: 15,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );

    overlay.insert(overlayEntry);

    Future.delayed(duration, () {
      overlayEntry.remove();
    });
  }

  static void showError({
    required BuildContext context,
    required String message,
  }) {
    show(context: context, message: message, isError: true);
  }
}
