// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'get_muscles_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

GetMusclesResponse _$GetMusclesResponseFromJson(Map<String, dynamic> json) =>
    GetMusclesResponse(
      muscles:
          (json['muscles'] as List<dynamic>?)
              ?.map((e) => e as String)
              .toList() ??
          [],
    );

Map<String, dynamic> _$GetMusclesResponseToJson(GetMusclesResponse instance) =>
    <String, dynamic>{'muscles': instance.muscles};
