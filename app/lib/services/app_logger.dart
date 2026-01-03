import 'package:flutter/foundation.dart';
import 'package:logger/logger.dart';
import 'console_logger_stub.dart'
    if (dart.library.js_interop) 'console_logger_web.dart' as console;

class AppLogger {
  static final _logger = Logger(
    printer: PrettyPrinter(
      methodCount: 0,
      errorMethodCount: 5,
      lineLength: 80,
      colors: true,
      printEmojis: false,
      dateTimeFormat: DateTimeFormat.onlyTimeAndSinceStart,
    ),
    output: ConsoleOutput(),
    level: Level.debug,
  );

  static void _logWeb(String level, String message, dynamic error, StackTrace? stackTrace) {
    console.consoleLog('[$level] ${DateTime.now().toIso8601String()}: $message');
    if (error != null) console.consoleLog('  error: $error');
    if (stackTrace != null) console.consoleLog('  stack: $stackTrace');
  }

  static void debug(String message, [dynamic error, StackTrace? stackTrace]) =>
      kIsWeb ? _logWeb('DEBUG', message, error, stackTrace) : _logger.d(message, error: error, stackTrace: stackTrace);

  static void info(String message, [dynamic error, StackTrace? stackTrace]) =>
      kIsWeb ? _logWeb('INFO', message, error, stackTrace) : _logger.i(message, error: error, stackTrace: stackTrace);

  static void warning(String message, [dynamic error, StackTrace? stackTrace]) =>
      kIsWeb ? _logWeb('WARN', message, error, stackTrace) : _logger.w(message, error: error, stackTrace: stackTrace);

  static void error(String message, [dynamic error, StackTrace? stackTrace]) =>
      kIsWeb ? _logWeb('ERROR', message, error, stackTrace) : _logger.e(message, error: error, stackTrace: stackTrace);
}
