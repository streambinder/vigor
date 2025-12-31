import 'activity.dart';

extension ActivityExt on Activity {
  /// Returns the display name from detail if available, otherwise falls back to name (ID)
  String get displayName {
    final detailName = detail['name'] as String?;
    return (detailName != null && detailName.isNotEmpty) ? detailName : name;
  }
}
