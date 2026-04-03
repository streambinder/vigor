// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'health_weight.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

HealthWeight _$HealthWeightFromJson(Map<String, dynamic> json) => HealthWeight(
  id: json['id'] as String? ?? '',
  userId: json['user_id'] as String? ?? '',
  weight: (json['weight'] as num).toDouble(),
  source: json['source'] as String? ?? '',
  sourceApp: json['source_app'] as String? ?? '',
  measuredAt: DateTime.parse(json['measured_at'] as String),
  hCRecordId: json['hc_record_id'] as String?,
  syncedAt: DateTime.parse(json['synced_at'] as String),
  createdAt: DateTime.parse(json['created_at'] as String),
  updatedAt: DateTime.parse(json['updated_at'] as String),
);

Map<String, dynamic> _$HealthWeightToJson(HealthWeight instance) =>
    <String, dynamic>{
      'id': instance.id,
      'user_id': instance.userId,
      'weight': instance.weight,
      'source': instance.source,
      'source_app': instance.sourceApp,
      'measured_at': HealthWeight._dateTimeToJson(instance.measuredAt),
      'hc_record_id': instance.hCRecordId,
      'synced_at': HealthWeight._dateTimeToJson(instance.syncedAt),
      'created_at': HealthWeight._dateTimeToJson(instance.createdAt),
      'updated_at': HealthWeight._dateTimeToJson(instance.updatedAt),
    };
