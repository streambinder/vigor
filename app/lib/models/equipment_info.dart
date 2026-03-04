/// lightweight equipment item from the /equipment API
class EquipmentInfo {
  final String id;
  final bool isWeighted;

  const EquipmentInfo({required this.id, required this.isWeighted});

  factory EquipmentInfo.fromJson(Map<String, dynamic> json) => EquipmentInfo(
    id: json['id'] as String,
    isWeighted: json['is_weighted'] as bool? ?? false,
  );
}
