/// lightweight equipment item from the /equipment API
class EquipmentInfo {
  final String id;
  final bool isWeighted;
  final List<String> aliases;

  const EquipmentInfo({required this.id, required this.isWeighted, this.aliases = const []});

  factory EquipmentInfo.fromJson(Map<String, dynamic> json) => EquipmentInfo(
    id: json['id'] as String,
    isWeighted: json['is_weighted'] as bool? ?? false,
    aliases: (json['aliases'] as List<dynamic>?)?.cast<String>() ?? const [],
  );
}
