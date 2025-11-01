import 'package:logger/logger.dart';

/// Centralized logging service for the application
/// Provides consistent logging across all services and screens
class AppLogger {
  static final Logger _logger = Logger(
    printer: PrettyPrinter(
      methodCount: 0,
      errorMethodCount: 5,
      lineLength: 80,
      colors: true,
      printEmojis: false,
      printTime: true,
    ),
    level: Level.debug,
  );

  /// Get a logger instance with optional tag prefix
  static Logger getLogger([String? tag]) {
    if (tag != null) {
      return Logger(
        printer: _TaggedPrinter(tag),
        level: Level.debug,
      );
    }
    return _logger;
  }

  /// Log debug message
  static void debug(String message, [dynamic error, StackTrace? stackTrace]) {
    _logger.d(message, error: error, stackTrace: stackTrace);
  }

  /// Log info message
  static void info(String message, [dynamic error, StackTrace? stackTrace]) {
    _logger.i(message, error: error, stackTrace: stackTrace);
  }

  /// Log warning message
  static void warning(String message, [dynamic error, StackTrace? stackTrace]) {
    _logger.w(message, error: error, stackTrace: stackTrace);
  }

  /// Log error message
  static void error(String message, [dynamic error, StackTrace? stackTrace]) {
    _logger.e(message, error: error, stackTrace: stackTrace);
  }
}

/// Custom printer that adds a tag prefix to log messages
class _TaggedPrinter extends LogPrinter {
  final String tag;
  final PrettyPrinter _printer = PrettyPrinter(
    methodCount: 0,
    errorMethodCount: 5,
    lineLength: 80,
    colors: true,
    printEmojis: false,
    printTime: true,
  );

  _TaggedPrinter(this.tag);

  @override
  List<String> log(LogEvent event) {
    final taggedEvent = LogEvent(
      event.level,
      '[$tag] ${event.message}',
      error: event.error,
      stackTrace: event.stackTrace,
    );
    return _printer.log(taggedEvent);
  }
}
