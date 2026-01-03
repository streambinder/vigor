// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'get_goals_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

GetGoalsResponse _$GetGoalsResponseFromJson(Map<String, dynamic> json) =>
    GetGoalsResponse(
      goals:
          (json['goals'] as List<dynamic>?)?.map((e) => e as String).toList() ??
          [],
    );

Map<String, dynamic> _$GetGoalsResponseToJson(GetGoalsResponse instance) =>
    <String, dynamic>{'goals': instance.goals};
