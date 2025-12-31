// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'get_gyms_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

GetGymsResponse _$GetGymsResponseFromJson(Map<String, dynamic> json) =>
    GetGymsResponse(
      gyms:
          (json['gyms'] as List<dynamic>?)
              ?.map((e) => Gym.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );

Map<String, dynamic> _$GetGymsResponseToJson(GetGymsResponse instance) =>
    <String, dynamic>{'gyms': instance.gyms.map((e) => e.toJson()).toList()};
