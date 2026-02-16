import 'dart:math';
import 'package:flutter/material.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/activity_ext.dart';
import '../models/training.dart';
import '../models/training_feedback.dart';
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
        feedback: TrainingFeedback(qualityReason: '', message: messagePrefix ?? ''),
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
  static Future<FeedbackResult?> showForUpdate(BuildContext context, Training training) async {
    final activities = _getWorkActivities(training);
    if (activities.isEmpty) {
      return FeedbackResult(
        feedback: TrainingFeedback(qualityReason: '', message: ''),
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
        initialFeedback: training.feedback,
        elapsedSeconds: training.completedIn,
      ),
    );
  }
}

class _FeedbackDialogContent extends StatefulWidget {
  final List<({String id, String exerciseId, String name, String? feedback})> activities;
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
    _exerciseFeedback = {
      for (final a in widget.activities)
        a.id: ExerciseFeedbackX.fromApiValue(a.feedback),
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
        quality: _trainingQuality,
        qualityReason: _trainingQuality == false ? _qualityReasonController.text.trim() : '',
        message: message,
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
