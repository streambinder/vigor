import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../design/tokens.dart';
import '../models/exercise.dart';
import '../models/flow_session.dart';
import '../models/flow_pose.dart';
import '../utils/exercise_modal.dart';
import '../widgets/cached_exercise_image.dart';
import '../services/service_locator.dart';
import '../timer/flow_timer_notifier.dart';
import '../utils/knowledge_labels.dart';
import '../generated/app_localizations.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/marquee_text.dart';
import '../widgets/timer/inline_timer_section.dart';
import 'main_navigation.dart';

class FlowDetailsScreen extends StatefulWidget {
  final FlowSession flowSession;

  const FlowDetailsScreen({super.key, required this.flowSession});

  @override
  State<FlowDetailsScreen> createState() => _FlowDetailsScreenState();
}

class _FlowDetailsScreenState extends State<FlowDetailsScreen> {
  late FlowSession _session;
  FlowTimerNotifier? _timerNotifier;

  final ScrollController _scrollController = ScrollController();
  final GlobalKey _timerKey = GlobalKey();
  double _timerHeight = 280;

  @override
  void initState() {
    super.initState();
    _session = widget.flowSession;
  }

  @override
  void dispose() {
    _scrollController.dispose();
    _timerNotifier?.removeListener(_onTimerUpdate);
    _timerNotifier?.dispose();
    super.dispose();
  }

  void _measureTimer() {
    final height = _timerKey.currentContext?.size?.height;
    if (height != null && height != _timerHeight) setState(() => _timerHeight = height);
  }

  void _startTimer() {
    if (_timerNotifier != null) return;
    setState(() {
      _timerNotifier = FlowTimerNotifier(poses: _session.poses);
      _timerNotifier!.initialize();
      _timerNotifier!.addListener(_onTimerUpdate);
    });
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
    if (_timerNotifier?.workoutCompleted == true) setState(() {});
  }

  Future<void> _completeFlow() async {
    if (_timerNotifier != null && _timerNotifier!.isSubmitting) return;
    _timerNotifier?.isSubmitting = true;
    _stopTimer();

    final response = await context.read<ServiceLocator>().flowService.completeFlow(_session.id);
    if (!mounted) return;

    if (response.isSuccess) {
      AdaptiveNotification.show(context: context, message: 'Flow session completed');
      _navigateBack();
    } else {
      AdaptiveNotification.showError(context: context, message: 'Failed to complete flow', rawError: response.error);
    }
  }

  Future<void> _deleteFlow() async {
    final l10n = AppLocalizations.of(context);
    final shouldDelete = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: 'Delete Flow',
      content: 'Delete "${_session.name}"? This cannot be undone.',
      actions: [
        AdaptiveDialogAction(label: l10n.cancel, onPressed: () => Navigator.of(context).pop(false)),
        AdaptiveDialogAction(label: l10n.delete, isDestructive: true, onPressed: () => Navigator.of(context).pop(true)),
      ],
    );

    if (shouldDelete != true || !mounted) return;

    final response = await context.read<ServiceLocator>().flowService.deleteFlow(_session.id);
    if (!mounted) return;
    if (response.isSuccess) {
      _navigateBack();
      AdaptiveNotification.show(context: context, message: 'Flow session deleted');
    } else {
      AdaptiveNotification.showError(context: context, message: 'Failed to delete flow', rawError: response.error);
    }
  }

  void _confirmStopTimer() {
    final l10n = AppLocalizations.of(context);
    showDialog(
      context: context,
      barrierDismissible: true,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Stop Flow?'),
        content: const Text('Timer progress will be lost.'),
        actions: [
          TextButton(onPressed: () => Navigator.of(dialogContext).pop(), child: Text(l10n.cancel)),
          TextButton(
            onPressed: () { Navigator.of(dialogContext).pop(); _stopTimer(); },
            child: Text(l10n.stop),
          ),
        ],
      ),
    );
  }

  void _confirmTimerExit() {
    final l10n = AppLocalizations.of(context);
    showDialog(
      context: context,
      barrierDismissible: true,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Stop Flow?'),
        content: const Text('Timer progress will be lost.'),
        actions: [
          TextButton(
            onPressed: () { Navigator.of(dialogContext).pop(); _stopTimer(); _navigateBack(); },
            child: Text(l10n.stop),
          ),
          TextButton(
            onPressed: () { Navigator.of(dialogContext).pop(); _completeFlow(); },
            child: Text(l10n.complete),
          ),
        ],
      ),
    );
  }

  void _navigateBack() {
    Navigator.of(context).popUntil((route) => route.isFirst);
    MainNavigation.navigateToTab(1);
  }

  String _formatDuration(int seconds) {
    final minutes = seconds ~/ 60;
    if (minutes < 60) return '${minutes}m';
    final hours = minutes ~/ 60;
    final rem = minutes % 60;
    return rem == 0 ? '${hours}h' : '${hours}h ${rem}m';
  }

  String _formatDate(DateTime date) {
    final l10n = AppLocalizations.of(context);
    final diff = DateTime.now().difference(date).inDays;
    if (diff == 0) return l10n.today;
    if (diff == 1) return l10n.yesterday;
    if (diff < 7) return l10n.daysAgo(diff);
    return '${date.day}/${date.month}';
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final timerActive = _timerNotifier != null;

    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, result) {
        if (didPop) return;
        if (_timerNotifier != null) { _confirmTimerExit(); } else { _navigateBack(); }
      },
      child: AdaptiveScaffold(
        appBar: AdaptiveAppBar(
          title: MarqueeText(text: _session.name),
          automaticallyImplyLeading: false,
          leading: IconButton(
            icon: const Icon(Icons.arrow_back_ios, color: VigorColors.stone),
            onPressed: () => _timerNotifier != null ? _confirmTimerExit() : _navigateBack(),
          ),
          actions: [
            if (_session.completedAt == null)
              PopupMenuButton<String>(
                icon: const Icon(Icons.more_vert, color: VigorColors.stone),
                onSelected: (value) { if (value == 'delete') _deleteFlow(); },
                itemBuilder: (context) => [
                  PopupMenuItem(
                    value: 'delete',
                    child: Row(
                      children: [
                        const Icon(Icons.delete, size: 20, color: VigorColors.error),
                        const SizedBox(width: VigorSpacing.sm),
                        Text(l10n.delete, style: VigorTypography.body.copyWith(color: VigorColors.error)),
                      ],
                    ),
                  ),
                ],
              ),
          ],
        ),
        body: CustomScrollView(
          controller: _scrollController,
          slivers: [
            SliverPadding(
              padding: const EdgeInsets.only(left: VigorSpacing.lg, right: VigorSpacing.lg, top: VigorSpacing.lg),
              sliver: SliverList.list(children: [
                Padding(
                  padding: const EdgeInsets.only(bottom: VigorSpacing.xl),
                  child: _buildHeaderWithActions(l10n, isDark),
                ),
              ]),
            ),
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
                      onDone: _completeFlow,
                      fallbackIcon: Icons.self_improvement,
                    ),
                  ),
                ),
              ),
            SliverPadding(
              padding: const EdgeInsets.only(left: VigorSpacing.lg, right: VigorSpacing.lg, bottom: VigorSpacing.lg),
              sliver: SliverList.list(children: [
                Padding(
                  padding: const EdgeInsets.only(bottom: VigorSpacing.md),
                  child: _buildPosesHeader(l10n),
                ),
                _buildPosesCard(isDark),
                const SizedBox(height: VigorSpacing.lg),
              ]),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildHeaderWithActions(AppLocalizations l10n, bool isDark) {
    final indigoColor = VigorColors.indigoAdaptive(context);
    final wellnessColor = VigorColors.byakurokuAdaptive(context);
    final isCompleted = _session.completedAt != null;

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
                if (_session.description.isNotEmpty) ...[
                  const SizedBox(height: VigorSpacing.md),
                  Text(_session.description, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context), height: 1.5)),
                ],
                if (_session.request.isNotEmpty) ...[
                  const SizedBox(height: VigorSpacing.md),
                  _buildInlineRequest(l10n),
                ],
                if (_session.muscles.isNotEmpty) ...[
                  const SizedBox(height: VigorSpacing.md),
                  Wrap(
                    spacing: VigorSpacing.xs,
                    runSpacing: VigorSpacing.xs,
                    children: _session.muscles.map((m) => _buildChip(Icons.accessibility_new, KnowledgeLabels.muscleLabel(m, l10n))).toList(),
                  ),
                ],
              ],
            ),
          ),
          Row(
            children: [
              Expanded(child: _timerNotifier != null
                ? _buildActionButton(icon: Icons.stop, label: l10n.stop, color: indigoColor, onPressed: _confirmStopTimer)
                : _buildActionButton(icon: Icons.timer, label: l10n.timer, color: indigoColor, onPressed: isCompleted ? null : _startTimer),
              ),
              if (_timerNotifier != null)
                Expanded(child: _buildActionButton(
                  icon: Icons.check_circle_outline, label: l10n.complete,
                  color: wellnessColor, onPressed: _completeFlow,
                ))
              else if (!isCompleted)
                Expanded(child: _buildActionButton(
                  icon: Icons.check_circle_outline, label: l10n.complete,
                  color: wellnessColor, onPressed: _completeFlow,
                )),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildMetadataChips(AppLocalizations l10n) {
    return Wrap(
      spacing: VigorSpacing.sm,
      runSpacing: VigorSpacing.sm,
      children: [
        _buildMethodologyBadge('FLOW'),
        _buildMetaChip(Icons.schedule, _formatDuration(_session.duration)),
        _buildMetaChip(Icons.calendar_today, _formatDate(_session.completedAt ?? _session.createdAt)),
        if (_session.completedAt != null) _buildMetaChip(Icons.check_circle_outline, l10n.completed),
      ],
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
            child: Text(_session.request, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context), fontStyle: FontStyle.italic)),
          ),
        ],
      ),
    );
  }

  Widget _buildPosesHeader(AppLocalizations l10n) {
    return Row(
      children: [
        const Icon(Icons.self_improvement, color: VigorColors.stone, size: 24),
        const SizedBox(width: VigorSpacing.sm),
        Text('Poses', style: VigorTypography.headline.copyWith(fontSize: 22, color: VigorColors.textPrimary(context))),
      ],
    );
  }

  // single card containing all poses, styled like a training routine card with one block
  Widget _buildPosesCard(bool isDark) {
    return Container(
      decoration: BoxDecoration(
        color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
        borderRadius: VigorRadius.radiusLg,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // routine-style header bar
          Container(
            width: double.infinity,
            padding: VigorSpacing.paddingMd,
            decoration: BoxDecoration(
              color: VigorColors.stone.withValues(alpha: 0.05),
              borderRadius: const BorderRadius.only(topLeft: Radius.circular(VigorRadius.lg), topRight: Radius.circular(VigorRadius.lg)),
            ),
            child: Center(
              child: Text('FLOW', style: VigorTypography.headline.copyWith(color: VigorColors.stone, fontSize: 14)),
            ),
          ),
          // poses as activity rows
          Padding(
            padding: VigorSpacing.paddingMd,
            child: Column(
              children: _session.poses.asMap().entries.map((entry) => Padding(
                padding: EdgeInsets.only(bottom: entry.key < _session.poses.length - 1 ? VigorSpacing.md : 0),
                child: _buildPoseRow(entry.value),
              )).toList(),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPoseRow(FlowPose pose) {
    final imageUrl = pose.detail['reference'] as String?;
    Exercise? exercise;
    try {
      if (pose.detail.isNotEmpty) exercise = Exercise.fromJson(pose.detail);
    } catch (_) {}

    return GestureDetector(
      onTap: exercise != null ? () => ExerciseModal.show(context, exercise!) : null,
      behavior: HitTestBehavior.opaque,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          CachedExerciseImage(
            imageUrl: imageUrl,
            width: 72,
            height: 72,
            borderRadius: VigorRadius.radiusMd,
            placeholderIcon: Icons.self_improvement,
          ),
          const SizedBox(width: VigorSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(pose.name, style: VigorTypography.body.copyWith(fontWeight: FontWeight.w600, color: VigorColors.textPrimary(context))),
                const SizedBox(height: VigorSpacing.sm),
                Wrap(
                  spacing: VigorSpacing.xs,
                  runSpacing: VigorSpacing.xs,
                  children: [
                    _buildTag(Icons.timer, '${pose.duration}s'),
                    if (pose.rest > 0) _buildTag(Icons.hourglass_bottom, '${pose.rest}s rest'),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
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

  Widget _buildMethodologyBadge(String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: VigorColors.stone.withValues(alpha: 0.1),
        borderRadius: VigorRadius.radiusXs,
      ),
      child: Text(label, style: VigorTypography.caption.copyWith(color: VigorColors.stone, fontWeight: FontWeight.w600, fontSize: 10)),
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

  @override double get maxExtent => height;
  @override double get minExtent => height;

  @override
  Widget build(BuildContext context, double shrinkOffset, bool overlapsContent) {
    final shadowOpacity = (shrinkOffset / 40).clamp(0.0, 1.0);
    return Stack(
      clipBehavior: Clip.none,
      children: [
        if (shadowOpacity > 0)
          Positioned.fill(
            top: VigorSpacing.lg,
            child: DecoratedBox(
              decoration: BoxDecoration(
                boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.15 * shadowOpacity), blurRadius: 24 * shadowOpacity, offset: Offset(0, 2 * shadowOpacity))],
              ),
            ),
          ),
        ClipRect(
          child: OverflowBox(minHeight: 0, maxHeight: double.infinity, alignment: Alignment.topLeft, child: child),
        ),
      ],
    );
  }

  @override
  bool shouldRebuild(_SliverTimerDelegate oldDelegate) => height != oldDelegate.height;
}
