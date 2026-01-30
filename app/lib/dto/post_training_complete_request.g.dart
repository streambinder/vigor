// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_training_complete_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostTrainingCompleteRequest _$PostTrainingCompleteRequestFromJson(
  Map<String, dynamic> json,
) => PostTrainingCompleteRequest(
  feedback: json['feedback'] as String? ?? '',
  activityFeedback:
      (json['activityFeedback'] as Map<String, dynamic>?)?.map(
        (k, e) => MapEntry(k, e as String),
      ) ??
      {},
  activityReports:
      (json['activityReports'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList() ??
      [],
);

Map<String, dynamic> _$PostTrainingCompleteRequestToJson(
  PostTrainingCompleteRequest instance,
) => <String, dynamic>{
  'feedback': instance.feedback,
  'activityFeedback': instance.activityFeedback,
  'activityReports': instance.activityReports,
};
