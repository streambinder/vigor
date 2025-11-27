import 'dart:ui';
import 'package:flutter/material.dart';
import '../../utils/platform_helper.dart';
import '../../theme/liquid_glass_theme.dart';

/// Platform-adaptive primary button
/// Uses Liquid Glass effect on iOS and FilledButton on other platforms
class AdaptiveButton extends StatelessWidget {
  final VoidCallback? onPressed;
  final Widget child;
  final bool isDestructive;
  final bool useGradient;

  const AdaptiveButton({
    super.key,
    required this.onPressed,
    required this.child,
    this.isDestructive = false,
    this.useGradient = false,
  });

  @override
  Widget build(BuildContext context) {
    if (PlatformHelper.useLiquidGlass) {
      return _buildLiquidGlassButton(context);
    }
    return _buildMaterialButton(context);
  }

  Widget _buildLiquidGlassButton(BuildContext context) {
    final color = isDestructive
        ? LiquidGlassTheme.errorColor
        : LiquidGlassTheme.primaryColor;

    return GestureDetector(
      onTap: onPressed,
      child: ClipRRect(
        borderRadius: BorderRadius.circular(14),
        child: BackdropFilter(
          filter: ImageFilter.blur(sigmaX: 10, sigmaY: 10),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
            decoration: useGradient
                ? LiquidGlassTheme.vibrantGradient(
                    colors: [
                      color,
                      color.withOpacity(0.8),
                    ],
                    borderRadius: 14,
                  )
                : BoxDecoration(
                    color: color.withOpacity(onPressed == null ? 0.5 : 0.9),
                    borderRadius: BorderRadius.circular(14),
                    border: Border.all(
                      color: Colors.white.withOpacity(0.2),
                      width: 1,
                    ),
                    boxShadow: onPressed != null
                        ? LiquidGlassTheme.glowShadow
                        : null,
                  ),
            child: DefaultTextStyle(
              style: const TextStyle(
                color: Colors.white,
                fontSize: 17,
                fontWeight: FontWeight.w600,
                letterSpacing: -0.4,
              ),
              child: child,
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildMaterialButton(BuildContext context) {
    return FilledButton(
      onPressed: onPressed,
      style: isDestructive
          ? FilledButton.styleFrom(
              backgroundColor: Theme.of(context).colorScheme.error,
              foregroundColor: Theme.of(context).colorScheme.onError,
            )
          : null,
      child: child,
    );
  }
}

/// Platform-adaptive text button
class AdaptiveTextButton extends StatelessWidget {
  final VoidCallback? onPressed;
  final Widget child;
  final bool isDestructive;

  const AdaptiveTextButton({
    super.key,
    required this.onPressed,
    required this.child,
    this.isDestructive = false,
  });

  @override
  Widget build(BuildContext context) {
    if (PlatformHelper.useLiquidGlass) {
      return GestureDetector(
        onTap: onPressed,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          child: DefaultTextStyle(
            style: TextStyle(
              color: isDestructive
                  ? LiquidGlassTheme.errorColor
                  : LiquidGlassTheme.primaryColor,
              fontSize: 17,
              fontWeight: FontWeight.w400,
              letterSpacing: -0.4,
            ),
            child: child,
          ),
        ),
      );
    }

    return TextButton(
      onPressed: onPressed,
      style: isDestructive
          ? TextButton.styleFrom(
              foregroundColor: Theme.of(context).colorScheme.error,
            )
          : null,
      child: child,
    );
  }
}

/// Platform-adaptive icon button
class AdaptiveIconButton extends StatelessWidget {
  final VoidCallback? onPressed;
  final Widget icon;
  final String? tooltip;
  final Color? color;

  const AdaptiveIconButton({
    super.key,
    required this.onPressed,
    required this.icon,
    this.tooltip,
    this.color,
  });

  @override
  Widget build(BuildContext context) {
    if (PlatformHelper.useLiquidGlass) {
      return GestureDetector(
        onTap: onPressed,
        child: Padding(
          padding: const EdgeInsets.all(8),
          child: IconTheme(
            data: IconThemeData(
              color: color ?? LiquidGlassTheme.primaryColor,
              size: 24,
            ),
            child: icon,
          ),
        ),
      );
    }

    return IconButton(
      onPressed: onPressed,
      icon: icon,
      tooltip: tooltip,
      color: color,
    );
  }
}
