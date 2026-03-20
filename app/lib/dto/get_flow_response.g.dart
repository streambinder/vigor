// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'get_flow_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

GetFlowResponse _$GetFlowResponseFromJson(Map<String, dynamic> json) =>
    GetFlowResponse(
      sessions:
          (json['sessions'] as List<dynamic>?)
              ?.map((e) => FlowSession.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );

Map<String, dynamic> _$GetFlowResponseToJson(GetFlowResponse instance) =>
    <String, dynamic>{
      'sessions': instance.sessions.map((e) => e.toJson()).toList(),
    };
