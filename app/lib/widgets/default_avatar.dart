import 'package:flutter/material.dart';

import '../design/tokens.dart';

/// Default avatar shown when the user has no profile picture.
///
/// A flat subtle gradient circle with a single uppercase letter derived from
/// the user name, falling back to the email and then to '?'.
class DefaultAvatar extends StatelessWidget {
  final String? name;
  final String? email;
  final double radius;

  const DefaultAvatar({super.key, this.name, this.email, this.radius = 20});

  String get _letter {
    final trimmedName = (name ?? '').trim();
    if (trimmedName.isNotEmpty) return trimmedName[0].toUpperCase();
    final trimmedEmail = (email ?? '').trim();
    if (trimmedEmail.isNotEmpty) return trimmedEmail[0].toUpperCase();
    return '?';
  }

  @override
  Widget build(BuildContext context) {
    final base = VigorColors.indigoAdaptive(context);
    return Container(
      width: radius * 2,
      height: radius * 2,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            Color.lerp(base, Colors.white, 0.12) ?? base,
            Color.lerp(base, Colors.black, 0.18) ?? base,
          ],
        ),
      ),
      alignment: Alignment.center,
      child: Text(
        _letter,
        style: TextStyle(
          color: Colors.white,
          fontSize: radius,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }
}
