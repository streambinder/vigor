// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_flow_complete_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostFlowCompleteRequest _$PostFlowCompleteRequestFromJson(
  Map<String, dynamic> json,
) => PostFlowCompleteRequest(
  completedIn: (json['completed_in'] as num?)?.toInt(),
);

Map<String, dynamic> _$PostFlowCompleteRequestToJson(
  PostFlowCompleteRequest instance,
) => <String, dynamic>{'completed_in': instance.completedIn};
