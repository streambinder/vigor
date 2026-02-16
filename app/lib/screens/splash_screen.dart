import 'package:flutter/material.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/vigor_logo.dart';

class SplashScreen extends StatelessWidget {
  const SplashScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const AdaptiveScaffold(
      body: Center(child: VigorLogo(size: 100)),
    );
  }
}
