// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'get_methodologies_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

GetMethodologiesResponse _$GetMethodologiesResponseFromJson(
  Map<String, dynamic> json,
) => GetMethodologiesResponse(
  methodologies:
      (json['methodologies'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList() ??
      [],
);

Map<String, dynamic> _$GetMethodologiesResponseToJson(
  GetMethodologiesResponse instance,
) => <String, dynamic>{'methodologies': instance.methodologies};
