import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';

import '../design/tokens.dart';
import '../models/exercise.dart';
import '../services/exercise_player_service.dart';
import 'cached_exercise_image.dart';

/// Row thumbnail for an exercise: the still for animated exercises, the
/// GIF otherwise, the no-image placeholder when neither exists.
///
/// The stills are fetched once per app run through the player service and
/// shared by every row. While they load, or when the fetch fails, the
/// thumbnail falls back to the GIF behavior so rows never break.
class ExerciseThumbnail extends StatelessWidget {
  final Exercise? exercise;
  final ExercisePlayerService playerService;
  final double width;
  final double height;
  final BorderRadius? borderRadius;

  const ExerciseThumbnail({
    super.key,
    required this.exercise,
    required this.playerService,
    required this.width,
    required this.height,
    this.borderRadius,
  });

  @override
  Widget build(BuildContext context) {
    final ex = exercise;
    return SizedBox(
      width: width,
      height: height,
      child: ex == null
          ? _placeholder()
          : FutureBuilder<Map<String, String>>(
              future: playerService.animatedExercises(),
              builder: (context, snapshot) {
                final still = snapshot.data?[ex.id];
                if (still != null) {
                  return _still(still);
                }
                if (CachedExerciseImage.isValidUrl(ex.reference)) {
                  return CachedExerciseImage(
                    imageUrl: ex.reference,
                    width: width,
                    height: height,
                    borderRadius: borderRadius,
                  );
                }
                return _placeholder();
              },
            ),
    );
  }

  Widget _still(String svg) {
    return ClipRRect(
      borderRadius: borderRadius ?? BorderRadius.zero,
      child: Container(
        color: VigorColors.stone.withValues(alpha: 0.1),
        child: Stack(
          children: [
            Center(
              child: Padding(
                padding: const EdgeInsets.all(6),
                child: SvgPicture.string(
                  svg,
                  width: width,
                  height: height,
                  fit: BoxFit.contain,
                ),
              ),
            ),
            Positioned(
              right: 2,
              bottom: 2,
              child: Container(
                decoration: const BoxDecoration(
                  color: Colors.black54,
                  shape: BoxShape.circle,
                ),
                padding: const EdgeInsets.all(2),
                child: const Icon(
                  Icons.play_arrow,
                  size: 14,
                  color: Colors.white,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _placeholder() {
    return Container(
      decoration: BoxDecoration(
        color: VigorColors.stone.withValues(alpha: 0.1),
        borderRadius: borderRadius,
      ),
      child: Center(
        child: Icon(
          Icons.fitness_center,
          size: width * 0.4,
          color: VigorColors.stone.withValues(alpha: 0.5),
        ),
      ),
    );
  }
}
