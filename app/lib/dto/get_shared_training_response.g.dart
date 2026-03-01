// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'get_shared_training_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

GetSharedTrainingResponse _$GetSharedTrainingResponseFromJson(
  Map<String, dynamic> json,
) => GetSharedTrainingResponse(
  training: json['training'] == null
      ? null
      : Training.fromJson(json['training'] as Map<String, dynamic>),
  owner: SharedTrainingOwner.fromJson(json['owner'] as Map<String, dynamic>),
);

Map<String, dynamic> _$GetSharedTrainingResponseToJson(
  GetSharedTrainingResponse instance,
) => <String, dynamic>{
  'training': instance.training?.toJson(),
  'owner': instance.owner.toJson(),
};
