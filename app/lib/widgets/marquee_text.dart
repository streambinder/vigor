import 'package:flutter/material.dart';

/// A text widget that scrolls horizontally when content overflows.
/// Uses a marquee animation with pause at edges for readability.
class MarqueeText extends StatefulWidget {
  final String text;
  final TextStyle? style;
  final Duration pauseDuration;
  final Duration scrollDuration;

  const MarqueeText({
    super.key,
    required this.text,
    this.style,
    this.pauseDuration = const Duration(seconds: 2),
    this.scrollDuration = const Duration(seconds: 4),
  });

  @override
  State<MarqueeText> createState() => _MarqueeTextState();
}

class _MarqueeTextState extends State<MarqueeText> with SingleTickerProviderStateMixin {
  late ScrollController _scrollController;
  bool _needsScroll = false;
  bool _isScrolling = false;

  @override
  void initState() {
    super.initState();
    _scrollController = ScrollController();
    WidgetsBinding.instance.addPostFrameCallback((_) => _checkOverflow());
  }

  @override
  void didUpdateWidget(MarqueeText oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.text != widget.text) {
      _needsScroll = false;
      _isScrolling = false;
      _scrollController.jumpTo(0);
      WidgetsBinding.instance.addPostFrameCallback((_) => _checkOverflow());
    }
  }

  void _checkOverflow() {
    if (!mounted) return;
    final maxScroll = _scrollController.position.maxScrollExtent;
    if (maxScroll > 0 && !_needsScroll) {
      setState(() => _needsScroll = true);
      _startScrolling();
    }
  }

  Future<void> _startScrolling() async {
    if (!mounted || _isScrolling || !_needsScroll) return;
    _isScrolling = true;

    while (mounted && _needsScroll) {
      await Future.delayed(widget.pauseDuration);
      if (!mounted) return;

      final maxScroll = _scrollController.position.maxScrollExtent;
      if (maxScroll <= 0) break;

      // scroll to end
      await _scrollController.animateTo(
        maxScroll,
        duration: widget.scrollDuration,
        curve: Curves.linear,
      );
      if (!mounted) return;

      await Future.delayed(widget.pauseDuration);
      if (!mounted) return;

      // scroll back to start
      await _scrollController.animateTo(
        0,
        duration: widget.scrollDuration,
        curve: Curves.linear,
      );
    }
    _isScrolling = false;
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      controller: _scrollController,
      scrollDirection: Axis.horizontal,
      physics: const NeverScrollableScrollPhysics(),
      child: Text(widget.text, style: widget.style, maxLines: 1),
    );
  }
}
