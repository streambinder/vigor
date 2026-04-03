// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'health_sync_weight.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

HealthSyncWeight _$HealthSyncWeightFromJson(Map<String, dynamic> json) =>
    HealthSyncWeight(
      hCRecordId: json['hc_record_id'] as String? ?? '',
      sourceApp: json['source_app'] as String? ?? '',
      measuredAt: (json['measured_at'] as num).toInt(),
      weight: (json['weight'] as num).toDouble(),
    );

Map<String, dynamic> _$HealthSyncWeightToJson(HealthSyncWeight instance) =>
    <String, dynamic>{
      'hc_record_id': instance.hCRecordId,
      'source_app': instance.sourceApp,
      'measured_at': instance.measuredAt,
      'weight': instance.weight,
    };
