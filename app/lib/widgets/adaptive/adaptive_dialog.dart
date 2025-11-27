import 'dart:ui';
import 'package:flutter/material.dart';
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
    return Dialog(
      backgroundColor: Colors.transparent,
      elevation: 0,
      child: ClipRRect(
        borderRadius: BorderRadius.circular(20),
        child: BackdropFilter(
          filter: ImageFilter.blur(
            sigmaX: LiquidGlassTheme.glassBlur,
            sigmaY: LiquidGlassTheme.glassBlur,
          ),
          child: Container(
            decoration: LiquidGlassTheme.glassDecoration(
              borderRadius: 20,
              opacity: 0.95,
            ),
            padding: const EdgeInsets.all(20),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (title != null) ...[
                  Text(
                    title!,
                    style: LiquidGlassTheme.headlineStyle,
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 12),
                ],
                if (content != null) ...[
                  Text(
                    content!,
                    style: LiquidGlassTheme.bodyStyle,
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 20),
                ],
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                  children: actions.map((action) {
                    return Expanded(
                      child: Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 4),
                        child: GestureDetector(
                          onTap: action.onPressed,
                          child: Container(
                            padding: const EdgeInsets.symmetric(vertical: 12),
                            decoration: BoxDecoration(
                              color: action.isDestructive
                                  ? LiquidGlassTheme.errorColor.withOpacity(0.1)
                                  : action.isDefault
                                      ? LiquidGlassTheme.primaryColor
                                          .withOpacity(0.9)
                                      : Colors.transparent,
                              borderRadius: BorderRadius.circular(10),
                              border: !action.isDefault
                                  ? Border.all(
                                      color: action.isDestructive
                                          ? LiquidGlassTheme.errorColor
                                          : Colors.white.withOpacity(0.3),
                                    )
                                  : null,
                            ),
                            child: Text(
                              action.label,
                              style: TextStyle(
                                color: action.isDefault
                                    ? Colors.white
                                    : action.isDestructive
                                        ? LiquidGlassTheme.errorColor
                                        : LiquidGlassTheme.primaryColor,
                                fontSize: 17,
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
                    ? TextButton.styleFrom(
                        foregroundColor: Theme.of(context).colorScheme.error,
                      )
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
