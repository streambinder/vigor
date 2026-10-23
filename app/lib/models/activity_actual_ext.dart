import 'activity_actual.dart';

/// weight for display, matching the activity tag format
/// (10.0 → "10", 7.5 → "7.5")
String formatWeightValue(double value) =>
    value == value.truncateToDouble() ? '${value.toInt()}' : value.toStringAsFixed(1);

extension ActivityActualExt on ActivityActual {
  bool get isSkipped => skipped ?? false;

  /// wire form for the feedback request: null fields stay out, so
  /// the backend reads them as the prescribed value standing
  Map<String, dynamic> toApiJson() {
    final json = toJson();
    json.removeWhere((_, value) => value == null);
    return json;
  }
}
