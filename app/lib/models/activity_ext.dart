import 'activity.dart';

extension ActivityExt on Activity {
  /// Returns the display name from detail if available, otherwise falls back to name (ID)
  String get displayName {
    final detailName = detail['name'] as String?;
    return (detailName != null && detailName.isNotEmpty) ? detailName : name;
  }

  /// Format weight_kg for display, dropping decimal for whole numbers (10.0 → "10", 7.5 → "7.5")
  String get weightKgDisplay => weightKg == weightKg.truncateToDouble()
      ? '${weightKg.toInt()}'
      : weightKg.toStringAsFixed(1);
}
