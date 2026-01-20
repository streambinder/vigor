import 'package:flutter/material.dart';
import '../../design/tokens.dart';

class TimerControls extends StatelessWidget {
  final bool isPaused;
  final bool canGoBack;
  final Color phaseColor;
  final VoidCallback onPauseToggle;
  final VoidCallback onSkipForward;
  final VoidCallback? onSkipBackward;

  const TimerControls({
    super.key,
    required this.isPaused,
    required this.canGoBack,
    required this.phaseColor,
    required this.onPauseToggle,
    required this.onSkipForward,
    this.onSkipBackward,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        FloatingActionButton(
          heroTag: 'back',
          onPressed: canGoBack ? onSkipBackward : null,
          backgroundColor: canGoBack ? phaseColor.withValues(alpha: 0.7) : VigorColors.stone,
          child: const Icon(
            Icons.skip_previous,
            color: Colors.white,
            size: 28,
          ),
        ),
        const SizedBox(width: VigorSpacing.lg),
        FloatingActionButton.large(
          heroTag: 'pause',
          onPressed: onPauseToggle,
          backgroundColor: phaseColor,
          child: Icon(
            isPaused ? Icons.play_arrow : Icons.pause,
            color: Colors.white,
            size: 36,
          ),
        ),
        const SizedBox(width: VigorSpacing.lg),
        FloatingActionButton(
          heroTag: 'forward',
          onPressed: onSkipForward,
          backgroundColor: phaseColor.withValues(alpha: 0.7),
          child: const Icon(
            Icons.skip_next,
            color: Colors.white,
            size: 28,
          ),
        ),
      ],
    );
  }
}
