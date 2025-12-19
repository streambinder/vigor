// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appName => 'Vigor';

  @override
  String get appTagline => 'Ex Sapientia Vis';

  @override
  String get navHome => 'Home';

  @override
  String get navActivity => 'Activity';

  @override
  String get navProfile => 'Profile';

  @override
  String get storageErrorTitle => 'Vigor - Storage Error';

  @override
  String get storageUnavailable => 'Storage Unavailable';

  @override
  String get storageErrorMessage =>
      'This app requires secure storage to protect your data. Please check your browser settings and try again.';

  @override
  String get signInWithGoogle => 'Sign in with Google';

  @override
  String get signingIn => 'Signing in...';

  @override
  String get failedToInitializeGoogleSignIn =>
      'Failed to initialize Google Sign In';

  @override
  String signInError(String message) {
    return 'Sign-in error: $message';
  }

  @override
  String get googleSignInFailed => 'Google sign-in failed';

  @override
  String get failedToGetAuthToken => 'Failed to get authentication token';

  @override
  String errorProcessingSignIn(String message) {
    return 'Error processing sign-in: $message';
  }

  @override
  String get googleSignInInitializing =>
      'Google Sign In is still initializing...';

  @override
  String get readyToTrain => 'Ready to train?';

  @override
  String get generateTrainingDescription =>
      'Generate a personalized training based on your profile and goals';

  @override
  String get generateTraining => 'Generate Training';

  @override
  String get refresh => 'Refresh';

  @override
  String get logout => 'Logout';

  @override
  String get userDataRefreshed => 'User data refreshed';

  @override
  String get editProfile => 'Edit Profile';

  @override
  String get settings => 'Settings';

  @override
  String get deleteGym => 'Delete Gym';

  @override
  String deleteGymConfirmation(String name) {
    return 'Are you sure you want to delete \"$name\"?';
  }

  @override
  String get cancel => 'Cancel';

  @override
  String get delete => 'Delete';

  @override
  String get logoutConfirmation => 'Are you sure you want to logout?';

  @override
  String get deleteAccount => 'Delete Account';

  @override
  String get deleteAccountConfirmation =>
      'Are you sure you want to delete your account? This action cannot be undone.';

  @override
  String get accountDeletedSuccessfully => 'Account deleted successfully';

  @override
  String get failedToDeleteAccount => 'Failed to delete account';

  @override
  String get failedToLoadGyms => 'Failed to load gyms';

  @override
  String get gymAddedSuccessfully => 'Gym added successfully';

  @override
  String get failedToAddGym => 'Failed to add gym';

  @override
  String get gymUpdatedSuccessfully => 'Gym updated successfully';

  @override
  String get failedToUpdateGym => 'Failed to update gym';

  @override
  String get gymDeletedSuccessfully => 'Gym deleted successfully';

  @override
  String get failedToDeleteGym => 'Failed to delete gym';

  @override
  String get birthdate => 'Birthdate';

  @override
  String get gender => 'Gender';

  @override
  String get language => 'Language';

  @override
  String get height => 'Height';

  @override
  String get weight => 'Weight';

  @override
  String get heightUnit => 'cm';

  @override
  String get weightUnit => 'kg';

  @override
  String heightWithUnit(double value) {
    return '$value cm';
  }

  @override
  String weightWithUnit(double value) {
    return '$value kg';
  }

  @override
  String get goals => 'Goals';

  @override
  String get injuries => 'Injuries';

  @override
  String get limitations => 'Limitations';

  @override
  String get favorites => 'Favorites';

  @override
  String get exercises => 'Exercises';

  @override
  String get equipment => 'Equipment';

  @override
  String startedDate(String date) {
    return 'Started: $date';
  }

  @override
  String yearLabel(int year) {
    return 'Year: $year';
  }

  @override
  String get myGyms => 'My Gyms';

  @override
  String get addGym => 'Add Gym';

  @override
  String get noGymsAddedYet => 'No gyms added yet';

  @override
  String get addYourFirstGym => 'Add Your First Gym';

  @override
  String get removeDefault => 'Remove Default';

  @override
  String get setAsDefault => 'Set as Default';

  @override
  String get edit => 'Edit';

  @override
  String get quickActions => 'Quick Actions';

  @override
  String get dangerZone => 'Danger Zone';

  @override
  String get completeYourProfile => 'Complete Your Profile';

  @override
  String get updateYourProfileInfo => 'Update your profile information below.';

  @override
  String get pleaseCompleteProfile =>
      'Please complete your profile. Fields marked with * are required.';

  @override
  String get firstName => 'First Name';

  @override
  String get lastName => 'Last Name';

  @override
  String get birthDate => 'Birth Date';

  @override
  String get male => 'Male';

  @override
  String get female => 'Female';

  @override
  String get selectLanguage => 'Select language';

  @override
  String get heightCm => 'Height (cm)';

  @override
  String get weightKg => 'Weight (kg)';

  @override
  String get required => 'Required';

  @override
  String get invalid => 'Invalid';

  @override
  String get pleaseSelectBirthDate => 'Please select your birth date';

  @override
  String get pleaseAddAtLeastOneGoal => 'Please add at least one goal';

  @override
  String get pleaseSelectLanguage => 'Please select your language';

  @override
  String get addAGoal => 'Add a goal';

  @override
  String get injuryDescription => 'Injury description';

  @override
  String get year => 'Year';

  @override
  String get addALimitation => 'Add a limitation';

  @override
  String get favoriteExercisesHint => 'e.g., squats, pull-ups, running';

  @override
  String get favoriteEquipmentHint => 'e.g., dumbbells, barbell, kettlebell';

  @override
  String get saveChanges => 'Save Changes';

  @override
  String get saveProfile => 'Save Profile';

  @override
  String get optionalLeaveEmpty => '(Optional - leave empty if none)';

  @override
  String get optionalExercisesPrefer =>
      '(Optional - exercises you enjoy or prefer)';

  @override
  String get optionalEquipmentPrefer =>
      '(Optional - equipment you prefer using)';

  @override
  String get optionalWorkoutTypesPrefer =>
      '(Optional - workout styles you prefer)';

  @override
  String get favoriteExercises => 'Favorite Exercises';

  @override
  String get favoriteEquipment => 'Favorite Equipment';

  @override
  String get favoriteWorkoutTypes => 'Preferred Workout Types';

  @override
  String get workoutTypeStrength => 'Strength';

  @override
  String get workoutTypeCircuit => 'Circuit';

  @override
  String get workoutTypeEmom => 'EMOM';

  @override
  String get workoutTypeAmrap => 'AMRAP';

  @override
  String get workoutTypeHiit => 'HIIT';

  @override
  String get workoutTypeForTime => 'For Time';

  @override
  String get workoutTypeEndurance => 'Endurance';

  @override
  String get workoutTypeMobility => 'Mobility';

  @override
  String get failedToUpdateProfile => 'Failed to update profile';

  @override
  String get activity => 'Activity';

  @override
  String get noTrainingsYet => 'No trainings yet';

  @override
  String get generateFirstTraining =>
      'Generate your first training from the Home tab';

  @override
  String get noTrainingAvailable =>
      'No training available. Start generating one.';

  @override
  String get availableTrainings => 'Available trainings';

  @override
  String get pastTrainings => 'Past trainings';

  @override
  String get stale => 'Stale';

  @override
  String get copied => 'Copied';

  @override
  String durationMin(int minutes) {
    return '$minutes min';
  }

  @override
  String durationHr(int hours) {
    return '$hours hr';
  }

  @override
  String durationHrMin(int hours, int minutes) {
    return '$hours hr $minutes min';
  }

  @override
  String get failedToLoadTrainings => 'Failed to load trainings';

  @override
  String get startTraining => 'Start Training';

  @override
  String get cloneTraining => 'Clone Training';

  @override
  String get addPartner => 'Add Partner';

  @override
  String get shareWithUser => 'Share with User';

  @override
  String get deleteTraining => 'Delete Training';

  @override
  String get leaveTraining => 'Leave Training';

  @override
  String get showAiReasoning => 'Show AI reasoning';

  @override
  String get reportIssue => 'Report issue';

  @override
  String deleteTrainingConfirmation(String name) {
    return 'Are you sure you want to delete \"$name\"? This action cannot be undone.';
  }

  @override
  String leaveTrainingConfirmation(String name) {
    return 'Are you sure you want to leave \"$name\"? You will no longer see this training.';
  }

  @override
  String addPartnerConfirmation(String userName, String trainingName) {
    return 'Add $userName as a partner to \"$trainingName\"?';
  }

  @override
  String cloneTrainingConfirmation(String name) {
    return 'Clone \"$name\" to your trainings?';
  }

  @override
  String shareTrainingConfirmation(String trainingName, String userName) {
    return 'Share \"$trainingName\" with $userName?';
  }

  @override
  String get trainingDeletedSuccessfully => 'Training deleted successfully';

  @override
  String get failedToDeleteTraining => 'Failed to delete training';

  @override
  String get leftTrainingSuccessfully => 'Left training successfully';

  @override
  String get partnerAddedSuccessfully => 'Partner added successfully';

  @override
  String get failedToAddPartner => 'Failed to add partner';

  @override
  String get trainingSharedSuccessfully => 'Training shared successfully';

  @override
  String get failedToShareTraining => 'Failed to share training';

  @override
  String get trainingCloned => 'Training cloned';

  @override
  String get failedToCloneTraining => 'Failed to clone training';

  @override
  String get trainingMarkedAsComplete => 'Training marked as complete';

  @override
  String get failedToCompleteTraining => 'Failed to complete training';

  @override
  String get reportSubmitted => 'Report submitted';

  @override
  String get failedToSubmitReport => 'Failed to submit report';

  @override
  String get reasoning => 'Reasoning';

  @override
  String get strategy => 'Strategy';

  @override
  String get typeSelection => 'Type Selection';

  @override
  String get progression => 'Progression';

  @override
  String get constraints => 'Constraints';

  @override
  String get researchApplied => 'Research Applied';

  @override
  String get targetMuscles => 'Target Muscles';

  @override
  String get naming => 'Naming';

  @override
  String get trainingRoutines => 'Training Routines';

  @override
  String get noEquipment => 'No equipment';

  @override
  String blockNumber(int number) {
    return 'Block $number';
  }

  @override
  String repeatsCount(int count) {
    return '${count}x';
  }

  @override
  String durationSeconds(int seconds) {
    return '${seconds}s';
  }

  @override
  String restSeconds(int seconds) {
    return '${seconds}s rest';
  }

  @override
  String repsCount(int count) {
    return '$count reps';
  }

  @override
  String weightKgValue(double value) {
    return '$value kg';
  }

  @override
  String get markAsComplete => 'Mark as Complete';

  @override
  String get references => 'References';

  @override
  String get describeIssue => 'Describe the issue with this training...';

  @override
  String get submit => 'Submit';

  @override
  String get close => 'Close';

  @override
  String get add => 'Add';

  @override
  String get update => 'Update';

  @override
  String get clone => 'Clone';

  @override
  String get share => 'Share';

  @override
  String get leave => 'Leave';

  @override
  String get tapToStart => 'Tap to start';

  @override
  String get trainingCompleted => 'Training Completed!';

  @override
  String greatJobCompleting(String name) {
    return 'Great job completing $name';
  }

  @override
  String get done => 'Done';

  @override
  String get complete => 'Complete';

  @override
  String routineCounter(int current, int total) {
    return 'Routine $current/$total';
  }

  @override
  String blockCounter(int current, int total) {
    return 'Block $current/$total';
  }

  @override
  String get exitTraining => 'Exit Training?';

  @override
  String get whatWouldYouLikeToDo => 'What would you like to do?';

  @override
  String get exit => 'Exit';

  @override
  String get continueTraining => 'Continue';

  @override
  String get failedToMarkComplete => 'Failed to mark training as complete';

  @override
  String get durationMinutes => 'Duration (minutes)';

  @override
  String get bodyweight => 'Bodyweight';

  @override
  String get gym => 'Gym';

  @override
  String get custom => 'Custom';

  @override
  String get noEquipmentBodyweightOnly =>
      'No equipment - bodyweight exercises only';

  @override
  String get noGymsDefinedCreateOne =>
      'No gyms defined. Create one in your profile settings.';

  @override
  String get selectAGym => 'Select a gym';

  @override
  String get addEquipment => 'Add Equipment';

  @override
  String get addEquipmentAvailable => 'Add the equipment you have available';

  @override
  String get includeWarmupCooldown => 'Include warm-up & cooldown';

  @override
  String get equipmentPlaceholder => 'e.g., Barbell, Dumbbells';

  @override
  String get customPromptOptional => 'Custom Prompt (optional)';

  @override
  String get focusOnUpperBody => 'e.g., Focus on upper body';

  @override
  String get trainingPartnersOptional => 'Training Partners (optional)';

  @override
  String get generatingTraining => 'Generating your training...';

  @override
  String get thisMayTakeAMoment => 'This may take a moment';

  @override
  String get trainingGeneratedSuccessfully =>
      'Training generated successfully!';

  @override
  String get failedToGenerateTraining => 'Failed to generate training';

  @override
  String get generate => 'Generate';

  @override
  String get editGym => 'Edit Gym';

  @override
  String get gymName => 'Gym Name';

  @override
  String get gymNamePlaceholder => 'e.g., Home Gym, LA Fitness';

  @override
  String get noEquipmentAddedYet => 'No equipment added yet';

  @override
  String get pleaseEnterGymName => 'Please enter a gym name';

  @override
  String get addAllEquipment => 'Add All';

  @override
  String get failedToLoadEquipment => 'Failed to load equipment';

  @override
  String get selectUser => 'Select User';

  @override
  String get searchByName => 'Search by name';

  @override
  String get noUsersAvailable => 'No users available';

  @override
  String get noMatchingUsers => 'No matching users';

  @override
  String get instructions => 'Instructions';

  @override
  String get howWasYourTraining => 'How was your training?';

  @override
  String get anyAdditionalComments => 'Any additional comments?';

  @override
  String get tooEasy => 'Too easy';

  @override
  String get tooHard => 'Too hard';

  @override
  String get flag => 'Flag';

  @override
  String get profile => 'Profile';

  @override
  String get languageEnglish => 'English';

  @override
  String get languageItaliano => 'Italiano';

  @override
  String get languageEspanol => 'Español';

  @override
  String get languageFrancais => 'Français';

  @override
  String get languageDeutsch => 'Deutsch';

  @override
  String get languagePortugues => 'Português';

  @override
  String get languageRussian => 'Русский';

  @override
  String get languageChinese => '中文';

  @override
  String get languageJapanese => '日本語';

  @override
  String get languageKorean => '한국어';
}
