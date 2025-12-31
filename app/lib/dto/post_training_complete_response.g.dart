// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_training_complete_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PostTrainingCompleteResponse _$PostTrainingCompleteResponseFromJson(
  Map<String, dynamic> json,
) => PostTrainingCompleteResponse(
  message: json['message'] as String? ?? '',
  training: Training.fromJson(json['training'] as Map<String, dynamic>),
);

Map<String, dynamic> _$PostTrainingCompleteResponseToJson(
  PostTrainingCompleteResponse instance,
) => <String, dynamic>{
  'message': instance.message,
  'training': instance.training.toJson(),
};
