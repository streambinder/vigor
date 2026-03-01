// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_training_complete_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostTrainingCompleteRequest _$PostTrainingCompleteRequestFromJson(
  Map<String, dynamic> json,
) => PostTrainingCompleteRequest(
  quality: json['quality'] as bool?,
  qualityReason: json['qualityReason'] as String? ?? '',
  message: json['message'] as String? ?? '',
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
  'quality': instance.quality,
  'qualityReason': instance.qualityReason,
  'message': instance.message,
  'activityFeedback': instance.activityFeedback,
  'activityReports': instance.activityReports,
  'completedIn': instance.completedIn,
};
