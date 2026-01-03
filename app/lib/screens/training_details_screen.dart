import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:url_launcher/url_launcher.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/training.dart';
import '../models/routine.dart';
import '../models/block.dart';
import '../models/activity.dart';
import '../models/exercise.dart';
import '../providers/auth_provider.dart';
import '../services/service_locator.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/cached_exercise_image.dart';
import '../widgets/user_select_dialog.dart';
import '../utils/exercise_modal.dart';
import '../utils/feedback_modal.dart';
import 'main_navigation.dart';
import 'tabata_timer_screen.dart';

class TrainingDetailsScreen extends StatefulWidget {
  final Training training;

  const TrainingDetailsScreen({super.key, required this.training});

  @override
  State<TrainingDetailsScreen> createState() => _TrainingDetailsScreenState();
}

class _TrainingDetailsScreenState extends State<TrainingDetailsScreen> {
  late Training _training;
  int _partnerCount = 0;

  Training get training => _training;

  @override
  void initState() {
    super.initState();
    _training = widget.training;
    _loadPartners();
  }

  Future<void> _loadPartners() async {
    final response = await context.read<ServiceLocator>().trainingService.getPartners(training.id);
    if (response.isSuccess && mounted) {
      setState(() => _partnerCount = response.data?.length ?? 0);
    }
  }

  String _formatDate(DateTime date) => '${date.day.toString().padLeft(2, '0')}/${date.month.toString().padLeft(2, '0')}/${date.year}';

  String _formatDuration(int seconds) {
    final minutes = seconds ~/ 60;
    if (minutes < 60) return '$minutes min';
    final hours = minutes ~/ 60;
    final remainingMinutes = minutes % 60;
    return remainingMinutes == 0 ? '$hours hr' : '$hours hr $remainingMinutes min';
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
          Navigator.of(context).pop(true);
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
    final result = await FeedbackModal.show(context, training);
    if (result == null) return;

    final response = await trainingService.completeTraining(
      training.id,
      feedback: result.feedback,
      activityFeedback: result.activityFeedback,
      activityReports: result.activityReports,
    );

    if (context.mounted) {
      if (response.isSuccess) {
        Navigator.of(context).pop(true);
        AdaptiveNotification.show(context: context, message: l10n.trainingMarkedAsComplete);
      } else {
        AdaptiveNotification.showError(context: context, message: l10n.failedToCompleteTraining, rawError: response.error);
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
        Navigator.of(context).pop(true);
      } else {
        AdaptiveNotification.showError(context: context, message: l10n.failedToCloneTraining, rawError: response.error);
      }
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
    final r = training.reasoning;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    showDialog(
      context: context,
      barrierDismissible: true,
      builder: (ctx) => AlertDialog(
        backgroundColor: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
        title: Row(
          children: [
            RepaintBoundary(
              child: ShaderMask(
                shaderCallback: (bounds) => const LinearGradient(colors: [VigorColors.orange, VigorColors.electricBlue]).createShader(bounds),
                child: const Icon(Icons.psychology, color: Colors.white),
              ),
            ),
            const SizedBox(width: VigorSpacing.sm),
            Text(l10n.reasoning, style: VigorTypography.headline.copyWith(color: VigorColors.textPrimary(ctx))),
          ],
        ),
        content: SizedBox(
          width: double.maxFinite,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                if (r.constraints.isNotEmpty) _buildReasoningSection(title: l10n.constraints, items: r.constraints, color: VigorColors.warning),
                if (r.typeSelection.isNotEmpty) _buildReasoningText(title: l10n.typeSelection, text: r.typeSelection, color: VigorColors.electricBlue),
                _buildReasoningText(title: l10n.strategy, text: r.strategy, color: VigorColors.success),
                if (r.progression.summary.isNotEmpty || r.progression.adjustments.isNotEmpty)
                  _buildProgressionSection(l10n, r),
                if (r.factsApplied.isNotEmpty) _buildReasoningSection(title: l10n.researchApplied, items: r.factsApplied, color: Colors.purple),
                if (r.targetMuscles.isNotEmpty) _buildMuscleChips(l10n, r.targetMuscles),
                if (r.exercises.isNotEmpty) _buildReasoningSection(title: l10n.exercises, items: r.exercises, color: VigorColors.orange),
              ],
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: Text(l10n.close, style: const TextStyle(color: VigorColors.orange)),
          ),
        ],
      ),
    );
  }

  Widget _buildReasoningSection({required String title, required List<String> items, required Color color}) {
    return Container(
      margin: const EdgeInsets.only(bottom: VigorSpacing.md),
      padding: VigorSpacing.paddingMd,
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: VigorRadius.radiusMd,
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(title, style: VigorTypography.label.copyWith(color: color, fontWeight: FontWeight.w600)),
          const SizedBox(height: VigorSpacing.sm),
          ...items.map((item) => Padding(
            padding: const EdgeInsets.only(bottom: VigorSpacing.xs),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  margin: const EdgeInsets.only(top: 6, right: VigorSpacing.sm),
                  width: 5,
                  height: 5,
                  decoration: BoxDecoration(color: color, shape: BoxShape.circle),
                ),
                Expanded(child: Text(item, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context)))),
              ],
            ),
          )),
        ],
      ),
    );
  }

  Widget _buildReasoningText({required String title, required String text, required Color color}) {
    return Container(
      margin: const EdgeInsets.only(bottom: VigorSpacing.md),
      padding: VigorSpacing.paddingMd,
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: VigorRadius.radiusMd,
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(title, style: VigorTypography.label.copyWith(color: color, fontWeight: FontWeight.w600)),
          const SizedBox(height: VigorSpacing.sm),
          Text(text, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
        ],
      ),
    );
  }

  Widget _buildProgressionSection(AppLocalizations l10n, reasoning) {
    return Container(
      margin: const EdgeInsets.only(bottom: VigorSpacing.md),
      padding: VigorSpacing.paddingMd,
      decoration: BoxDecoration(
        color: VigorColors.electricBlue.withValues(alpha: 0.1),
        borderRadius: VigorRadius.radiusMd,
        border: Border.all(color: VigorColors.electricBlue.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(l10n.progression, style: VigorTypography.label.copyWith(color: VigorColors.electricBlue, fontWeight: FontWeight.w600)),
          const SizedBox(height: VigorSpacing.sm),
          if (reasoning.progression.summary.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(bottom: VigorSpacing.sm),
              child: Text(reasoning.progression.summary, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
            ),
          ...reasoning.progression.adjustments.map((a) => Padding(
            padding: const EdgeInsets.only(bottom: VigorSpacing.xs),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  margin: const EdgeInsets.only(top: 6, right: VigorSpacing.sm),
                  width: 5,
                  height: 5,
                  decoration: const BoxDecoration(color: VigorColors.electricBlue, shape: BoxShape.circle),
                ),
                Expanded(child: Text('${a.exercise}: ${a.adjustment} (${a.reason})', style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context)))),
              ],
            ),
          )),
        ],
      ),
    );
  }

  Widget _buildMuscleChips(AppLocalizations l10n, List<String> muscles) {
    return Container(
      margin: const EdgeInsets.only(bottom: VigorSpacing.md),
      padding: VigorSpacing.paddingMd,
      decoration: BoxDecoration(
        color: Colors.red.withValues(alpha: 0.1),
        borderRadius: VigorRadius.radiusMd,
        border: Border.all(color: Colors.red.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(l10n.targetMuscles, style: VigorTypography.label.copyWith(color: Colors.red.shade700, fontWeight: FontWeight.w600)),
          const SizedBox(height: VigorSpacing.sm),
          Wrap(
            spacing: VigorSpacing.xs,
            runSpacing: VigorSpacing.xs,
            children: muscles.map((m) => Container(
              padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
              decoration: BoxDecoration(color: Colors.red.shade700, borderRadius: VigorRadius.radiusFull),
              child: Text(m, style: VigorTypography.caption.copyWith(color: Colors.white, fontWeight: FontWeight.w500)),
            )).toList(),
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
    final l10n = AppLocalizations.of(context);
    final response = await context.read<ServiceLocator>().trainingService.shuffleActivity(activity.id);

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      final newActivity = response.data!;
      for (final routine in _training.routines) {
        for (final block in routine.blocks) {
          final idx = block.activities.indexWhere((a) => a.id == activity.id);
          if (idx >= 0) {
            block.activities[idx] = newActivity;
            setState(() {});
            AdaptiveNotification.show(context: context, message: l10n.exerciseShuffled);
            return;
          }
        }
      }
    } else {
      AdaptiveNotification.showError(context: context, message: l10n.failedToShuffleExercise, rawError: response.error);
    }
  }

  Future<void> _launchUrl(String url) async {
    final uri = Uri.tryParse(url);
    if (uri != null && await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final currentUserId = context.read<AuthProvider>().currentUser?.id ?? '';
    final isOwner = training.userId == currentUserId;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, result) {
        if (!didPop) _navigateToActivityScreen(context);
      },
      child: AdaptiveScaffold(
        appBar: AdaptiveAppBar(
          title: Text(training.name),
          automaticallyImplyLeading: false,
          leading: IconButton(
            icon: const Icon(Icons.arrow_back_ios, color: VigorColors.orange),
            onPressed: () => _navigateToActivityScreen(context),
          ),
        ),
        body: ListView(
          padding: VigorSpacing.paddingLg,
          children: [
            _buildHeader(l10n, isDark),
            const SizedBox(height: VigorSpacing.lg),
            _buildMetadataChips(l10n),
            if (training.references.isNotEmpty) ...[
              const SizedBox(height: VigorSpacing.md),
              _buildReferencesSection(l10n),
            ],
            const SizedBox(height: VigorSpacing.lg),
            _buildPrimaryActions(l10n, isOwner),
            const SizedBox(height: VigorSpacing.sm),
            _buildSecondaryActions(l10n, isOwner),
            const SizedBox(height: VigorSpacing.sm),
            _buildDangerActions(l10n, isOwner),
            const SizedBox(height: VigorSpacing.xl),
            _buildRoutinesHeader(l10n),
            const SizedBox(height: VigorSpacing.md),
            ...training.routines.map((routine) => Padding(
              padding: const EdgeInsets.only(bottom: VigorSpacing.md),
              child: _buildRoutineCard(routine, isDark),
            )),
            const SizedBox(height: VigorSpacing.xxl),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader(AppLocalizations l10n, bool isDark) {
    return Container(
      padding: VigorSpacing.paddingLg,
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [VigorColors.orange.withValues(alpha: 0.15), VigorColors.electricBlue.withValues(alpha: 0.15)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: VigorRadius.radiusLg,
        border: Border.all(color: isDark ? Colors.white.withValues(alpha: 0.1) : Colors.black.withValues(alpha: 0.05)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(training.description, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context), height: 1.5)),
          if (training.equipment.isNotEmpty) ...[
            const SizedBox(height: VigorSpacing.md),
            Wrap(
              spacing: VigorSpacing.xs,
              runSpacing: VigorSpacing.xs,
              children: training.equipment.map((eq) => Container(
                padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
                decoration: BoxDecoration(
                  color: VigorColors.electricBlue.withValues(alpha: 0.15),
                  borderRadius: VigorRadius.radiusFull,
                  border: Border.all(color: VigorColors.electricBlue.withValues(alpha: 0.3)),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.fitness_center, size: 12, color: VigorColors.electricBlue),
                    const SizedBox(width: VigorSpacing.xs),
                    Text(eq, style: VigorTypography.caption.copyWith(color: VigorColors.electricBlue)),
                  ],
                ),
              )).toList(),
            ),
          ],
          if (training.goals.isNotEmpty) ...[
            const SizedBox(height: VigorSpacing.md),
            Wrap(
              spacing: VigorSpacing.xs,
              runSpacing: VigorSpacing.xs,
              children: training.goals.map((goal) => Container(
                padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
                decoration: BoxDecoration(
                  color: VigorColors.success.withValues(alpha: 0.15),
                  borderRadius: VigorRadius.radiusFull,
                  border: Border.all(color: VigorColors.success.withValues(alpha: 0.3)),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.track_changes, size: 12, color: VigorColors.success),
                    const SizedBox(width: VigorSpacing.xs),
                    Text(goal, style: VigorTypography.caption.copyWith(color: VigorColors.success)),
                  ],
                ),
              )).toList(),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildMetadataChips(AppLocalizations l10n) {
    return Wrap(
      spacing: VigorSpacing.sm,
      runSpacing: VigorSpacing.sm,
      children: [
        if (_partnerCount > 0) _buildMetaChip(Icons.people, '${1 + _partnerCount}', VigorColors.electricBlue),
        if (training.gym != null) _buildMetaChip(Icons.location_on, training.gym!.name, VigorColors.success),
        _buildMetaChip(Icons.tune, training.type, VigorColors.orange),
        _buildMetaChip(Icons.schedule, _formatDuration(training.duration), VigorColors.warning),
        _buildMetaChip(Icons.calendar_today, _formatDate(training.completedAt ?? training.createdAt), VigorColors.textSecondary(context)),
        if (training.parentId != null) _buildMetaChip(Icons.copy, l10n.copied, Colors.purple),
      ],
    );
  }

  Widget _buildMetaChip(IconData icon, String text, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.sm),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: VigorRadius.radiusFull,
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: color),
          const SizedBox(width: VigorSpacing.xs),
          Text(text, style: VigorTypography.caption.copyWith(color: color, fontWeight: FontWeight.w500)),
        ],
      ),
    );
  }

  Widget _buildReferencesSection(AppLocalizations l10n) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Container(
      decoration: BoxDecoration(
        color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
        borderRadius: VigorRadius.radiusMd,
        border: Border.all(color: isDark ? Colors.white.withValues(alpha: 0.1) : Colors.black.withValues(alpha: 0.05)),
      ),
      child: Theme(
        data: Theme.of(context).copyWith(dividerColor: Colors.transparent),
        child: ExpansionTile(
          leading: const Icon(Icons.science, size: 20, color: VigorColors.electricBlue),
          title: Text(l10n.references, style: VigorTypography.label.copyWith(color: VigorColors.textSecondary(context))),
          childrenPadding: const EdgeInsets.only(left: VigorSpacing.md, right: VigorSpacing.md, bottom: VigorSpacing.md),
          children: training.references.map((url) => GestureDetector(
            onTap: () => _launchUrl(url),
            child: Padding(
              padding: const EdgeInsets.symmetric(vertical: VigorSpacing.xs),
              child: Row(
                children: [
                  const Icon(Icons.link, size: 14, color: VigorColors.electricBlue),
                  const SizedBox(width: VigorSpacing.sm),
                  Expanded(
                    child: Text(url, style: VigorTypography.caption.copyWith(color: VigorColors.electricBlue, decoration: TextDecoration.underline), overflow: TextOverflow.ellipsis),
                  ),
                ],
              ),
            ),
          )).toList(),
        ),
      ),
    );
  }

  Widget _buildPrimaryActions(AppLocalizations l10n, bool isOwner) {
    return Row(
      children: [
        Expanded(child: _buildGradientButton(Icons.timer, l10n.startTraining, () async {
          final completed = await Navigator.of(context).push<bool>(MaterialPageRoute(builder: (context) => TabataTimerScreen(training: training)));
          if (completed == true && mounted) Navigator.of(context).pop(true);
        })),
        if (training.completedAt == null) ...[
          const SizedBox(width: VigorSpacing.sm),
          Expanded(child: _buildGradientButton(Icons.check_circle_outline, l10n.markAsComplete, isOwner ? () => _completeTraining(context) : null)),
        ],
      ],
    );
  }

  Widget _buildGradientButton(IconData icon, String label, VoidCallback? onPressed) {
    final isDisabled = onPressed == null;
    return GestureDetector(
      onTap: onPressed,
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: VigorSpacing.md),
        decoration: BoxDecoration(
          gradient: isDisabled ? null : const LinearGradient(colors: [VigorColors.orange, VigorColors.electricBlue]),
          color: isDisabled ? VigorColors.textMuted(context).withValues(alpha: 0.3) : null,
          borderRadius: VigorRadius.radiusMd,
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(icon, size: 18, color: isDisabled ? VigorColors.textMuted(context) : Colors.white),
            const SizedBox(width: VigorSpacing.sm),
            Flexible(child: Text(label, style: VigorTypography.label.copyWith(color: isDisabled ? VigorColors.textMuted(context) : Colors.white), overflow: TextOverflow.ellipsis)),
          ],
        ),
      ),
    );
  }

  Widget _buildSecondaryActions(AppLocalizations l10n, bool isOwner) {
    return Column(
      children: [
        Row(
          children: [
            Expanded(child: _buildOutlineButton(Icons.copy, l10n.cloneTraining, VigorColors.orange, () => _cloneTraining(context))),
            const SizedBox(width: VigorSpacing.sm),
            Expanded(child: _buildOutlineButton(Icons.share, l10n.shareWithUser, VigorColors.electricBlue, () => _showCopyTrainingDialog(context))),
          ],
        ),
        const SizedBox(height: VigorSpacing.sm),
        Row(
          children: [
            if (isOwner) ...[
              Expanded(child: _buildOutlineButton(Icons.person_add, l10n.addPartner, VigorColors.success, () => _showAddPartnerDialog(context))),
              const SizedBox(width: VigorSpacing.sm),
            ],
            Expanded(child: _buildOutlineButton(Icons.psychology, l10n.showAiReasoning, Colors.purple, () => _showReasoningDialog(context))),
          ],
        ),
      ],
    );
  }

  Widget _buildDangerActions(AppLocalizations l10n, bool isOwner) {
    return Row(
      children: [
        Expanded(child: _buildOutlineButton(Icons.flag_outlined, l10n.reportIssue, VigorColors.warning, () => _showReportDialog(context))),
        const SizedBox(width: VigorSpacing.sm),
        Expanded(child: _buildOutlineButton(Icons.delete, isOwner ? l10n.deleteTraining : l10n.leaveTraining, VigorColors.error, () => _deleteTraining(context))),
      ],
    );
  }

  Widget _buildOutlineButton(IconData icon, String label, Color color, VoidCallback onPressed) {
    return GestureDetector(
      onTap: onPressed,
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: VigorSpacing.sm + 2),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.1),
          borderRadius: VigorRadius.radiusMd,
          border: Border.all(color: color.withValues(alpha: 0.4)),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(icon, size: 16, color: color),
            const SizedBox(width: VigorSpacing.xs),
            Flexible(child: Text(label, style: VigorTypography.caption.copyWith(color: color, fontWeight: FontWeight.w600), overflow: TextOverflow.ellipsis)),
          ],
        ),
      ),
    );
  }

  Widget _buildRoutinesHeader(AppLocalizations l10n) {
    return Row(
      children: [
        RepaintBoundary(
          child: ShaderMask(
            shaderCallback: (bounds) => const LinearGradient(colors: [VigorColors.orange, VigorColors.electricBlue]).createShader(bounds),
            child: const Icon(Icons.list_alt, color: Colors.white, size: 24),
          ),
        ),
        const SizedBox(width: VigorSpacing.sm),
        Text(l10n.trainingRoutines, style: VigorTypography.headline.copyWith(fontSize: 20, color: VigorColors.textPrimary(context))),
      ],
    );
  }

  Widget _buildRoutineCard(Routine routine, bool isDark) {
    return Container(
      decoration: BoxDecoration(
        color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
        borderRadius: VigorRadius.radiusLg,
        border: Border.all(color: isDark ? Colors.white.withValues(alpha: 0.1) : Colors.black.withValues(alpha: 0.05)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // routine header
          Container(
            padding: VigorSpacing.paddingMd,
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [VigorColors.orange.withValues(alpha: 0.1), Colors.transparent],
                begin: Alignment.centerLeft,
                end: Alignment.centerRight,
              ),
              borderRadius: const BorderRadius.only(topLeft: Radius.circular(VigorRadius.lg), topRight: Radius.circular(VigorRadius.lg)),
            ),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.xs),
                  decoration: const BoxDecoration(gradient: LinearGradient(colors: [VigorColors.orange, VigorColors.electricBlue]), borderRadius: VigorRadius.radiusFull),
                  child: Text(routine.type.toUpperCase(), style: VigorTypography.label.copyWith(color: Colors.white, fontWeight: FontWeight.bold)),
                ),
                const Spacer(),
                if (routine.rest > 0)
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
                    decoration: BoxDecoration(color: VigorColors.warning.withValues(alpha: 0.15), borderRadius: VigorRadius.radiusFull, border: Border.all(color: VigorColors.warning.withValues(alpha: 0.3))),
                    child: Row(
                      children: [
                        const Icon(Icons.timer, size: 12, color: VigorColors.warning),
                        const SizedBox(width: VigorSpacing.xs),
                        Text('${routine.rest}s', style: VigorTypography.caption.copyWith(color: VigorColors.warning, fontWeight: FontWeight.w600)),
                      ],
                    ),
                  ),
              ],
            ),
          ),
          // blocks
          Padding(
            padding: VigorSpacing.paddingMd,
            child: Column(
              children: routine.blocks.asMap().entries.map((entry) => Padding(
                padding: EdgeInsets.only(bottom: entry.key < routine.blocks.length - 1 ? VigorSpacing.md : 0),
                child: _buildBlockCard(entry.value, entry.key + 1, routine.blocks.length, isDark),
              )).toList(),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildBlockCard(Block block, int blockNumber, int totalBlocks, bool isDark) {
    final showBlockLabel = totalBlocks > 1;

    return Container(
      padding: VigorSpacing.paddingMd,
      decoration: BoxDecoration(
        color: isDark ? Colors.white.withValues(alpha: 0.03) : Colors.grey.shade50,
        borderRadius: VigorRadius.radiusMd,
        border: Border.all(color: isDark ? Colors.white.withValues(alpha: 0.08) : Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (showBlockLabel || block.repeats > 1 || block.rest > 0) ...[
            Row(
              children: [
                if (showBlockLabel)
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
                    decoration: BoxDecoration(color: VigorColors.orange.withValues(alpha: 0.15), borderRadius: VigorRadius.radiusSm),
                    child: Text('Block $blockNumber', style: VigorTypography.caption.copyWith(fontWeight: FontWeight.bold, color: VigorColors.orange)),
                  ),
                const Spacer(),
                if (block.repeats > 1)
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
                    decoration: BoxDecoration(color: VigorColors.electricBlue.withValues(alpha: 0.15), borderRadius: VigorRadius.radiusSm),
                    child: Row(
                      children: [
                        const Icon(Icons.repeat, size: 12, color: VigorColors.electricBlue),
                        const SizedBox(width: VigorSpacing.xs),
                        Text('${block.repeats}x', style: VigorTypography.caption.copyWith(fontWeight: FontWeight.bold, color: VigorColors.electricBlue)),
                      ],
                    ),
                  ),
                if (block.rest > 0) ...[
                  const SizedBox(width: VigorSpacing.sm),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
                    decoration: BoxDecoration(color: VigorColors.warning.withValues(alpha: 0.15), borderRadius: VigorRadius.radiusSm),
                    child: Text('${block.rest}s', style: VigorTypography.caption.copyWith(fontWeight: FontWeight.bold, color: VigorColors.warning)),
                  ),
                ],
              ],
            ),
            const SizedBox(height: VigorSpacing.md),
          ],
          ...block.activities.asMap().entries.map((entry) => Padding(
            padding: EdgeInsets.only(bottom: entry.key < block.activities.length - 1 ? VigorSpacing.md : 0),
            child: _buildActivityRow(entry.value, isDark),
          )),
        ],
      ),
    );
  }

  Widget _buildActivityRow(Activity activity, bool isDark) {
    final exercise = _parseExercise(activity.detail);
    final hasValidImage = exercise != null && CachedExerciseImage.isValidUrl(exercise.reference);

    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // thumbnail with cached image
        GestureDetector(
          onTap: exercise != null ? () => ExerciseModal.show(context, exercise) : null,
          child: Container(
            width: 72,
            height: 72,
            decoration: BoxDecoration(
              borderRadius: VigorRadius.radiusMd,
              gradient: !hasValidImage
                  ? LinearGradient(colors: [VigorColors.orange.withValues(alpha: 0.2), VigorColors.electricBlue.withValues(alpha: 0.2)])
                  : null,
              border: Border.all(color: isDark ? Colors.white.withValues(alpha: 0.15) : Colors.grey.shade300),
            ),
            child: hasValidImage
                ? CachedExerciseImage(
                    imageUrl: exercise.reference,
                    width: 72,
                    height: 72,
                    borderRadius: VigorRadius.radiusMd,
                  )
                : _buildPlaceholderIcon(),
          ),
        ),
        const SizedBox(width: VigorSpacing.md),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(activity.name, style: VigorTypography.body.copyWith(fontWeight: FontWeight.w600, color: VigorColors.textPrimary(context))),
              if (exercise != null && (exercise.equipment.isNotEmpty || exercise.muscles.isNotEmpty || activity.modifiers.isNotEmpty)) ...[
                const SizedBox(height: VigorSpacing.xs),
                _buildExerciseDetails(exercise, activity.modifiers),
              ],
              const SizedBox(height: VigorSpacing.sm),
              _buildActivityTags(activity),
            ],
          ),
        ),
        if (training.completedAt == null)
          GestureDetector(
            onTap: () => _shuffleActivity(activity),
            child: Container(
              padding: VigorSpacing.paddingSm,
              decoration: BoxDecoration(color: VigorColors.orange.withValues(alpha: 0.15), borderRadius: VigorRadius.radiusFull),
              child: const Icon(Icons.refresh, size: 18, color: VigorColors.orange),
            ),
          ),
      ],
    );
  }

  Widget _buildPlaceholderIcon() {
    return Center(
      child: RepaintBoundary(
        child: ShaderMask(
          shaderCallback: (bounds) => const LinearGradient(colors: [VigorColors.orange, VigorColors.electricBlue]).createShader(bounds),
          child: const Icon(Icons.fitness_center, size: 28, color: Colors.white),
        ),
      ),
    );
  }

  Widget _buildExerciseDetails(Exercise exercise, List<String> modifiers) {
    return Wrap(
      spacing: VigorSpacing.sm,
      runSpacing: VigorSpacing.xs,
      children: [
        if (exercise.muscles.isNotEmpty) _buildDetailChip(Icons.accessibility_new, exercise.muscles.take(2).join(' · '), Colors.red.shade600),
        if (exercise.equipment.isNotEmpty) _buildDetailChip(Icons.fitness_center, exercise.equipment.join(' · '), VigorColors.electricBlue),
        if (modifiers.isNotEmpty) _buildDetailChip(Icons.tune, modifiers.join(' · '), VigorColors.success),
      ],
    );
  }

  Widget _buildDetailChip(IconData icon, String text, Color color) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 12, color: color),
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
        if (activity.reps > 0) _buildTag(Icons.repeat, '${activity.reps}', Colors.purple),
        if (activity.weightKg > 0) _buildTag(Icons.fitness_center, '${activity.weightKg}kg', Colors.red.shade700),
        if (activity.duration > 0) _buildTag(Icons.timer, _formatTime(activity.duration), VigorColors.electricBlue),
        if (activity.rest > 0) _buildTag(Icons.hourglass_bottom, '${activity.rest}s', VigorColors.warning),
      ],
    );
  }

  Widget _buildTag(IconData icon, String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: VigorRadius.radiusSm,
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 11, color: color),
          const SizedBox(width: VigorSpacing.xs),
          Text(label, style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: color)),
        ],
      ),
    );
  }
}
