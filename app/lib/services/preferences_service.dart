import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class PreferencesService {
  static const String _defaultGymIdKey = 'default_gym_id';
  static const String _localeKey = 'app_locale';
  static const String _themeModeKey = 'theme_mode';
  static const String _defaultDurationKey = 'default_duration';
  static const String _intervalJingleKey = 'interval_jingle';
  static const String _duckOtherAudioKey = 'duck_other_audio';
  static const String _warmupCooldownKey = 'warmup_cooldown';
  static const String _useRecommendedDurationKey = 'use_recommended_duration';
  static const String _hcChangesTokenKey = 'hc_changes_token';
  static const String _hcLastSyncMsKey = 'hc_last_sync_ms';
  static const String _hcConnectedKey = 'hc_connected';
  static const String _hcOnboardingDismissedMsKey = 'hc_onboarding_dismissed_ms';
  static const int defaultDurationFallback = 60;

  SharedPreferences? _prefs;

  Future<void> initialize() async {
    _prefs = await SharedPreferences.getInstance();
  }

  String? get defaultGymId => _prefs?.getString(_defaultGymIdKey);

  Future<void> setDefaultGymId(String? id) async {
    if (id == null) {
      await _prefs?.remove(_defaultGymIdKey);
    } else {
      await _prefs?.setString(_defaultGymIdKey, id);
    }
  }

  Future<void> clearDefaultGymIfMatches(String id) async {
    if (defaultGymId == id) {
      await setDefaultGymId(null);
    }
  }

  String? get locale => _prefs?.getString(_localeKey);

  Future<void> setLocale(String locale) async {
    await _prefs?.setString(_localeKey, locale);
  }

  // Theme mode: 'system' (default), 'light', 'dark'
  String get themeModeString => _prefs?.getString(_themeModeKey) ?? 'system';

  ThemeMode get themeMode {
    switch (themeModeString) {
      case 'light':
        return ThemeMode.light;
      case 'dark':
        return ThemeMode.dark;
      default:
        return ThemeMode.system;
    }
  }

  Future<void> setThemeMode(String mode) async {
    await _prefs?.setString(_themeModeKey, mode);
  }

  int get defaultDuration => _prefs?.getInt(_defaultDurationKey) ?? defaultDurationFallback;

  Future<void> setDefaultDuration(int minutes) async {
    await _prefs?.setInt(_defaultDurationKey, minutes);
  }

  bool get intervalJingle => _prefs?.getBool(_intervalJingleKey) ?? true;

  Future<void> setIntervalJingle(bool enabled) async {
    await _prefs?.setBool(_intervalJingleKey, enabled);
  }

  bool get duckOtherAudio => _prefs?.getBool(_duckOtherAudioKey) ?? true;

  Future<void> setDuckOtherAudio(bool enabled) async {
    await _prefs?.setBool(_duckOtherAudioKey, enabled);
  }

  bool get warmupCooldown => _prefs?.getBool(_warmupCooldownKey) ?? true;

  Future<void> setWarmupCooldown(bool enabled) async {
    await _prefs?.setBool(_warmupCooldownKey, enabled);
  }

  bool get useRecommendedDuration => _prefs?.getBool(_useRecommendedDurationKey) ?? true;

  Future<void> setUseRecommendedDuration(bool enabled) async {
    await _prefs?.setBool(_useRecommendedDurationKey, enabled);
  }

  // health connect sync state

  String? get hcChangesToken => _prefs?.getString(_hcChangesTokenKey);

  Future<void> setHcChangesToken(String? token) async {
    if (token == null) {
      await _prefs?.remove(_hcChangesTokenKey);
    } else {
      await _prefs?.setString(_hcChangesTokenKey, token);
    }
  }

  int? get hcLastSyncMs => _prefs?.getInt(_hcLastSyncMsKey);

  Future<void> setHcLastSyncMs(int? ms) async {
    if (ms == null) {
      await _prefs?.remove(_hcLastSyncMsKey);
    } else {
      await _prefs?.setInt(_hcLastSyncMsKey, ms);
    }
  }

  /// clear all health-related keys (called on logout for multi-user scoping)
  Future<void> clearHealthData() async {
    await _prefs?.remove(_hcChangesTokenKey);
    await _prefs?.remove(_hcLastSyncMsKey);
    await _prefs?.remove(_hcConnectedKey);
    await _prefs?.remove(_hcOnboardingDismissedMsKey);
  }

  // health connect connection state

  bool get hcConnected => _prefs?.getBool(_hcConnectedKey) ?? false;

  Future<void> setHcConnected(bool connected) async {
    await _prefs?.setBool(_hcConnectedKey, connected);
  }

  // health onboarding dismissal

  int? get hcOnboardingDismissedMs => _prefs?.getInt(_hcOnboardingDismissedMsKey);

  Future<void> setHcOnboardingDismissedMs(int ms) async {
    await _prefs?.setInt(_hcOnboardingDismissedMsKey, ms);
  }

  /// true if onboarding was dismissed less than 30 days ago
  bool get hcOnboardingRecentlyDismissed {
    final dismissedMs = hcOnboardingDismissedMs;
    if (dismissedMs == null) return false;
    final elapsed = DateTime.now().millisecondsSinceEpoch - dismissedMs;
    return elapsed < const Duration(days: 30).inMilliseconds;
  }
}
