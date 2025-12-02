import 'dart:js_interop';

@JS('console.log')
external void _log(JSString message);

void consoleLog(String message) => _log(message.toJS);
