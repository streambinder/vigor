import 'package:flutter/material.dart';
import '../config/api_config.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/exercise.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';
import '../widgets/adaptive/adaptive.dart';

enum _ExerciseTab { cues, instructions }

class ExerciseModal {
  static void show(BuildContext context, Exercise exercise) {
    showDialog(
      context: context,
      barrierDismissible: true,
      builder: (context) => Dialog(
        backgroundColor: Colors.transparent,
        child: _ExerciseModalContent(exercise: exercise),
      ),
    );
  }
}

class _ExerciseModalContent extends StatefulWidget {
  final Exercise exercise;
  const _ExerciseModalContent({required this.exercise});

  @override
  State<_ExerciseModalContent> createState() => _ExerciseModalContentState();
}

class _ExerciseModalContentState extends State<_ExerciseModalContent> {
  late _ExerciseTab _selectedTab;

  @override
  void initState() {
    super.initState();
    // default to cues when available, otherwise instructions
    _selectedTab = widget.exercise.cues.isNotEmpty
        ? _ExerciseTab.cues
        : _ExerciseTab.instructions;
  }

  static bool _isValidImageUrl(String? url) {
    if (url == null || url.isEmpty) return false;
    final uri = Uri.tryParse(url);
    if (uri == null) return false;
    return uri.hasScheme && (uri.scheme == 'http' || uri.scheme == 'https');
  }

  // proxy external image URLs through backend to avoid CORS issues on web
  static String _proxyImageUrl(String url) {
    return '${ApiConfig.baseUrl}/proxy/image?url=${Uri.encodeComponent(url)}';
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final exercise = widget.exercise;
    final hasCues = exercise.cues.isNotEmpty;
    final hasInstructions = exercise.instructions.isNotEmpty;
    final showTabs = hasCues && hasInstructions;

    return Container(
      constraints: const BoxConstraints(maxWidth: 600),
      decoration: PlatformHelper.useLiquidGlass
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
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              if (_isValidImageUrl(exercise.reference))
                Image.network(
                  _proxyImageUrl(exercise.reference),
                  fit: BoxFit.contain,
                  // frameBuilder works reliably on all platforms (unlike
                  // loadingBuilder which may not fire on android for proxied GIFs)
                  frameBuilder: (context, child, frame, wasSynchronouslyLoaded) {
                    if (wasSynchronouslyLoaded || frame != null) return child;
                    return Container(
                      height: 200,
                      color: VigorColors.stone.withValues(alpha: 0.1),
                      child: Center(
                        child: Icon(
                          Icons.fitness_center,
                          size: 200 * 0.4,
                          color: VigorColors.stone.withValues(alpha: 0.5),
                        ),
                      ),
                    );
                  },
                  errorBuilder: (context, error, stackTrace) {
                    return Container(
                      height: 200,
                      color: VigorColors.stone.withValues(alpha: 0.1),
                      child: Center(
                        child: Icon(
                          Icons.fitness_center,
                          size: 200 * 0.4,
                          color: VigorColors.stone.withValues(alpha: 0.5),
                        ),
                      ),
                    );
                  },
                ),
              Padding(
                padding: const EdgeInsets.all(24.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      exercise.name,
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 22)
                          : Theme.of(context).textTheme.headlineSmall,
                    ),
                    if (hasCues || hasInstructions) ...[
                      const SizedBox(height: 16),
                      if (showTabs) ...[
                        SizedBox(
                          width: double.infinity,
                          child: SegmentedButton<_ExerciseTab>(
                            segments: [
                              ButtonSegment(
                                value: _ExerciseTab.cues,
                                label: Text(
                                  l10n.cues,
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ),
                              ButtonSegment(
                                value: _ExerciseTab.instructions,
                                label: Text(
                                  l10n.instructions,
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ),
                            ],
                            selected: {_selectedTab},
                            onSelectionChanged: (selected) {
                              setState(() => _selectedTab = selected.first);
                            },
                            showSelectedIcon: false,
                          ),
                        ),
                        const SizedBox(height: 16),
                      ],
                      if (!showTabs && hasInstructions)
                        _buildSectionHeader(context, l10n.instructions),
                      if (!showTabs && hasCues)
                        _buildSectionHeader(context, l10n.cues),
                      if (_selectedTab == _ExerciseTab.cues && hasCues)
                        _buildCuesList(exercise.cues, context),
                      if (_selectedTab == _ExerciseTab.instructions && hasInstructions)
                        _buildInstructionsList(exercise.instructions, context),
                    ],
                    const SizedBox(height: 16),
                    Align(
                      alignment: Alignment.centerRight,
                      child: AdaptiveTextButton(
                        onPressed: () => Navigator.of(context).pop(),
                        child: Text(l10n.close),
                      ),
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

  Widget _buildSectionHeader(BuildContext context, String title) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Text(
        title,
        style: PlatformHelper.useLiquidGlass
            ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 16)
            : Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
      ),
    );
  }

  Widget _buildCuesList(List<String> cues, BuildContext context) {
    return Column(
      children: cues.map((cue) => Padding(
        padding: const EdgeInsets.only(bottom: 8.0),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SizedBox(
              width: 24,
              child: Text(
                '•',
                style: VigorTypography.label.copyWith(
                  fontWeight: FontWeight.bold,
                  color: VigorColors.indigoAdaptive(context),
                ),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                cue,
                style: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.bodyStyle
                    : Theme.of(context).textTheme.bodyMedium,
              ),
            ),
          ],
        ),
      )).toList(),
    );
  }

  Widget _buildInstructionsList(List<String> instructions, BuildContext context) {
    return Column(
      children: instructions.asMap().entries.map((entry) => Padding(
        padding: const EdgeInsets.only(bottom: 8.0),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SizedBox(
              width: 24,
              child: Text(
                '${entry.key + 1}.',
                style: VigorTypography.label.copyWith(
                  fontWeight: FontWeight.bold,
                  color: VigorColors.indigoAdaptive(context),
                ),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                entry.value,
                style: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.bodyStyle
                    : Theme.of(context).textTheme.bodyMedium,
              ),
            ),
          ],
        ),
      )).toList(),
    );
  }
}
