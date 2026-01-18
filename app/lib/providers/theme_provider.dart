import 'package:flutter/material.dart';
import '../services/preferences_service.dart';

class ThemeProvider extends ChangeNotifier {
  final PreferencesService _prefs;

  ThemeProvider(this._prefs);

  ThemeMode get themeMode => _prefs.themeMode;
  String get themeModeString => _prefs.themeModeString;

  Future<void> setThemeMode(String mode) async {
    await _prefs.setThemeMode(mode);
    notifyListeners();
  }
}
