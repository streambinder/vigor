import '../models/profile.dart';
import '../models/profile_data.dart' as profile_models;

/// Helper utilities for profile validation
class ProfileHelper {
  /// Check if all required Profile fields are filled
  /// Returns a map of missing field names (json tags) -> human-readable names
  static Map<String, String> getMissingRequiredFields(Profile profile) {
    final Map<String, String> missing = {};

    // Check top-level required fields from Profile.requiredFields
    for (final fieldName in Profile.requiredFields) {
      switch (fieldName) {
        case 'birthdate':
          // Check if birthdate is invalid or a default value
          // Reject: Unix epoch (1970-01-01), Go zero time (0001-01-01),
          // dates before 1900, or dates in the future
          final now = DateTime.now();
          if (profile.birthdate.year < 1900 ||
              profile.birthdate.isAfter(now) ||
              (profile.birthdate.year == 1970 &&
                  profile.birthdate.month == 1 &&
                  profile.birthdate.day == 1)) {
            missing[fieldName] = 'Birth Date';
          }
          break;
        case 'gender':
          if (profile.gender.isEmpty) {
            missing[fieldName] = 'Gender';
          }
          break;
        case 'language':
          if (profile.language.isEmpty) {
            missing[fieldName] = 'Language';
          }
          break;
        case 'height':
          if (profile.height <= 0) {
            missing[fieldName] = 'Height';
          }
          break;
        case 'weight':
          if (profile.weight <= 0) {
            missing[fieldName] = 'Weight';
          }
          break;
      }
    }

    // Check nested profileData required fields
    // Only goals is required; injuries and limitations are optional
    try {
      if (profile.data.isEmpty) {
        missing['goals'] = 'Goals';
      } else {
        final data = profile_models.profileData.fromJson(profile.data);
        if (data.goals.isEmpty) {
          missing['goals'] = 'Goals';
        }
      }
    } catch (e) {
      missing['goals'] = 'Goals';
    }

    return missing;
  }

  /// Check if profile is complete (all required fields filled)
  static bool isProfileComplete(Profile profile) {
    return getMissingRequiredFields(profile).isEmpty;
  }

  /// Get human-readable field names for missing fields
  static List<String> getMissingFieldNames(Profile profile) {
    return getMissingRequiredFields(profile).values.toList();
  }
}
