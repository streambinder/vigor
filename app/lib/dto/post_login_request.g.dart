// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_login_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostLoginRequest _$PostLoginRequestFromJson(Map<String, dynamic> json) =>
    PostLoginRequest(
      email: json['email'] as String? ?? '',
      password: json['password'] as String? ?? '',
    );

Map<String, dynamic> _$PostLoginRequestToJson(PostLoginRequest instance) =>
    <String, dynamic>{'email': instance.email, 'password': instance.password};
