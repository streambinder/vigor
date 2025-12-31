// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'get_training_partners_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

GetTrainingPartnersResponse _$GetTrainingPartnersResponseFromJson(
  Map<String, dynamic> json,
) => GetTrainingPartnersResponse(
  partners:
      (json['partners'] as List<dynamic>?)
          ?.map((e) => Partner.fromJson(e as Map<String, dynamic>))
          .toList() ??
      [],
);

Map<String, dynamic> _$GetTrainingPartnersResponseToJson(
  GetTrainingPartnersResponse instance,
) => <String, dynamic>{
  'partners': instance.partners.map((e) => e.toJson()).toList(),
};
