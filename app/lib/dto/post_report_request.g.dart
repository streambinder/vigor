// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_report_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostReportRequest _$PostReportRequestFromJson(Map<String, dynamic> json) =>
    PostReportRequest(
      trainingId: json['training_id'] as String? ?? '',
      content: json['content'] as String? ?? '',
    );

Map<String, dynamic> _$PostReportRequestToJson(PostReportRequest instance) =>
    <String, dynamic>{
      'training_id': instance.trainingId,
      'content': instance.content,
    };
