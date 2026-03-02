import 'dart:math';
import 'package:flutter/material.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/activity_ext.dart';
import '../models/exercise.dart';
import '../models/training.dart';
import '../models/training_feedback.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/exercise_modal.dart';
import '../utils/platform_helper.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/cached_exercise_image.dart';

// matches backend model.Feedback* constants, maps to -2..+2 slider
enum ExerciseFeedback { none, impossible, tooHard, ok, easy, tooEasy }

extension ExerciseFeedbackX on ExerciseFeedback {
  String toApiValue() => switch (this) {
        ExerciseFeedback.impossible => 'impossible',
        ExerciseFeedback.tooHard => 'too_hard',
        ExerciseFeedback.ok => 'ok',
        ExerciseFeedback.easy => 'easy',
        ExerciseFeedback.tooEasy => 'too_easy',
        ExerciseFeedback.none => '',
      };

  static ExerciseFeedback fromApiValue(String? value) => switch (value) {
        'impossible' => ExerciseFeedback.impossible,
        'too_hard' => ExerciseFeedback.tooHard,
        'ok' => ExerciseFeedback.ok,
        'easy' => ExerciseFeedback.easy,
        'too_easy' => ExerciseFeedback.tooEasy,
        // migrate legacy values
        'hard' => ExerciseFeedback.tooHard,
        'skipped' => ExerciseFeedback.impossible,
        _ => ExerciseFeedback.none,
      };

  /// slider position: -2 (impossible) to +2 (too easy), null if unset
  int? get sliderValue => switch (this) {
        ExerciseFeedback.impossible => -2,
        ExerciseFeedback.tooHard => -1,
        ExerciseFeedback.ok => 0,
        ExerciseFeedback.easy => 1,
        ExerciseFeedback.tooEasy => 2,
        ExerciseFeedback.none => null,
      };

  static ExerciseFeedback fromSliderValue(int value) => switch (value) {
        -2 => ExerciseFeedback.impossible,
        -1 => ExerciseFeedback.tooHard,
        0 => ExerciseFeedback.ok,
        1 => ExerciseFeedback.easy,
        2 => ExerciseFeedback.tooEasy,
        _ => ExerciseFeedback.none,
      };

  /// short label for the current slider position
  String label(AppLocalizations l10n) => switch (this) {
        ExerciseFeedback.impossible => l10n.impossible,
        ExerciseFeedback.tooHard => l10n.tooHard,
        ExerciseFeedback.ok => l10n.ok,
        ExerciseFeedback.easy => l10n.easy,
        ExerciseFeedback.tooEasy => l10n.tooEasy,
        ExerciseFeedback.none => '',
      };
}

class FeedbackResult {
  final TrainingFeedback feedback;
  final Map<String, String> activityFeedback;
  final List<String> activityReports; // activity IDs flagged by user
  final int? completedIn; // actual duration in seconds

  FeedbackResult({
    required this.feedback,
    required this.activityFeedback,
    required this.activityReports,
    this.completedIn,
  });
}

class FeedbackModal {
  /// extracts unique activities from work routines
  static List<({String id, String exerciseId, String name, Exercise? exercise})> _getWorkActivities(Training training) {
    final seen = <String>{};
    final activities = <({String id, String exerciseId, String name, Exercise? exercise})>[];
    for (final routine in training.routines) {
      if (routine.type != 'work') continue;
      for (final block in routine.blocks) {
        for (final activity in block.activities) {
          if (!seen.contains(activity.exerciseId)) {
            seen.add(activity.exerciseId);
            Exercise? exercise;
            try { exercise = Exercise.fromJson(activity.detail); } catch (_) {}
            activities.add((id: activity.id, exerciseId: activity.exerciseId, name: activity.displayName, exercise: exercise));
          }
        }
      }
    }
    return activities;
  }

  /// shows the feedback modal and returns the result, or null if cancelled
  /// [messagePrefix] is prepended to user message (e.g. methodology stats)
  /// [elapsedSeconds] pre-fills the duration field from timer tracking
  static Future<FeedbackResult?> show(
    BuildContext context,
    Training training, {
    String? messagePrefix,
    int? elapsedSeconds,
  }) async {
    final activities = _getWorkActivities(training);
    if (activities.isEmpty) {
      return FeedbackResult(
        feedback: TrainingFeedback(id: '', trainingId: '', userId: '', qualityReason: '', message: messagePrefix ?? '', activityFeedback: {}, createdAt: DateTime.now()),
        activityFeedback: {},
        activityReports: [],
      );
    }

    return showDialog<FeedbackResult>(
      context: context,
      barrierDismissible: false,
      builder: (context) => _FeedbackDialogContent(
        activities: activities,
        showReports: true,
        messagePrefix: messagePrefix,
        elapsedSeconds: elapsedSeconds,
      ),
    );
  }

  /// shows the feedback modal for updating existing feedback (no reports, pre-populated values)
  static Future<FeedbackResult?> showForUpdate(BuildContext context, Training training, {TrainingFeedback? existingFeedback}) async {
    final activities = _getWorkActivities(training);
    if (activities.isEmpty) {
      return FeedbackResult(
        feedback: TrainingFeedback(id: '', trainingId: '', userId: '', qualityReason: '', message: '', activityFeedback: {}, createdAt: DateTime.now()),
        activityFeedback: {},
        activityReports: [],
      );
    }

    return showDialog<FeedbackResult>(
      context: context,
      barrierDismissible: false,
      builder: (context) => _FeedbackDialogContent(
        activities: activities,
        showReports: false,
        initialFeedback: existingFeedback,
        elapsedSeconds: training.completedIn,
      ),
    );
  }
}

class _FeedbackDialogContent extends StatefulWidget {
  final List<({String id, String exerciseId, String name, Exercise? exercise})> activities;
  final bool showReports;
  final TrainingFeedback? initialFeedback;
  final String? messagePrefix;
  final int? elapsedSeconds;

  const _FeedbackDialogContent({
    required this.activities,
    required this.showReports,
    this.initialFeedback,
    this.messagePrefix,
    this.elapsedSeconds,
  });

  @override
  State<_FeedbackDialogContent> createState() => _FeedbackDialogContentState();
}

class _FeedbackDialogContentState extends State<_FeedbackDialogContent> {
  final _messageController = TextEditingController();
  final _qualityReasonController = TextEditingController();
  late double _durationMinutes;
  late final Map<String, ExerciseFeedback> _exerciseFeedback; // keyed by activity id
  late final Set<String> _flaggedActivities; // activity IDs that are flagged

  bool? _trainingQuality; // null = unselected, true = good, false = bad

  static const _minDuration = 10.0;
  static const _maxDuration = 180.0;

  @override
  void initState() {
    super.initState();
    // populate from existing structured feedback when updating
    final initial = widget.initialFeedback;
    if (initial != null) {
      _trainingQuality = initial.quality;
      _qualityReasonController.text = initial.qualityReason;
      _messageController.text = initial.message;
    }
    _durationMinutes = widget.elapsedSeconds != null && widget.elapsedSeconds! > 0
        ? min(_maxDuration, max(_minDuration, (widget.elapsedSeconds! / 60).roundToDouble()))
        : _minDuration;
    final afb = widget.initialFeedback?.activityFeedback ?? {};
    _exerciseFeedback = {
      for (final a in widget.activities)
        a.id: (afb[a.exerciseId] ?? '').isNotEmpty
            ? ExerciseFeedbackX.fromApiValue(afb[a.exerciseId])
            : ExerciseFeedback.ok,
    };
    _flaggedActivities = {};
  }

  @override
  void dispose() {
    _messageController.dispose();
    _qualityReasonController.dispose();
    super.dispose();
  }

  bool get _isValid {
    if (_trainingQuality == null) return false;
    // bad quality requires a reason
    if (_trainingQuality == false && _qualityReasonController.text.trim().isEmpty) return false;
    // if there's a prefix (methodology stats), form is valid even without user input
    if (widget.messagePrefix != null && widget.messagePrefix!.isNotEmpty) {
      return true;
    }
    final hasExplicitFeedback = _exerciseFeedback.values.any(
      (f) => f != ExerciseFeedback.none,
    );
    final hasGeneralFeedback = _messageController.text.trim().isNotEmpty;
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

    // build message: prefix + user text
    final userMessage = _messageController.text.trim().replaceAll(RegExp(r'[\r\n]+'), ' ');
    final prefix = widget.messagePrefix;
    final message = prefix != null && prefix.isNotEmpty ? '$prefix$userMessage' : userMessage;

    Navigator.of(context).pop(FeedbackResult(
      feedback: TrainingFeedback(
        id: '',
        trainingId: '',
        userId: '',
        quality: _trainingQuality,
        qualityReason: _trainingQuality == false ? _qualityReasonController.text.trim() : '',
        message: message,
        activityFeedback: activityFeedback,
        createdAt: DateTime.now(),
      ),
      activityFeedback: activityFeedback,
      activityReports: activityReports,
      completedIn: _durationMinutes > 0 ? (_durationMinutes * 60).round() : null,
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
                  l10n.feedback,
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
                      // training quality rating
                      Builder(builder: (context) {
                        final primaryColor = useLiquidGlass
                            ? LiquidGlassTheme.primaryColor
                            : Theme.of(context).colorScheme.primary;
                        return Padding(
                          padding: const EdgeInsets.only(bottom: 4),
                          child: Row(
                            children: [
                              Expanded(
                                child: Text(
                                  l10n.trainingQuality,
                                  style: useLiquidGlass
                                      ? LiquidGlassTheme.bodyStyle
                                      : Theme.of(context).textTheme.bodyMedium,
                                ),
                              ),
                              IconButton(
                                icon: Icon(
                                  Icons.thumb_up_outlined,
                                  color: _trainingQuality == true ? primaryColor : VigorColors.stone,
                                ),
                                onPressed: () => setState(() => _trainingQuality = _trainingQuality == true ? null : true),
                                tooltip: l10n.good,
                              ),
                              IconButton(
                                icon: Icon(
                                  Icons.thumb_down_outlined,
                                  color: _trainingQuality == false ? primaryColor : VigorColors.stone,
                                ),
                                onPressed: () => setState(() => _trainingQuality = _trainingQuality == false ? null : false),
                                tooltip: l10n.bad,
                              ),
                            ],
                          ),
                        );
                      }),
                      Padding(
                        padding: const EdgeInsets.only(bottom: 12),
                        child: Text(
                          l10n.trainingQualityHint,
                          style: (useLiquidGlass
                                  ? LiquidGlassTheme.bodyStyle
                                  : Theme.of(context).textTheme.bodySmall)
                              ?.copyWith(color: Theme.of(context).colorScheme.onSurfaceVariant),
                        ),
                      ),
                      // reason field when quality is bad
                      if (_trainingQuality == false)
                        Padding(
                          padding: const EdgeInsets.only(bottom: 12),
                          child: TextField(
                            controller: _qualityReasonController,
                            maxLines: 1,
                            decoration: InputDecoration(
                              hintText: l10n.qualityReasonHint,
                              border: OutlineInputBorder(
                                borderRadius: BorderRadius.circular(12),
                              ),
                            ),
                            onChanged: (_) => setState(() {}),
                          ),
                        ),
                      // methodology stats (read-only, auto-generated)
                      if (widget.messagePrefix != null && widget.messagePrefix!.isNotEmpty)
                        Padding(
                          padding: const EdgeInsets.only(bottom: 12),
                          child: TextField(
                            controller: TextEditingController(text: widget.messagePrefix!.trim()),
                            readOnly: true,
                            maxLines: 1,
                            style: TextStyle(
                              color: Theme.of(context).colorScheme.onSurfaceVariant,
                            ),
                            decoration: InputDecoration(
                              filled: true,
                              fillColor: Theme.of(context).colorScheme.surfaceContainerHighest,
                              border: OutlineInputBorder(
                                borderRadius: BorderRadius.circular(12),
                                borderSide: BorderSide.none,
                              ),
                            ),
                          ),
                        ),
                      Padding(
                        padding: const EdgeInsets.only(bottom: 12),
                        child: Column(
                          children: [
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  l10n.actualDuration,
                                  style: useLiquidGlass
                                      ? LiquidGlassTheme.bodyStyle
                                      : Theme.of(context).textTheme.bodyMedium,
                                ),
                                Text(
                                  '${_durationMinutes.round()}′',
                                  style: (useLiquidGlass
                                          ? LiquidGlassTheme.bodyStyle
                                          : Theme.of(context).textTheme.bodyMedium)
                                      ?.copyWith(fontWeight: FontWeight.w600),
                                ),
                              ],
                            ),
                            Slider(
                              value: _durationMinutes.clamp(_minDuration, _maxDuration),
                              min: _minDuration,
                              max: _maxDuration,
                              divisions: 34,
                              label: '${_durationMinutes.round()} min',
                              onChanged: (v) => setState(() => _durationMinutes = v),
                            ),
                          ],
                        ),
                      ),
                      TextField(
                        controller: _messageController,
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
                      Padding(
                        padding: const EdgeInsets.only(bottom: 8),
                        child: Text(
                          l10n.exercises,
                          style: useLiquidGlass
                              ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 16)
                              : Theme.of(context).textTheme.titleMedium,
                        ),
                      ),
                      ...widget.activities.map((a) => _ExerciseFeedbackRow(
                            name: a.name,
                            exercise: a.exercise,
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
  final Exercise? exercise;
  final ExerciseFeedback feedback;
  final bool isFlagged;
  final bool showFlag;
  final ValueChanged<ExerciseFeedback> onFeedbackChanged;
  final ValueChanged<bool> onFlagChanged;

  const _ExerciseFeedbackRow({
    required this.name,
    this.exercise,
    required this.feedback,
    required this.isFlagged,
    required this.showFlag,
    required this.onFeedbackChanged,
    required this.onFlagChanged,
  });

  @override
  Widget build(BuildContext context) {
    final useLiquidGlass = PlatformHelper.useLiquidGlass;
    final l10n = AppLocalizations.of(context);
    final isActive = feedback != ExerciseFeedback.none;
    // invert: slider -2 = too easy (left), +2 = impossible (right)
    final sliderValue = -(feedback.sliderValue ?? 0).toDouble();

    final activeColor = switch (feedback) {
      ExerciseFeedback.tooEasy || ExerciseFeedback.easy => VigorColors.indigo,
      ExerciseFeedback.ok || ExerciseFeedback.none => VigorColors.stone,
      ExerciseFeedback.tooHard || ExerciseFeedback.impossible => VigorColors.persimmon,
    };

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              if (exercise != null)
                Padding(
                  padding: const EdgeInsets.only(right: VigorSpacing.sm),
                  child: GestureDetector(
                    onTap: () => ExerciseModal.show(context, exercise!),
                    child: CachedExerciseImage(
                      imageUrl: exercise!.reference,
                      width: 28,
                      height: 28,
                      isCircular: true,
                    ),
                  ),
                ),
              Expanded(
                child: Text(
                  name,
                  style: useLiquidGlass
                      ? LiquidGlassTheme.bodyStyle
                      : Theme.of(context).textTheme.bodyMedium,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              Text(
                isActive ? feedback.label(l10n) : l10n.ok,
                style: (useLiquidGlass
                        ? LiquidGlassTheme.bodyStyle
                        : Theme.of(context).textTheme.bodySmall)
                    ?.copyWith(
                  color: isActive ? activeColor : VigorColors.stone,
                  fontWeight: FontWeight.w600,
                ),
              ),
              // flag (independent from feedback, only shown in completion mode)
              if (showFlag)
                IconButton(
                  icon: Icon(
                    isFlagged ? Icons.flag : Icons.flag_outlined,
                    color: isFlagged ? VigorColors.crimson : VigorColors.stone,
                    size: 20,
                  ),
                  onPressed: () => onFlagChanged(!isFlagged),
                  tooltip: 'Flag issue',
                  visualDensity: VisualDensity.compact,
                ),
            ],
          ),
          Opacity(
            opacity: isActive ? 1.0 : 0.35,
            child: SliderTheme(
              data: SliderTheme.of(context).copyWith(
                trackHeight: 3,
                thumbShape: const RoundSliderThumbShape(enabledThumbRadius: 7),
                overlayShape: const RoundSliderOverlayShape(overlayRadius: 14),
                activeTrackColor: activeColor,
                thumbColor: activeColor,
                inactiveTrackColor: activeColor.withValues(alpha: 0.2),
              ),
              child: Slider(
                value: sliderValue,
                min: -2,
                max: 2,
                divisions: 4,
                onChanged: (v) => onFeedbackChanged(ExerciseFeedbackX.fromSliderValue(-v.round())),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
