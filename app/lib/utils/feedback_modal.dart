import 'package:flutter/material.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/activity_ext.dart';
import '../models/training.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';
import '../widgets/adaptive/adaptive.dart';

// matches backend model.Feedback* constants
enum ExerciseFeedback { none, tooEasy, easy, ok, hard, tooHard, skipped }

extension ExerciseFeedbackX on ExerciseFeedback {
  String toApiValue() => switch (this) {
        ExerciseFeedback.tooEasy => 'too_easy',
        ExerciseFeedback.easy => 'easy',
        ExerciseFeedback.ok => 'ok',
        ExerciseFeedback.hard => 'hard',
        ExerciseFeedback.tooHard => 'too_hard',
        ExerciseFeedback.skipped => 'skipped',
        ExerciseFeedback.none => '',
      };

  static ExerciseFeedback fromApiValue(String? value) => switch (value) {
        'too_easy' => ExerciseFeedback.tooEasy,
        'easy' => ExerciseFeedback.easy,
        'ok' => ExerciseFeedback.ok,
        'hard' => ExerciseFeedback.hard,
        'too_hard' => ExerciseFeedback.tooHard,
        'skipped' => ExerciseFeedback.skipped,
        _ => ExerciseFeedback.none,
      };
}

class FeedbackResult {
  final String feedback;
  final Map<String, String> activityFeedback;
  final List<String> activityReports; // activity IDs flagged by user

  FeedbackResult({
    required this.feedback,
    required this.activityFeedback,
    required this.activityReports,
  });
}

class FeedbackModal {
  /// extracts unique activities from work routines
  static List<({String id, String exerciseId, String name, String? feedback})> _getWorkActivities(Training training) {
    final seen = <String>{};
    final activities = <({String id, String exerciseId, String name, String? feedback})>[];
    for (final routine in training.routines) {
      if (routine.type != 'work') continue;
      for (final block in routine.blocks) {
        for (final activity in block.activities) {
          if (!seen.contains(activity.exerciseId)) {
            seen.add(activity.exerciseId);
            activities.add((id: activity.id, exerciseId: activity.exerciseId, name: activity.displayName, feedback: activity.feedback));
          }
        }
      }
    }
    return activities;
  }

  /// shows the feedback modal and returns the result, or null if cancelled
  static Future<FeedbackResult?> show(BuildContext context, Training training) async {
    final activities = _getWorkActivities(training);
    if (activities.isEmpty) {
      return FeedbackResult(feedback: '', activityFeedback: {}, activityReports: []);
    }

    return showDialog<FeedbackResult>(
      context: context,
      barrierDismissible: false,
      builder: (context) => _FeedbackDialogContent(
        activities: activities,
        showReports: true,
      ),
    );
  }

  /// shows the feedback modal for updating existing feedback (no reports, pre-populated values)
  static Future<FeedbackResult?> showForUpdate(BuildContext context, Training training) async {
    final activities = _getWorkActivities(training);
    if (activities.isEmpty) {
      return FeedbackResult(feedback: '', activityFeedback: {}, activityReports: []);
    }

    return showDialog<FeedbackResult>(
      context: context,
      barrierDismissible: false,
      builder: (context) => _FeedbackDialogContent(
        activities: activities,
        showReports: false,
        initialFeedback: training.feedback,
      ),
    );
  }
}

class _FeedbackDialogContent extends StatefulWidget {
  final List<({String id, String exerciseId, String name, String? feedback})> activities;
  final bool showReports;
  final String? initialFeedback;

  const _FeedbackDialogContent({
    required this.activities,
    required this.showReports,
    this.initialFeedback,
  });

  @override
  State<_FeedbackDialogContent> createState() => _FeedbackDialogContentState();
}

class _FeedbackDialogContentState extends State<_FeedbackDialogContent> {
  final _feedbackController = TextEditingController();
  late final Map<String, ExerciseFeedback> _exerciseFeedback; // keyed by activity id
  late final Set<String> _flaggedActivities; // activity IDs that are flagged

  @override
  void initState() {
    super.initState();
    _feedbackController.text = widget.initialFeedback ?? '';
    _exerciseFeedback = {
      for (final a in widget.activities)
        a.id: ExerciseFeedbackX.fromApiValue(a.feedback),
    };
    _flaggedActivities = {};
  }

  @override
  void dispose() {
    _feedbackController.dispose();
    super.dispose();
  }

  bool get _isValid {
    final hasExplicitFeedback = _exerciseFeedback.values.any(
      (f) => f != ExerciseFeedback.none,
    );
    final hasGeneralFeedback = _feedbackController.text.trim().isNotEmpty;
    final hasFlags = widget.showReports && _flaggedActivities.isNotEmpty;
    return hasExplicitFeedback || hasGeneralFeedback || hasFlags;
  }

  void _complete() {
    if (!_isValid) return;
    final activityFeedback = <String, String>{};
    final activityReports = <String>[];

    for (final a in widget.activities) {
      final fb = _exerciseFeedback[a.id]!;
      final apiValue = fb.toApiValue();
      if (apiValue.isNotEmpty) {
        activityFeedback[a.exerciseId] = apiValue;
      }
      if (_flaggedActivities.contains(a.id)) {
        activityReports.add(a.id);
      }
    }

    final feedback = _feedbackController.text.trim().replaceAll(RegExp(r'[\r\n]+'), ' ');
    Navigator.of(context).pop(FeedbackResult(
      feedback: feedback,
      activityFeedback: activityFeedback,
      activityReports: activityReports,
    ));
  }

  @override
  Widget build(BuildContext context) {
    final useLiquidGlass = PlatformHelper.useLiquidGlass;
    final l10n = AppLocalizations.of(context);

    return Dialog(
      backgroundColor: Colors.transparent,
      child: Container(
        constraints: const BoxConstraints(maxWidth: 500, maxHeight: 600),
        decoration: useLiquidGlass
            ? LiquidGlassTheme.glassDecoration(
                borderRadius: 20,
                border: Border.all(
                  color: Colors.white.withValues(alpha: 0.2),
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
                  l10n.howWasYourTraining,
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
                          hintText: l10n.anyAdditionalComments,
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(12),
                          ),
                        ),
                        onChanged: (_) => setState(() {}),
                      ),
                      const SizedBox(height: 16),
                      ...widget.activities.map((a) => _ExerciseFeedbackRow(
                            name: a.name,
                            feedback: _exerciseFeedback[a.id]!,
                            isFlagged: _flaggedActivities.contains(a.id),
                            showFlag: widget.showReports,
                            onFeedbackChanged: (feedback) {
                              setState(() => _exerciseFeedback[a.id] = feedback);
                            },
                            onFlagChanged: (flagged) {
                              setState(() {
                                if (flagged) {
                                  _flaggedActivities.add(a.id);
                                } else {
                                  _flaggedActivities.remove(a.id);
                                }
                              });
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
                      child: Text(l10n.cancel),
                    ),
                    const SizedBox(width: 16),
                    AdaptiveButton(
                      onPressed: _isValid ? _complete : null,
                      child: Text(l10n.complete),
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
  final bool isFlagged;
  final bool showFlag;
  final ValueChanged<ExerciseFeedback> onFeedbackChanged;
  final ValueChanged<bool> onFlagChanged;

  const _ExerciseFeedbackRow({
    required this.name,
    required this.feedback,
    required this.isFlagged,
    required this.showFlag,
    required this.onFeedbackChanged,
    required this.onFlagChanged,
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
          // too easy (thumbs up)
          IconButton(
            icon: Icon(
              Icons.thumb_up_outlined,
              color: feedback == ExerciseFeedback.tooEasy ? primaryColor : VigorColors.stone,
            ),
            onPressed: () => onFeedbackChanged(
              feedback == ExerciseFeedback.tooEasy ? ExerciseFeedback.none : ExerciseFeedback.tooEasy,
            ),
            tooltip: 'Too easy',
          ),
          // too hard (thumbs down)
          IconButton(
            icon: Icon(
              Icons.thumb_down_outlined,
              color: feedback == ExerciseFeedback.tooHard ? primaryColor : VigorColors.stone,
            ),
            onPressed: () => onFeedbackChanged(
              feedback == ExerciseFeedback.tooHard ? ExerciseFeedback.none : ExerciseFeedback.tooHard,
            ),
            tooltip: 'Too hard',
          ),
          // flag (independent from feedback, only shown in completion mode)
          if (showFlag)
            IconButton(
              icon: Icon(
                isFlagged ? Icons.flag : Icons.flag_outlined,
                color: isFlagged ? VigorColors.crimson : VigorColors.stone,
              ),
              onPressed: () => onFlagChanged(!isFlagged),
              tooltip: 'Flag issue',
            ),
        ],
      ),
    );
  }
}
