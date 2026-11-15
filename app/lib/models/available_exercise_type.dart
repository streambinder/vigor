/// One training Type with its count of servable work-pool exercises,
/// as returned by GET /exercises/available-types.
class AvailableExerciseType {
  final String id;
  final int count;

  const AvailableExerciseType({required this.id, required this.count});

  factory AvailableExerciseType.fromJson(Map<String, dynamic> json) {
    return AvailableExerciseType(
      id: json['id'] as String? ?? '',
      count: (json['count'] as num?)?.toInt() ?? 0,
    );
  }
}
