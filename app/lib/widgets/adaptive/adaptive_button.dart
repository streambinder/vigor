import 'dart:ui';
import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../utils/platform_helper.dart';
import '../../theme/liquid_glass_theme.dart';

/// Platform-adaptive primary button
/// Uses Liquid Glass effect on iOS and FilledButton on other platforms
class AdaptiveButton extends StatelessWidget {
  final VoidCallback? onPressed;
  final Widget child;
  final bool isDestructive;
  final bool useGradient;
  final bool isSecondary;

  const AdaptiveButton({
    super.key,
    required this.onPressed,
    required this.child,
    this.isDestructive = false,
    this.useGradient = false,
    this.isSecondary = false,
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
        ? VigorColors.error
        : isSecondary
            ? VigorColors.indigo
            : VigorColors.persimmon;

    return GestureDetector(
      onTap: onPressed,
      child: ClipRRect(
        borderRadius: VigorRadius.button,
        child: BackdropFilter(
          filter: ImageFilter.blur(sigmaX: 10, sigmaY: 10),
          child: AnimatedContainer(
            duration: VigorAnimation.fast,
            padding: VigorSpacing.buttonPadding,
            decoration: useGradient
                ? LiquidGlassTheme.vibrantGradient(
                    colors: [color, color.withValues(alpha: 0.8)],
                    borderRadius: VigorRadius.sm,
                  )
                : BoxDecoration(
                    color: color.withValues(alpha: onPressed == null ? 0.5 : 0.9),
                    borderRadius: VigorRadius.button,
                    border: Border.all(
                      color: Colors.white.withValues(alpha: 0.2),
                      width: 1,
                    ),
                    boxShadow: onPressed != null
                        ? (isSecondary
                            ? VigorShadows.indigoGlow
                            : VigorShadows.persimmonGlow)
                        : null,
                  ),
            child: DefaultTextStyle(
              style: VigorTypography.label.copyWith(
                color: Colors.white,
                fontWeight: FontWeight.w600,
              ),
              child: child,
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildMaterialButton(BuildContext context) {
    if (isSecondary) {
      return FilledButton(
        onPressed: onPressed,
        style: FilledButton.styleFrom(
          backgroundColor: VigorColors.indigo,
          foregroundColor: Colors.white,
        ),
        child: child,
      );
    }
    return FilledButton(
      onPressed: onPressed,
      style: isDestructive
          ? FilledButton.styleFrom(
              backgroundColor: VigorColors.error,
              foregroundColor: Colors.white,
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
  final bool isSecondary;

  const AdaptiveTextButton({
    super.key,
    required this.onPressed,
    required this.child,
    this.isDestructive = false,
    this.isSecondary = false,
  });

  @override
  Widget build(BuildContext context) {
    if (PlatformHelper.useLiquidGlass) {
      final color = isDestructive
          ? VigorColors.error
          : isSecondary
              ? VigorColors.indigo
              : VigorColors.persimmon;

      return GestureDetector(
        onTap: onPressed,
        child: Padding(
          padding: const EdgeInsets.symmetric(
            horizontal: VigorSpacing.sm,
            vertical: VigorSpacing.sm,
          ),
          child: DefaultTextStyle(
            style: VigorTypography.label.copyWith(color: color),
            child: child,
          ),
        ),
      );
    }

    final color = isDestructive
        ? VigorColors.error
        : isSecondary
            ? VigorColors.indigoAdaptive(context)
            : VigorColors.indigoAdaptive(context);

    return TextButton(
      onPressed: onPressed,
      style: TextButton.styleFrom(foregroundColor: color),
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
          padding: VigorSpacing.paddingSm,
          child: IconTheme(
            data: IconThemeData(
              color: color ?? VigorColors.stone,
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
      color: color ?? VigorColors.stone,
    );
  }
}
