// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'get_equipment_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

GetEquipmentResponse _$GetEquipmentResponseFromJson(
  Map<String, dynamic> json,
) => GetEquipmentResponse(
  equipment:
      (json['equipment'] as List<dynamic>?)?.map((e) => e as String).toList() ??
      [],
);

Map<String, dynamic> _$GetEquipmentResponseToJson(
  GetEquipmentResponse instance,
) => <String, dynamic>{'equipment': instance.equipment};
