import 'dart:async';
import 'package:audioplayers/audioplayers.dart';

/// Audio service for playing interval whistle sounds
/// Supports audio ducking to lower other app audio during playback
class AudioService {
  static final AudioService _instance = AudioService._internal();
  factory AudioService() => _instance;
  AudioService._internal();

  final AudioPlayer _player = AudioPlayer();
  final _source = AssetSource('audio/whistle.mp3');
  bool _initialized = false;
  StreamSubscription? _completeSubscription;

  Future<void> initialize() async {
    if (_initialized) return;
    await _player.setSource(_source);
    await _player.setVolume(1.0);
    // re-prepare source after playback completes so next play works reliably —
    // without this, seek+resume silently fails on completed players
    _completeSubscription = _player.onPlayerComplete.listen((_) {
      _player.setSource(_source);
    });
    _initialized = true;
  }

  /// configure audio ducking — lowers other app audio during whistle playback
  Future<void> setDuckOtherAudio(bool enabled) async {
    if (enabled) {
      await _player.setAudioContext(AudioContext(
        android: const AudioContextAndroid(
          audioFocus: AndroidAudioFocus.gainTransientMayDuck,
        ),
        iOS: AudioContextIOS(
          category: AVAudioSessionCategory.playback,
          options: const {
            AVAudioSessionOptions.duckOthers,
            AVAudioSessionOptions.mixWithOthers,
          },
        ),
      ));
    } else {
      await _player.setAudioContext(AudioContext(
        android: const AudioContextAndroid(
          audioFocus: AndroidAudioFocus.none,
        ),
        iOS: AudioContextIOS(
          category: AVAudioSessionCategory.playback,
          options: const {AVAudioSessionOptions.mixWithOthers},
        ),
      ));
    }
  }

  Future<void> playWhistle() async {
    if (!_initialized) await initialize();
    // stop + play from scratch — seek+resume is unreliable after the player
    // enters completed state or after iOS audio session interruptions
    await _player.stop();
    await _player.play(_source);
  }

  Future<void> stopWhistle() async {
    await _player.stop();
  }

  /// re-prepare audio after app returns from background — iOS deactivates the
  /// audio session on backgrounding, so the player needs a fresh source
  Future<void> reactivate() async {
    if (!_initialized) return;
    await _player.setSource(_source);
  }

  void dispose() {
    _completeSubscription?.cancel();
    _player.dispose();
  }
}
