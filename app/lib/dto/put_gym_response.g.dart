// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'put_gym_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PutGymResponse _$PutGymResponseFromJson(Map<String, dynamic> json) =>
    PutGymResponse(
      message: json['message'] as String? ?? '',
      gym: Gym.fromJson(json['gym'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$PutGymResponseToJson(PutGymResponse instance) =>
    <String, dynamic>{
      'message': instance.message,
      'gym': instance.gym.toJson(),
    };
