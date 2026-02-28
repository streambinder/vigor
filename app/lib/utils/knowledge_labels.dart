import '../generated/app_localizations.dart';

/// centralizes ID-to-l10n mapping for knowledge entities (goals, methodologies, families)
class KnowledgeLabels {
  KnowledgeLabels._();

  // -- goals --

  static String goalLabel(String id, AppLocalizations l10n) {
    return switch (id) {
      'hypertrophy' => l10n.goalHypertrophy,
      'fat loss' => l10n.goalFatLoss,
      'toning' => l10n.goalToning,
      'posture' => l10n.goalPosture,
      'rehabilitation' => l10n.goalRehabilitation,
      'wellness' => l10n.goalWellness,
      'flexibility' => l10n.goalFlexibility,
      'sports' => l10n.goalSports,
      _ => _titleCase(id, ' '),
    };
  }

  static String goalDescription(String id, AppLocalizations l10n) {
    return switch (id) {
      'hypertrophy' => l10n.goalHypertrophyDescription,
      'fat loss' => l10n.goalFatLossDescription,
      'toning' => l10n.goalToningDescription,
      'posture' => l10n.goalPostureDescription,
      'rehabilitation' => l10n.goalRehabilitationDescription,
      'wellness' => l10n.goalWellnessDescription,
      'flexibility' => l10n.goalFlexibilityDescription,
      'sports' => l10n.goalSportsDescription,
      _ => '',
    };
  }

  // -- methodologies --

  static String methodologyLabel(String id, AppLocalizations l10n) {
    return switch (id) {
      'auto' => l10n.methodologyAuto,
      'strength' => l10n.workoutTypeStrength,
      'supersets' => l10n.workoutTypeSupersets,
      'circuit' => l10n.workoutTypeCircuit,
      'emom' => l10n.workoutTypeEmom,
      'amrap' => l10n.workoutTypeAmrap,
      'hiit' => l10n.workoutTypeHiit,
      'for_time' => l10n.workoutTypeForTime,
      'endurance' => l10n.workoutTypeEndurance,
      'mobility' => l10n.workoutTypeMobility,
      _ => id,
    };
  }

  static String methodologyDescription(String id, AppLocalizations l10n) {
    return switch (id) {
      'strength' => l10n.workoutTypeStrengthDescription,
      'supersets' => l10n.workoutTypeSupersetsDescription,
      'circuit' => l10n.workoutTypeCircuitDescription,
      'emom' => l10n.workoutTypeEmomDescription,
      'amrap' => l10n.workoutTypeAmrapDescription,
      'hiit' => l10n.workoutTypeHiitDescription,
      'for_time' => l10n.workoutTypeForTimeDescription,
      'endurance' => l10n.workoutTypeEnduranceDescription,
      'mobility' => l10n.workoutTypeMobilityDescription,
      _ => '',
    };
  }

  // -- movement families --

  static String familyLabel(String id, AppLocalizations l10n) {
    return switch (id) {
      'horizontal_push' => l10n.familyHorizontalPush,
      'horizontal_pull' => l10n.familyHorizontalPull,
      'vertical_push' => l10n.familyVerticalPush,
      'vertical_pull' => l10n.familyVerticalPull,
      'squat' => l10n.familySquat,
      'hinge' => l10n.familyHinge,
      'core' => l10n.familyCore,
      'carry' => l10n.familyCarry,
      'cardio' => l10n.familyCardio,
      'mobility' => l10n.familyMobility,
      'balance' => l10n.familyBalance,
      _ => _titleCase(id, '_'),
    };
  }

  /// preferred UI display order for movement families
  static const familyDisplayOrder = [
    'horizontal_push',
    'horizontal_pull',
    'vertical_push',
    'vertical_pull',
    'squat',
    'hinge',
    'core',
    'cardio',
    'mobility',
    'balance',
    'carry',
  ];

  static String _titleCase(String s, String separator) {
    return s
        .split(separator)
        .map((w) => w.isEmpty ? '' : '${w[0].toUpperCase()}${w.substring(1)}')
        .join(' ');
  }
}
