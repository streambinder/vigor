import 'package:shared_preferences/shared_preferences.dart';

class PreferencesService {
  static const String _defaultGymNameKey = 'default_gym_name';

  SharedPreferences? _prefs;

  Future<void> initialize() async {
    _prefs = await SharedPreferences.getInstance();
  }

  String? get defaultGymName => _prefs?.getString(_defaultGymNameKey);

  Future<void> setDefaultGymName(String? name) async {
    if (name == null) {
      await _prefs?.remove(_defaultGymNameKey);
    } else {
      await _prefs?.setString(_defaultGymNameKey, name);
    }
  }

  Future<void> clearDefaultGymIfMatches(String name) async {
    if (defaultGymName == name) {
      await setDefaultGymName(null);
    }
  }
}
