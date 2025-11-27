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
    try {
      if (profile.data.isEmpty) {
        // If data is completely empty, all nested fields are missing
        missing['goals'] = 'Goals';
        missing['injuries'] = 'Injuries';
        missing['limitations'] = 'Limitations';
      } else {
        final data = profile_models.profileData.fromJson(profile.data);

        // Check goals
        if (data.goals.isEmpty) {
          missing['goals'] = 'Goals';
        }

        // Check injuries (can be empty as user might not have injuries)
        // We'll mark it as missing if null, but empty list is okay

        // Check limitations (can be empty as user might not have limitations)
        // We'll mark it as missing if null, but empty list is okay
      }
    } catch (e) {
      // If data parsing fails, consider all nested fields missing
      missing['goals'] = 'Goals';
      missing['injuries'] = 'Injuries';
      missing['limitations'] = 'Limitations';
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
