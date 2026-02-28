import 'package:audioplayers/audioplayers.dart';

/// Audio service for playing interval whistle sounds
/// Supports audio ducking to lower other app audio during playback
class AudioService {
  static final AudioService _instance = AudioService._internal();
  factory AudioService() => _instance;
  AudioService._internal();

  final AudioPlayer _player = AudioPlayer();
  bool _initialized = false;

  Future<void> initialize() async {
    if (_initialized) return;
    await _player.setSource(AssetSource('audio/whistle.mp3'));
    await _player.setVolume(1.0);
    _initialized = true;
  }

  /// configure audio ducking — lowers other app audio during whistle playback
  Future<void> setDuckOtherAudio(bool enabled) async {
    if (enabled) {
      await _player.setAudioContext(AudioContext(
        android: AudioContextAndroid(
          audioFocus: AndroidAudioFocus.gainTransientMayDuck,
        ),
        iOS: AudioContextIOS(
          category: AVAudioSessionCategory.playback,
          options: {
            AVAudioSessionOptions.duckOthers,
            AVAudioSessionOptions.mixWithOthers,
          },
        ),
      ));
    } else {
      await _player.setAudioContext(AudioContext(
        android: AudioContextAndroid(
          audioFocus: AndroidAudioFocus.gain,
        ),
        iOS: AudioContextIOS(
          category: AVAudioSessionCategory.playback,
          options: {AVAudioSessionOptions.mixWithOthers},
        ),
      ));
    }
  }

  Future<void> playWhistle() async {
    if (!_initialized) await initialize();
    await _player.stop();
    await _player.resume();
  }

  void dispose() {
    _player.dispose();
  }
}
