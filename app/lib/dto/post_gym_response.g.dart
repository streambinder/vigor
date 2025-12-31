// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_gym_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostGymResponse _$PostGymResponseFromJson(Map<String, dynamic> json) =>
    PostGymResponse(
      message: json['message'] as String? ?? '',
      gym: Gym.fromJson(json['gym'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$PostGymResponseToJson(PostGymResponse instance) =>
    <String, dynamic>{
      'message': instance.message,
      'gym': instance.gym.toJson(),
    };
