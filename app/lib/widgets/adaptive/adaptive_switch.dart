import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../utils/platform_helper.dart';

/// Platform-adaptive switch
/// Uses CupertinoSwitch on iOS and Material Switch on other platforms
class AdaptiveSwitch extends StatelessWidget {
  final bool value;
  final ValueChanged<bool>? onChanged;

  const AdaptiveSwitch({
    super.key,
    required this.value,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    if (PlatformHelper.useLiquidGlass) {
      return CupertinoSwitch(
        value: value,
        onChanged: onChanged,
        activeTrackColor: VigorColors.persimmon,
      );
    }
    return Switch(
      value: value,
      onChanged: onChanged,
    );
  }
}
