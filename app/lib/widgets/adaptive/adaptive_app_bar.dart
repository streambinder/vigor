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

    return _buildMaterialAppBar(context);
  }

  Widget _buildMaterialAppBar(BuildContext context) {
    final textColor = VigorColors.textPrimary(context);
    final hasCustomLeading = leading != null;
    final willHaveLeading = hasCustomLeading ||
        (automaticallyImplyLeading && Navigator.of(context).canPop());

    // wrap actions to compensate for IconButton's internal padding (8px)
    // so icons visually align with body content at VigorSpacing.lg (24px)
    final wrappedActions = actions != null
        ? [
            Padding(
              padding: const EdgeInsets.only(right: VigorSpacing.lg - 8),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: actions!,
              ),
            ),
          ]
        : null;

    // wrap custom leading to add left margin so icon aligns with body content
    final wrappedLeading = hasCustomLeading
        ? Padding(
            padding: const EdgeInsets.only(left: VigorSpacing.lg - 8),
            child: leading,
          )
        : null;

    return Column(
      children: [
        const SizedBox(height: VigorSpacing.sm),
        Expanded(
          child: AppBar(
            title: title != null
                ? DefaultTextStyle(
                    style: VigorTypography.headline.copyWith(color: textColor),
                    child: title!,
                  )
                : null,
            titleSpacing: willHaveLeading ? VigorSpacing.sm : VigorSpacing.lg,
            actions: wrappedActions,
            leading: wrappedLeading,
            leadingWidth: hasCustomLeading ? kToolbarHeight + VigorSpacing.lg - 8 : null,
            automaticallyImplyLeading: hasCustomLeading ? false : automaticallyImplyLeading,
            backgroundColor: Colors.transparent,
            elevation: 0,
            toolbarHeight: kToolbarHeight,
          ),
        ),
      ],
    );
  }

  Widget _buildLiquidGlassAppBar(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final textColor = VigorColors.textPrimary(context);
    // cache canPop result to avoid race conditions between padding calc and widget render
    final canPop = Navigator.of(context).canPop();
    final showBackButton = leading == null && automaticallyImplyLeading && canPop;
    final hasLeading = leading != null || showBackButton;
    final hasActions = actions != null && actions!.isNotEmpty;

    // compensate for IconButton's internal 8px padding so icons align with body content
    final leftPadding = hasLeading ? VigorSpacing.lg - 8 : VigorSpacing.lg;
    final rightPadding = hasActions ? VigorSpacing.lg - 8 : VigorSpacing.lg;

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
            minimum: const EdgeInsets.only(top: VigorSpacing.sm),
            child: Container(
              height: kToolbarHeight,
              padding: EdgeInsets.only(left: leftPadding, right: rightPadding),
              child: Row(
                children: [
                  if (leading != null)
                    leading!
                  else if (showBackButton)
                    IconButton(
                      icon: const Icon(Icons.arrow_back_ios),
                      onPressed: () => Navigator.of(context).pop(),
                      color: VigorColors.stone,
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
  Size get preferredSize => const Size.fromHeight(kToolbarHeight + VigorSpacing.sm);
}
