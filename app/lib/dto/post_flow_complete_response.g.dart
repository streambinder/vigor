// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_flow_complete_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostFlowCompleteResponse _$PostFlowCompleteResponseFromJson(
  Map<String, dynamic> json,
) => PostFlowCompleteResponse(
  message: json['message'] as String? ?? '',
  session: FlowSession.fromJson(json['session'] as Map<String, dynamic>),
);

Map<String, dynamic> _$PostFlowCompleteResponseToJson(
  PostFlowCompleteResponse instance,
) => <String, dynamic>{
  'message': instance.message,
  'session': instance.session.toJson(),
};
