import 'package:flutter/material.dart';

/// Text that tweens between numeric values with an ease-in-out curve,
/// counting from the previously shown value whenever the input changes.
class AnimatedNumberText extends StatefulWidget {
  const AnimatedNumberText({
    super.key,
    required this.value,
    required this.formatValue,
    this.placeholder = '—',
    this.style,
    this.duration = const Duration(milliseconds: 900),
  });

  /// Current value; null renders [placeholder] statically.
  final double? value;

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

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(vsync: this, duration: widget.duration);
    _animation = const AlwaysStoppedAnimation(0);
    final value = widget.value;
    if (value != null) _tweenTo(value, from: 0);
  }

  @override
  void didUpdateWidget(AnimatedNumberText oldWidget) {
    super.didUpdateWidget(oldWidget);
    final value = widget.value;
    if (value == null) {
      _controller.stop();
      return;
    }
    if (value != oldWidget.value) {
      _tweenTo(value, from: oldWidget.value == null ? 0 : _animation.value);
    }
  }

  void _tweenTo(double target, {required double from}) {
    if (from == target) {
      _animation = AlwaysStoppedAnimation(target);
      return;
    }
    _animation = Tween<double>(begin: from, end: target).animate(
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
