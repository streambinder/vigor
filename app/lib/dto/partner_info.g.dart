// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'partner_info.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PartnerInfo _$PartnerInfoFromJson(Map<String, dynamic> json) => PartnerInfo(
  id: json['id'] as String? ?? '',
  trainingId: json['training_id'] as String? ?? '',
  userId: json['user_id'] as String? ?? '',
  firstName: json['first_name'] as String? ?? '',
  lastName: json['last_name'] as String? ?? '',
  createdAt: json['created_at'] as String? ?? '',
);

Map<String, dynamic> _$PartnerInfoToJson(PartnerInfo instance) =>
    <String, dynamic>{
      'id': instance.id,
      'training_id': instance.trainingId,
      'user_id': instance.userId,
      'first_name': instance.firstName,
      'last_name': instance.lastName,
      'created_at': instance.createdAt,
    };
