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
      'Secure storage required. Check browser settings.';

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
      'Create a workout tailored to your goals';

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
  String get other => 'Other';

  @override
  String get deleteGym => 'Delete Gym';

  @override
  String deleteGymConfirmation(String name) {
    return 'Delete \"$name\"?';
  }

  @override
  String get cancel => 'Cancel';

  @override
  String get delete => 'Delete';

  @override
  String get logoutConfirmation => 'Logout?';

  @override
  String get deleteAccount => 'Delete Account';

  @override
  String get deleteAccountConfirmation =>
      'Delete your account? This cannot be undone.';

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
  String get conditions => 'Conditions';

  @override
  String get favorites => 'Favorites';

  @override
  String get personalDetails => 'Personal Details';

  @override
  String get healthAndGoals => 'Health & Goals';

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
  String get myGyms => 'Gyms';

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
  String get updateYourProfileInfo => 'Update your info below.';

  @override
  String get pleaseCompleteProfile => 'Complete your profile. * = required.';

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
  String get addACondition => 'Add a condition';

  @override
  String get favoriteExercisesHint => 'e.g., squats, pull-ups, running';

  @override
  String get favoriteEquipmentHint => 'e.g., dumbbells, barbell, kettlebell';

  @override
  String get saveChanges => 'Save Changes';

  @override
  String get saveProfile => 'Save Profile';

  @override
  String get optionalLeaveEmpty => '(Optional)';

  @override
  String get optionalExercisesPrefer => '(Optional)';

  @override
  String get optionalEquipmentPrefer => '(Optional)';

  @override
  String get optionalWorkoutTypesPrefer => '(Optional)';

  @override
  String get favoriteExercises => 'Favorite Exercises';

  @override
  String get favoriteEquipment => 'Favorite Equipment';

  @override
  String get favoriteWorkoutTypes => 'Preferred Workout Types';

  @override
  String get workoutTypeStrength => 'Strength';

  @override
  String get workoutTypeStrengthDescription =>
      'Build maximal force with heavy loads and full rest';

  @override
  String get workoutTypeSupersets => 'Supersets';

  @override
  String get workoutTypeSupersetsDescription =>
      'Pair opposing muscles back-to-back for time-efficient training';

  @override
  String get workoutTypeCircuit => 'Circuit';

  @override
  String get workoutTypeCircuitDescription =>
      'Move through stations with minimal rest for conditioning';

  @override
  String get workoutTypeEmom => 'EMOM';

  @override
  String get workoutTypeEmomDescription =>
      'Every Minute On the Minute: complete reps then rest until next minute';

  @override
  String get workoutTypeAmrap => 'AMRAP';

  @override
  String get workoutTypeAmrapDescription =>
      'As Many Rounds As Possible within a time cap';

  @override
  String get workoutTypeHiit => 'HIIT';

  @override
  String get workoutTypeHiitDescription =>
      'Alternate high-intensity bursts with short recovery';

  @override
  String get workoutTypeForTime => 'For Time';

  @override
  String get workoutTypeForTimeDescription =>
      'Complete the workout as fast as possible';

  @override
  String get workoutTypeEndurance => 'Endurance';

  @override
  String get workoutTypeEnduranceDescription =>
      'Sustained effort at moderate intensity for aerobic capacity';

  @override
  String get workoutTypeMobility => 'Mobility';

  @override
  String get workoutTypeMobilityDescription =>
      'Improve range of motion and joint health';

  @override
  String get methodologyOptional => 'Methodology (optional)';

  @override
  String get methodologyAuto => 'Auto';

  @override
  String get goalsOptional => 'Goals (optional)';

  @override
  String get musclesOptional => 'Muscles (optional)';

  @override
  String get musclesAuto => 'All';

  @override
  String get advancedSettings => 'Advanced';

  @override
  String get failedToUpdateProfile => 'Failed to update profile';

  @override
  String get activity => 'Activity';

  @override
  String get noTrainingsYet => 'No trainings yet';

  @override
  String get generateFirstTraining => 'Create your first training from Home';

  @override
  String get noTrainingAvailable => 'No training yet. Generate one.';

  @override
  String get availableTrainings => 'Available trainings';

  @override
  String get pastTrainings => 'Past trainings';

  @override
  String get stale => 'Stale';

  @override
  String get today => 'Today';

  @override
  String get yesterday => 'Yesterday';

  @override
  String get available => 'Available';

  @override
  String get completed => 'Completed';

  @override
  String get completedSingular => 'Complete';

  @override
  String get noPastTrainings => 'No completed trainings yet';

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
    return 'Delete \"$name\"? This cannot be undone.';
  }

  @override
  String leaveTrainingConfirmation(String name) {
    return 'Leave \"$name\"? You won\'t see it anymore.';
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
  String get trainingMarkedAsComplete => 'Training completed';

  @override
  String get failedToCompleteTraining => 'Failed to complete training';

  @override
  String get feedback => 'Feedback';

  @override
  String get feedbackUpdated => 'Feedback updated';

  @override
  String get failedToUpdateFeedback => 'Failed to update feedback';

  @override
  String get reportSubmitted => 'Report submitted';

  @override
  String get failedToSubmitReport => 'Failed to submit report';

  @override
  String get shuffleExercise => 'Shuffle exercise';

  @override
  String get exerciseShuffled => 'Exercise shuffled';

  @override
  String get failedToShuffleExercise => 'Failed to shuffle exercise';

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
  String get trainingRoutines => 'Training';

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
  String get updateFeedback => 'Update Feedback';

  @override
  String get references => 'References';

  @override
  String get literature => 'Literature';

  @override
  String get request => 'Request';

  @override
  String get describeIssue => 'Describe the issue...';

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
  String get tapWhenDone => 'Tap when done';

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
  String get whatWouldYouLikeToDo => 'What next?';

  @override
  String get exit => 'Exit';

  @override
  String get continueTraining => 'Continue';

  @override
  String get failedToMarkComplete => 'Failed to complete training';

  @override
  String get durationMinutes => 'Duration (minutes)';

  @override
  String get bodyweight => 'Bodyweight';

  @override
  String get gym => 'Gym';

  @override
  String get custom => 'Other';

  @override
  String get noEquipmentBodyweightOnly => 'Bodyweight only';

  @override
  String get noGymsDefinedCreateOne => 'No gyms. Create one in profile.';

  @override
  String get selectAGym => 'Select a gym';

  @override
  String get addEquipment => 'Add Equipment';

  @override
  String get addEquipmentAvailable => 'Add your available equipment';

  @override
  String get includeWarmupCooldown => 'Include warm-up & cooldown';

  @override
  String get equipmentPlaceholder => 'e.g., Barbell, Dumbbells';

  @override
  String get customPromptOptional => 'Custom Prompt (optional)';

  @override
  String get focusOnUpperBody => 'e.g., Focus on upper body';

  @override
  String get trainingPartnersOptional => 'Partner';

  @override
  String get generatingTraining => 'Generating your training...';

  @override
  String get thisMayTakeAMoment => 'This may take a moment';

  @override
  String generationFailedRetrying(int attempt) {
    return 'Generation failed, retrying #$attempt...';
  }

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
  String get availableWeights => 'Available Weights';

  @override
  String get availableWeightsHint =>
      'Configure weight options for weighted modifiers at this gym.';

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
  String get cues => 'Cues';

  @override
  String get howWasYourTraining => 'How was your training?';

  @override
  String get anyAdditionalComments => 'Any additional comments?';

  @override
  String get actualDuration => 'Actual duration (min)';

  @override
  String get impossible => 'Can\'t do';

  @override
  String get tooHard => 'Too hard';

  @override
  String get ok => 'OK';

  @override
  String get easy => 'Easy';

  @override
  String get tooEasy => 'Too easy';

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

  @override
  String nextExercise(String name) {
    return 'Next: $name';
  }

  @override
  String get rest => 'Rest';

  @override
  String get upcoming => 'Upcoming';

  @override
  String get yourProgress => 'Your Progress';

  @override
  String trainingsCompleted(int count) {
    return '$count trainings completed';
  }

  @override
  String get completedTrainings => 'completed trainings';

  @override
  String get partneredTrainings => 'partnered trainings';

  @override
  String get movementFamilies => 'Movement Families';

  @override
  String get muscleActivity => 'Muscle Activity';

  @override
  String get failedToLoadProgress => 'Failed to load progress';

  @override
  String get noProgressYet => 'Complete trainings to see your progress';

  @override
  String get calibration => 'Calibration';

  @override
  String get calibrationGlobal => 'Global';

  @override
  String get calibrationNeeded =>
      'Complete your first training for Vigor to calibrate recommendations to your level';

  @override
  String get calibrationDescription =>
      'During calibration, the platform collects data from your feedback to make an initial assessment of your fitness level';

  @override
  String get calibrationInProgress =>
      'Your trainings are getting smarter with each session';

  @override
  String get calibrationTrainingNote =>
      'This training may not fully match your goals — the system is still learning your fitness level and prioritizing movement variety to build a complete profile';

  @override
  String calibrationFamiliesLearned(int calibrated, int total) {
    return '$calibrated/$total movement patterns learned';
  }

  @override
  String get capabilities => 'Capabilities';

  @override
  String get muscleHeatMap => 'Muscle Heat Map';

  @override
  String get heatResting => 'Resting';

  @override
  String get heatRecovered => 'Recovered';

  @override
  String get heatActive => 'Active';

  @override
  String get heatWarm => 'Warm';

  @override
  String get heatHot => 'Hot';

  @override
  String get noTrainingsCompletedYet => 'Start training to see something here';

  @override
  String get theme => 'Theme';

  @override
  String get themeAuto => 'Auto';

  @override
  String get themeLight => 'Light';

  @override
  String get themeDark => 'Dark';

  @override
  String get themeAutoDescription => 'Follow system settings';

  @override
  String get appearance => 'Appearance';

  @override
  String get trainingDefaults => 'Defaults';

  @override
  String get defaultDuration => 'Training Duration';

  @override
  String get warmupCooldown => 'Warmup and cooldown';

  @override
  String get timer => 'Timer';

  @override
  String get intervalJingle => 'Whistle at interval change';

  @override
  String get duckOtherAudio => 'Lower other audio';

  @override
  String get duckOtherAudioDescription =>
      'Reduce volume of music and other apps during whistle';

  @override
  String get goalHypertrophy => 'Muscle Building';

  @override
  String get goalHypertrophyDescription =>
      'Build muscle size with targeted resistance training';

  @override
  String get goalFatLoss => 'Fat Loss';

  @override
  String get goalFatLossDescription =>
      'Burn calories and reduce body fat with high-energy workouts';

  @override
  String get goalToning => 'Toning';

  @override
  String get goalToningDescription =>
      'Develop lean muscle definition and a fit appearance';

  @override
  String get goalPosture => 'Posture';

  @override
  String get goalPostureDescription =>
      'Strengthen your back and core for better alignment';

  @override
  String get goalRehabilitation => 'Rehabilitation';

  @override
  String get goalRehabilitationDescription =>
      'Safe, controlled exercises for injury recovery';

  @override
  String get goalWellness => 'Wellness';

  @override
  String get goalWellnessDescription =>
      'Balanced workouts for overall health and stress relief';

  @override
  String get goalFlexibility => 'Flexibility';

  @override
  String get goalFlexibilityDescription =>
      'Improve range of motion with stretching and mobility';

  @override
  String get goalSports => 'Sports Performance';

  @override
  String get goalSportsDescription =>
      'Boost athletic ability with power and agility training';

  @override
  String get thisWeek => 'This Week';

  @override
  String get trainingPlan => 'Training Plan';

  @override
  String get sessionsPerWeek => 'Sessions per week';

  @override
  String get sessionDuration => 'Session duration';

  @override
  String get preferredTime => 'Preferred time';

  @override
  String get recommendedTime => 'Recommended time';

  @override
  String get methodologyMix => 'Methodology mix';

  @override
  String get pastWeeks => 'Past weeks';

  @override
  String get weeklyTarget => 'Weekly Target';

  @override
  String daysLeft(int count) {
    return '$count days left';
  }

  @override
  String get recommended => 'Recommended';

  @override
  String get duration => 'Duration';

  @override
  String get familyHorizontalPush => 'Push';

  @override
  String get familyHorizontalPull => 'Pull';

  @override
  String get familyVerticalPush => 'Overhead';

  @override
  String get familyVerticalPull => 'Pull-up';

  @override
  String get familySquat => 'Squat';

  @override
  String get familyHinge => 'Hinge';

  @override
  String get familyCore => 'Core';

  @override
  String get familyCarry => 'Carry';

  @override
  String get familyCardio => 'Cardio';

  @override
  String get familyMobility => 'Mobility';

  @override
  String get familyBalance => 'Balance';

  @override
  String get trainingQuality => 'How did you like this training?';

  @override
  String get trainingQualityHint => 'Helps us evaluate AI training quality';

  @override
  String get qualityReasonHint => 'What could be improved?';

  @override
  String get good => 'Good';

  @override
  String get bad => 'Bad';

  @override
  String get loadingMsg1 => 'Analyzing your profile...';

  @override
  String get loadingMsg2 => 'Selecting exercises...';

  @override
  String get loadingMsg3 => 'Building your routine...';

  @override
  String get loadingMsg4 => 'Calculating training volume...';

  @override
  String get loadingMsg5 => 'Optimizing rest intervals...';

  @override
  String get loadingMsg6 => 'Cross-referencing exercise research...';

  @override
  String get loadingMsg7 => 'Balancing muscle groups...';

  @override
  String get loadingMsg8 => 'Crafting progression path...';

  @override
  String get loadingMsg9 => 'Fine-tuning intensity...';

  @override
  String get loadingMsg10 => 'Reviewing movement patterns...';

  @override
  String get loadingMsg11 => 'Assessing recovery needs...';

  @override
  String get loadingMsg12 => 'Picking exercise variations...';

  @override
  String get loadingMsg13 => 'Structuring training blocks...';

  @override
  String get loadingMsg14 => 'Timing work intervals...';

  @override
  String get loadingMsg15 => 'Applying exercise science...';

  @override
  String get loadingMsg16 => 'Designing warm-up sequence...';

  @override
  String get loadingMsg17 => 'Mapping movement families...';

  @override
  String get loadingMsg18 => 'Evaluating load distribution...';

  @override
  String get loadingMsg19 => 'Personalizing your session...';

  @override
  String get loadingMsg20 => 'Almost ready...';

  @override
  String loadingMsgGoal(String goal) {
    return 'Optimizing for $goal...';
  }

  @override
  String get loadingMsgInjuries => 'Adapting around your injuries...';

  @override
  String get loadingMsgFavorites => 'Prioritizing your favorite exercises...';

  @override
  String get loadingMsgConditions => 'Adapting for your conditions...';

  @override
  String loadingMsgMethodology(String methodology) {
    return 'Designing a $methodology session...';
  }

  @override
  String get loadingMsgPartners => 'Coordinating partner workout...';

  @override
  String loadingMsgGym(String gym) {
    return 'Loading $gym equipment...';
  }

  @override
  String get loadingMsgHistory => 'Analyzing your recent sessions...';

  @override
  String get loadingRetryMsg1 => 'hmm, that didn\'t come out right — retrying';

  @override
  String get loadingRetryMsg2 => 'let me try that again...';

  @override
  String get loadingRetryMsg3 => 'not quite — giving it another shot';

  @override
  String get loadingRetryMsg4 => 'one more try, hang tight';

  @override
  String get loadingRetryMsg5 => 'oops, recalibrating...';

  @override
  String get loadingRetryMsg6 => 'almost had it — trying again';

  @override
  String nSelected(int count) {
    return '$count selected';
  }

  @override
  String deleteSelectedTrainings(int count) {
    return 'Delete $count trainings?';
  }

  @override
  String get trainingsDeletedSuccessfully => 'Trainings deleted successfully';

  @override
  String get shareByLink => 'Share Link';

  @override
  String get addToMyTrainings => 'Add to My Trainings';

  @override
  String get goToTraining => 'Go to Training';

  @override
  String sharedBy(String name) {
    return 'Shared by $name';
  }

  @override
  String get linkCopied => 'Link copied to clipboard';

  @override
  String get trainingNotFound => 'Training not found';

  @override
  String get trainingAddedSuccessfully => 'Training added successfully';

  @override
  String get failedToAddTraining => 'Failed to add training';

  @override
  String get loginToAdd => 'Log in to add this training';

  @override
  String get pendingFeedbacks => 'Pending feedbacks';

  @override
  String get pendingFeedbacksDescription =>
      'Some of your trainings are marked as complete but have no feedback from you. Feedback helps improve your future training recommendations.';
}
