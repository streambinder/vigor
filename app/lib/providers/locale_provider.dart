import 'package:flutter/material.dart';
import '../services/preferences_service.dart';

/// maps profile language field values to dart locales
const _languageToLocale = {
  'english': Locale('en'),
  'italiano': Locale('it'),
  'español': Locale('es'),
  'français': Locale('fr'),
  'deutsch': Locale('de'),
  'português': Locale('pt'),
  'русский': Locale('ru'),
  '中文': Locale('zh'),
  '日本語': Locale('ja'),
  '한국어': Locale('ko'),
};

/// supported locales for the app
const supportedLocales = [
  Locale('en'),
  Locale('it'),
  Locale('es'),
  Locale('fr'),
  Locale('de'),
  Locale('pt'),
  Locale('ru'),
  Locale('zh'),
  Locale('ja'),
  Locale('ko'),
];

class LocaleProvider extends ChangeNotifier {
  final PreferencesService _prefs;
  Locale _locale = const Locale('en');

  LocaleProvider(this._prefs) {
    _loadSavedLocale();
  }

  Locale get locale => _locale;

  void _loadSavedLocale() {
    final saved = _prefs.locale;
    if (saved != null) {
      final parts = saved.split('_');
      _locale = Locale(parts[0], parts.length > 1 ? parts[1] : null);
    }
  }

  /// sets locale from profile language field (e.g. "italiano")
  void setFromProfileLanguage(String? language) {
    if (language == null || language.isEmpty) return;
    final newLocale = _languageToLocale[language.toLowerCase()];
    if (newLocale != null && supportedLocales.contains(newLocale)) {
      setLocale(newLocale);
    }
  }

  void setLocale(Locale locale) {
    if (!supportedLocales.contains(locale)) return;
    if (_locale == locale) return;

    _locale = locale;
    _prefs.setLocale(locale.toString());
    notifyListeners();
  }
}
