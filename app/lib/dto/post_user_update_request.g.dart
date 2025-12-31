// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_user_update_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostUserUpdateRequest _$PostUserUpdateRequestFromJson(
  Map<String, dynamic> json,
) => PostUserUpdateRequest(
  firstName: json['first_name'] as String? ?? '',
  lastName: json['last_name'] as String? ?? '',
  birthdate: json['birthdate'] as String? ?? '',
  gender: json['gender'] as String? ?? '',
  language: json['language'] as String? ?? '',
  height: (json['height'] as num).toDouble(),
  weight: (json['weight'] as num).toDouble(),
  data: json['data'] as Map<String, dynamic>,
);

Map<String, dynamic> _$PostUserUpdateRequestToJson(
  PostUserUpdateRequest instance,
) => <String, dynamic>{
  'first_name': instance.firstName,
  'last_name': instance.lastName,
  'birthdate': instance.birthdate,
  'gender': instance.gender,
  'language': instance.language,
  'height': instance.height,
  'weight': instance.weight,
  'data': instance.data,
};
