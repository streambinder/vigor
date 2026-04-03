// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'health_sync_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

HealthSyncRequest _$HealthSyncRequestFromJson(
  Map<String, dynamic> json,
) => HealthSyncRequest(
  metrics:
      (json['metrics'] as List<dynamic>?)
          ?.map((e) => HealthSyncMetric.fromJson(e as Map<String, dynamic>))
          .toList() ??
      [],
  sessions:
      (json['sessions'] as List<dynamic>?)
          ?.map((e) => HealthSyncSession.fromJson(e as Map<String, dynamic>))
          .toList() ??
      [],
  weights:
      (json['weights'] as List<dynamic>?)
          ?.map((e) => HealthSyncWeight.fromJson(e as Map<String, dynamic>))
          .toList() ??
      [],
  hRSamples:
      (json['hr_samples'] as List<dynamic>?)
          ?.map((e) => HealthSyncHRSample.fromJson(e as Map<String, dynamic>))
          .toList() ??
      [],
  deletedRecordIDs:
      (json['deleted_record_ids'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList() ??
      [],
);

Map<String, dynamic> _$HealthSyncRequestToJson(HealthSyncRequest instance) =>
    <String, dynamic>{
      'metrics': instance.metrics.map((e) => e.toJson()).toList(),
      'sessions': instance.sessions.map((e) => e.toJson()).toList(),
      'weights': instance.weights.map((e) => e.toJson()).toList(),
      'hr_samples': instance.hRSamples.map((e) => e.toJson()).toList(),
      'deleted_record_ids': instance.deletedRecordIDs,
    };
