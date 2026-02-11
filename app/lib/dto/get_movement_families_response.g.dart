// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'get_movement_families_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

GetMovementFamiliesResponse _$GetMovementFamiliesResponseFromJson(
  Map<String, dynamic> json,
) => GetMovementFamiliesResponse(
  movementFamilies:
      (json['movement_families'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList() ??
      [],
);

Map<String, dynamic> _$GetMovementFamiliesResponseToJson(
  GetMovementFamiliesResponse instance,
) => <String, dynamic>{'movement_families': instance.movementFamilies};
