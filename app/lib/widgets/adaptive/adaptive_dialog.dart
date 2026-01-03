import 'dart:ui';
import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../utils/platform_helper.dart';
import '../../theme/liquid_glass_theme.dart';

/// Platform-adaptive alert dialog
/// Uses Liquid Glass effect on iOS and Material AlertDialog on other platforms
class AdaptiveAlertDialog extends StatelessWidget {
  final String? title;
  final String? content;
  final List<AdaptiveDialogAction> actions;

  const AdaptiveAlertDialog({
    super.key,
    this.title,
    this.content,
    required this.actions,
  });

  static Future<T?> show<T>({
    required BuildContext context,
    String? title,
    String? content,
    required List<AdaptiveDialogAction> actions,
  }) {
    return showDialog<T>(
      context: context,
      barrierDismissible: true,
      builder: (context) => AdaptiveAlertDialog(
        title: title,
        content: content,
        actions: actions,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    if (PlatformHelper.useLiquidGlass) {
      return _buildLiquidGlassDialog(context);
    }
    return _buildMaterialDialog(context);
  }

  Widget _buildLiquidGlassDialog(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final textColor = VigorColors.textPrimary(context);
    final secondaryColor = VigorColors.textSecondary(context);

    return Dialog(
      backgroundColor: Colors.transparent,
      elevation: 0,
      child: ClipRRect(
        borderRadius: VigorRadius.modal,
        child: BackdropFilter(
          filter: ImageFilter.blur(
            sigmaX: VigorColors.glassBlur,
            sigmaY: VigorColors.glassBlur,
          ),
          child: Container(
            decoration: LiquidGlassTheme.glassDecoration(
              borderRadius: VigorRadius.lg,
              opacity: 0.95,
              isDark: isDark,
            ),
            padding: VigorSpacing.paddingLg,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (title != null) ...[
                  Text(
                    title!,
                    style: VigorTypography.headline.copyWith(color: textColor),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: VigorSpacing.sm),
                ],
                if (content != null) ...[
                  Text(
                    content!,
                    style: VigorTypography.body.copyWith(color: secondaryColor),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: VigorSpacing.lg),
                ],
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                  children: actions.map((action) {
                    return Expanded(
                      child: Padding(
                        padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.xs),
                        child: GestureDetector(
                          onTap: action.onPressed,
                          child: Container(
                            padding: const EdgeInsets.symmetric(vertical: VigorSpacing.sm),
                            decoration: BoxDecoration(
                              color: action.isDestructive
                                  ? VigorColors.error.withValues(alpha: 0.1)
                                  : action.isDefault
                                      ? VigorColors.orange.withValues(alpha: 0.9)
                                      : Colors.transparent,
                              borderRadius: VigorRadius.radiusSm,
                              border: !action.isDefault
                                  ? Border.all(
                                      color: action.isDestructive
                                          ? VigorColors.error
                                          : VigorColors.border(context),
                                    )
                                  : null,
                            ),
                            child: Text(
                              action.label,
                              style: VigorTypography.label.copyWith(
                                color: action.isDefault
                                    ? Colors.white
                                    : action.isDestructive
                                        ? VigorColors.error
                                        : VigorColors.orange,
                                fontWeight: action.isDefault
                                    ? FontWeight.w600
                                    : FontWeight.w400,
                              ),
                              textAlign: TextAlign.center,
                            ),
                          ),
                        ),
                      ),
                    );
                  }).toList(),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildMaterialDialog(BuildContext context) {
    return AlertDialog(
      title: title != null ? Text(title!) : null,
      content: content != null ? Text(content!) : null,
      actions: actions
          .map((action) => TextButton(
                onPressed: action.onPressed,
                style: action.isDestructive
                    ? TextButton.styleFrom(foregroundColor: VigorColors.error)
                    : action.isDefault
                        ? TextButton.styleFrom(foregroundColor: VigorColors.orange)
                        : null,
                child: Text(action.label),
              ))
          .toList(),
    );
  }
}

/// Dialog action configuration
class AdaptiveDialogAction {
  final String label;
  final VoidCallback onPressed;
  final bool isDestructive;
  final bool isDefault;

  const AdaptiveDialogAction({
    required this.label,
    required this.onPressed,
    this.isDestructive = false,
    this.isDefault = false,
  });
}
