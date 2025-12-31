// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_login_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostLoginResponse _$PostLoginResponseFromJson(Map<String, dynamic> json) =>
    PostLoginResponse(
      accessToken: json['access_token'] as String? ?? '',
      refreshToken: json['refresh_token'] as String? ?? '',
    );

Map<String, dynamic> _$PostLoginResponseToJson(PostLoginResponse instance) =>
    <String, dynamic>{
      'access_token': instance.accessToken,
      'refresh_token': instance.refreshToken,
    };
