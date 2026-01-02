import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';

/// Vigor logo widget using the brand V icon.
/// Use [withBackground] for full logo with dark background (splash screens).
/// Use default for transparent V icon.
class VigorLogo extends StatelessWidget {
  final double size;
  final bool withBackground;

  const VigorLogo({
    super.key,
    required this.size,
    this.withBackground = false,
  });

  @override
  Widget build(BuildContext context) {
    final asset = withBackground
        ? 'assets/icons/vigor-app-icon.svg'
        : 'assets/icons/vigor-icon.svg';

    return SvgPicture.asset(
      asset,
      width: size,
      height: size,
    );
  }
}
