import 'package:flutter/foundation.dart';
import 'package:logger/logger.dart';

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
    filter: ProductionFilter(),
    printer: SimplePrinter(printTime: true, colors: false),
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

  static void debug(String message, [dynamic error, StackTrace? stackTrace]) =>
      _logger.d(message, error: error, stackTrace: stackTrace);

  static void info(String message, [dynamic error, StackTrace? stackTrace]) =>
      _logger.i(message, error: error, stackTrace: stackTrace);

  static void warning(String message, [dynamic error, StackTrace? stackTrace]) =>
      _logger.w(message, error: error, stackTrace: stackTrace);

  static void error(String message, [dynamic error, StackTrace? stackTrace]) =>
      _logger.e(message, error: error, stackTrace: stackTrace);
}
