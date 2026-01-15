import 'dart:ui';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../design/tokens.dart';
import '../../utils/platform_helper.dart';
import '../../theme/liquid_glass_theme.dart';

/// Platform-adaptive text field
/// Uses Liquid Glass effect on iOS and Material TextField on other platforms
class AdaptiveTextField extends StatelessWidget {
  final TextEditingController? controller;
  final String? placeholder;
  final String? labelText;
  final TextInputType? keyboardType;
  final bool obscureText;
  final ValueChanged<String>? onChanged;
  final ValueChanged<String>? onSubmitted;
  final String? errorText;
  final Widget? prefix;
  final Widget? suffix;
  final int? maxLines;
  final int? minLines;
  final int? maxLength;
  final List<TextInputFormatter>? inputFormatters;
  final bool readOnly;
  final VoidCallback? onTap;
  final FocusNode? focusNode;

  const AdaptiveTextField({
    super.key,
    this.controller,
    this.placeholder,
    this.labelText,
    this.keyboardType,
    this.obscureText = false,
    this.onChanged,
    this.onSubmitted,
    this.errorText,
    this.prefix,
    this.suffix,
    this.maxLines = 1,
    this.minLines,
    this.maxLength,
    this.inputFormatters,
    this.readOnly = false,
    this.onTap,
    this.focusNode,
  });

  @override
  Widget build(BuildContext context) {
    if (PlatformHelper.useLiquidGlass) {
      return _buildLiquidGlassField(context);
    }
    return _buildMaterialField(context);
  }

  Widget _buildLiquidGlassField(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final textColor = VigorColors.textPrimary(context);
    final hintColor = VigorColors.textMuted(context);
    final labelColor = VigorColors.textSecondary(context);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        if (labelText != null) ...[
          Text(
            labelText!,
            style: VigorTypography.caption.copyWith(color: labelColor),
          ),
          const SizedBox(height: VigorSpacing.sm),
        ],
        ClipRRect(
          borderRadius: VigorRadius.input,
          child: BackdropFilter(
            filter: ImageFilter.blur(sigmaX: 10, sigmaY: 10),
            child: Container(
              decoration: LiquidGlassTheme.glassDecoration(
                borderRadius: VigorRadius.sm,
                isDark: isDark,
                border: Border.all(
                  color: errorText != null
                      ? VigorColors.error.withValues(alpha: 0.5)
                      : VigorColors.border(context),
                  width: 1.5,
                ),
              ),
              child: TextField(
                controller: controller,
                keyboardType: keyboardType,
                obscureText: obscureText,
                onChanged: onChanged,
                onSubmitted: onSubmitted,
                maxLines: maxLines,
                minLines: minLines,
                maxLength: maxLength,
                inputFormatters: inputFormatters,
                readOnly: readOnly,
                onTap: onTap,
                focusNode: focusNode,
                style: VigorTypography.body.copyWith(color: textColor),
                cursorColor: VigorColors.orange,
                decoration: InputDecoration(
                  hintText: placeholder ?? labelText,
                  hintStyle: VigorTypography.body.copyWith(color: hintColor),
                  prefixIcon: prefix,
                  suffixIcon: suffix,
                  border: InputBorder.none,
                  contentPadding: VigorSpacing.inputPadding,
                ),
              ),
            ),
          ),
        ),
        if (errorText != null) ...[
          const SizedBox(height: VigorSpacing.xs),
          Text(
            errorText!,
            style: VigorTypography.caption.copyWith(color: VigorColors.error),
          ),
        ],
      ],
    );
  }

  Widget _buildMaterialField(BuildContext context) {
    return TextField(
      controller: controller,
      decoration: InputDecoration(
        labelText: labelText,
        hintText: placeholder,
        errorText: errorText,
        prefixIcon: prefix,
        suffixIcon: suffix,
      ),
      keyboardType: keyboardType,
      obscureText: obscureText,
      onChanged: onChanged,
      onSubmitted: onSubmitted,
      maxLines: maxLines,
      minLines: minLines,
      maxLength: maxLength,
      inputFormatters: inputFormatters,
      readOnly: readOnly,
      onTap: onTap,
      focusNode: focusNode,
    );
  }
}
