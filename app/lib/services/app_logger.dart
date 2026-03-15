import 'package:flutter/foundation.dart';
import 'package:logger/logger.dart';
import 'console_logger_stub.dart'
    if (dart.library.js_interop) 'console_logger_web.dart' as console;

/// in-memory ring buffer that also forwards to console
class _MemoryOutput extends LogOutput {
  static const int maxEntries = 2000;
  final List<String> _buffer = [];
  final LogOutput _consoleOutput = ConsoleOutput();
  // guard to prevent duplicate captures when logger's ConsoleOutput calls debugPrint
  bool _outputting = false;

  List<String> get entries => List.unmodifiable(_buffer);

  void clear() => _buffer.clear();

  void addRaw(String line) {
    // skip lines already captured by output() — ConsoleOutput calls debugPrint internally
    if (_outputting) return;
    if (_buffer.length >= maxEntries) _buffer.removeAt(0);
    _buffer.add(line);
  }

  @override
  void output(OutputEvent event) {
    _outputting = true;
    _consoleOutput.output(event);
    _outputting = false;
    for (final line in event.lines) {
      if (_buffer.length >= maxEntries) _buffer.removeAt(0);
      _buffer.add(line);
    }
  }
}

class AppLogger {
  static final _memoryOutput = _MemoryOutput();

  static final _logger = Logger(
    // DevelopmentFilter (default) swallows all logs in release mode via assert() trick
    filter: ProductionFilter(),
    printer: PrettyPrinter(
      methodCount: 0,
      errorMethodCount: 5,
      lineLength: 80,
      colors: true,
      printEmojis: false,
      dateTimeFormat: DateTimeFormat.onlyTimeAndSinceStart,
    ),
    output: _memoryOutput,
    level: Level.debug,
  );

  static List<String> get logs => _memoryOutput.entries;
  static void clearLogs() => _memoryOutput.clear();

  /// call once in main() to intercept all print/debugPrint output into the log buffer
  static void captureAllLogs() {
    final originalDebugPrint = debugPrint;
    debugPrint = (String? message, {int? wrapWidth}) {
      if (message != null) _memoryOutput.addRaw(message);
      originalDebugPrint(message, wrapWidth: wrapWidth);
    };
  }

  static void _logWeb(String level, String message, dynamic error, StackTrace? stackTrace) {
    final entry = '[$level] ${DateTime.now().toIso8601String()}: $message'
        '${error != null ? '\n  error: $error' : ''}'
        '${stackTrace != null ? '\n  stack: $stackTrace' : ''}';
    console.consoleLog(entry);
    if (_memoryOutput._buffer.length >= _MemoryOutput.maxEntries) _memoryOutput._buffer.removeAt(0);
    _memoryOutput._buffer.add(entry);
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
