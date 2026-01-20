import 'package:audioplayers/audioplayers.dart';

/// Simple audio service for playing interval completion jingles
class AudioService {
  static final AudioService _instance = AudioService._internal();
  factory AudioService() => _instance;
  AudioService._internal();

  final AudioPlayer _player = AudioPlayer();
  bool _initialized = false;

  Future<void> initialize() async {
    if (_initialized) return;
    await _player.setSource(AssetSource('audio/jingle.mp3'));
    await _player.setVolume(0.7);
    _initialized = true;
  }

  Future<void> playJingle() async {
    if (!_initialized) await initialize();
    await _player.stop();
    await _player.resume();
  }

  void dispose() {
    _player.dispose();
  }
}
