// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_training_complete_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostTrainingCompleteRequest _$PostTrainingCompleteRequestFromJson(
  Map<String, dynamic> json,
) => PostTrainingCompleteRequest(
  feedback: TrainingFeedback.fromJson(json['feedback'] as Map<String, dynamic>),
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
  completedIn: (json['completedIn'] as num?)?.toInt(),
);

Map<String, dynamic> _$PostTrainingCompleteRequestToJson(
  PostTrainingCompleteRequest instance,
) => <String, dynamic>{
  'feedback': instance.feedback.toJson(),
  'activityFeedback': instance.activityFeedback,
  'activityReports': instance.activityReports,
  'completedIn': instance.completedIn,
};
