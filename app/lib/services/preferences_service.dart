import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class PreferencesService {
  static const String _defaultGymIdKey = 'default_gym_id';
  static const String _localeKey = 'app_locale';
  static const String _themeModeKey = 'theme_mode';
  static const String _defaultDurationKey = 'default_duration';
  static const String _intervalJingleKey = 'interval_jingle';
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
}
