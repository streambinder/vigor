// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_user_update_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostUserUpdateResponse _$PostUserUpdateResponseFromJson(
  Map<String, dynamic> json,
) => PostUserUpdateResponse(
  message: json['message'] as String? ?? '',
  profile: Profile.fromJson(json['profile'] as Map<String, dynamic>),
);

Map<String, dynamic> _$PostUserUpdateResponseToJson(
  PostUserUpdateResponse instance,
) => <String, dynamic>{
  'message': instance.message,
  'profile': instance.profile.toJson(),
};
