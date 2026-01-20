import 'package:flutter/material.dart';
import '../../design/tokens.dart';

class RestDisplay extends StatelessWidget {
  final int remainingSeconds;

  const RestDisplay({
    super.key,
    required this.remainingSeconds,
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(
            Icons.local_drink,
            size: 64,
            color: VigorColors.indigo,
          ),
          const SizedBox(height: VigorSpacing.lg),
          Text(
            '$remainingSeconds',
            style: VigorTypography.dataDisplay.copyWith(
              color: VigorColors.indigo,
            ),
          ),
        ],
      ),
    );
  }
}
