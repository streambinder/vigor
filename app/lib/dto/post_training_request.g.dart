// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_training_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostTrainingRequest _$PostTrainingRequestFromJson(
  Map<String, dynamic> json,
) => PostTrainingRequest(
  duration: (json['duration'] as num).toInt(),
  equipment:
      (json['equipment'] as List<dynamic>?)?.map((e) => e as String).toList() ??
      [],
  gym: json['gym'] as String? ?? '',
  prompt: json['prompt'] as String? ?? '',
  partners:
      (json['partners'] as List<dynamic>?)?.map((e) => e as String).toList() ??
      [],
  skipWarmupCooldown: json['skipWarmupCooldown'] as bool,
  methodology: json['methodology'] as String? ?? '',
);

Map<String, dynamic> _$PostTrainingRequestToJson(
  PostTrainingRequest instance,
) => <String, dynamic>{
  'duration': instance.duration,
  'equipment': instance.equipment,
  'gym': instance.gym,
  'prompt': instance.prompt,
  'partners': instance.partners,
  'skipWarmupCooldown': instance.skipWarmupCooldown,
  'methodology': instance.methodology,
};
