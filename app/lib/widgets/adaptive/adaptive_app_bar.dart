import 'dart:ui';
import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../utils/platform_helper.dart';

/// Platform-adaptive app bar
/// Uses Liquid Glass effect on iOS and Material AppBar on other platforms
class AdaptiveAppBar extends StatelessWidget implements PreferredSizeWidget {
  final Widget? title;
  final List<Widget>? actions;
  final Widget? leading;
  final bool automaticallyImplyLeading;
  final bool useLiquidGlass;

  const AdaptiveAppBar({
    super.key,
    this.title,
    this.actions,
    this.leading,
    this.automaticallyImplyLeading = true,
    this.useLiquidGlass = true,
  });

  @override
  Widget build(BuildContext context) {
    if (PlatformHelper.useLiquidGlass && useLiquidGlass) {
      return _buildLiquidGlassAppBar(context);
    }

    return AppBar(
      title: title,
      actions: actions,
      leading: leading,
      automaticallyImplyLeading: automaticallyImplyLeading,
    );
  }

  Widget _buildLiquidGlassAppBar(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final textColor = VigorColors.textPrimary(context);

    return ClipRRect(
      child: BackdropFilter(
        filter: ImageFilter.blur(sigmaX: 10, sigmaY: 10),
        child: Container(
          decoration: BoxDecoration(
            color: isDark
                ? Colors.black.withValues(alpha: 0.3)
                : Colors.white.withValues(alpha: 0.3),
            border: Border(
              bottom: BorderSide(
                color: isDark
                    ? Colors.white.withValues(alpha: 0.1)
                    : Colors.white.withValues(alpha: 0.2),
                width: 0.5,
              ),
            ),
          ),
          child: SafeArea(
            bottom: false,
            child: Container(
              height: kToolbarHeight,
              padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm),
              child: Row(
                children: [
                  if (leading != null)
                    leading!
                  else if (automaticallyImplyLeading &&
                      Navigator.of(context).canPop())
                    IconButton(
                      icon: const Icon(Icons.arrow_back_ios),
                      onPressed: () => Navigator.of(context).pop(),
                      color: VigorColors.orange,
                    ),
                  if (title != null)
                    Expanded(
                      child: DefaultTextStyle(
                        style: VigorTypography.headline.copyWith(color: textColor),
                        child: title!,
                      ),
                    ),
                  if (actions != null) ...actions!,
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  @override
  Size get preferredSize =>
      const Size.fromHeight(kToolbarHeight + 44);
}
