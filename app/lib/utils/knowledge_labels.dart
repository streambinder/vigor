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

  // -- equipment --

  static String equipmentLabel(String id, AppLocalizations l10n) {
    return switch (id) {
      'partner' => l10n.equipmentPartner,
      'balance board' => l10n.equipmentBalanceBoard,
      'band' => l10n.equipmentBand,
      'barbell' => l10n.equipmentBarbell,
      'bench' => l10n.equipmentBench,
      'box' => l10n.equipmentBox,
      'bosu ball' => l10n.equipmentBosuBall,
      'cable' => l10n.equipmentCable,
      'dip station' => l10n.equipmentDipStation,
      'dumbbell' => l10n.equipmentDumbbell,
      'elliptical machine' => l10n.equipmentEllipticalMachine,
      'ez barbell' => l10n.equipmentEzBarbell,
      'hammer' => l10n.equipmentHammer,
      'kettlebell' => l10n.equipmentKettlebell,
      'leverage machine' => l10n.equipmentLeverageMachine,
      'medicine ball' => l10n.equipmentMedicineBall,
      'olympic barbell' => l10n.equipmentOlympicBarbell,
      'pull-up bar' => l10n.equipmentPullUpBar,
      'resistance band' => l10n.equipmentResistanceBand,
      'rings' => l10n.equipmentRings,
      'roller' => l10n.equipmentRoller,
      'rope' => l10n.equipmentRope,
      'rowing machine' => l10n.equipmentRowingMachine,
      'skierg machine' => l10n.equipmentSkiergMachine,
      'sled machine' => l10n.equipmentSledMachine,
      'smith machine' => l10n.equipmentSmithMachine,
      'stability ball' => l10n.equipmentStabilityBall,
      'stationary bike' => l10n.equipmentStationaryBike,
      'stepmill machine' => l10n.equipmentStepmillMachine,
      'tire' => l10n.equipmentTire,
      'treadmill' => l10n.equipmentTreadmill,
      'trap bar' => l10n.equipmentTrapBar,
      'trx' => l10n.equipmentTrx,
      'upper body ergometer' => l10n.equipmentUpperBodyErgometer,
      'wheel roller' => l10n.equipmentWheelRoller,
      // modifiers are mixed with equipment in gym/favorites lists
      'weighted vest' => l10n.modifierWeightedVest,
      'parallettes' => l10n.modifierParallettes,
      'ankle weights' => l10n.modifierAnkleWeights,
      'dip belt' => l10n.modifierDipBelt,
      'push up bars' => l10n.modifierPushUpBars,
      'resistance bands' => l10n.modifierResistanceBands,
      'weight' => l10n.modifierWeight,
      'wrist weights' => l10n.modifierWristWeights,
      _ => _titleCase(id, ' '),
    };
  }

  // -- muscles --

  static String muscleLabel(String id, AppLocalizations l10n) {
    return switch (id) {
      'chest' => l10n.muscleChest,
      'back' => l10n.muscleBack,
      'shoulders' => l10n.muscleShoulders,
      'arms' => l10n.muscleArms,
      'core' => l10n.muscleCore,
      'glutes' => l10n.muscleGlutes,
      'legs' => l10n.muscleLegs,
      _ => _titleCase(id, '_'),
    };
  }

  // -- modifiers --

  static String modifierLabel(String id, AppLocalizations l10n) {
    return switch (id) {
      'weighted vest' => l10n.modifierWeightedVest,
      'parallettes' => l10n.modifierParallettes,
      'ankle weights' => l10n.modifierAnkleWeights,
      'dip belt' => l10n.modifierDipBelt,
      'push up bars' => l10n.modifierPushUpBars,
      'resistance bands' => l10n.modifierResistanceBands,
      'weight' => l10n.modifierWeight,
      'wrist weights' => l10n.modifierWristWeights,
      _ => _titleCase(id, ' '),
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
