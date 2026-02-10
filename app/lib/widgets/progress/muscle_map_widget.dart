import 'package:flutter/material.dart';
import 'package:flutter/services.dart' show rootBundle;
import 'package:flutter_svg/flutter_svg.dart';
import '../../design/tokens.dart';
import '../../models/muscle_impact.dart';

/// A widget that displays a human body figure with muscles colored based on training heat.
/// Inspired by MuscleWiki's interactive body map.
class MuscleMapWidget extends StatefulWidget {
  /// Map of muscle ID to MuscleImpact (contains heat value 0-100)
  final Map<String, MuscleImpact> muscles;

  /// Whether to show toggle for front/back views (default: true)
  /// If false, shows both views side by side
  final bool showToggle;

  /// Height constraint for the widget
  final double? height;

  /// Callback when a muscle group is tapped
  final void Function(String muscleId)? onMuscleTap;

  /// User gender for selecting body model ('male' or 'female')
  final String? gender;

  const MuscleMapWidget({
    super.key,
    required this.muscles,
    this.showToggle = true,
    this.height,
    this.onMuscleTap,
    this.gender,
  });

  @override
  State<MuscleMapWidget> createState() => _MuscleMapWidgetState();
}

class _MuscleMapWidgetState extends State<MuscleMapWidget> {
  bool _showFront = true;
  String? _frontSvg;
  String? _backSvg;
  bool _isLoading = true;

  // mapping from Vigor muscle IDs to MuscleWiki SVG group IDs
  static const _muscleToSvgFront = {
    'chest': ['chest'],
    'back': ['traps'],
    'shoulders': ['front-shoulders'],
    'arms': ['biceps', 'forearms'],
    'core': ['abdominals', 'obliques'],
    'legs': ['quads', 'calves'],
  };

  static const _muscleToSvgBack = {
    'back': ['traps', 'lats', 'lowerback', 'traps-middle'],
    'shoulders': ['rear-shoulders'],
    'arms': ['triceps', 'forearms'],
    'glutes': ['glutes'],
    'legs': ['hamstrings', 'calves'],
  };

  // all SVG muscle group IDs (for applying base color)
  static const _allSvgMusclesFront = [
    'chest', 'traps', 'front-shoulders', 'biceps', 'forearms',
    'abdominals', 'obliques', 'quads', 'calves', 'hands'
  ];

  static const _allSvgMusclesBack = [
    'traps', 'lats', 'lowerback', 'traps-middle', 'rear-shoulders',
    'triceps', 'forearms', 'glutes', 'hamstrings', 'calves', 'hands'
  ];

  @override
  void initState() {
    super.initState();
    _loadSvgAssets();
  }

  @override
  void didUpdateWidget(MuscleMapWidget oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.gender != widget.gender) {
      _loadSvgAssets();
    }
  }

  Future<void> _loadSvgAssets() async {
    final isFemale = widget.gender?.toLowerCase() == 'female';
    try {
      String front;
      String back;
      if (isFemale) {
        // try loading female variants, fall back to default
        try {
          front = await rootBundle.loadString('assets/body_front_female.svg');
        } catch (_) {
          front = await rootBundle.loadString('assets/body_front.svg');
        }
        try {
          back = await rootBundle.loadString('assets/body_back_female.svg');
        } catch (_) {
          back = await rootBundle.loadString('assets/body_back.svg');
        }
      } else {
        front = await rootBundle.loadString('assets/body_front.svg');
        back = await rootBundle.loadString('assets/body_back.svg');
      }
      if (mounted) {
        setState(() {
          _frontSvg = front;
          _backSvg = back;
          _isLoading = false;
        });
      }
    } catch (e) {
      debugPrint('Failed to load SVG assets: $e');
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return SizedBox(
        height: widget.height ?? 350,
        child: const Center(child: CircularProgressIndicator()),
      );
    }

    if (_frontSvg == null || _backSvg == null) {
      return SizedBox(
        height: widget.height ?? 350,
        child: Center(
          child: Text(
            'Failed to load body map',
            style: VigorTypography.body.copyWith(color: VigorColors.textSecondary(context)),
          ),
        ),
      );
    }

    final isDark = Theme.of(context).brightness == Brightness.dark;

    if (widget.showToggle) {
      return Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          _buildToggle(context),
          const SizedBox(height: VigorSpacing.sm),
          Flexible(
            child: _buildBodyView(context, isDark, _showFront),
          ),
        ],
      );
    }

    // side by side view
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(child: _buildBodyView(context, isDark, true)),
        const SizedBox(width: VigorSpacing.xs),
        Expanded(child: _buildBodyView(context, isDark, false)),
      ],
    );
  }

  Widget _buildToggle(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        _buildToggleButton(context, 'Front', _showFront, () => setState(() => _showFront = true)),
        const SizedBox(width: VigorSpacing.sm),
        _buildToggleButton(context, 'Back', !_showFront, () => setState(() => _showFront = false)),
      ],
    );
  }

  Widget _buildToggleButton(BuildContext context, String label, bool isSelected, VoidCallback onTap) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        decoration: BoxDecoration(
          color: isSelected ? VigorColors.indigoAdaptive(context) : VigorColors.surface(context),
          borderRadius: VigorRadius.radiusSm,
          border: Border.all(color: VigorColors.border(context)),
        ),
        child: Text(
          label,
          style: VigorTypography.label.copyWith(
            color: isSelected ? Colors.white : VigorColors.textSecondary(context),
            fontWeight: isSelected ? FontWeight.w600 : FontWeight.w400,
          ),
        ),
      ),
    );
  }

  Widget _buildBodyView(BuildContext context, bool isDark, bool isFront) {
    final svgString = isFront ? _frontSvg! : _backSvg!;
    final muscleMapping = isFront ? _muscleToSvgFront : _muscleToSvgBack;
    final allMuscles = isFront ? _allSvgMusclesFront : _allSvgMusclesBack;

    // apply colors to the SVG by replacing placeholders
    final coloredSvg = _applyMuscleColors(svgString, muscleMapping, allMuscles, isDark);

    return SizedBox(
      height: widget.height ?? 350,
      child: SvgPicture.string(
        coloredSvg,
        fit: BoxFit.contain,
      ),
    );
  }

  String _applyMuscleColors(String svgString, Map<String, List<String>> muscleMapping, List<String> allMuscles, bool isDark) {
    var result = svgString;

    // clip head by adjusting viewBox: original is "0 0 660.46 1206.46", skip first ~195 y units
    result = result.replaceFirst(
      RegExp(r'viewBox="[^"]*"'),
      'viewBox="0 195 660.46 1011.46"',
    );

    // brighten body outline stroke in dark mode (original is #484a68)
    if (isDark) {
      result = result.replaceAll('stroke="#484a68"', 'stroke="#B8BCC4"');
    }

    // build a map of svgId -> heat value
    final svgHeatMap = <String, double>{};
    for (final entry in muscleMapping.entries) {
      final vigorMuscleId = entry.key;
      final svgGroupIds = entry.value;
      final impact = widget.muscles[vigorMuscleId];
      final heat = impact?.heat ?? 0.0;
      for (final svgId in svgGroupIds) {
        svgHeatMap[svgId] = heat;
      }
    }

    // replace placeholders for each muscle group
    for (final svgId in allMuscles) {
      final heat = svgHeatMap[svgId];
      final colorData = heat != null
          ? _heatToColorAndOpacity(heat, isDark)
          : const {'color': 'transparent', 'opacity': '0'};

      result = result.replaceAll('{{FILL_$svgId}}', colorData['color']!);
      result = result.replaceAll('{{OPACITY_$svgId}}', colorData['opacity']!);
    }

    return result;
  }

  /// 5-step gradient: indigo (cool) -> persimmon (active) -> crimson (hot)
  /// heat is 0-100 from backend
  Map<String, String> _heatToColorAndOpacity(double heat, bool isDark) {
    final coolColor = isDark ? '#5A9ABF' : '#2B4C5D';
    const persimmon = '#E65D38';
    const crimson = '#8F1D21';

    if (heat <= 10) return {'color': coolColor, 'opacity': '0.15'};
    if (heat <= 30) return {'color': coolColor, 'opacity': '0.35'};
    if (heat <= 55) return {'color': persimmon, 'opacity': '0.30'};
    if (heat <= 80) return {'color': persimmon, 'opacity': '0.55'};
    return {'color': crimson, 'opacity': '0.70'};
  }
}
