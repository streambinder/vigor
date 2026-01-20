import '../models/routine.dart';
import '../models/block.dart';
import 'timer_mode.dart';

/// Represents a segment of the workout with its timer mode
/// Used for hybrid mode switching between warmup/work/cooldown
class RoutineSegment {
  final Routine routine;
  final TimerMode mode;
  final List<Block> blocks;

  RoutineSegment({
    required this.routine,
    required this.mode,
    required this.blocks,
  });

  /// Build segments from training structure
  /// Warmup/cooldown always use interval mode
  /// Work routine uses methodology-appropriate mode
  static List<RoutineSegment> buildSegments(
    List<Routine> routines,
    String methodology,
  ) {
    final segments = <RoutineSegment>[];

    for (final routine in routines) {
      final mode = _modeForRoutine(routine, methodology);
      segments.add(RoutineSegment(
        routine: routine,
        mode: mode,
        blocks: routine.blocks,
      ));
    }

    return segments;
  }

  static TimerMode _modeForRoutine(Routine routine, String methodology) {
    // warmup and cooldown always use interval mode
    final type = routine.type.toLowerCase();
    if (type == 'warmup' || type == 'warm-up' || type == 'cooldown' || type == 'cool-down') {
      return TimerMode.interval;
    }
    // work routine uses methodology-specific mode
    return TimerModeExt.fromMethodology(methodology);
  }
}
