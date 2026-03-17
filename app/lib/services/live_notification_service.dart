import 'dart:io';
import 'package:flutter/services.dart';

/// service to manage live status bar notifications for training timers
/// android 16+ uses live update notifications with status bar chips
/// ios 16.2+ uses live activities with dynamic island
class LiveNotificationService {
  static const _androidChannel = MethodChannel('it.davidepucci.vigor/timer_notification');

  bool _isActive = false;
  Function()? onTimerStop;
  Function()? onTimerComplete;

  LiveNotificationService() {
    _androidChannel.setMethodCallHandler(_handleMethodCall);
  }

  Future<dynamic> _handleMethodCall(MethodCall call) async {
    switch (call.method) {
      case 'onTimerStop':
        onTimerStop?.call();
        break;
      case 'onTimerComplete':
        onTimerComplete?.call();
        break;
    }
  }

  /// update the live notification with current timer state
  Future<void> updateTimerNotification({
    required String trainingName,
    required int totalElapsedSeconds,
    required String contentText,
    required int intervalRemainingSeconds,
    required bool isDurationBased,
    required double progress,
    required String stopLabel,
    required String completeLabel,
  }) async {
    if (!_isActive) {
      _isActive = true;
    }

    try {
      if (Platform.isAndroid) {
        await _androidChannel.invokeMethod('updateTimerNotification', {
          'trainingName': trainingName,
          'totalElapsedSeconds': totalElapsedSeconds,
          'contentText': contentText,
          'intervalRemainingSeconds': intervalRemainingSeconds,
          'isDurationBased': isDurationBased,
          'progress': progress,
          'stopLabel': stopLabel,
          'completeLabel': completeLabel,
        });
      } else if (Platform.isIOS) {
        await _androidChannel.invokeMethod('updateTimerNotification', {
          'trainingName': trainingName,
          'exerciseName': contentText,
          'remainingSeconds': intervalRemainingSeconds,
          'progress': progress,
        });
      }
    } catch (e) {
      // silently fail on older platforms that don't support live notifications
      // or when permissions are denied
    }
  }

  /// cancel the live notification (call when timer completes or user exits)
  Future<void> cancelTimerNotification() async {
    if (!_isActive) return;

    _isActive = false;

    try {
      if (Platform.isAndroid) {
        await _androidChannel.invokeMethod('cancelTimerNotification');
      } else if (Platform.isIOS) {
        // ios live activity cancellation will go here
      }
    } catch (e) {
      // silently fail
    }
  }

  bool get isActive => _isActive;
}
