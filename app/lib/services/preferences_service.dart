import 'package:shared_preferences/shared_preferences.dart';

class PreferencesService {
  static const String _defaultGymIdKey = 'default_gym_id';
  static const String _localeKey = 'app_locale';

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
}
