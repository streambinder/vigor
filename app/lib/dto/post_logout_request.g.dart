// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_logout_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostLogoutRequest _$PostLogoutRequestFromJson(Map<String, dynamic> json) =>
    PostLogoutRequest(refreshToken: json['refresh_token'] as String? ?? '');

Map<String, dynamic> _$PostLogoutRequestToJson(PostLogoutRequest instance) =>
    <String, dynamic>{'refresh_token': instance.refreshToken};
