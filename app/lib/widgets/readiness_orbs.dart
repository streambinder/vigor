import 'dart:math';

import 'package:flutter/material.dart';

/// Readiness orbs painted behind the training counter circle, orbkit style:
/// soft radial-gradient orbs drift on slow deterministic orbits around the
/// circle, blending the readiness level color with hue-shifted companions.
class ReadinessOrbs extends StatefulWidget {
  final Color color;
  final double circleRadius;

  const ReadinessOrbs({super.key, required this.color, required this.circleRadius});

  @override
  State<ReadinessOrbs> createState() => _ReadinessOrbsState();
}

class _ReadinessOrbsState extends State<ReadinessOrbs>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 14),
    )..repeat();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    // fade the effect in on first paint, then cross-fade the color whenever
    // the target changes (grey placeholder -> readiness level color)
    return TweenAnimationBuilder<double>(
      tween: Tween(begin: 0.0, end: 1.0),
      duration: const Duration(milliseconds: 600),
      curve: Curves.easeOut,
      builder: (context, fade, _) => Opacity(
        opacity: fade,
        child: TweenAnimationBuilder<Color?>(
          tween: ColorTween(begin: widget.color, end: widget.color),
          duration: const Duration(milliseconds: 600),
          builder: (context, color, _) => AnimatedBuilder(
            animation: _controller,
            builder: (context, _) => CustomPaint(
              painter: ReadinessOrbsPainter(
                color: color ?? widget.color,
                circleRadius: widget.circleRadius,
                time: _controller.value,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class ReadinessOrbsPainter extends CustomPainter {
  static const int orbCount = 6;

  final Color color;
  final double circleRadius;
  final double time;

  ReadinessOrbsPainter({
    required this.color,
    required this.circleRadius,
    required this.time,
  });

  List<Color> palette() {
    final hsl = HSLColor.fromColor(color);
    return [
      color,
      Color.lerp(color, Colors.white, 0.45)!,
      hsl
          .withHue((hsl.hue + 28) % 360)
          .withLightness(max(0.0, min(1.0, hsl.lightness + 0.08)))
          .toColor(),
      hsl
          .withHue((hsl.hue - 32 + 360) % 360)
          .withLightness(max(0.0, min(1.0, hsl.lightness - 0.06)))
          .toColor(),
    ];
  }

  void paintOrb(Canvas canvas, Offset center, double radius, Color color) {
    final alpha = color.a;
    canvas.drawCircle(
      center,
      radius,
      Paint()
        ..shader = RadialGradient(
          colors: [
            color.withValues(alpha: alpha * 0.9),
            color.withValues(alpha: 0),
          ],
        ).createShader(Rect.fromCircle(center: center, radius: radius)),
    );
  }

  @override
  void paint(Canvas canvas, Size size) {
    final center = size.center(Offset.zero);
    final palette = this.palette();

    // ambient bed: one large soft orb hugging the circle edge
    paintOrb(
      canvas,
      center,
      circleRadius * 1.05,
      palette[0].withValues(alpha: 0.22),
    );

    for (var i = 0; i < orbCount; i++) {
      // deterministic orbit seeded from the orb index, like orbkit drift
      final seed = i * 2.399963;
      final turns = 0.35 + 0.12 * (i % 3);
      final angle = seed + time * 2 * pi * turns;
      final orbitRadius =
          circleRadius * (1.04 + 0.10 * sin(time * 2 * pi * 2 + seed));
      final orbCenter = Offset(
        center.dx + cos(angle) * orbitRadius,
        center.dy + sin(angle) * orbitRadius * 0.86,
      );
      final orbRadius =
          circleRadius * (0.26 + 0.06 * sin(time * 2 * pi + seed * 1.7));
      final orbColor = palette[i % palette.length];
      final alpha = 0.50 + 0.20 * sin(time * 2 * pi * 1.5 + seed);
      paintOrb(
        canvas,
        orbCenter,
        orbRadius,
        orbColor.withValues(alpha: alpha),
      );
    }
  }

  @override
  bool shouldRepaint(ReadinessOrbsPainter oldDelegate) =>
      color != oldDelegate.color ||
      circleRadius != oldDelegate.circleRadius ||
      time != oldDelegate.time;
}
