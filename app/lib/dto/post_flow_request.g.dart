// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_flow_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostFlowRequest _$PostFlowRequestFromJson(Map<String, dynamic> json) =>
    PostFlowRequest(
      duration: (json['duration'] as num).toInt(),
      muscles: (json['muscles'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList(),
      prompt: json['prompt'] as String?,
    );

Map<String, dynamic> _$PostFlowRequestToJson(PostFlowRequest instance) =>
    <String, dynamic>{
      'duration': instance.duration,
      'muscles': instance.muscles,
      'prompt': instance.prompt,
    };
