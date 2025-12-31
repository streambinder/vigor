// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_register_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostRegisterRequest _$PostRegisterRequestFromJson(Map<String, dynamic> json) =>
    PostRegisterRequest(
      email: json['email'] as String? ?? '',
      password: json['password'] as String? ?? '',
    );

Map<String, dynamic> _$PostRegisterRequestToJson(
  PostRegisterRequest instance,
) => <String, dynamic>{'email': instance.email, 'password': instance.password};
