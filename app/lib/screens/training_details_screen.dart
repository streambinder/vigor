import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_markdown/flutter_markdown.dart';
import 'package:provider/provider.dart';
import 'package:share_plus/share_plus.dart';
import 'package:url_launcher/url_launcher.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/training.dart';
import '../models/routine.dart';
import '../models/block.dart';
import '../models/activity.dart';
import '../models/activity_ext.dart';
import '../models/exercise.dart';
import '../dto/partner_info.dart';
import '../models/training_feedback.dart';
import '../providers/auth_provider.dart';
import '../services/app_event.dart';
import '../services/preferences_service.dart';
import '../services/service_locator.dart';
import '../timer/workout_timer_notifier.dart';
import '../utils/knowledge_labels.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/cached_exercise_image.dart';
import '../widgets/marquee_text.dart';
import '../widgets/user_select_dialog.dart';
import '../utils/exercise_modal.dart';
import '../utils/feedback_modal.dart';
import '../widgets/timer/inline_timer_section.dart';
import 'main_navigation.dart';

class TrainingDetailsScreen extends StatefulWidget {
  final Training training;

  const TrainingDetailsScreen({super.key, required this.training});

  @override
  State<TrainingDetailsScreen> createState() => _TrainingDetailsScreenState();
}

class _TrainingDetailsScreenState extends State<TrainingDetailsScreen> with AppEventSubscriber<TrainingDetailsScreen> {
  late Training _training;
  List<PartnerInfo> _partners = [];
  TrainingFeedback? _userFeedback;
  Map<String, dynamic>? _healthSessionData;
  final Set<String> _shufflingActivityIds = {};

  // inline timer state
  WorkoutTimerNotifier? _timerNotifier;
  final ScrollController _scrollController = ScrollController();
  final GlobalKey _timerKey = GlobalKey();
  double _timerHeight = 280;

  Training get training => _training;

  @override
  void initState() {
    super.initState();
    _training = widget.training;
    subscribeToEvents(context.read<ServiceLocator>().events, _onAppEvent);
    _loadPartners();
    if (_training.completedAt != null) _loadUserFeedback();
    if (_training.hasHealthSession) _loadHealthSession();
  }

  void _onAppEvent(AppEvent event) {
    if (!mounted) return;
    if (event is FeedbackSubmitted && event.trainingId == _training.id) {
      _loadUserFeedback();
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    _timerNotifier?.removeListener(_onTimerUpdate);
    _timerNotifier?.dispose();
    super.dispose();
  }

  Future<void> _loadPartners() async {
    final currentUserId = context.read<AuthProvider>().currentUser?.id;
    final response = await context.read<ServiceLocator>().trainingService.getPartners(training.id);
    if (response.isSuccess && mounted) {
      setState(() => _partners = (response.data ?? [])
          .where((p) => p.userId != currentUserId)
          .toList());
    }
  }

  Future<void> _loadUserFeedback() async {
    final response = await context.read<ServiceLocator>().trainingService.getUserFeedback(training.id);
    if (response.isSuccess && mounted) {
      setState(() => _userFeedback = response.data);
    }
  }

  /// fetch linked health exercise session data (HR samples, avg/max HR, zones)
  Future<void> _loadHealthSession() async {
    final locator = context.read<ServiceLocator>();
    final response = await locator.trainingService.getHealthSession(training.id);
    if (response.isSuccess && response.data != null && mounted) {
      setState(() => _healthSessionData = response.data);
    }
  }

  /// re-sync local training from the API after timer completion.
  /// the service's onDataChanged fires refreshTrainings() but it's not awaited,
  /// so we fetch the list ourselves to guarantee fresh data.
  Future<void> _refreshTraining() async {
    final locator = context.read<ServiceLocator>();
    final response = await locator.trainingService.getTrainings();
    if (response.isSuccess && mounted) {
      locator.trainingsNotifier.value = response.data;
      final updated = response.data
          ?.cast<Training?>()
          .firstWhere((t) => t!.id == _training.id, orElse: () => null);
      if (updated != null) {
        setState(() => _training = updated);
        _loadUserFeedback();
      }
    }
  }

  // --- inline timer lifecycle ---

  void _measureTimer() {
    final height = _timerKey.currentContext?.size?.height;
    if (height != null && height != _timerHeight) {
      setState(() => _timerHeight = height);
    }
  }

  void _startTimer() {
    if (_timerNotifier != null) return;
    setState(() {
      _timerNotifier = WorkoutTimerNotifier(
        training: _training,
        prefs: context.read<PreferencesService>(),
        serviceLocator: context.read<ServiceLocator>(),
        context: context,
      );
      // notification actions just trigger the same flow as ui buttons
      _timerNotifier!.onNotificationStop = () {
        // bring app to foreground and trigger stop confirmation
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (mounted) _confirmStopTimer();
        });
      };
      _timerNotifier!.onNotificationComplete = () {
        // bring app to foreground and trigger complete flow
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (mounted) _completeFromTimer(navigateOnDone: false);
        });
      };
      _timerNotifier!.initialize();
      _timerNotifier!.addListener(_onTimerUpdate);
    });
    // measure timer height + scroll to it after insertion
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _measureTimer();
      final keyContext = _timerKey.currentContext;
      if (keyContext != null) {
        Scrollable.ensureVisible(keyContext, duration: VigorAnimation.medium, curve: VigorAnimation.defaultCurve, alignment: 0.3);
      }
    });
  }

  void _stopTimer() {
    _timerNotifier?.removeListener(_onTimerUpdate);
    _timerNotifier?.dispose();
    setState(() => _timerNotifier = null);
  }

  void _onTimerUpdate() {
    if (!mounted) return;
    WidgetsBinding.instance.addPostFrameCallback((_) => _measureTimer());
    // only rebuild when discrete state changes (completion), not every tick.
    // the compact bar uses its own ListenableBuilder for per-tick updates.
    if (_timerNotifier?.workoutCompleted == true) {
      setState(() {});
    }
  }

  // --- completion + exit flows ---

  Future<void> _showFeedbackAndComplete({bool navigateOnDone = false}) async {
    if (_timerNotifier == null || _timerNotifier!.isSubmitting) return;
    final result = await FeedbackModal.show(
      context,
      _training,
      messagePrefix: _timerNotifier!.methodologyStats,
      elapsedSeconds: _timerNotifier!.accumulatedElapsedSeconds,
    );
    if (result == null || _timerNotifier == null || !mounted) return;
    _timerNotifier!.isSubmitting = true;
    await _markTrainingComplete(result);
    if (!mounted) return;
    _stopTimer();
    if (navigateOnDone) {
      _navigateToActivityScreen(context);
    } else {
      _refreshTraining();
    }
  }

  Future<void> _markTrainingComplete(FeedbackResult result) async {
    final response = await context.read<ServiceLocator>().trainingService.completeTraining(
      _training.id,
      feedback: result.feedback,
      activityFeedback: result.activityFeedback,
      completedIn: result.completedIn,
    );
    if (!response.isSuccess && mounted) {
      AdaptiveNotification.showError(
        context: context,
        message: AppLocalizations.of(context).failedToMarkComplete,
        rawError: response.error,
      );
    }
  }

  void _confirmStopTimer() {
    final l10n = AppLocalizations.of(context);
    showDialog(
      context: context,
      barrierDismissible: true,
      builder: (dialogContext) => AlertDialog(
        title: Text(l10n.stopTraining),
        content: Text(l10n.stopTrainingConfirm),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(dialogContext).pop(),
            child: Text(l10n.cancel),
          ),
          TextButton(
            onPressed: () {
              Navigator.of(dialogContext).pop();
              _stopTimer();
            },
            child: Text(l10n.stop),
          ),
        ],
      ),
    );
  }


  void _completeFromTimer({bool navigateOnDone = false}) {
    _timerNotifier?.captureEarlyExitStats();
    _showFeedbackAndComplete(navigateOnDone: navigateOnDone);
  }

  /// back button / swipe-back during active timer: offer stop or complete,
  /// both navigating to the training listing afterwards
  void _confirmTimerExit() {
    final l10n = AppLocalizations.of(context);
    showDialog(
      context: context,
      barrierDismissible: true,
      builder: (dialogContext) => AlertDialog(
        title: Text(l10n.stopTraining),
        content: Text(l10n.stopTrainingConfirm),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.of(dialogContext).pop();
              _stopTimer();
              _navigateToActivityScreen(context);
            },
            child: Text(l10n.stop),
          ),
          TextButton(
            onPressed: () {
              Navigator.of(dialogContext).pop();
              _completeFromTimer(navigateOnDone: true);
            },
            child: Text(l10n.complete),
          ),
        ],
      ),
    );
  }

  String _formatDate(DateTime date) {
    final l10n = AppLocalizations.of(context);
    final localDate = date.toLocal();
    final now = DateTime.now();
    final diff = DateTime(now.year, now.month, now.day).difference(DateTime(localDate.year, localDate.month, localDate.day)).inDays;
    if (diff == 0) return l10n.today;
    if (diff == 1) return l10n.yesterday;
    if (diff < 7) return l10n.daysAgo(diff);
    return '${localDate.day}/${localDate.month}';
  }

  String _formatDuration(int seconds) {
    final minutes = seconds ~/ 60;
    if (minutes < 60) return '${minutes}m';
    final hours = minutes ~/ 60;
    final remainingMinutes = minutes % 60;
    return remainingMinutes == 0 ? '${hours}h' : '${hours}h ${remainingMinutes}m';
  }

  String _formatTime(int seconds) {
    if (seconds < 60) return '${seconds}s';
    final minutes = seconds ~/ 60;
    final remainingSeconds = seconds % 60;
    return remainingSeconds == 0 ? '${minutes}m' : '${minutes}m ${remainingSeconds}s';
  }

  Exercise? _parseExercise(Map<String, dynamic> detail) {
    if (detail.isEmpty) return null;
    try {
      return Exercise.fromJson(detail);
    } catch (_) {
      return null;
    }
  }

  Future<void> _deleteTraining(BuildContext context) async {
    final l10n = AppLocalizations.of(context);
    final currentUserId = context.read<AuthProvider>().currentUser?.id ?? '';
    final isOwner = training.userId == currentUserId;
    final title = isOwner ? l10n.deleteTraining : l10n.leaveTraining;
    final content = isOwner ? l10n.deleteTrainingConfirmation(training.name) : l10n.leaveTrainingConfirmation(training.name);
    final actionLabel = isOwner ? l10n.delete : l10n.leave;
    final successMessage = isOwner ? l10n.trainingDeletedSuccessfully : l10n.leftTrainingSuccessfully;

    final shouldDelete = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: title,
      content: content,
      actions: [
        AdaptiveDialogAction(label: l10n.cancel, onPressed: () => Navigator.of(context).pop(false)),
        AdaptiveDialogAction(label: actionLabel, isDestructive: true, onPressed: () => Navigator.of(context).pop(true)),
      ],
    );

    if (shouldDelete == true && context.mounted) {
      final response = await context.read<ServiceLocator>().trainingService.deleteTraining(training.id);

      if (context.mounted) {
        if (response.isSuccess) {
          _navigateToActivityScreen(context);
          AdaptiveNotification.show(context: context, message: successMessage);
        } else {
          AdaptiveNotification.showError(context: context, message: l10n.failedToDeleteTraining, rawError: response.error);
        }
      }
    }
  }

  Future<void> _completeTraining(BuildContext context) async {
    final l10n = AppLocalizations.of(context);
    final trainingService = context.read<ServiceLocator>().trainingService;
    final result = await FeedbackModal.show(context, training, elapsedSeconds: training.duration);
    if (result == null) return;

    final response = await trainingService.completeTraining(
      training.id,
      feedback: result.feedback,
      activityFeedback: result.activityFeedback,
      activityReports: result.activityReports,
      completedIn: result.completedIn,
    );

    if (context.mounted) {
      if (response.isSuccess) {
        _navigateToActivityScreen(context);
        AdaptiveNotification.show(context: context, message: l10n.trainingMarkedAsComplete);
      } else {
        AdaptiveNotification.showError(context: context, message: l10n.failedToCompleteTraining, rawError: response.error);
      }
    }
  }

  Future<void> _updateFeedback(BuildContext context) async {
    final l10n = AppLocalizations.of(context);
    final trainingService = context.read<ServiceLocator>().trainingService;
    final result = await FeedbackModal.showForUpdate(context, training, existingFeedback: _userFeedback);
    if (result == null) return;

    final response = await trainingService.updateFeedback(
      training.id,
      feedback: result.feedback,
      activityFeedback: result.activityFeedback,
      completedIn: result.completedIn,
    );

    if (context.mounted) {
      if (response.isSuccess) {
        _navigateToActivityScreen(context);
        AdaptiveNotification.show(context: context, message: l10n.feedbackUpdated);
      } else {
        AdaptiveNotification.showError(context: context, message: l10n.failedToUpdateFeedback, rawError: response.error);
      }
    }
  }

  void _navigateToActivityScreen(BuildContext context) {
    Navigator.of(context).popUntil((route) => route.isFirst);
    MainNavigation.navigateToTab(1);
  }

  Future<void> _showAddPartnerDialog(BuildContext context) async {
    final l10n = AppLocalizations.of(context);
    final user = await showUserSelectDialog(context: context, title: l10n.addPartner);
    if (user == null || !context.mounted) return;

    final shouldAdd = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: l10n.addPartner,
      content: l10n.addPartnerConfirmation(user.displayName, training.name),
      actions: [
        AdaptiveDialogAction(label: l10n.cancel, onPressed: () => Navigator.of(context).pop(false)),
        AdaptiveDialogAction(label: l10n.add, onPressed: () => Navigator.of(context).pop(true)),
      ],
    );

    if (shouldAdd != true || !context.mounted) return;

    final response = await context.read<ServiceLocator>().trainingService.addPartner(training.id, user.id);

    if (context.mounted) {
      if (response.isSuccess) {
        _loadPartners();
        AdaptiveNotification.show(context: context, message: l10n.partnerAddedSuccessfully);
      } else {
        AdaptiveNotification.showError(context: context, message: l10n.failedToAddPartner, rawError: response.error);
      }
    }
  }

  Future<void> _cloneTraining(BuildContext context) async {
    final l10n = AppLocalizations.of(context);
    final currentUserId = context.read<AuthProvider>().currentUser?.id ?? '';
    if (currentUserId.isEmpty) return;

    final shouldClone = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: l10n.cloneTraining,
      content: l10n.cloneTrainingConfirmation(training.name),
      actions: [
        AdaptiveDialogAction(label: l10n.cancel, onPressed: () => Navigator.of(context).pop(false)),
        AdaptiveDialogAction(label: l10n.clone, onPressed: () => Navigator.of(context).pop(true)),
      ],
    );

    if (shouldClone != true || !context.mounted) return;

    final response = await context.read<ServiceLocator>().trainingService.copyTraining(training.id, currentUserId);

    if (context.mounted) {
      if (response.isSuccess) {
        AdaptiveNotification.show(context: context, message: l10n.trainingCloned);
        _navigateToActivityScreen(context);
      } else {
        AdaptiveNotification.showError(context: context, message: l10n.failedToCloneTraining, rawError: response.error);
      }
    }
  }

  Future<void> _shareByLink(BuildContext context) async {
    final l10n = AppLocalizations.of(context);
    final response = await context.read<ServiceLocator>().trainingService.shareTraining(training.id);

    if (!context.mounted) return;

    if (response.isSuccess && response.data != null) {
      final url = response.data!['url']!;
      try {
        await Share.share(url);
      } catch (_) {
        // fallback: copy to clipboard (e.g. web where share sheet unavailable)
        await Clipboard.setData(ClipboardData(text: url));
        if (context.mounted) {
          AdaptiveNotification.show(context: context, message: l10n.linkCopied);
        }
      }
    } else {
      AdaptiveNotification.showError(context: context, message: l10n.failedToShareTraining, rawError: response.error);
    }
  }

  Future<void> _showCopyTrainingDialog(BuildContext context) async {
    final l10n = AppLocalizations.of(context);
    final user = await showUserSelectDialog(context: context, title: l10n.shareWithUser);
    if (user == null || !context.mounted) return;

    final shouldShare = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: l10n.shareWithUser,
      content: l10n.shareTrainingConfirmation(training.name, user.displayName),
      actions: [
        AdaptiveDialogAction(label: l10n.cancel, onPressed: () => Navigator.of(context).pop(false)),
        AdaptiveDialogAction(label: l10n.share, onPressed: () => Navigator.of(context).pop(true)),
      ],
    );

    if (shouldShare != true || !context.mounted) return;

    final response = await context.read<ServiceLocator>().trainingService.copyTraining(training.id, user.id);

    if (context.mounted) {
      if (response.isSuccess) {
        AdaptiveNotification.show(context: context, message: l10n.trainingSharedSuccessfully);
      } else {
        AdaptiveNotification.showError(context: context, message: l10n.failedToShareTraining, rawError: response.error);
      }
    }
  }

  void _showReasoningDialog(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final reasoningText = training.prompt.reasoning.output?.isNotEmpty == true
        ? training.prompt.reasoning.output!
        : training.prompt.reasoning.prompt.user;
    final reasoningModel = training.prompt.reasoning.model;
    final structuringModel = training.prompt.structuring.model;

    showDialog(
      context: context,
      barrierDismissible: true,
      builder: (ctx) => AlertDialog(
        backgroundColor: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
        title: Row(
          children: [
            const Icon(Icons.psychology, color: VigorColors.stone),
            const SizedBox(width: VigorSpacing.sm),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(l10n.reasoning, style: VigorTypography.headline.copyWith(color: VigorColors.textPrimary(ctx))),
                  const SizedBox(height: VigorSpacing.xs),
                  Text(
                    '$reasoningModel → $structuringModel',
                    style: VigorTypography.caption.copyWith(color: VigorColors.stone),
                  ),
                ],
              ),
            ),
          ],
        ),
        content: SizedBox(
          width: double.maxFinite,
          child: Container(
            padding: VigorSpacing.paddingMd,
            decoration: BoxDecoration(
              color: VigorColors.stone.withValues(alpha: 0.08),
              borderRadius: VigorRadius.radiusMd,
            ),
            child: Markdown(
              data: reasoningText,
              selectable: true,
              shrinkWrap: true,
              styleSheet: MarkdownStyleSheet(
                p: VigorTypography.body.copyWith(
                  color: VigorColors.textPrimary(ctx),
                  fontSize: 14,
                ),
                h1: VigorTypography.headline.copyWith(
                  color: VigorColors.textPrimary(ctx),
                  fontSize: 18,
                ),
                h2: VigorTypography.headline.copyWith(
                  color: VigorColors.textPrimary(ctx),
                  fontSize: 16,
                ),
                h3: VigorTypography.headline.copyWith(
                  color: VigorColors.textPrimary(ctx),
                  fontSize: 14,
                ),
                code: VigorTypography.body.copyWith(
                  fontFamily: 'monospace',
                  fontSize: 13,
                  color: VigorColors.indigo,
                ),
                listBullet: VigorTypography.body.copyWith(
                  color: VigorColors.stone,
                ),
              ),
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: Text(l10n.close, style: const TextStyle(color: VigorColors.indigo)),
          ),
        ],
      ),
    );
  }


  Future<void> _showReportDialog(BuildContext context) async {
    final l10n = AppLocalizations.of(context);
    final controller = TextEditingController();
    final result = await showDialog<String>(
      context: context,
      barrierDismissible: true,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.reportIssue),
        content: TextField(
          controller: controller,
          maxLines: 4,
          decoration: InputDecoration(
            hintText: l10n.describeIssue,
            border: const OutlineInputBorder(borderRadius: VigorRadius.radiusMd),
          ),
          autofocus: true,
        ),
        actions: [
          TextButton(onPressed: () => Navigator.of(ctx).pop(), child: Text(l10n.cancel)),
          TextButton(onPressed: () => Navigator.of(ctx).pop(controller.text.trim()), child: Text(l10n.submit)),
        ],
      ),
    );

    if (result == null || result.isEmpty || !context.mounted) return;

    final response = await context.read<ServiceLocator>().trainingService.createReport(training.id, result);

    if (context.mounted) {
      if (response.isSuccess) {
        AdaptiveNotification.show(context: context, message: l10n.reportSubmitted);
      } else {
        AdaptiveNotification.showError(context: context, message: l10n.failedToSubmitReport, rawError: response.error);
      }
    }
  }

  Future<void> _shuffleActivity(Activity activity) async {
    if (_shufflingActivityIds.contains(activity.id)) return;
    final l10n = AppLocalizations.of(context);
    setState(() => _shufflingActivityIds.add(activity.id));
    final response = await context.read<ServiceLocator>().trainingService.shuffleActivity(activity.id);

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      final newActivity = response.data!;
      for (final routine in _training.routines) {
        for (final block in routine.blocks) {
          final idx = block.activities.indexWhere((a) => a.id == activity.id);
          if (idx >= 0) {
            block.activities[idx] = newActivity;
            setState(() => _shufflingActivityIds.remove(activity.id));
            AdaptiveNotification.show(context: context, message: l10n.exerciseShuffled);
            return;
          }
        }
      }
      setState(() => _shufflingActivityIds.remove(activity.id));
    } else {
      setState(() => _shufflingActivityIds.remove(activity.id));
      AdaptiveNotification.showError(context: context, message: l10n.failedToShuffleExercise, rawError: response.error);
    }
  }

  Future<void> _launchUrl(String url) async {
    final uri = Uri.tryParse(url);
    if (uri != null) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final currentUserId = context.read<AuthProvider>().currentUser?.id ?? '';
    final isOwner = training.userId == currentUserId;
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final timerActive = _timerNotifier != null;

    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, result) {
        if (didPop) return;
        if (_timerNotifier != null) {
          _confirmTimerExit();
        } else {
          _navigateToActivityScreen(context);
        }
      },
      child: AdaptiveScaffold(
        appBar: AdaptiveAppBar(
          title: MarqueeText(text: training.name),
          automaticallyImplyLeading: false,
          leading: IconButton(
            icon: const Icon(Icons.arrow_back_ios, color: VigorColors.stone),
            onPressed: () {
              if (_timerNotifier != null) {
                _confirmTimerExit();
              } else {
                _navigateToActivityScreen(context);
              }
            },
          ),
          actions: [
            _buildMenuButton(l10n, isOwner),
          ],
        ),
        body: ValueListenableBuilder<bool>(
          valueListenable: context.read<ServiceLocator>().isCalibratingNotifier,
          builder: (context, isCalibrating, _) {
            return CustomScrollView(
              slivers: [
                // pre-timer items: calibration note (if any) + header card
                SliverPadding(
                  padding: const EdgeInsets.only(left: VigorSpacing.lg, right: VigorSpacing.lg, top: VigorSpacing.lg),
                  sliver: SliverList.list(children: [
                    if (isCalibrating)
                      Padding(
                        padding: const EdgeInsets.only(bottom: VigorSpacing.md),
                        child: _buildCalibrationNote(l10n),
                      ),
                    Padding(
                      padding: const EdgeInsets.only(bottom: VigorSpacing.xl),
                      child: _buildHeaderWithActions(l10n, isDark, isOwner),
                    ),
                  ]),
                ),
                // sticky timer when active
                if (timerActive)
                  SliverPersistentHeader(
                    pinned: true,
                    delegate: _SliverTimerDelegate(
                      height: _timerHeight + VigorSpacing.md + VigorSpacing.xl,
                      child: Padding(
                        padding: const EdgeInsets.only(left: VigorSpacing.lg, right: VigorSpacing.lg, top: VigorSpacing.md, bottom: VigorSpacing.xl),
                        child: InlineTimerSection(
                          key: _timerKey,
                          notifier: _timerNotifier!,
                          onDone: _showFeedbackAndComplete,
                        ),
                      ),
                    ),
                  ),
                // post-timer items: health section, routines header, routine cards, bottom spacer
                SliverPadding(
                  padding: const EdgeInsets.only(left: VigorSpacing.lg, right: VigorSpacing.lg, bottom: VigorSpacing.lg),
                  sliver: SliverList.list(children: [
                    if (_healthSessionData != null)
                      Padding(
                        padding: const EdgeInsets.only(bottom: VigorSpacing.xl),
                        child: _buildHealthSessionSection(l10n, isDark),
                      ),
                    Padding(
                      padding: const EdgeInsets.only(bottom: VigorSpacing.md),
                      child: _buildRoutinesHeader(l10n),
                    ),
                    ...training.routines.map((routine) => Padding(
                      padding: const EdgeInsets.only(bottom: VigorSpacing.md),
                      child: _buildRoutineCard(routine, isDark),
                    )),
                    const SizedBox(height: VigorSpacing.lg),
                  ]),
                ),
              ],
            );
          },
        ),
      ),
    );
  }

  Widget _buildCalibrationNote(AppLocalizations l10n) {
    return AdaptiveCard(
      glassColor: VigorColors.indigoAdaptive(context).withValues(alpha: 0.08),
      padding: VigorSpacing.paddingMd,
      child: Text(
        l10n.calibrationTrainingNote,
        style: VigorTypography.caption.copyWith(
          color: VigorColors.textSecondary(context),
          height: 1.4,
        ),
      ),
    );
  }

  Widget _buildHeaderWithActions(AppLocalizations l10n, bool isDark, bool isOwner) {
    return Container(
      decoration: BoxDecoration(
        color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
        borderRadius: VigorRadius.radiusLg,
      ),
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: VigorSpacing.paddingLg,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildMetadataChips(l10n),
                const SizedBox(height: VigorSpacing.md),
                Text(training.description, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context), height: 1.5)),
                if (training.request.isNotEmpty) ...[
                  const SizedBox(height: VigorSpacing.md),
                  _buildInlineRequest(l10n),
                ],
                if (training.references.isNotEmpty) ...[
                  const SizedBox(height: VigorSpacing.md),
                  _buildInlineReferences(l10n),
                ],
                if (training.equipment.isNotEmpty) ...[
                  const SizedBox(height: VigorSpacing.md),
                  Wrap(
                    spacing: VigorSpacing.xs,
                    runSpacing: VigorSpacing.xs,
                    children: training.equipment.map((eq) => _buildChip(Icons.fitness_center, KnowledgeLabels.equipmentLabel(eq, l10n))).toList(),
                  ),
                ],
                if (training.goals.isNotEmpty) ...[
                  const SizedBox(height: VigorSpacing.md),
                  Wrap(
                    spacing: VigorSpacing.xs,
                    runSpacing: VigorSpacing.xs,
                    children: training.goals.map((goal) => _buildChip(Icons.track_changes, KnowledgeLabels.goalLabel(goal, l10n))).toList(),
                  ),
                ],
                if (training.muscles.isNotEmpty) ...[
                  const SizedBox(height: VigorSpacing.md),
                  Wrap(
                    spacing: VigorSpacing.xs,
                    runSpacing: VigorSpacing.xs,
                    children: training.muscles.map((muscle) => _buildChip(Icons.accessibility_new, KnowledgeLabels.muscleLabel(muscle, l10n))).toList(),
                  ),
                ],
              ],
            ),
          ),
          _buildPrimaryActions(l10n, isOwner),
        ],
      ),
    );
  }

  Widget _buildInlineRequest(AppLocalizations l10n) {
    return Theme(
      data: Theme.of(context).copyWith(dividerColor: Colors.transparent),
      child: ExpansionTile(
        tilePadding: EdgeInsets.zero,
        childrenPadding: const EdgeInsets.symmetric(vertical: VigorSpacing.xs),
        leading: const Icon(Icons.chat_bubble_outline, size: 18, color: VigorColors.stone),
        title: Text(l10n.request, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context))),
        children: [
          Padding(
            padding: const EdgeInsets.symmetric(vertical: VigorSpacing.xs),
            child: Text(
              training.request,
              style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context), fontStyle: FontStyle.italic),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildInlineReferences(AppLocalizations l10n) {
    final linkColor = VigorColors.indigoAdaptive(context);
    return Theme(
      data: Theme.of(context).copyWith(dividerColor: Colors.transparent),
      child: ExpansionTile(
        tilePadding: EdgeInsets.zero,
        childrenPadding: const EdgeInsets.symmetric(vertical: VigorSpacing.xs),
        leading: const Icon(Icons.menu_book, size: 18, color: VigorColors.stone),
        title: Text(l10n.literature, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context))),
        children: training.references.map((ref) => Padding(
          padding: const EdgeInsets.symmetric(vertical: VigorSpacing.xs),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Text(ref.excerpt, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context)), maxLines: 3, overflow: TextOverflow.ellipsis),
              ),
              const SizedBox(width: VigorSpacing.sm),
              GestureDetector(
                onTap: () => _launchUrl(ref.url),
                child: Icon(Icons.open_in_new, size: 16, color: linkColor),
              ),
            ],
          ),
        )).toList(),
      ),
    );
  }

  Widget _buildChip(IconData icon, String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
      decoration: BoxDecoration(
        color: VigorColors.stone.withValues(alpha: 0.1),
        borderRadius: VigorRadius.radiusXs,
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: VigorColors.stone),
          const SizedBox(width: VigorSpacing.xs),
          Text(label, style: VigorTypography.caption.copyWith(color: VigorColors.stone)),
        ],
      ),
    );
  }

  Widget _buildMetadataChips(AppLocalizations l10n) {
    return Wrap(
      spacing: VigorSpacing.sm,
      runSpacing: VigorSpacing.sm,
      children: [
        _buildMethodologyBadge(KnowledgeLabels.methodologyLabel(training.methodology, l10n)),
        if (training.completedAt != null && training.hasHealthSession) _buildMetaChip(Icons.monitor_heart, l10n.healthMetrics),
        _buildMetaChip(Icons.schedule, _formatDuration(training.completedIn ?? training.duration)),
        _buildMetaChip(Icons.calendar_today, _formatDate(training.completedAt ?? training.createdAt)),
        if (_partners.isNotEmpty) ..._partners.map((p) => _buildMetaChip(Icons.person, '${p.firstName} ${p.lastName}'.trim())),
        if (training.gym != null) _buildMetaChip(Icons.location_on, training.gym!.name),
        if (training.parentId != null) _buildMetaChip(Icons.copy, l10n.copied),
      ],
    );
  }

  Widget _buildMethodologyBadge(String methodology) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: VigorColors.stone.withValues(alpha: 0.1),
        borderRadius: VigorRadius.radiusXs,
      ),
      child: Text(
        methodology.toUpperCase(),
        style: VigorTypography.caption.copyWith(
          color: VigorColors.stone,
          fontWeight: FontWeight.w600,
          fontSize: 10,
        ),
      ),
    );
  }

  Widget _buildMetaChip(IconData icon, String text) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
      decoration: BoxDecoration(
        color: VigorColors.stone.withValues(alpha: 0.1),
        borderRadius: VigorRadius.radiusXs,
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: VigorColors.stone),
          const SizedBox(width: 4),
          Text(text, style: VigorTypography.data.copyWith(color: VigorColors.textSecondary(context), fontSize: 11)),
        ],
      ),
    );
  }

  /// avg/max HR, duration, and calories from linked health session
  Widget _buildHealthSessionSection(AppLocalizations l10n, bool isDark) {
    if (_healthSessionData == null) return const SizedBox.shrink();
    final data = _healthSessionData!;
    final avgHr = data['avg_hr'] as int?;
    final maxHr = data['max_hr'] as int?;
    final calories = data['calories'] as num?;
    final startedAt = data['started_at'] as String?;
    final endedAt = data['ended_at'] as String?;
    final accentColor = VigorColors.indigoAdaptive(context);

    // compute session duration
    int? durationMins;
    if (startedAt != null && endedAt != null) {
      final start = DateTime.tryParse(startedAt);
      final end = DateTime.tryParse(endedAt);
      if (start != null && end != null) {
        durationMins = end.difference(start).inMinutes;
      }
    }

    final stats = <Widget>[];
    if (avgHr != null) stats.add(Expanded(child: _buildHrStat(l10n.avgHr, '$avgHr', l10n.bpm, accentColor)));
    if (maxHr != null) stats.add(Expanded(child: _buildHrStat(l10n.maxHr, '$maxHr', l10n.bpm, VigorColors.persimmon)));
    if (durationMins != null && durationMins > 0) stats.add(Expanded(child: _buildHrStat(l10n.duration, '$durationMins', 'min', VigorColors.stone)));
    if (calories != null && calories > 0) stats.add(Expanded(child: _buildHrStat(l10n.healthDailyCalories, '${calories.round()}', 'kcal', VigorColors.stone)));

    if (stats.isEmpty) return const SizedBox.shrink();

    return Container(
      decoration: BoxDecoration(
        color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
        borderRadius: VigorRadius.radiusLg,
      ),
      padding: VigorSpacing.paddingLg,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.monitor_heart, size: 20, color: accentColor),
              const SizedBox(width: VigorSpacing.sm),
              Text(l10n.heartRate, style: VigorTypography.headline.copyWith(fontSize: 16, color: VigorColors.textPrimary(context))),
            ],
          ),
          const SizedBox(height: VigorSpacing.md),
          Row(children: stats),
        ],
      ),
    );
  }

  Widget _buildHrStat(String label, String value, String unit, Color color) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: VigorTypography.caption.copyWith(color: VigorColors.stone)),
        const SizedBox(height: 2),
        Row(
          crossAxisAlignment: CrossAxisAlignment.baseline,
          textBaseline: TextBaseline.alphabetic,
          children: [
            Text(value, style: VigorTypography.dataLarge.copyWith(color: color)),
            const SizedBox(width: 4),
            Text(unit, style: VigorTypography.caption.copyWith(color: VigorColors.stone)),
          ],
        ),
      ],
    );
  }

  Widget _buildPrimaryActions(AppLocalizations l10n, bool isOwner) {
    final isCompleted = training.completedAt != null;
    final indigoColor = VigorColors.indigoAdaptive(context);

    return Row(
      children: [
        // stop / start toggle
        Expanded(child: _timerNotifier != null
          ? _buildActionButton(
              icon: Icons.stop,
              label: l10n.stop,
              color: indigoColor,
              onPressed: _confirmStopTimer,
            )
          : _buildActionButton(
              icon: Icons.timer,
              label: l10n.timer,
              color: indigoColor,
              onPressed: _startTimer,
            ),
        ),
        if (_timerNotifier != null)
          // complete = feedback modal + complete training (with timer stats)
          Expanded(child: _buildActionButton(
            icon: Icons.check_circle_outline,
            label: l10n.complete,
            color: VigorColors.persimmon,
            onPressed: _completeFromTimer,
          ))
        else if (!isCompleted)
          Expanded(child: _buildActionButton(
            icon: Icons.check_circle_outline,
            label: l10n.complete,
            color: VigorColors.persimmon,
            onPressed: () => _completeTraining(context),
          ))
        else
          Expanded(child: _buildActionButton(
            icon: Icons.rate_review,
            label: l10n.feedback,
            color: indigoColor,
            onPressed: () => _updateFeedback(context),
          )),
      ],
    );
  }

  Widget _buildMenuButton(AppLocalizations l10n, bool isOwner) {
    return PopupMenuButton<String>(
      icon: const Icon(Icons.more_vert, color: VigorColors.stone),
      onSelected: (value) {
        switch (value) {
          case 'clone':
            _cloneTraining(context);
            break;
          case 'share':
            _shareByLink(context);
            break;
          case 'share_with_user':
            _showCopyTrainingDialog(context);
            break;
          case 'add_partner':
            _showAddPartnerDialog(context);
            break;
          case 'reasoning':
            _showReasoningDialog(context);
            break;
          case 'report':
            _showReportDialog(context);
            break;
          case 'delete':
            _deleteTraining(context);
            break;
        }
      },
      itemBuilder: (context) => [
        PopupMenuItem(
          value: 'clone',
          child: Row(
            children: [
              const Icon(Icons.copy, size: 20, color: VigorColors.stone),
              const SizedBox(width: VigorSpacing.sm),
              Text(l10n.clone, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
            ],
          ),
        ),
        PopupMenuItem(
          value: 'share',
          child: Row(
            children: [
              const Icon(Icons.link, size: 20, color: VigorColors.stone),
              const SizedBox(width: VigorSpacing.sm),
              Text(l10n.shareByLink, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
            ],
          ),
        ),
        PopupMenuItem(
          value: 'share_with_user',
          child: Row(
            children: [
              const Icon(Icons.share, size: 20, color: VigorColors.stone),
              const SizedBox(width: VigorSpacing.sm),
              Text(l10n.shareWithUser, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
            ],
          ),
        ),
        if (isOwner)
          PopupMenuItem(
            value: 'add_partner',
            child: Row(
              children: [
                const Icon(Icons.person_add, size: 20, color: VigorColors.stone),
                const SizedBox(width: VigorSpacing.sm),
                Text(l10n.addPartner, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
              ],
            ),
          ),
        PopupMenuItem(
          value: 'reasoning',
          child: Row(
            children: [
              const Icon(Icons.psychology, size: 20, color: VigorColors.stone),
              const SizedBox(width: VigorSpacing.sm),
              Text(l10n.showAiReasoning, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
            ],
          ),
        ),
        const PopupMenuDivider(),
        PopupMenuItem(
          value: 'report',
          child: Row(
            children: [
              const Icon(Icons.flag_outlined, size: 20, color: VigorColors.stone),
              const SizedBox(width: VigorSpacing.sm),
              Text(l10n.reportIssue, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
            ],
          ),
        ),
        PopupMenuItem(
          value: 'delete',
          child: Row(
            children: [
              const Icon(Icons.delete, size: 20, color: VigorColors.error),
              const SizedBox(width: VigorSpacing.sm),
              Text(isOwner ? l10n.delete : l10n.leave, style: VigorTypography.body.copyWith(color: VigorColors.error)),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildActionButton({required IconData icon, required String label, required Color color, VoidCallback? onPressed}) {
    final isDisabled = onPressed == null;
    return GestureDetector(
      onTap: onPressed,
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: VigorSpacing.md),
        color: isDisabled ? VigorColors.stone.withValues(alpha: 0.2) : color,
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(icon, size: 18, color: isDisabled ? VigorColors.stone : Colors.white),
            const SizedBox(width: VigorSpacing.sm),
            Flexible(child: Text(label, style: VigorTypography.label.copyWith(color: isDisabled ? VigorColors.stone : Colors.white), overflow: TextOverflow.ellipsis)),
          ],
        ),
      ),
    );
  }

  Widget _buildRoutinesHeader(AppLocalizations l10n) {
    return Row(
      children: [
        const Icon(Icons.list_alt, color: VigorColors.stone, size: 24),
        const SizedBox(width: VigorSpacing.sm),
        Text(l10n.trainingRoutines, style: VigorTypography.headline.copyWith(fontSize: 22, color: VigorColors.textPrimary(context))),
      ],
    );
  }

  Widget _buildRoutineCard(Routine routine, bool isDark) {
    return Container(
      decoration: BoxDecoration(
        color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
        borderRadius: VigorRadius.radiusLg,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // routine header
          Stack(
            children: [
              Container(
                width: double.infinity,
                padding: VigorSpacing.paddingMd,
                decoration: BoxDecoration(
                  color: VigorColors.stone.withValues(alpha: 0.05),
                  borderRadius: const BorderRadius.only(topLeft: Radius.circular(VigorRadius.lg), topRight: Radius.circular(VigorRadius.lg)),
                ),
                child: Center(
                  child: Text(routine.type.toUpperCase(), style: VigorTypography.headline.copyWith(color: VigorColors.stone, fontSize: 14)),
                ),
              ),
              if (routine.rest > 0)
                Positioned(
                  top: VigorSpacing.sm,
                  right: VigorSpacing.sm,
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
                    decoration: BoxDecoration(
                      color: VigorColors.stone.withValues(alpha: 0.1),
                      borderRadius: VigorRadius.radiusFull,
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(Icons.timer, size: 12, color: VigorColors.stone),
                        const SizedBox(width: VigorSpacing.xs),
                        Text('${routine.rest}s', style: VigorTypography.data.copyWith(color: VigorColors.stone, fontWeight: FontWeight.w600)),
                      ],
                    ),
                  ),
                ),
            ],
          ),
          // blocks
          Column(
            children: routine.blocks.asMap().entries.expand((entry) => [
              if (entry.key > 0) Divider(height: 1, thickness: 1, color: VigorColors.stone.withValues(alpha: 0.2)),
              _buildBlockCard(entry.value, entry.key + 1, routine.blocks.length, isDark),
            ]).toList(),
          ),
        ],
      ),
    );
  }

  Widget _buildBlockCard(Block block, int blockNumber, int totalBlocks, bool isDark) {
    final showBlockLabel = totalBlocks > 1;
    final hasChips = block.repeats > 1 || block.rest > 0;

    return Stack(
      clipBehavior: Clip.none,
      children: [
        Padding(
          padding: VigorSpacing.paddingMd,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              if (showBlockLabel) ...[
                Center(
                  child: Text('Block $blockNumber', style: VigorTypography.caption.copyWith(fontWeight: FontWeight.bold, color: VigorColors.stone)),
                ),
                const SizedBox(height: VigorSpacing.md),
              ],
              ...block.activities.asMap().entries.map((entry) => Padding(
                padding: EdgeInsets.only(bottom: entry.key < block.activities.length - 1 ? VigorSpacing.md : 0),
                child: _buildActivityRow(entry.value, isDark, isFirstInBlock: entry.key == 0 && hasChips && totalBlocks == 1),
              )),
            ],
          ),
        ),
        if (hasChips)
          Positioned(
            top: VigorSpacing.sm,
            right: VigorSpacing.sm,
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (block.repeats > 1)
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
                    decoration: BoxDecoration(
                      color: VigorColors.stone.withValues(alpha: 0.1),
                      borderRadius: VigorRadius.radiusSm,
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(Icons.repeat, size: 12, color: VigorColors.stone),
                        const SizedBox(width: VigorSpacing.xs),
                        Text('${block.repeats}x', style: VigorTypography.data.copyWith(fontWeight: FontWeight.bold, color: VigorColors.stone)),
                      ],
                    ),
                  ),
                if (block.repeats > 1 && block.rest > 0)
                  const SizedBox(width: VigorSpacing.xs),
                if (block.rest > 0)
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
                    decoration: BoxDecoration(
                      color: VigorColors.stone.withValues(alpha: 0.1),
                      borderRadius: VigorRadius.radiusSm,
                    ),
                    child: Text('${block.rest}s', style: VigorTypography.data.copyWith(fontWeight: FontWeight.bold, color: VigorColors.stone)),
                  ),
              ],
            ),
          ),
      ],
    );
  }

  Widget _buildActivityRow(Activity activity, bool isDark, {bool isFirstInBlock = false}) {
    final exercise = _parseExercise(activity.detail);
    final hasValidImage = exercise != null && CachedExerciseImage.isValidUrl(exercise.reference);

    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(
          child: GestureDetector(
            onTap: exercise != null ? () => ExerciseModal.show(context, exercise) : null,
            behavior: HitTestBehavior.opaque,
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // thumbnail
                SizedBox(
                  width: 72,
                  height: 72,
                  child: hasValidImage
                      ? CachedExerciseImage(
                          imageUrl: exercise.reference,
                          width: 72,
                          height: 72,
                          borderRadius: VigorRadius.radiusMd,
                        )
                      : Container(
                          decoration: BoxDecoration(
                            color: VigorColors.stone.withValues(alpha: 0.1),
                            borderRadius: VigorRadius.radiusMd,
                          ),
                          child: Center(child: Icon(Icons.fitness_center, size: 72 * 0.4, color: VigorColors.stone.withValues(alpha: 0.5))),
                        ),
                ),
                const SizedBox(width: VigorSpacing.md),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      MarqueeText(text: activity.name, style: VigorTypography.body.copyWith(fontWeight: FontWeight.w600, color: VigorColors.textPrimary(context))),
                      if (exercise != null && (exercise.equipment.isNotEmpty || exercise.muscles.isNotEmpty || activity.modifiers.isNotEmpty)) ...[
                        const SizedBox(height: VigorSpacing.xs),
                        _buildExerciseDetails(exercise, activity.modifiers),
                      ],
                      const SizedBox(height: VigorSpacing.sm),
                      _buildActivityTags(activity),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
        if (training.completedAt == null)
          ValueListenableBuilder<bool>(
            valueListenable: context.read<ServiceLocator>().isCalibratingNotifier,
            builder: (context, isCalibrating, _) {
              if (isCalibrating) return const SizedBox.shrink();
              return SizedBox(
                height: 72,
                child: Padding(
                  // move refresh button down if first exercise in single-block routine with chips
                  padding: EdgeInsets.only(top: isFirstInBlock ? VigorSpacing.xl : 0),
                  child: Center(
                    child: _ShuffleButton(
                      isLoading: _shufflingActivityIds.contains(activity.id),
                      onTap: () => _shuffleActivity(activity),
                    ),
                  ),
                ),
              );
            },
          ),
      ],
    );
  }

  Widget _buildExerciseDetails(Exercise exercise, List<String> modifiers) {
    final l10n = AppLocalizations.of(context);
    return Wrap(
      spacing: VigorSpacing.sm,
      runSpacing: VigorSpacing.xs,
      children: [
        if (exercise.muscles.isNotEmpty) _buildDetailChip(Icons.accessibility_new, exercise.muscles.take(2).map((m) => KnowledgeLabels.muscleLabel(m, l10n)).join(' · ')),
        if (exercise.equipment.isNotEmpty) _buildDetailChip(Icons.fitness_center, exercise.equipment.map((e) => KnowledgeLabels.equipmentLabel(e, l10n)).join(' · ')),
        if (modifiers.isNotEmpty) _buildDetailChip(Icons.tune, modifiers.map((m) => KnowledgeLabels.modifierLabel(m, l10n)).join(' · ')),
      ],
    );
  }

  Widget _buildDetailChip(IconData icon, String text) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 12, color: VigorColors.stone),
        const SizedBox(width: VigorSpacing.xs),
        Flexible(child: Text(text, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context)), overflow: TextOverflow.ellipsis)),
      ],
    );
  }

  Widget _buildActivityTags(Activity activity) {
    return Wrap(
      spacing: VigorSpacing.xs,
      runSpacing: VigorSpacing.xs,
      children: [
        if (activity.reps > 0) _buildTag(Icons.repeat, '${activity.reps}'),
        if (activity.weightKg > 0) _buildTag(Icons.fitness_center, '${activity.weightKgDisplay}kg'),
        if (activity.duration > 0) _buildTag(Icons.timer, _formatTime(activity.duration)),
        if (activity.rest > 0) _buildTag(Icons.hourglass_bottom, '${activity.rest}s'),
      ],
    );
  }

  Widget _buildTag(IconData icon, String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
      decoration: BoxDecoration(
        color: VigorColors.stone.withValues(alpha: 0.1),
        borderRadius: VigorRadius.radiusSm,
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 11, color: VigorColors.stone),
          const SizedBox(width: VigorSpacing.xs),
          Text(label, style: VigorTypography.data.copyWith(fontSize: 11, fontWeight: FontWeight.w600, color: VigorColors.stone)),
        ],
      ),
    );
  }
}

class _SliverTimerDelegate extends SliverPersistentHeaderDelegate {
  final double height;
  final Widget child;

  const _SliverTimerDelegate({required this.height, required this.child});

  @override
  double get maxExtent => height;

  @override
  double get minExtent => height;

  @override
  Widget build(BuildContext context, double shrinkOffset, bool overlapsContent) {
    final shadowOpacity = (shrinkOffset / 40).clamp(0.0, 1.0);
    return Stack(
      clipBehavior: Clip.none,
      children: [
        // shadow layer — only visible when pinned, sits behind content
        if (shadowOpacity > 0)
          Positioned.fill(
            top: VigorSpacing.lg,
            child: DecoratedBox(
              decoration: BoxDecoration(
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.15 * shadowOpacity),
                    blurRadius: 24 * shadowOpacity,
                    offset: Offset(0, 2 * shadowOpacity),
                  ),
                ],
              ),
            ),
          ),
        // let the child size itself naturally (ignoring the tight height constraint
        // from the sliver) so _timerKey can measure intrinsic height, then clip any
        // overflow that occurs during the first frame before measurement catches up
        ClipRect(
          child: OverflowBox(
            minHeight: 0,
            maxHeight: double.infinity,
            alignment: Alignment.topLeft,
            child: child,
          ),
        ),
      ],
    );
  }

  @override
  bool shouldRebuild(_SliverTimerDelegate oldDelegate) => height != oldDelegate.height;
}

class _ShuffleButton extends StatefulWidget {
  final bool isLoading;
  final VoidCallback onTap;

  const _ShuffleButton({required this.isLoading, required this.onTap});

  @override
  State<_ShuffleButton> createState() => _ShuffleButtonState();
}

class _ShuffleButtonState extends State<_ShuffleButton> with SingleTickerProviderStateMixin {
  late final AnimationController _controller = AnimationController(
    vsync: this,
    duration: const Duration(milliseconds: 800),
  );

  @override
  void initState() {
    super.initState();
    if (widget.isLoading) _controller.repeat();
  }

  @override
  void didUpdateWidget(_ShuffleButton oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.isLoading && !_controller.isAnimating) {
      _controller.repeat();
    } else if (!widget.isLoading && _controller.isAnimating) {
      // finish current rotation cleanly instead of snapping
      _controller.forward().whenComplete(() {
        if (!widget.isLoading) _controller.stop();
      });
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: widget.isLoading ? null : widget.onTap,
      child: Container(
        padding: VigorSpacing.paddingSm,
        decoration: BoxDecoration(
          color: VigorColors.stone.withValues(alpha: 0.1),
          borderRadius: VigorRadius.radiusFull,
        ),
        child: RotationTransition(
          turns: _controller,
          child: const Icon(Icons.refresh, size: 18, color: VigorColors.stone),
        ),
      ),
    );
  }
}
