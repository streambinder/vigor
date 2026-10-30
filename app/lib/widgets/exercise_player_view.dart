import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:webview_flutter/webview_flutter.dart';

import '../models/exercise.dart';
import '../services/exercise_player_service.dart';

/// WebView viewport rendering the exercise animation in Three.js.
///
/// Loads the bundled player page, then hands it the animation document
/// and the selected model bytes as base64 chunks: one large script
/// string can exceed the platform channel limit, so the transfer is
/// chunked. Any failure shows the fallback widget, so the modal
/// degrades to the GIF instead of an error screen.
class ExercisePlayerView extends StatefulWidget {
  final Exercise exercise;
  final ExercisePlayerService playerService;
  final Widget fallback;
  final double height;

  const ExercisePlayerView({
    super.key,
    required this.exercise,
    required this.playerService,
    required this.fallback,
    this.height = 300,
  });

  @override
  State<ExercisePlayerView> createState() => _ExercisePlayerViewState();
}

class _ExercisePlayerViewState extends State<ExercisePlayerView> {
  static const int _chunkSize = 128 * 1024;

  late final WebViewController _controller;
  final Completer<void> _pageReady = Completer<void>();
  bool _ready = false;
  bool _failed = false;

  @override
  void initState() {
    super.initState();
    _controller = WebViewController()
      ..setJavaScriptMode(JavaScriptMode.unrestricted)
      ..setBackgroundColor(Colors.transparent)
      ..addJavaScriptChannel(
        'PlayerReady',
        onMessageReceived: (_) {
          if (mounted) setState(() => _ready = true);
        },
      )
      ..addJavaScriptChannel(
        'PlayerError',
        onMessageReceived: (_) {
          if (mounted) setState(() => _failed = true);
        },
      )
      ..setNavigationDelegate(
        NavigationDelegate(
          onPageFinished: (_) {
            if (!_pageReady.isCompleted) _pageReady.complete();
          },
          onWebResourceError: (_) {
            if (mounted) setState(() => _failed = true);
          },
        ),
      )
      ..loadFlutterAsset('assets/player/player.html');
    _start();
  }

  Future<void> _start() async {
    try {
      await _pageReady.future;
      final animation =
          await widget.playerService.animation(widget.exercise.id);
      final modelId = await widget.playerService.selectedModelId();
      final bytes =
          modelId == null ? null : await widget.playerService.modelBytes(modelId);
      if (!mounted) return;
      if (animation == null || bytes == null) {
        setState(() => _failed = true);
        return;
      }
      await _transfer(animation, bytes);
    } catch (_) {
      if (mounted) setState(() => _failed = true);
    }
  }

  Future<void> _transfer(String animation, Uint8List model) async {
    final animB64 = base64Encode(utf8.encode(animation));
    final modelB64 = base64Encode(model);
    for (var i = 0; i < animB64.length; i += _chunkSize) {
      final end = (i + _chunkSize < animB64.length)
          ? i + _chunkSize
          : animB64.length;
      await _controller.runJavaScript(
        "window.appendAnimationChunk('${animB64.substring(i, end)}')",
      );
    }
    for (var i = 0; i < modelB64.length; i += _chunkSize) {
      final end = (i + _chunkSize < modelB64.length)
          ? i + _chunkSize
          : modelB64.length;
      await _controller.runJavaScript(
        "window.appendModelChunk('${modelB64.substring(i, end)}')",
      );
    }
    await _controller.runJavaScript('window.startPlayer()');
  }

  @override
  Widget build(BuildContext context) {
    if (_failed) {
      return widget.fallback;
    }
    return SizedBox(
      height: widget.height,
      child: Stack(
        children: [
          WebViewWidget(controller: _controller),
          if (!_ready)
            const Center(child: CircularProgressIndicator()),
        ],
      ),
    );
  }
}
