import 'dart:ui';
import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../utils/platform_helper.dart';

/// Platform-adaptive list tile
/// Uses Liquid Glass effect on iOS and Material ListTile on other platforms
class AdaptiveListTile extends StatelessWidget {
  final Widget? leading;
  final Widget title;
  final Widget? subtitle;
  final Widget? trailing;
  final VoidCallback? onTap;
  final EdgeInsetsGeometry? contentPadding;

  const AdaptiveListTile({
    super.key,
    this.leading,
    required this.title,
    this.subtitle,
    this.trailing,
    this.onTap,
    this.contentPadding,
  });

  @override
  Widget build(BuildContext context) {
    if (PlatformHelper.useLiquidGlass) {
      return _buildLiquidGlassTile(context);
    }
    return _buildMaterialTile(context);
  }

  Widget _buildLiquidGlassTile(BuildContext context) {
    final textColor = VigorColors.textPrimary(context);
    final secondaryColor = VigorColors.textSecondary(context);

    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: contentPadding ?? VigorSpacing.listTilePadding,
        child: Row(
          children: [
            if (leading != null) ...[
              leading!,
              SizedBox(width: VigorSpacing.sm),
            ],
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  DefaultTextStyle(
                    style: VigorTypography.body.copyWith(color: textColor),
                    child: title,
                  ),
                  if (subtitle != null) ...[
                    SizedBox(height: VigorSpacing.xs),
                    DefaultTextStyle(
                      style: VigorTypography.caption.copyWith(color: secondaryColor),
                      child: subtitle!,
                    ),
                  ],
                ],
              ),
            ),
            if (trailing != null) ...[
              SizedBox(width: VigorSpacing.sm),
              trailing!,
            ] else if (onTap != null) ...[
              SizedBox(width: VigorSpacing.sm),
              Icon(
                Icons.chevron_right,
                size: 20,
                color: secondaryColor,
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildMaterialTile(BuildContext context) {
    return ListTile(
      leading: leading,
      title: title,
      subtitle: subtitle,
      trailing: trailing,
      onTap: onTap,
      contentPadding: contentPadding,
    );
  }
}
