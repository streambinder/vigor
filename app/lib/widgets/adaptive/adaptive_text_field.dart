import 'dart:ui';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
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
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        if (labelText != null) ...[
          Text(
            labelText!,
            style: LiquidGlassTheme.captionStyle,
          ),
          const SizedBox(height: 8),
        ],
        ClipRRect(
          borderRadius: BorderRadius.circular(12),
          child: BackdropFilter(
            filter: ImageFilter.blur(sigmaX: 10, sigmaY: 10),
            child: Container(
              decoration: LiquidGlassTheme.glassDecoration(
                borderRadius: 12,
                border: Border.all(
                  color: errorText != null
                      ? LiquidGlassTheme.errorColor.withOpacity(0.5)
                      : Colors.white.withOpacity(
                          LiquidGlassTheme.glassBorderOpacity,
                        ),
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
                inputFormatters: inputFormatters,
                readOnly: readOnly,
                onTap: onTap,
                focusNode: focusNode,
                style: LiquidGlassTheme.bodyStyle,
                decoration: InputDecoration(
                  hintText: placeholder ?? labelText,
                  hintStyle: LiquidGlassTheme.bodyStyle.copyWith(
                    color: LiquidGlassTheme.captionStyle.color,
                  ),
                  prefixIcon: prefix,
                  suffixIcon: suffix,
                  border: InputBorder.none,
                  contentPadding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 14,
                  ),
                ),
              ),
            ),
          ),
        ),
        if (errorText != null) ...[
          const SizedBox(height: 6),
          Text(
            errorText!,
            style: LiquidGlassTheme.captionStyle.copyWith(
              color: LiquidGlassTheme.errorColor,
            ),
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
      inputFormatters: inputFormatters,
      readOnly: readOnly,
      onTap: onTap,
      focusNode: focusNode,
    );
  }
}
