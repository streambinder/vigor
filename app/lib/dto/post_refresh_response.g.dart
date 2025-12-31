// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_refresh_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostRefreshResponse _$PostRefreshResponseFromJson(Map<String, dynamic> json) =>
    PostRefreshResponse(
      accessToken: json['access_token'] as String? ?? '',
      refreshToken: json['refresh_token'] as String? ?? '',
    );

Map<String, dynamic> _$PostRefreshResponseToJson(
  PostRefreshResponse instance,
) => <String, dynamic>{
  'access_token': instance.accessToken,
  'refresh_token': instance.refreshToken,
};
