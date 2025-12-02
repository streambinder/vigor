import 'dart:js_interop';

@JS('window.ENV.API_URL')
external JSString? get _apiUrl;

String? getApiUrlFromWindow() {
  try {
    return _apiUrl?.toDart;
  } catch (_) {
    return null;
  }
}
