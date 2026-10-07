import 'package:flutter/material.dart';

/// Text that tweens between numeric values with an ease-in-out curve,
/// counting from the previously shown value whenever the input changes.
///
/// The caller seeds the first frame with [seedValue] — the last value shown
/// in an earlier session — so a reopen tweens seed → current only when the
/// value actually changed, and never counts up from zero. With no seed the
/// first value is shown as-is.
class AnimatedNumberText extends StatefulWidget {
  const AnimatedNumberText({
    super.key,
    required this.value,
    required this.formatValue,
    this.seedValue,
    this.placeholder = '—',
    this.style,
    this.duration = const Duration(milliseconds: 900),
  });

  /// Current value; null renders [placeholder] statically.
  final double? value;

  /// Last value shown in a previous session; the first tween starts here.
  final double? seedValue;

  /// Formats the interpolated value for display.
  final String Function(double value) formatValue;

  /// Shown while [value] is null.
  final String placeholder;

  /// Style applied to the rendered text.
  final TextStyle? style;

  /// Duration of a single value-to-value tween.
  final Duration duration;

  @override
  State<AnimatedNumberText> createState() => _AnimatedNumberTextState();
}

class _AnimatedNumberTextState extends State<AnimatedNumberText>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late Animation<double> _animation;

  /// last numeric value handed to this widget for display: the next tween
  /// starts from here, so equal consecutive values never animate
  double? _displayed;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(vsync: this, duration: widget.duration);
    _animation = const AlwaysStoppedAnimation(0);
    _displayed = widget.seedValue;
    final value = widget.value;
    if (value != null) _showValue(value);
  }

  @override
  void didUpdateWidget(AnimatedNumberText oldWidget) {
    super.didUpdateWidget(oldWidget);
    final value = widget.value;
    if (value == null) {
      _controller.stop();
      return;
    }
    if (value != oldWidget.value) _showValue(value);
  }

  /// renders [value]: tweens from the previously shown value when there is
  /// one and it differs, snaps directly when there is none or it is unchanged
  void _showValue(double value) {
    final from = _displayed;
    _displayed = value;
    if (from == null || from == value) {
      _controller.stop();
      _animation = AlwaysStoppedAnimation(value);
      return;
    }
    _animation = Tween<double>(begin: from, end: value).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );
    _controller.forward(from: 0);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final value = widget.value;
    if (value == null) return Text(widget.placeholder, style: widget.style);
    return AnimatedBuilder(
      animation: _animation,
      builder: (context, _) => Text(
        widget.formatValue(_animation.value),
        style: widget.style,
      ),
    );
  }
}
