// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'put_gym_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PutGymRequest _$PutGymRequestFromJson(Map<String, dynamic> json) =>
    PutGymRequest(
      name: json['name'] as String?,
      equipment: (json['equipment'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList(),
      modifierVariants: json['modifier_variants'] as Map<String, dynamic>?,
    );

Map<String, dynamic> _$PutGymRequestToJson(PutGymRequest instance) =>
    <String, dynamic>{
      'name': instance.name,
      'equipment': instance.equipment,
      'modifier_variants': instance.modifierVariants,
    };
