// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_refresh_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostRefreshRequest _$PostRefreshRequestFromJson(Map<String, dynamic> json) =>
    PostRefreshRequest(refreshToken: json['refresh_token'] as String? ?? '');

Map<String, dynamic> _$PostRefreshRequestToJson(PostRefreshRequest instance) =>
    <String, dynamic>{'refresh_token': instance.refreshToken};
