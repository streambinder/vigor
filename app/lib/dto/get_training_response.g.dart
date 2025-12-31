// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'get_training_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

GetTrainingResponse _$GetTrainingResponseFromJson(Map<String, dynamic> json) =>
    GetTrainingResponse(
      trainings:
          (json['trainings'] as List<dynamic>?)
              ?.map((e) => Training.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );

Map<String, dynamic> _$GetTrainingResponseToJson(
  GetTrainingResponse instance,
) => <String, dynamic>{
  'trainings': instance.trainings.map((e) => e.toJson()).toList(),
};
