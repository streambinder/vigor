// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_gym_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostGymRequest _$PostGymRequestFromJson(Map<String, dynamic> json) =>
    PostGymRequest(
      name: json['name'] as String? ?? '',
      equipment:
          (json['equipment'] as List<dynamic>?)
              ?.map((e) => e as String)
              .toList() ??
          [],
    );

Map<String, dynamic> _$PostGymRequestToJson(PostGymRequest instance) =>
    <String, dynamic>{'name': instance.name, 'equipment': instance.equipment};
