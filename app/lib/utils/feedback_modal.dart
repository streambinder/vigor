import 'package:flutter/material.dart';
import '../models/training.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';
import '../widgets/adaptive/adaptive.dart';

enum ExerciseFeedback { none, tooEasy, tooHard }

class FeedbackResult {
  final String feedback;
  final Map<String, String> activityFeedback;

  FeedbackResult({required this.feedback, required this.activityFeedback});
}

class FeedbackModal {
  static const _workTypes = ['cardio', 'strength', 'skill'];

  /// extracts unique exercise names from work routines only (excludes warmup/cooldown and non-work activities)
  static List<String> _getWorkExerciseNames(Training training) {
    final seen = <String>{};
    final names = <String>[];
    for (final routine in training.routines) {
      if (routine.type != 'work') continue;
      for (final block in routine.blocks) {
        for (final activity in block.activities) {
          final detailType = activity.detail['type'] as String? ?? '';
          if (_workTypes.contains(detailType) && !seen.contains(activity.name)) {
            seen.add(activity.name);
            names.add(activity.name);
          }
        }
      }
    }
    return names;
  }

  /// shows the feedback modal and returns the result, or null if cancelled
  static Future<FeedbackResult?> show(BuildContext context, Training training) async {
    final exerciseNames = _getWorkExerciseNames(training);
    if (exerciseNames.isEmpty) {
      // no work exercises to rate, return empty result
      return FeedbackResult(feedback: '', activityFeedback: {});
    }

    return showDialog<FeedbackResult>(
      context: context,
      barrierDismissible: false,
      builder: (context) => _FeedbackDialogContent(exerciseNames: exerciseNames),
    );
  }
}

class _FeedbackDialogContent extends StatefulWidget {
  final List<String> exerciseNames;

  const _FeedbackDialogContent({required this.exerciseNames});

  @override
  State<_FeedbackDialogContent> createState() => _FeedbackDialogContentState();
}

class _FeedbackDialogContentState extends State<_FeedbackDialogContent> {
  final _feedbackController = TextEditingController();
  late final Map<String, ExerciseFeedback> _exerciseFeedback;

  @override
  void initState() {
    super.initState();
    _exerciseFeedback = {
      for (final name in widget.exerciseNames) name: ExerciseFeedback.none,
    };
  }

  @override
  void dispose() {
    _feedbackController.dispose();
    super.dispose();
  }

  bool get _isValid {
    // at least one explicit feedback (thumbs up/down) OR general text provided
    final hasExplicitFeedback = _exerciseFeedback.values.any(
      (f) => f != ExerciseFeedback.none,
    );
    final hasGeneralFeedback = _feedbackController.text.trim().isNotEmpty;
    return hasExplicitFeedback || hasGeneralFeedback;
  }

  void _complete() {
    if (!_isValid) return;
    final activityFeedback = <String, String>{};
    for (final entry in _exerciseFeedback.entries) {
      final value = switch (entry.value) {
        ExerciseFeedback.tooEasy => 'too easy',
        ExerciseFeedback.tooHard => 'too hard',
        ExerciseFeedback.none => 'ok',
      };
      activityFeedback[entry.key] = value;
    }
    // strip newlines from feedback text
    final feedback = _feedbackController.text.trim().replaceAll(RegExp(r'[\r\n]+'), ' ');
    Navigator.of(context).pop(FeedbackResult(
      feedback: feedback,
      activityFeedback: activityFeedback,
    ));
  }

  @override
  Widget build(BuildContext context) {
    final useLiquidGlass = PlatformHelper.useLiquidGlass;

    return Dialog(
      backgroundColor: Colors.transparent,
      child: Container(
        constraints: const BoxConstraints(maxWidth: 500, maxHeight: 600),
        decoration: useLiquidGlass
            ? LiquidGlassTheme.glassDecoration(
                borderRadius: 20,
                border: Border.all(
                  color: Colors.white.withOpacity(0.2),
                  width: 1.5,
                ),
              )
            : BoxDecoration(
                color: Theme.of(context).colorScheme.surface,
                borderRadius: BorderRadius.circular(20),
              ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Padding(
                padding: const EdgeInsets.fromLTRB(24, 24, 24, 16),
                child: Text(
                  'How was your workout?',
                  style: useLiquidGlass
                      ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 20)
                      : Theme.of(context).textTheme.titleLarge,
                ),
              ),
              Flexible(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.symmetric(horizontal: 24),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      TextField(
                        controller: _feedbackController,
                        maxLines: 1,
                        decoration: InputDecoration(
                          hintText: 'Any additional comments?',
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(12),
                          ),
                        ),
                        onChanged: (_) => setState(() {}),
                      ),
                      const SizedBox(height: 16),
                      ...widget.exerciseNames.map((name) => _ExerciseFeedbackRow(
                            name: name,
                            feedback: _exerciseFeedback[name]!,
                            onChanged: (feedback) {
                              setState(() => _exerciseFeedback[name] = feedback);
                            },
                          )),
                    ],
                  ),
                ),
              ),
              Padding(
                padding: const EdgeInsets.all(24),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    AdaptiveTextButton(
                      onPressed: () => Navigator.of(context).pop(null),
                      child: const Text('Cancel'),
                    ),
                    const SizedBox(width: 16),
                    AdaptiveButton(
                      onPressed: _isValid ? _complete : null,
                      child: const Text('Complete'),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ExerciseFeedbackRow extends StatelessWidget {
  final String name;
  final ExerciseFeedback feedback;
  final ValueChanged<ExerciseFeedback> onChanged;

  const _ExerciseFeedbackRow({
    required this.name,
    required this.feedback,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    final useLiquidGlass = PlatformHelper.useLiquidGlass;
    final primaryColor = useLiquidGlass
        ? LiquidGlassTheme.primaryColor
        : Theme.of(context).colorScheme.primary;

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        children: [
          Expanded(
            child: Text(
              name,
              style: useLiquidGlass
                  ? LiquidGlassTheme.bodyStyle
                  : Theme.of(context).textTheme.bodyMedium,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          const SizedBox(width: 8),
          IconButton(
            icon: Icon(
              Icons.thumb_up,
              color: feedback == ExerciseFeedback.tooEasy
                  ? primaryColor
                  : Colors.grey,
            ),
            onPressed: () => onChanged(
              feedback == ExerciseFeedback.tooEasy
                  ? ExerciseFeedback.none
                  : ExerciseFeedback.tooEasy,
            ),
            tooltip: 'Too easy',
          ),
          IconButton(
            icon: Icon(
              Icons.thumb_down,
              color: feedback == ExerciseFeedback.tooHard
                  ? primaryColor
                  : Colors.grey,
            ),
            onPressed: () => onChanged(
              feedback == ExerciseFeedback.tooHard
                  ? ExerciseFeedback.none
                  : ExerciseFeedback.tooHard,
            ),
            tooltip: 'Too hard',
          ),
        ],
      ),
    );
  }
}
