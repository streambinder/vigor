// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'shared_training_owner.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

SharedTrainingOwner _$SharedTrainingOwnerFromJson(Map<String, dynamic> json) =>
    SharedTrainingOwner(
      userId: json['user_id'] as String? ?? '',
      firstName: json['first_name'] as String? ?? '',
      lastName: json['last_name'] as String? ?? '',
    );

Map<String, dynamic> _$SharedTrainingOwnerToJson(
  SharedTrainingOwner instance,
) => <String, dynamic>{
  'user_id': instance.userId,
  'first_name': instance.firstName,
  'last_name': instance.lastName,
};
