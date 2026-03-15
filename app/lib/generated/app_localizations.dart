import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_de.dart';
import 'app_localizations_en.dart';
import 'app_localizations_es.dart';
import 'app_localizations_fr.dart';
import 'app_localizations_it.dart';
import 'app_localizations_ja.dart';
import 'app_localizations_ko.dart';
import 'app_localizations_pt.dart';
import 'app_localizations_ru.dart';
import 'app_localizations_zh.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'generated/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
    : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations)!;
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
        delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('de'),
    Locale('en'),
    Locale('es'),
    Locale('fr'),
    Locale('it'),
    Locale('ja'),
    Locale('ko'),
    Locale('pt'),
    Locale('ru'),
    Locale('zh'),
  ];

  /// No description provided for @appName.
  ///
  /// In en, this message translates to:
  /// **'Vigor'**
  String get appName;

  /// No description provided for @appTagline.
  ///
  /// In en, this message translates to:
  /// **'Ex Sapientia Vis'**
  String get appTagline;

  /// No description provided for @navHome.
  ///
  /// In en, this message translates to:
  /// **'Home'**
  String get navHome;

  /// No description provided for @navActivity.
  ///
  /// In en, this message translates to:
  /// **'Activity'**
  String get navActivity;

  /// No description provided for @navProfile.
  ///
  /// In en, this message translates to:
  /// **'Profile'**
  String get navProfile;

  /// No description provided for @storageErrorTitle.
  ///
  /// In en, this message translates to:
  /// **'Vigor - Storage Error'**
  String get storageErrorTitle;

  /// No description provided for @storageUnavailable.
  ///
  /// In en, this message translates to:
  /// **'Storage Unavailable'**
  String get storageUnavailable;

  /// No description provided for @storageErrorMessage.
  ///
  /// In en, this message translates to:
  /// **'Secure storage required. Check browser settings.'**
  String get storageErrorMessage;

  /// No description provided for @signInWithGoogle.
  ///
  /// In en, this message translates to:
  /// **'Sign in with Google'**
  String get signInWithGoogle;

  /// No description provided for @signingIn.
  ///
  /// In en, this message translates to:
  /// **'Signing in...'**
  String get signingIn;

  /// No description provided for @failedToInitializeGoogleSignIn.
  ///
  /// In en, this message translates to:
  /// **'Failed to initialize Google Sign In'**
  String get failedToInitializeGoogleSignIn;

  /// No description provided for @signInError.
  ///
  /// In en, this message translates to:
  /// **'Sign-in error: {message}'**
  String signInError(String message);

  /// No description provided for @googleSignInFailed.
  ///
  /// In en, this message translates to:
  /// **'Google sign-in failed'**
  String get googleSignInFailed;

  /// No description provided for @failedToGetAuthToken.
  ///
  /// In en, this message translates to:
  /// **'Failed to get authentication token'**
  String get failedToGetAuthToken;

  /// No description provided for @errorProcessingSignIn.
  ///
  /// In en, this message translates to:
  /// **'Error processing sign-in: {message}'**
  String errorProcessingSignIn(String message);

  /// No description provided for @googleSignInInitializing.
  ///
  /// In en, this message translates to:
  /// **'Google Sign In is still initializing...'**
  String get googleSignInInitializing;

  /// No description provided for @readyToTrain.
  ///
  /// In en, this message translates to:
  /// **'Ready to train?'**
  String get readyToTrain;

  /// No description provided for @generateTrainingDescription.
  ///
  /// In en, this message translates to:
  /// **'Create a workout tailored to your goals'**
  String get generateTrainingDescription;

  /// No description provided for @generateTraining.
  ///
  /// In en, this message translates to:
  /// **'Generate Training'**
  String get generateTraining;

  /// No description provided for @refresh.
  ///
  /// In en, this message translates to:
  /// **'Refresh'**
  String get refresh;

  /// No description provided for @logout.
  ///
  /// In en, this message translates to:
  /// **'Logout'**
  String get logout;

  /// No description provided for @userDataRefreshed.
  ///
  /// In en, this message translates to:
  /// **'User data refreshed'**
  String get userDataRefreshed;

  /// No description provided for @editProfile.
  ///
  /// In en, this message translates to:
  /// **'Edit Profile'**
  String get editProfile;

  /// No description provided for @settings.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get settings;

  /// No description provided for @other.
  ///
  /// In en, this message translates to:
  /// **'Other'**
  String get other;

  /// No description provided for @deleteGym.
  ///
  /// In en, this message translates to:
  /// **'Delete Gym'**
  String get deleteGym;

  /// No description provided for @deleteGymConfirmation.
  ///
  /// In en, this message translates to:
  /// **'Delete \"{name}\"?'**
  String deleteGymConfirmation(String name);

  /// No description provided for @cancel.
  ///
  /// In en, this message translates to:
  /// **'Cancel'**
  String get cancel;

  /// No description provided for @delete.
  ///
  /// In en, this message translates to:
  /// **'Delete'**
  String get delete;

  /// No description provided for @logoutConfirmation.
  ///
  /// In en, this message translates to:
  /// **'Logout?'**
  String get logoutConfirmation;

  /// No description provided for @deleteAccount.
  ///
  /// In en, this message translates to:
  /// **'Delete Account'**
  String get deleteAccount;

  /// No description provided for @deleteAccountConfirmation.
  ///
  /// In en, this message translates to:
  /// **'Delete your account? This cannot be undone.'**
  String get deleteAccountConfirmation;

  /// No description provided for @accountDeletedSuccessfully.
  ///
  /// In en, this message translates to:
  /// **'Account deleted successfully'**
  String get accountDeletedSuccessfully;

  /// No description provided for @failedToDeleteAccount.
  ///
  /// In en, this message translates to:
  /// **'Failed to delete account'**
  String get failedToDeleteAccount;

  /// No description provided for @failedToLoadGyms.
  ///
  /// In en, this message translates to:
  /// **'Failed to load gyms'**
  String get failedToLoadGyms;

  /// No description provided for @gymAddedSuccessfully.
  ///
  /// In en, this message translates to:
  /// **'Gym added successfully'**
  String get gymAddedSuccessfully;

  /// No description provided for @failedToAddGym.
  ///
  /// In en, this message translates to:
  /// **'Failed to add gym'**
  String get failedToAddGym;

  /// No description provided for @gymUpdatedSuccessfully.
  ///
  /// In en, this message translates to:
  /// **'Gym updated successfully'**
  String get gymUpdatedSuccessfully;

  /// No description provided for @failedToUpdateGym.
  ///
  /// In en, this message translates to:
  /// **'Failed to update gym'**
  String get failedToUpdateGym;

  /// No description provided for @gymDeletedSuccessfully.
  ///
  /// In en, this message translates to:
  /// **'Gym deleted successfully'**
  String get gymDeletedSuccessfully;

  /// No description provided for @failedToDeleteGym.
  ///
  /// In en, this message translates to:
  /// **'Failed to delete gym'**
  String get failedToDeleteGym;

  /// No description provided for @birthdate.
  ///
  /// In en, this message translates to:
  /// **'Birthdate'**
  String get birthdate;

  /// No description provided for @gender.
  ///
  /// In en, this message translates to:
  /// **'Gender'**
  String get gender;

  /// No description provided for @language.
  ///
  /// In en, this message translates to:
  /// **'Language'**
  String get language;

  /// No description provided for @height.
  ///
  /// In en, this message translates to:
  /// **'Height'**
  String get height;

  /// No description provided for @weight.
  ///
  /// In en, this message translates to:
  /// **'Weight'**
  String get weight;

  /// No description provided for @heightUnit.
  ///
  /// In en, this message translates to:
  /// **'cm'**
  String get heightUnit;

  /// No description provided for @weightUnit.
  ///
  /// In en, this message translates to:
  /// **'kg'**
  String get weightUnit;

  /// No description provided for @heightWithUnit.
  ///
  /// In en, this message translates to:
  /// **'{value} cm'**
  String heightWithUnit(double value);

  /// No description provided for @weightWithUnit.
  ///
  /// In en, this message translates to:
  /// **'{value} kg'**
  String weightWithUnit(double value);

  /// No description provided for @goals.
  ///
  /// In en, this message translates to:
  /// **'Goals'**
  String get goals;

  /// No description provided for @injuries.
  ///
  /// In en, this message translates to:
  /// **'Injuries'**
  String get injuries;

  /// No description provided for @limitations.
  ///
  /// In en, this message translates to:
  /// **'Limitations'**
  String get limitations;

  /// No description provided for @conditions.
  ///
  /// In en, this message translates to:
  /// **'Conditions'**
  String get conditions;

  /// No description provided for @favorites.
  ///
  /// In en, this message translates to:
  /// **'Favorites'**
  String get favorites;

  /// No description provided for @personalDetails.
  ///
  /// In en, this message translates to:
  /// **'Personal Details'**
  String get personalDetails;

  /// No description provided for @healthAndGoals.
  ///
  /// In en, this message translates to:
  /// **'Health & Goals'**
  String get healthAndGoals;

  /// No description provided for @exercises.
  ///
  /// In en, this message translates to:
  /// **'Exercises'**
  String get exercises;

  /// No description provided for @equipment.
  ///
  /// In en, this message translates to:
  /// **'Equipment'**
  String get equipment;

  /// No description provided for @startedDate.
  ///
  /// In en, this message translates to:
  /// **'Started: {date}'**
  String startedDate(String date);

  /// No description provided for @yearLabel.
  ///
  /// In en, this message translates to:
  /// **'Year: {year}'**
  String yearLabel(int year);

  /// No description provided for @myGyms.
  ///
  /// In en, this message translates to:
  /// **'Gyms'**
  String get myGyms;

  /// No description provided for @addGym.
  ///
  /// In en, this message translates to:
  /// **'Add Gym'**
  String get addGym;

  /// No description provided for @noGymsAddedYet.
  ///
  /// In en, this message translates to:
  /// **'No gyms added yet'**
  String get noGymsAddedYet;

  /// No description provided for @addYourFirstGym.
  ///
  /// In en, this message translates to:
  /// **'Add Your First Gym'**
  String get addYourFirstGym;

  /// No description provided for @removeDefault.
  ///
  /// In en, this message translates to:
  /// **'Remove Default'**
  String get removeDefault;

  /// No description provided for @setAsDefault.
  ///
  /// In en, this message translates to:
  /// **'Set as Default'**
  String get setAsDefault;

  /// No description provided for @edit.
  ///
  /// In en, this message translates to:
  /// **'Edit'**
  String get edit;

  /// No description provided for @quickActions.
  ///
  /// In en, this message translates to:
  /// **'Quick Actions'**
  String get quickActions;

  /// No description provided for @dangerZone.
  ///
  /// In en, this message translates to:
  /// **'Danger Zone'**
  String get dangerZone;

  /// No description provided for @completeYourProfile.
  ///
  /// In en, this message translates to:
  /// **'Complete Your Profile'**
  String get completeYourProfile;

  /// No description provided for @updateYourProfileInfo.
  ///
  /// In en, this message translates to:
  /// **'Update your info below.'**
  String get updateYourProfileInfo;

  /// No description provided for @pleaseCompleteProfile.
  ///
  /// In en, this message translates to:
  /// **'Complete your profile. * = required.'**
  String get pleaseCompleteProfile;

  /// No description provided for @firstName.
  ///
  /// In en, this message translates to:
  /// **'First Name'**
  String get firstName;

  /// No description provided for @lastName.
  ///
  /// In en, this message translates to:
  /// **'Last Name'**
  String get lastName;

  /// No description provided for @birthDate.
  ///
  /// In en, this message translates to:
  /// **'Birth Date'**
  String get birthDate;

  /// No description provided for @male.
  ///
  /// In en, this message translates to:
  /// **'Male'**
  String get male;

  /// No description provided for @female.
  ///
  /// In en, this message translates to:
  /// **'Female'**
  String get female;

  /// No description provided for @selectLanguage.
  ///
  /// In en, this message translates to:
  /// **'Select language'**
  String get selectLanguage;

  /// No description provided for @heightCm.
  ///
  /// In en, this message translates to:
  /// **'Height (cm)'**
  String get heightCm;

  /// No description provided for @weightKg.
  ///
  /// In en, this message translates to:
  /// **'Weight (kg)'**
  String get weightKg;

  /// No description provided for @required.
  ///
  /// In en, this message translates to:
  /// **'Required'**
  String get required;

  /// No description provided for @invalid.
  ///
  /// In en, this message translates to:
  /// **'Invalid'**
  String get invalid;

  /// No description provided for @pleaseSelectBirthDate.
  ///
  /// In en, this message translates to:
  /// **'Please select your birth date'**
  String get pleaseSelectBirthDate;

  /// No description provided for @pleaseAddAtLeastOneGoal.
  ///
  /// In en, this message translates to:
  /// **'Please add at least one goal'**
  String get pleaseAddAtLeastOneGoal;

  /// No description provided for @pleaseSelectLanguage.
  ///
  /// In en, this message translates to:
  /// **'Please select your language'**
  String get pleaseSelectLanguage;

  /// No description provided for @addAGoal.
  ///
  /// In en, this message translates to:
  /// **'Add a goal'**
  String get addAGoal;

  /// No description provided for @injuryDescription.
  ///
  /// In en, this message translates to:
  /// **'Injury description'**
  String get injuryDescription;

  /// No description provided for @year.
  ///
  /// In en, this message translates to:
  /// **'Year'**
  String get year;

  /// No description provided for @addALimitation.
  ///
  /// In en, this message translates to:
  /// **'Add a limitation'**
  String get addALimitation;

  /// No description provided for @addACondition.
  ///
  /// In en, this message translates to:
  /// **'Add a condition'**
  String get addACondition;

  /// No description provided for @favoriteExercisesHint.
  ///
  /// In en, this message translates to:
  /// **'e.g., squats, pull-ups, running'**
  String get favoriteExercisesHint;

  /// No description provided for @favoriteEquipmentHint.
  ///
  /// In en, this message translates to:
  /// **'e.g., dumbbells, barbell, kettlebell'**
  String get favoriteEquipmentHint;

  /// No description provided for @saveChanges.
  ///
  /// In en, this message translates to:
  /// **'Save Changes'**
  String get saveChanges;

  /// No description provided for @saveProfile.
  ///
  /// In en, this message translates to:
  /// **'Save Profile'**
  String get saveProfile;

  /// No description provided for @optionalLeaveEmpty.
  ///
  /// In en, this message translates to:
  /// **'(Optional)'**
  String get optionalLeaveEmpty;

  /// No description provided for @optionalExercisesPrefer.
  ///
  /// In en, this message translates to:
  /// **'(Optional)'**
  String get optionalExercisesPrefer;

  /// No description provided for @optionalEquipmentPrefer.
  ///
  /// In en, this message translates to:
  /// **'(Optional)'**
  String get optionalEquipmentPrefer;

  /// No description provided for @optionalWorkoutTypesPrefer.
  ///
  /// In en, this message translates to:
  /// **'(Optional)'**
  String get optionalWorkoutTypesPrefer;

  /// No description provided for @favoriteExercises.
  ///
  /// In en, this message translates to:
  /// **'Favorite Exercises'**
  String get favoriteExercises;

  /// No description provided for @favoriteEquipment.
  ///
  /// In en, this message translates to:
  /// **'Favorite Equipment'**
  String get favoriteEquipment;

  /// No description provided for @favoriteWorkoutTypes.
  ///
  /// In en, this message translates to:
  /// **'Preferred Workout Types'**
  String get favoriteWorkoutTypes;

  /// No description provided for @workoutTypeStrength.
  ///
  /// In en, this message translates to:
  /// **'Strength'**
  String get workoutTypeStrength;

  /// No description provided for @workoutTypeStrengthDescription.
  ///
  /// In en, this message translates to:
  /// **'Build maximal force with heavy loads and full rest'**
  String get workoutTypeStrengthDescription;

  /// No description provided for @workoutTypeSupersets.
  ///
  /// In en, this message translates to:
  /// **'Supersets'**
  String get workoutTypeSupersets;

  /// No description provided for @workoutTypeSupersetsDescription.
  ///
  /// In en, this message translates to:
  /// **'Pair opposing muscles back-to-back for time-efficient training'**
  String get workoutTypeSupersetsDescription;

  /// No description provided for @workoutTypeCircuit.
  ///
  /// In en, this message translates to:
  /// **'Circuit'**
  String get workoutTypeCircuit;

  /// No description provided for @workoutTypeCircuitDescription.
  ///
  /// In en, this message translates to:
  /// **'Move through stations with minimal rest for conditioning'**
  String get workoutTypeCircuitDescription;

  /// No description provided for @workoutTypeEmom.
  ///
  /// In en, this message translates to:
  /// **'EMOM'**
  String get workoutTypeEmom;

  /// No description provided for @workoutTypeEmomDescription.
  ///
  /// In en, this message translates to:
  /// **'Every Minute On the Minute: complete reps then rest until next minute'**
  String get workoutTypeEmomDescription;

  /// No description provided for @workoutTypeAmrap.
  ///
  /// In en, this message translates to:
  /// **'AMRAP'**
  String get workoutTypeAmrap;

  /// No description provided for @workoutTypeAmrapDescription.
  ///
  /// In en, this message translates to:
  /// **'As Many Rounds As Possible within a time cap'**
  String get workoutTypeAmrapDescription;

  /// No description provided for @workoutTypeHiit.
  ///
  /// In en, this message translates to:
  /// **'HIIT'**
  String get workoutTypeHiit;

  /// No description provided for @workoutTypeHiitDescription.
  ///
  /// In en, this message translates to:
  /// **'Alternate high-intensity bursts with short recovery'**
  String get workoutTypeHiitDescription;

  /// No description provided for @workoutTypeForTime.
  ///
  /// In en, this message translates to:
  /// **'For Time'**
  String get workoutTypeForTime;

  /// No description provided for @workoutTypeForTimeDescription.
  ///
  /// In en, this message translates to:
  /// **'Complete the workout as fast as possible'**
  String get workoutTypeForTimeDescription;

  /// No description provided for @workoutTypeEndurance.
  ///
  /// In en, this message translates to:
  /// **'Endurance'**
  String get workoutTypeEndurance;

  /// No description provided for @workoutTypeEnduranceDescription.
  ///
  /// In en, this message translates to:
  /// **'Sustained effort at moderate intensity for aerobic capacity'**
  String get workoutTypeEnduranceDescription;

  /// No description provided for @workoutTypeMobility.
  ///
  /// In en, this message translates to:
  /// **'Mobility'**
  String get workoutTypeMobility;

  /// No description provided for @workoutTypeMobilityDescription.
  ///
  /// In en, this message translates to:
  /// **'Improve range of motion and joint health'**
  String get workoutTypeMobilityDescription;

  /// No description provided for @methodologyOptional.
  ///
  /// In en, this message translates to:
  /// **'Methodology (optional)'**
  String get methodologyOptional;

  /// No description provided for @methodologyAuto.
  ///
  /// In en, this message translates to:
  /// **'Auto'**
  String get methodologyAuto;

  /// No description provided for @goalsOptional.
  ///
  /// In en, this message translates to:
  /// **'Goals (optional)'**
  String get goalsOptional;

  /// No description provided for @musclesOptional.
  ///
  /// In en, this message translates to:
  /// **'Muscles (optional)'**
  String get musclesOptional;

  /// No description provided for @musclesAuto.
  ///
  /// In en, this message translates to:
  /// **'All'**
  String get musclesAuto;

  /// No description provided for @advancedSettings.
  ///
  /// In en, this message translates to:
  /// **'Advanced'**
  String get advancedSettings;

  /// No description provided for @failedToUpdateProfile.
  ///
  /// In en, this message translates to:
  /// **'Failed to update profile'**
  String get failedToUpdateProfile;

  /// No description provided for @activity.
  ///
  /// In en, this message translates to:
  /// **'Activity'**
  String get activity;

  /// No description provided for @noTrainingsYet.
  ///
  /// In en, this message translates to:
  /// **'No trainings yet'**
  String get noTrainingsYet;

  /// No description provided for @generateFirstTraining.
  ///
  /// In en, this message translates to:
  /// **'Create your first training from Home'**
  String get generateFirstTraining;

  /// No description provided for @noTrainingAvailable.
  ///
  /// In en, this message translates to:
  /// **'No training yet. Generate one.'**
  String get noTrainingAvailable;

  /// No description provided for @availableTrainings.
  ///
  /// In en, this message translates to:
  /// **'Available trainings'**
  String get availableTrainings;

  /// No description provided for @pastTrainings.
  ///
  /// In en, this message translates to:
  /// **'Past trainings'**
  String get pastTrainings;

  /// No description provided for @stale.
  ///
  /// In en, this message translates to:
  /// **'Stale'**
  String get stale;

  /// No description provided for @today.
  ///
  /// In en, this message translates to:
  /// **'Today'**
  String get today;

  /// No description provided for @yesterday.
  ///
  /// In en, this message translates to:
  /// **'Yesterday'**
  String get yesterday;

  /// No description provided for @daysAgo.
  ///
  /// In en, this message translates to:
  /// **'{count}d ago'**
  String daysAgo(int count);

  /// No description provided for @available.
  ///
  /// In en, this message translates to:
  /// **'Available'**
  String get available;

  /// No description provided for @completed.
  ///
  /// In en, this message translates to:
  /// **'Completed'**
  String get completed;

  /// No description provided for @completedSingular.
  ///
  /// In en, this message translates to:
  /// **'Complete'**
  String get completedSingular;

  /// No description provided for @noPastTrainings.
  ///
  /// In en, this message translates to:
  /// **'No completed trainings yet'**
  String get noPastTrainings;

  /// No description provided for @copied.
  ///
  /// In en, this message translates to:
  /// **'Copied'**
  String get copied;

  /// No description provided for @durationMin.
  ///
  /// In en, this message translates to:
  /// **'{minutes} min'**
  String durationMin(int minutes);

  /// No description provided for @durationHr.
  ///
  /// In en, this message translates to:
  /// **'{hours} hr'**
  String durationHr(int hours);

  /// No description provided for @durationHrMin.
  ///
  /// In en, this message translates to:
  /// **'{hours} hr {minutes} min'**
  String durationHrMin(int hours, int minutes);

  /// No description provided for @failedToLoadTrainings.
  ///
  /// In en, this message translates to:
  /// **'Failed to load trainings'**
  String get failedToLoadTrainings;

  /// No description provided for @startTraining.
  ///
  /// In en, this message translates to:
  /// **'Start Training'**
  String get startTraining;

  /// No description provided for @cloneTraining.
  ///
  /// In en, this message translates to:
  /// **'Clone Training'**
  String get cloneTraining;

  /// No description provided for @addPartner.
  ///
  /// In en, this message translates to:
  /// **'Add Partner'**
  String get addPartner;

  /// No description provided for @shareWithUser.
  ///
  /// In en, this message translates to:
  /// **'Share with User'**
  String get shareWithUser;

  /// No description provided for @deleteTraining.
  ///
  /// In en, this message translates to:
  /// **'Delete Training'**
  String get deleteTraining;

  /// No description provided for @leaveTraining.
  ///
  /// In en, this message translates to:
  /// **'Leave Training'**
  String get leaveTraining;

  /// No description provided for @showAiReasoning.
  ///
  /// In en, this message translates to:
  /// **'Show AI reasoning'**
  String get showAiReasoning;

  /// No description provided for @reportIssue.
  ///
  /// In en, this message translates to:
  /// **'Report issue'**
  String get reportIssue;

  /// No description provided for @deleteTrainingConfirmation.
  ///
  /// In en, this message translates to:
  /// **'Delete \"{name}\"? This cannot be undone.'**
  String deleteTrainingConfirmation(String name);

  /// No description provided for @leaveTrainingConfirmation.
  ///
  /// In en, this message translates to:
  /// **'Leave \"{name}\"? You won\'t see it anymore.'**
  String leaveTrainingConfirmation(String name);

  /// No description provided for @addPartnerConfirmation.
  ///
  /// In en, this message translates to:
  /// **'Add {userName} as a partner to \"{trainingName}\"?'**
  String addPartnerConfirmation(String userName, String trainingName);

  /// No description provided for @cloneTrainingConfirmation.
  ///
  /// In en, this message translates to:
  /// **'Clone \"{name}\" to your trainings?'**
  String cloneTrainingConfirmation(String name);

  /// No description provided for @shareTrainingConfirmation.
  ///
  /// In en, this message translates to:
  /// **'Share \"{trainingName}\" with {userName}?'**
  String shareTrainingConfirmation(String trainingName, String userName);

  /// No description provided for @trainingDeletedSuccessfully.
  ///
  /// In en, this message translates to:
  /// **'Training deleted successfully'**
  String get trainingDeletedSuccessfully;

  /// No description provided for @failedToDeleteTraining.
  ///
  /// In en, this message translates to:
  /// **'Failed to delete training'**
  String get failedToDeleteTraining;

  /// No description provided for @leftTrainingSuccessfully.
  ///
  /// In en, this message translates to:
  /// **'Left training successfully'**
  String get leftTrainingSuccessfully;

  /// No description provided for @partnerAddedSuccessfully.
  ///
  /// In en, this message translates to:
  /// **'Partner added successfully'**
  String get partnerAddedSuccessfully;

  /// No description provided for @failedToAddPartner.
  ///
  /// In en, this message translates to:
  /// **'Failed to add partner'**
  String get failedToAddPartner;

  /// No description provided for @trainingSharedSuccessfully.
  ///
  /// In en, this message translates to:
  /// **'Training shared successfully'**
  String get trainingSharedSuccessfully;

  /// No description provided for @failedToShareTraining.
  ///
  /// In en, this message translates to:
  /// **'Failed to share training'**
  String get failedToShareTraining;

  /// No description provided for @trainingCloned.
  ///
  /// In en, this message translates to:
  /// **'Training cloned'**
  String get trainingCloned;

  /// No description provided for @failedToCloneTraining.
  ///
  /// In en, this message translates to:
  /// **'Failed to clone training'**
  String get failedToCloneTraining;

  /// No description provided for @trainingMarkedAsComplete.
  ///
  /// In en, this message translates to:
  /// **'Training completed'**
  String get trainingMarkedAsComplete;

  /// No description provided for @failedToCompleteTraining.
  ///
  /// In en, this message translates to:
  /// **'Failed to complete training'**
  String get failedToCompleteTraining;

  /// No description provided for @feedback.
  ///
  /// In en, this message translates to:
  /// **'Feedback'**
  String get feedback;

  /// No description provided for @feedbackUpdated.
  ///
  /// In en, this message translates to:
  /// **'Feedback updated'**
  String get feedbackUpdated;

  /// No description provided for @failedToUpdateFeedback.
  ///
  /// In en, this message translates to:
  /// **'Failed to update feedback'**
  String get failedToUpdateFeedback;

  /// No description provided for @reportSubmitted.
  ///
  /// In en, this message translates to:
  /// **'Report submitted'**
  String get reportSubmitted;

  /// No description provided for @failedToSubmitReport.
  ///
  /// In en, this message translates to:
  /// **'Failed to submit report'**
  String get failedToSubmitReport;

  /// No description provided for @shuffleExercise.
  ///
  /// In en, this message translates to:
  /// **'Shuffle exercise'**
  String get shuffleExercise;

  /// No description provided for @exerciseShuffled.
  ///
  /// In en, this message translates to:
  /// **'Exercise shuffled'**
  String get exerciseShuffled;

  /// No description provided for @failedToShuffleExercise.
  ///
  /// In en, this message translates to:
  /// **'Failed to shuffle exercise'**
  String get failedToShuffleExercise;

  /// No description provided for @reasoning.
  ///
  /// In en, this message translates to:
  /// **'Reasoning'**
  String get reasoning;

  /// No description provided for @strategy.
  ///
  /// In en, this message translates to:
  /// **'Strategy'**
  String get strategy;

  /// No description provided for @typeSelection.
  ///
  /// In en, this message translates to:
  /// **'Type Selection'**
  String get typeSelection;

  /// No description provided for @progression.
  ///
  /// In en, this message translates to:
  /// **'Progression'**
  String get progression;

  /// No description provided for @constraints.
  ///
  /// In en, this message translates to:
  /// **'Constraints'**
  String get constraints;

  /// No description provided for @researchApplied.
  ///
  /// In en, this message translates to:
  /// **'Research Applied'**
  String get researchApplied;

  /// No description provided for @targetMuscles.
  ///
  /// In en, this message translates to:
  /// **'Target Muscles'**
  String get targetMuscles;

  /// No description provided for @naming.
  ///
  /// In en, this message translates to:
  /// **'Naming'**
  String get naming;

  /// No description provided for @trainingRoutines.
  ///
  /// In en, this message translates to:
  /// **'Training'**
  String get trainingRoutines;

  /// No description provided for @noEquipment.
  ///
  /// In en, this message translates to:
  /// **'No equipment'**
  String get noEquipment;

  /// No description provided for @blockNumber.
  ///
  /// In en, this message translates to:
  /// **'Block {number}'**
  String blockNumber(int number);

  /// No description provided for @repeatsCount.
  ///
  /// In en, this message translates to:
  /// **'{count}x'**
  String repeatsCount(int count);

  /// No description provided for @durationSeconds.
  ///
  /// In en, this message translates to:
  /// **'{seconds}s'**
  String durationSeconds(int seconds);

  /// No description provided for @restSeconds.
  ///
  /// In en, this message translates to:
  /// **'{seconds}s rest'**
  String restSeconds(int seconds);

  /// No description provided for @repsCount.
  ///
  /// In en, this message translates to:
  /// **'{count} reps'**
  String repsCount(int count);

  /// No description provided for @weightKgValue.
  ///
  /// In en, this message translates to:
  /// **'{value} kg'**
  String weightKgValue(double value);

  /// No description provided for @markAsComplete.
  ///
  /// In en, this message translates to:
  /// **'Mark as Complete'**
  String get markAsComplete;

  /// No description provided for @updateFeedback.
  ///
  /// In en, this message translates to:
  /// **'Update Feedback'**
  String get updateFeedback;

  /// No description provided for @references.
  ///
  /// In en, this message translates to:
  /// **'References'**
  String get references;

  /// No description provided for @literature.
  ///
  /// In en, this message translates to:
  /// **'Literature'**
  String get literature;

  /// No description provided for @request.
  ///
  /// In en, this message translates to:
  /// **'Request'**
  String get request;

  /// No description provided for @describeIssue.
  ///
  /// In en, this message translates to:
  /// **'Describe the issue...'**
  String get describeIssue;

  /// No description provided for @submit.
  ///
  /// In en, this message translates to:
  /// **'Submit'**
  String get submit;

  /// No description provided for @close.
  ///
  /// In en, this message translates to:
  /// **'Close'**
  String get close;

  /// No description provided for @add.
  ///
  /// In en, this message translates to:
  /// **'Add'**
  String get add;

  /// No description provided for @update.
  ///
  /// In en, this message translates to:
  /// **'Update'**
  String get update;

  /// No description provided for @clone.
  ///
  /// In en, this message translates to:
  /// **'Clone'**
  String get clone;

  /// No description provided for @share.
  ///
  /// In en, this message translates to:
  /// **'Share'**
  String get share;

  /// No description provided for @leave.
  ///
  /// In en, this message translates to:
  /// **'Leave'**
  String get leave;

  /// No description provided for @tapToStart.
  ///
  /// In en, this message translates to:
  /// **'Tap to start'**
  String get tapToStart;

  /// No description provided for @tapWhenDone.
  ///
  /// In en, this message translates to:
  /// **'Tap when done'**
  String get tapWhenDone;

  /// No description provided for @trainingCompleted.
  ///
  /// In en, this message translates to:
  /// **'Training Completed!'**
  String get trainingCompleted;

  /// No description provided for @greatJobCompleting.
  ///
  /// In en, this message translates to:
  /// **'Great job completing {name}'**
  String greatJobCompleting(String name);

  /// No description provided for @done.
  ///
  /// In en, this message translates to:
  /// **'Done'**
  String get done;

  /// No description provided for @complete.
  ///
  /// In en, this message translates to:
  /// **'Complete'**
  String get complete;

  /// No description provided for @routineCounter.
  ///
  /// In en, this message translates to:
  /// **'Routine {current}/{total}'**
  String routineCounter(int current, int total);

  /// No description provided for @blockCounter.
  ///
  /// In en, this message translates to:
  /// **'Block {current}/{total}'**
  String blockCounter(int current, int total);

  /// No description provided for @exitTraining.
  ///
  /// In en, this message translates to:
  /// **'Exit Training?'**
  String get exitTraining;

  /// No description provided for @whatWouldYouLikeToDo.
  ///
  /// In en, this message translates to:
  /// **'What next?'**
  String get whatWouldYouLikeToDo;

  /// No description provided for @exit.
  ///
  /// In en, this message translates to:
  /// **'Exit'**
  String get exit;

  /// No description provided for @continueTraining.
  ///
  /// In en, this message translates to:
  /// **'Continue'**
  String get continueTraining;

  /// No description provided for @stop.
  ///
  /// In en, this message translates to:
  /// **'Stop'**
  String get stop;

  /// No description provided for @stopTraining.
  ///
  /// In en, this message translates to:
  /// **'Stop Training?'**
  String get stopTraining;

  /// No description provided for @stopTrainingConfirm.
  ///
  /// In en, this message translates to:
  /// **'Timer progress will be lost.'**
  String get stopTrainingConfirm;

  /// No description provided for @failedToMarkComplete.
  ///
  /// In en, this message translates to:
  /// **'Failed to complete training'**
  String get failedToMarkComplete;

  /// No description provided for @durationMinutes.
  ///
  /// In en, this message translates to:
  /// **'Duration (minutes)'**
  String get durationMinutes;

  /// No description provided for @bodyweight.
  ///
  /// In en, this message translates to:
  /// **'Bodyweight'**
  String get bodyweight;

  /// No description provided for @gym.
  ///
  /// In en, this message translates to:
  /// **'Gym'**
  String get gym;

  /// No description provided for @custom.
  ///
  /// In en, this message translates to:
  /// **'Other'**
  String get custom;

  /// No description provided for @noEquipmentBodyweightOnly.
  ///
  /// In en, this message translates to:
  /// **'Bodyweight only'**
  String get noEquipmentBodyweightOnly;

  /// No description provided for @noGymsDefinedCreateOne.
  ///
  /// In en, this message translates to:
  /// **'No gyms. Create one in profile.'**
  String get noGymsDefinedCreateOne;

  /// No description provided for @selectAGym.
  ///
  /// In en, this message translates to:
  /// **'Select a gym'**
  String get selectAGym;

  /// No description provided for @addEquipment.
  ///
  /// In en, this message translates to:
  /// **'Add Equipment'**
  String get addEquipment;

  /// No description provided for @addEquipmentAvailable.
  ///
  /// In en, this message translates to:
  /// **'Add your available equipment'**
  String get addEquipmentAvailable;

  /// No description provided for @includeWarmupCooldown.
  ///
  /// In en, this message translates to:
  /// **'Include warm-up & cooldown'**
  String get includeWarmupCooldown;

  /// No description provided for @equipmentPlaceholder.
  ///
  /// In en, this message translates to:
  /// **'e.g., Barbell, Dumbbells'**
  String get equipmentPlaceholder;

  /// No description provided for @customPromptOptional.
  ///
  /// In en, this message translates to:
  /// **'Custom Prompt (optional)'**
  String get customPromptOptional;

  /// No description provided for @focusOnUpperBody.
  ///
  /// In en, this message translates to:
  /// **'e.g., Focus on upper body'**
  String get focusOnUpperBody;

  /// No description provided for @trainingPartnersOptional.
  ///
  /// In en, this message translates to:
  /// **'Partner'**
  String get trainingPartnersOptional;

  /// No description provided for @generatingTraining.
  ///
  /// In en, this message translates to:
  /// **'Generating your training...'**
  String get generatingTraining;

  /// No description provided for @thisMayTakeAMoment.
  ///
  /// In en, this message translates to:
  /// **'This may take a moment'**
  String get thisMayTakeAMoment;

  /// No description provided for @generationFailedRetrying.
  ///
  /// In en, this message translates to:
  /// **'Generation failed, retrying #{attempt}...'**
  String generationFailedRetrying(int attempt);

  /// No description provided for @trainingGeneratedSuccessfully.
  ///
  /// In en, this message translates to:
  /// **'Training generated successfully!'**
  String get trainingGeneratedSuccessfully;

  /// No description provided for @failedToGenerateTraining.
  ///
  /// In en, this message translates to:
  /// **'Failed to generate training'**
  String get failedToGenerateTraining;

  /// No description provided for @generate.
  ///
  /// In en, this message translates to:
  /// **'Generate'**
  String get generate;

  /// No description provided for @editGym.
  ///
  /// In en, this message translates to:
  /// **'Edit Gym'**
  String get editGym;

  /// No description provided for @gymName.
  ///
  /// In en, this message translates to:
  /// **'Gym Name'**
  String get gymName;

  /// No description provided for @gymNamePlaceholder.
  ///
  /// In en, this message translates to:
  /// **'e.g., Home Gym, LA Fitness'**
  String get gymNamePlaceholder;

  /// No description provided for @availableWeights.
  ///
  /// In en, this message translates to:
  /// **'Available Weights'**
  String get availableWeights;

  /// No description provided for @availableWeightsHint.
  ///
  /// In en, this message translates to:
  /// **'Configure weight options for weighted modifiers at this gym.'**
  String get availableWeightsHint;

  /// No description provided for @noEquipmentAddedYet.
  ///
  /// In en, this message translates to:
  /// **'No equipment added yet'**
  String get noEquipmentAddedYet;

  /// No description provided for @pleaseEnterGymName.
  ///
  /// In en, this message translates to:
  /// **'Please enter a gym name'**
  String get pleaseEnterGymName;

  /// No description provided for @addAllEquipment.
  ///
  /// In en, this message translates to:
  /// **'Add All'**
  String get addAllEquipment;

  /// No description provided for @failedToLoadEquipment.
  ///
  /// In en, this message translates to:
  /// **'Failed to load equipment'**
  String get failedToLoadEquipment;

  /// No description provided for @selectUser.
  ///
  /// In en, this message translates to:
  /// **'Select User'**
  String get selectUser;

  /// No description provided for @searchByName.
  ///
  /// In en, this message translates to:
  /// **'Search by name'**
  String get searchByName;

  /// No description provided for @noUsersAvailable.
  ///
  /// In en, this message translates to:
  /// **'No users available'**
  String get noUsersAvailable;

  /// No description provided for @noMatchingUsers.
  ///
  /// In en, this message translates to:
  /// **'No matching users'**
  String get noMatchingUsers;

  /// No description provided for @noMatchingEquipment.
  ///
  /// In en, this message translates to:
  /// **'No matching equipment'**
  String get noMatchingEquipment;

  /// No description provided for @instructions.
  ///
  /// In en, this message translates to:
  /// **'Instructions'**
  String get instructions;

  /// No description provided for @cues.
  ///
  /// In en, this message translates to:
  /// **'Cues'**
  String get cues;

  /// No description provided for @howWasYourTraining.
  ///
  /// In en, this message translates to:
  /// **'How was your training?'**
  String get howWasYourTraining;

  /// No description provided for @anyAdditionalComments.
  ///
  /// In en, this message translates to:
  /// **'Any additional comments?'**
  String get anyAdditionalComments;

  /// No description provided for @actualDuration.
  ///
  /// In en, this message translates to:
  /// **'Actual duration (min)'**
  String get actualDuration;

  /// No description provided for @impossible.
  ///
  /// In en, this message translates to:
  /// **'Can\'t do'**
  String get impossible;

  /// No description provided for @tooHard.
  ///
  /// In en, this message translates to:
  /// **'Too hard'**
  String get tooHard;

  /// No description provided for @ok.
  ///
  /// In en, this message translates to:
  /// **'OK'**
  String get ok;

  /// No description provided for @easy.
  ///
  /// In en, this message translates to:
  /// **'Easy'**
  String get easy;

  /// No description provided for @tooEasy.
  ///
  /// In en, this message translates to:
  /// **'Too easy'**
  String get tooEasy;

  /// No description provided for @flag.
  ///
  /// In en, this message translates to:
  /// **'Flag'**
  String get flag;

  /// No description provided for @profile.
  ///
  /// In en, this message translates to:
  /// **'Profile'**
  String get profile;

  /// No description provided for @languageEnglish.
  ///
  /// In en, this message translates to:
  /// **'English'**
  String get languageEnglish;

  /// No description provided for @languageItaliano.
  ///
  /// In en, this message translates to:
  /// **'Italiano'**
  String get languageItaliano;

  /// No description provided for @languageEspanol.
  ///
  /// In en, this message translates to:
  /// **'Español'**
  String get languageEspanol;

  /// No description provided for @languageFrancais.
  ///
  /// In en, this message translates to:
  /// **'Français'**
  String get languageFrancais;

  /// No description provided for @languageDeutsch.
  ///
  /// In en, this message translates to:
  /// **'Deutsch'**
  String get languageDeutsch;

  /// No description provided for @languagePortugues.
  ///
  /// In en, this message translates to:
  /// **'Português'**
  String get languagePortugues;

  /// No description provided for @languageRussian.
  ///
  /// In en, this message translates to:
  /// **'Русский'**
  String get languageRussian;

  /// No description provided for @languageChinese.
  ///
  /// In en, this message translates to:
  /// **'中文'**
  String get languageChinese;

  /// No description provided for @languageJapanese.
  ///
  /// In en, this message translates to:
  /// **'日本語'**
  String get languageJapanese;

  /// No description provided for @languageKorean.
  ///
  /// In en, this message translates to:
  /// **'한국어'**
  String get languageKorean;

  /// No description provided for @nextExercise.
  ///
  /// In en, this message translates to:
  /// **'Next: {name}'**
  String nextExercise(String name);

  /// No description provided for @rest.
  ///
  /// In en, this message translates to:
  /// **'Rest'**
  String get rest;

  /// No description provided for @upcoming.
  ///
  /// In en, this message translates to:
  /// **'Upcoming'**
  String get upcoming;

  /// No description provided for @yourProgress.
  ///
  /// In en, this message translates to:
  /// **'Your Progress'**
  String get yourProgress;

  /// No description provided for @trainingsCompleted.
  ///
  /// In en, this message translates to:
  /// **'{count} trainings completed'**
  String trainingsCompleted(int count);

  /// No description provided for @completedTrainings.
  ///
  /// In en, this message translates to:
  /// **'completed trainings'**
  String get completedTrainings;

  /// No description provided for @partneredTrainings.
  ///
  /// In en, this message translates to:
  /// **'partnered trainings'**
  String get partneredTrainings;

  /// No description provided for @movementFamilies.
  ///
  /// In en, this message translates to:
  /// **'Movement Families'**
  String get movementFamilies;

  /// No description provided for @muscleActivity.
  ///
  /// In en, this message translates to:
  /// **'Muscle Activity'**
  String get muscleActivity;

  /// No description provided for @failedToLoadProgress.
  ///
  /// In en, this message translates to:
  /// **'Failed to load progress'**
  String get failedToLoadProgress;

  /// No description provided for @noProgressYet.
  ///
  /// In en, this message translates to:
  /// **'Complete trainings to see your progress'**
  String get noProgressYet;

  /// No description provided for @calibration.
  ///
  /// In en, this message translates to:
  /// **'Calibration'**
  String get calibration;

  /// No description provided for @calibrationGlobal.
  ///
  /// In en, this message translates to:
  /// **'Global'**
  String get calibrationGlobal;

  /// No description provided for @calibrationNeeded.
  ///
  /// In en, this message translates to:
  /// **'Complete your first training for Vigor to calibrate recommendations to your level'**
  String get calibrationNeeded;

  /// No description provided for @calibrationDescription.
  ///
  /// In en, this message translates to:
  /// **'During calibration, the platform collects data from your feedback to make an initial assessment of your fitness level'**
  String get calibrationDescription;

  /// No description provided for @calibrationInProgress.
  ///
  /// In en, this message translates to:
  /// **'Your trainings are getting smarter with each session'**
  String get calibrationInProgress;

  /// No description provided for @calibrationTrainingNote.
  ///
  /// In en, this message translates to:
  /// **'This training may not fully match your goals — the system is still learning your fitness level and prioritizing movement variety to build a complete profile'**
  String get calibrationTrainingNote;

  /// No description provided for @calibrationFamiliesLearned.
  ///
  /// In en, this message translates to:
  /// **'{calibrated}/{total} movement patterns learned'**
  String calibrationFamiliesLearned(int calibrated, int total);

  /// No description provided for @capabilities.
  ///
  /// In en, this message translates to:
  /// **'Capabilities'**
  String get capabilities;

  /// No description provided for @muscleHeatMap.
  ///
  /// In en, this message translates to:
  /// **'Muscle Heat Map'**
  String get muscleHeatMap;

  /// No description provided for @heatResting.
  ///
  /// In en, this message translates to:
  /// **'Resting'**
  String get heatResting;

  /// No description provided for @heatRecovered.
  ///
  /// In en, this message translates to:
  /// **'Recovered'**
  String get heatRecovered;

  /// No description provided for @heatActive.
  ///
  /// In en, this message translates to:
  /// **'Active'**
  String get heatActive;

  /// No description provided for @heatWarm.
  ///
  /// In en, this message translates to:
  /// **'Warm'**
  String get heatWarm;

  /// No description provided for @heatHot.
  ///
  /// In en, this message translates to:
  /// **'Hot'**
  String get heatHot;

  /// No description provided for @noTrainingsCompletedYet.
  ///
  /// In en, this message translates to:
  /// **'Start training to see something here'**
  String get noTrainingsCompletedYet;

  /// No description provided for @theme.
  ///
  /// In en, this message translates to:
  /// **'Theme'**
  String get theme;

  /// No description provided for @themeAuto.
  ///
  /// In en, this message translates to:
  /// **'Auto'**
  String get themeAuto;

  /// No description provided for @themeLight.
  ///
  /// In en, this message translates to:
  /// **'Light'**
  String get themeLight;

  /// No description provided for @themeDark.
  ///
  /// In en, this message translates to:
  /// **'Dark'**
  String get themeDark;

  /// No description provided for @themeAutoDescription.
  ///
  /// In en, this message translates to:
  /// **'Follow system settings'**
  String get themeAutoDescription;

  /// No description provided for @appearance.
  ///
  /// In en, this message translates to:
  /// **'Appearance'**
  String get appearance;

  /// No description provided for @trainingDefaults.
  ///
  /// In en, this message translates to:
  /// **'Defaults'**
  String get trainingDefaults;

  /// No description provided for @defaultDuration.
  ///
  /// In en, this message translates to:
  /// **'Training Duration'**
  String get defaultDuration;

  /// No description provided for @warmupCooldown.
  ///
  /// In en, this message translates to:
  /// **'Warmup and cooldown'**
  String get warmupCooldown;

  /// No description provided for @timer.
  ///
  /// In en, this message translates to:
  /// **'Timer'**
  String get timer;

  /// No description provided for @intervalJingle.
  ///
  /// In en, this message translates to:
  /// **'Whistle at interval change'**
  String get intervalJingle;

  /// No description provided for @duckOtherAudio.
  ///
  /// In en, this message translates to:
  /// **'Lower other audio'**
  String get duckOtherAudio;

  /// No description provided for @duckOtherAudioDescription.
  ///
  /// In en, this message translates to:
  /// **'Reduce volume of music and other apps during whistle'**
  String get duckOtherAudioDescription;

  /// No description provided for @goalHypertrophy.
  ///
  /// In en, this message translates to:
  /// **'Muscle Building'**
  String get goalHypertrophy;

  /// No description provided for @goalHypertrophyDescription.
  ///
  /// In en, this message translates to:
  /// **'Build muscle size with targeted resistance training'**
  String get goalHypertrophyDescription;

  /// No description provided for @goalFatLoss.
  ///
  /// In en, this message translates to:
  /// **'Fat Loss'**
  String get goalFatLoss;

  /// No description provided for @goalFatLossDescription.
  ///
  /// In en, this message translates to:
  /// **'Burn calories and reduce body fat with high-energy workouts'**
  String get goalFatLossDescription;

  /// No description provided for @goalToning.
  ///
  /// In en, this message translates to:
  /// **'Toning'**
  String get goalToning;

  /// No description provided for @goalToningDescription.
  ///
  /// In en, this message translates to:
  /// **'Develop lean muscle definition and a fit appearance'**
  String get goalToningDescription;

  /// No description provided for @goalPosture.
  ///
  /// In en, this message translates to:
  /// **'Posture'**
  String get goalPosture;

  /// No description provided for @goalPostureDescription.
  ///
  /// In en, this message translates to:
  /// **'Strengthen your back and core for better alignment'**
  String get goalPostureDescription;

  /// No description provided for @goalRehabilitation.
  ///
  /// In en, this message translates to:
  /// **'Rehabilitation'**
  String get goalRehabilitation;

  /// No description provided for @goalRehabilitationDescription.
  ///
  /// In en, this message translates to:
  /// **'Safe, controlled exercises for injury recovery'**
  String get goalRehabilitationDescription;

  /// No description provided for @goalWellness.
  ///
  /// In en, this message translates to:
  /// **'Wellness'**
  String get goalWellness;

  /// No description provided for @goalWellnessDescription.
  ///
  /// In en, this message translates to:
  /// **'Balanced workouts for overall health and stress relief'**
  String get goalWellnessDescription;

  /// No description provided for @goalFlexibility.
  ///
  /// In en, this message translates to:
  /// **'Flexibility'**
  String get goalFlexibility;

  /// No description provided for @goalFlexibilityDescription.
  ///
  /// In en, this message translates to:
  /// **'Improve range of motion with stretching and mobility'**
  String get goalFlexibilityDescription;

  /// No description provided for @goalSports.
  ///
  /// In en, this message translates to:
  /// **'Sports Performance'**
  String get goalSports;

  /// No description provided for @goalSportsDescription.
  ///
  /// In en, this message translates to:
  /// **'Boost athletic ability with power and agility training'**
  String get goalSportsDescription;

  /// No description provided for @thisWeek.
  ///
  /// In en, this message translates to:
  /// **'This Week'**
  String get thisWeek;

  /// No description provided for @trainingPlan.
  ///
  /// In en, this message translates to:
  /// **'Training Plan'**
  String get trainingPlan;

  /// No description provided for @sessionsPerWeek.
  ///
  /// In en, this message translates to:
  /// **'Sessions per week'**
  String get sessionsPerWeek;

  /// No description provided for @sessionDuration.
  ///
  /// In en, this message translates to:
  /// **'Session duration'**
  String get sessionDuration;

  /// No description provided for @preferredTime.
  ///
  /// In en, this message translates to:
  /// **'Preferred time'**
  String get preferredTime;

  /// No description provided for @recommendedTime.
  ///
  /// In en, this message translates to:
  /// **'Recommended time'**
  String get recommendedTime;

  /// No description provided for @methodologyMix.
  ///
  /// In en, this message translates to:
  /// **'Methodology mix'**
  String get methodologyMix;

  /// No description provided for @pastWeeks.
  ///
  /// In en, this message translates to:
  /// **'Past weeks'**
  String get pastWeeks;

  /// No description provided for @weeklyTarget.
  ///
  /// In en, this message translates to:
  /// **'Weekly Target'**
  String get weeklyTarget;

  /// No description provided for @daysLeft.
  ///
  /// In en, this message translates to:
  /// **'{count} days left'**
  String daysLeft(int count);

  /// No description provided for @recommended.
  ///
  /// In en, this message translates to:
  /// **'Recommended'**
  String get recommended;

  /// No description provided for @duration.
  ///
  /// In en, this message translates to:
  /// **'Duration'**
  String get duration;

  /// No description provided for @familyHorizontalPush.
  ///
  /// In en, this message translates to:
  /// **'Push'**
  String get familyHorizontalPush;

  /// No description provided for @familyHorizontalPull.
  ///
  /// In en, this message translates to:
  /// **'Pull'**
  String get familyHorizontalPull;

  /// No description provided for @familyVerticalPush.
  ///
  /// In en, this message translates to:
  /// **'Overhead'**
  String get familyVerticalPush;

  /// No description provided for @familyVerticalPull.
  ///
  /// In en, this message translates to:
  /// **'Pull-up'**
  String get familyVerticalPull;

  /// No description provided for @familySquat.
  ///
  /// In en, this message translates to:
  /// **'Squat'**
  String get familySquat;

  /// No description provided for @familyHinge.
  ///
  /// In en, this message translates to:
  /// **'Hinge'**
  String get familyHinge;

  /// No description provided for @familyCore.
  ///
  /// In en, this message translates to:
  /// **'Core'**
  String get familyCore;

  /// No description provided for @familyCarry.
  ///
  /// In en, this message translates to:
  /// **'Carry'**
  String get familyCarry;

  /// No description provided for @familyCardio.
  ///
  /// In en, this message translates to:
  /// **'Cardio'**
  String get familyCardio;

  /// No description provided for @familyMobility.
  ///
  /// In en, this message translates to:
  /// **'Mobility'**
  String get familyMobility;

  /// No description provided for @familyBalance.
  ///
  /// In en, this message translates to:
  /// **'Balance'**
  String get familyBalance;

  /// No description provided for @trainingQuality.
  ///
  /// In en, this message translates to:
  /// **'How did you like this training?'**
  String get trainingQuality;

  /// No description provided for @trainingQualityHint.
  ///
  /// In en, this message translates to:
  /// **'Helps us evaluate AI training quality'**
  String get trainingQualityHint;

  /// No description provided for @qualityReasonHint.
  ///
  /// In en, this message translates to:
  /// **'What could be improved?'**
  String get qualityReasonHint;

  /// No description provided for @good.
  ///
  /// In en, this message translates to:
  /// **'Good'**
  String get good;

  /// No description provided for @bad.
  ///
  /// In en, this message translates to:
  /// **'Bad'**
  String get bad;

  /// No description provided for @loadingMsg1.
  ///
  /// In en, this message translates to:
  /// **'Analyzing your profile...'**
  String get loadingMsg1;

  /// No description provided for @loadingMsg2.
  ///
  /// In en, this message translates to:
  /// **'Selecting exercises...'**
  String get loadingMsg2;

  /// No description provided for @loadingMsg3.
  ///
  /// In en, this message translates to:
  /// **'Building your routine...'**
  String get loadingMsg3;

  /// No description provided for @loadingMsg4.
  ///
  /// In en, this message translates to:
  /// **'Calculating training volume...'**
  String get loadingMsg4;

  /// No description provided for @loadingMsg5.
  ///
  /// In en, this message translates to:
  /// **'Optimizing rest intervals...'**
  String get loadingMsg5;

  /// No description provided for @loadingMsg6.
  ///
  /// In en, this message translates to:
  /// **'Cross-referencing exercise research...'**
  String get loadingMsg6;

  /// No description provided for @loadingMsg7.
  ///
  /// In en, this message translates to:
  /// **'Balancing muscle groups...'**
  String get loadingMsg7;

  /// No description provided for @loadingMsg8.
  ///
  /// In en, this message translates to:
  /// **'Crafting progression path...'**
  String get loadingMsg8;

  /// No description provided for @loadingMsg9.
  ///
  /// In en, this message translates to:
  /// **'Fine-tuning intensity...'**
  String get loadingMsg9;

  /// No description provided for @loadingMsg10.
  ///
  /// In en, this message translates to:
  /// **'Reviewing movement patterns...'**
  String get loadingMsg10;

  /// No description provided for @loadingMsg11.
  ///
  /// In en, this message translates to:
  /// **'Assessing recovery needs...'**
  String get loadingMsg11;

  /// No description provided for @loadingMsg12.
  ///
  /// In en, this message translates to:
  /// **'Picking exercise variations...'**
  String get loadingMsg12;

  /// No description provided for @loadingMsg13.
  ///
  /// In en, this message translates to:
  /// **'Structuring training blocks...'**
  String get loadingMsg13;

  /// No description provided for @loadingMsg14.
  ///
  /// In en, this message translates to:
  /// **'Timing work intervals...'**
  String get loadingMsg14;

  /// No description provided for @loadingMsg15.
  ///
  /// In en, this message translates to:
  /// **'Applying exercise science...'**
  String get loadingMsg15;

  /// No description provided for @loadingMsg16.
  ///
  /// In en, this message translates to:
  /// **'Designing warm-up sequence...'**
  String get loadingMsg16;

  /// No description provided for @loadingMsg17.
  ///
  /// In en, this message translates to:
  /// **'Mapping movement families...'**
  String get loadingMsg17;

  /// No description provided for @loadingMsg18.
  ///
  /// In en, this message translates to:
  /// **'Evaluating load distribution...'**
  String get loadingMsg18;

  /// No description provided for @loadingMsg19.
  ///
  /// In en, this message translates to:
  /// **'Personalizing your session...'**
  String get loadingMsg19;

  /// No description provided for @loadingMsg20.
  ///
  /// In en, this message translates to:
  /// **'Almost ready...'**
  String get loadingMsg20;

  /// No description provided for @loadingMsgGoal.
  ///
  /// In en, this message translates to:
  /// **'Optimizing for {goal}...'**
  String loadingMsgGoal(String goal);

  /// No description provided for @loadingMsgInjuries.
  ///
  /// In en, this message translates to:
  /// **'Adapting around your injuries...'**
  String get loadingMsgInjuries;

  /// No description provided for @loadingMsgFavorites.
  ///
  /// In en, this message translates to:
  /// **'Prioritizing your favorite exercises...'**
  String get loadingMsgFavorites;

  /// No description provided for @loadingMsgConditions.
  ///
  /// In en, this message translates to:
  /// **'Adapting for your conditions...'**
  String get loadingMsgConditions;

  /// No description provided for @loadingMsgMethodology.
  ///
  /// In en, this message translates to:
  /// **'Designing a {methodology} session...'**
  String loadingMsgMethodology(String methodology);

  /// No description provided for @loadingMsgPartners.
  ///
  /// In en, this message translates to:
  /// **'Coordinating partner workout...'**
  String get loadingMsgPartners;

  /// No description provided for @loadingMsgGym.
  ///
  /// In en, this message translates to:
  /// **'Loading {gym} equipment...'**
  String loadingMsgGym(String gym);

  /// No description provided for @loadingMsgHistory.
  ///
  /// In en, this message translates to:
  /// **'Analyzing your recent sessions...'**
  String get loadingMsgHistory;

  /// No description provided for @loadingRetryMsg1.
  ///
  /// In en, this message translates to:
  /// **'hmm, that didn\'t come out right — retrying'**
  String get loadingRetryMsg1;

  /// No description provided for @loadingRetryMsg2.
  ///
  /// In en, this message translates to:
  /// **'let me try that again...'**
  String get loadingRetryMsg2;

  /// No description provided for @loadingRetryMsg3.
  ///
  /// In en, this message translates to:
  /// **'not quite — giving it another shot'**
  String get loadingRetryMsg3;

  /// No description provided for @loadingRetryMsg4.
  ///
  /// In en, this message translates to:
  /// **'one more try, hang tight'**
  String get loadingRetryMsg4;

  /// No description provided for @loadingRetryMsg5.
  ///
  /// In en, this message translates to:
  /// **'oops, recalibrating...'**
  String get loadingRetryMsg5;

  /// No description provided for @loadingRetryMsg6.
  ///
  /// In en, this message translates to:
  /// **'almost had it — trying again'**
  String get loadingRetryMsg6;

  /// No description provided for @nSelected.
  ///
  /// In en, this message translates to:
  /// **'{count} selected'**
  String nSelected(int count);

  /// No description provided for @deleteSelectedTrainings.
  ///
  /// In en, this message translates to:
  /// **'Delete {count} trainings?'**
  String deleteSelectedTrainings(int count);

  /// No description provided for @trainingsDeletedSuccessfully.
  ///
  /// In en, this message translates to:
  /// **'Trainings deleted successfully'**
  String get trainingsDeletedSuccessfully;

  /// No description provided for @shareByLink.
  ///
  /// In en, this message translates to:
  /// **'Share Link'**
  String get shareByLink;

  /// No description provided for @addToMyTrainings.
  ///
  /// In en, this message translates to:
  /// **'Add to My Trainings'**
  String get addToMyTrainings;

  /// No description provided for @goToTraining.
  ///
  /// In en, this message translates to:
  /// **'Go to Training'**
  String get goToTraining;

  /// No description provided for @sharedBy.
  ///
  /// In en, this message translates to:
  /// **'Shared by {name}'**
  String sharedBy(String name);

  /// No description provided for @linkCopied.
  ///
  /// In en, this message translates to:
  /// **'Link copied to clipboard'**
  String get linkCopied;

  /// No description provided for @trainingNotFound.
  ///
  /// In en, this message translates to:
  /// **'Training not found'**
  String get trainingNotFound;

  /// No description provided for @trainingAddedSuccessfully.
  ///
  /// In en, this message translates to:
  /// **'Training added successfully'**
  String get trainingAddedSuccessfully;

  /// No description provided for @failedToAddTraining.
  ///
  /// In en, this message translates to:
  /// **'Failed to add training'**
  String get failedToAddTraining;

  /// No description provided for @loginToAdd.
  ///
  /// In en, this message translates to:
  /// **'Log in to add this training'**
  String get loginToAdd;

  /// No description provided for @pendingFeedbacks.
  ///
  /// In en, this message translates to:
  /// **'Pending feedbacks'**
  String get pendingFeedbacks;

  /// No description provided for @pendingFeedbacksDescription.
  ///
  /// In en, this message translates to:
  /// **'Some of your trainings are marked as complete but have no feedback from you. Feedback helps improve your future training recommendations.'**
  String get pendingFeedbacksDescription;

  /// No description provided for @equipmentPartner.
  ///
  /// In en, this message translates to:
  /// **'Partner'**
  String get equipmentPartner;

  /// No description provided for @equipmentBalanceBoard.
  ///
  /// In en, this message translates to:
  /// **'Balance Board'**
  String get equipmentBalanceBoard;

  /// No description provided for @equipmentBand.
  ///
  /// In en, this message translates to:
  /// **'Band'**
  String get equipmentBand;

  /// No description provided for @equipmentBarbell.
  ///
  /// In en, this message translates to:
  /// **'Barbell'**
  String get equipmentBarbell;

  /// No description provided for @equipmentBench.
  ///
  /// In en, this message translates to:
  /// **'Bench'**
  String get equipmentBench;

  /// No description provided for @equipmentBox.
  ///
  /// In en, this message translates to:
  /// **'Box'**
  String get equipmentBox;

  /// No description provided for @equipmentBosuBall.
  ///
  /// In en, this message translates to:
  /// **'Bosu Ball'**
  String get equipmentBosuBall;

  /// No description provided for @equipmentCable.
  ///
  /// In en, this message translates to:
  /// **'Cable'**
  String get equipmentCable;

  /// No description provided for @equipmentDipStation.
  ///
  /// In en, this message translates to:
  /// **'Dip Station'**
  String get equipmentDipStation;

  /// No description provided for @equipmentDumbbell.
  ///
  /// In en, this message translates to:
  /// **'Dumbbell'**
  String get equipmentDumbbell;

  /// No description provided for @equipmentEllipticalMachine.
  ///
  /// In en, this message translates to:
  /// **'Elliptical Machine'**
  String get equipmentEllipticalMachine;

  /// No description provided for @equipmentEzBarbell.
  ///
  /// In en, this message translates to:
  /// **'EZ Barbell'**
  String get equipmentEzBarbell;

  /// No description provided for @equipmentHammer.
  ///
  /// In en, this message translates to:
  /// **'Hammer'**
  String get equipmentHammer;

  /// No description provided for @equipmentKettlebell.
  ///
  /// In en, this message translates to:
  /// **'Kettlebell'**
  String get equipmentKettlebell;

  /// No description provided for @equipmentLeverageMachine.
  ///
  /// In en, this message translates to:
  /// **'Leverage Machine'**
  String get equipmentLeverageMachine;

  /// No description provided for @equipmentMedicineBall.
  ///
  /// In en, this message translates to:
  /// **'Medicine Ball'**
  String get equipmentMedicineBall;

  /// No description provided for @equipmentOlympicBarbell.
  ///
  /// In en, this message translates to:
  /// **'Olympic Barbell'**
  String get equipmentOlympicBarbell;

  /// No description provided for @equipmentPullUpBar.
  ///
  /// In en, this message translates to:
  /// **'Pull-up Bar'**
  String get equipmentPullUpBar;

  /// No description provided for @equipmentResistanceBand.
  ///
  /// In en, this message translates to:
  /// **'Resistance Band'**
  String get equipmentResistanceBand;

  /// No description provided for @equipmentRings.
  ///
  /// In en, this message translates to:
  /// **'Rings'**
  String get equipmentRings;

  /// No description provided for @equipmentRoller.
  ///
  /// In en, this message translates to:
  /// **'Roller'**
  String get equipmentRoller;

  /// No description provided for @equipmentRope.
  ///
  /// In en, this message translates to:
  /// **'Rope'**
  String get equipmentRope;

  /// No description provided for @equipmentRowingMachine.
  ///
  /// In en, this message translates to:
  /// **'Rowing Machine'**
  String get equipmentRowingMachine;

  /// No description provided for @equipmentSkiergMachine.
  ///
  /// In en, this message translates to:
  /// **'SkiErg Machine'**
  String get equipmentSkiergMachine;

  /// No description provided for @equipmentSledMachine.
  ///
  /// In en, this message translates to:
  /// **'Sled Machine'**
  String get equipmentSledMachine;

  /// No description provided for @equipmentSmithMachine.
  ///
  /// In en, this message translates to:
  /// **'Smith Machine'**
  String get equipmentSmithMachine;

  /// No description provided for @equipmentStabilityBall.
  ///
  /// In en, this message translates to:
  /// **'Stability Ball'**
  String get equipmentStabilityBall;

  /// No description provided for @equipmentStationaryBike.
  ///
  /// In en, this message translates to:
  /// **'Stationary Bike'**
  String get equipmentStationaryBike;

  /// No description provided for @equipmentStepmillMachine.
  ///
  /// In en, this message translates to:
  /// **'Stepmill Machine'**
  String get equipmentStepmillMachine;

  /// No description provided for @equipmentTire.
  ///
  /// In en, this message translates to:
  /// **'Tire'**
  String get equipmentTire;

  /// No description provided for @equipmentTreadmill.
  ///
  /// In en, this message translates to:
  /// **'Treadmill'**
  String get equipmentTreadmill;

  /// No description provided for @equipmentTrapBar.
  ///
  /// In en, this message translates to:
  /// **'Trap Bar'**
  String get equipmentTrapBar;

  /// No description provided for @equipmentTrx.
  ///
  /// In en, this message translates to:
  /// **'TRX'**
  String get equipmentTrx;

  /// No description provided for @equipmentUpperBodyErgometer.
  ///
  /// In en, this message translates to:
  /// **'Upper Body Ergometer'**
  String get equipmentUpperBodyErgometer;

  /// No description provided for @equipmentWheelRoller.
  ///
  /// In en, this message translates to:
  /// **'Ab Wheel'**
  String get equipmentWheelRoller;

  /// No description provided for @muscleChest.
  ///
  /// In en, this message translates to:
  /// **'Chest'**
  String get muscleChest;

  /// No description provided for @muscleBack.
  ///
  /// In en, this message translates to:
  /// **'Back'**
  String get muscleBack;

  /// No description provided for @muscleShoulders.
  ///
  /// In en, this message translates to:
  /// **'Shoulders'**
  String get muscleShoulders;

  /// No description provided for @muscleArms.
  ///
  /// In en, this message translates to:
  /// **'Arms'**
  String get muscleArms;

  /// No description provided for @muscleCore.
  ///
  /// In en, this message translates to:
  /// **'Core'**
  String get muscleCore;

  /// No description provided for @muscleGlutes.
  ///
  /// In en, this message translates to:
  /// **'Glutes'**
  String get muscleGlutes;

  /// No description provided for @muscleLegs.
  ///
  /// In en, this message translates to:
  /// **'Legs'**
  String get muscleLegs;

  /// No description provided for @modifierWeightedVest.
  ///
  /// In en, this message translates to:
  /// **'Weighted Vest'**
  String get modifierWeightedVest;

  /// No description provided for @modifierParallettes.
  ///
  /// In en, this message translates to:
  /// **'Parallettes'**
  String get modifierParallettes;

  /// No description provided for @modifierAnkleWeights.
  ///
  /// In en, this message translates to:
  /// **'Ankle Weights'**
  String get modifierAnkleWeights;

  /// No description provided for @modifierDipBelt.
  ///
  /// In en, this message translates to:
  /// **'Dip Belt'**
  String get modifierDipBelt;

  /// No description provided for @modifierPushUpBars.
  ///
  /// In en, this message translates to:
  /// **'Push-up Bars'**
  String get modifierPushUpBars;

  /// No description provided for @modifierResistanceBands.
  ///
  /// In en, this message translates to:
  /// **'Resistance Bands'**
  String get modifierResistanceBands;

  /// No description provided for @modifierWeight.
  ///
  /// In en, this message translates to:
  /// **'Weight'**
  String get modifierWeight;

  /// No description provided for @modifierWristWeights.
  ///
  /// In en, this message translates to:
  /// **'Wrist Weights'**
  String get modifierWristWeights;

  /// No description provided for @healthData.
  ///
  /// In en, this message translates to:
  /// **'Health Data'**
  String get healthData;

  /// No description provided for @healthConnected.
  ///
  /// In en, this message translates to:
  /// **'Connected'**
  String get healthConnected;

  /// No description provided for @healthSynchronizing.
  ///
  /// In en, this message translates to:
  /// **'Synchronizing...'**
  String get healthSynchronizing;

  /// No description provided for @healthSynchronized.
  ///
  /// In en, this message translates to:
  /// **'Synchronized'**
  String get healthSynchronized;

  /// No description provided for @healthSynchronize.
  ///
  /// In en, this message translates to:
  /// **'Synchronize now'**
  String get healthSynchronize;

  /// No description provided for @healthNotConnected.
  ///
  /// In en, this message translates to:
  /// **'Not connected'**
  String get healthNotConnected;

  /// No description provided for @healthNativeOnly.
  ///
  /// In en, this message translates to:
  /// **'Available on iOS and Android'**
  String get healthNativeOnly;

  /// No description provided for @healthBuildingBaselines.
  ///
  /// In en, this message translates to:
  /// **'Building your personal baselines ({days}/14 days)'**
  String healthBuildingBaselines(int days);

  /// No description provided for @healthSyncNoData.
  ///
  /// In en, this message translates to:
  /// **'No health data found on device'**
  String get healthSyncNoData;

  /// No description provided for @healthSyncFailed.
  ///
  /// In en, this message translates to:
  /// **'Sync failed'**
  String get healthSyncFailed;

  /// No description provided for @healthSyncFailedDetail.
  ///
  /// In en, this message translates to:
  /// **'Data read from device but upload failed: {error}'**
  String healthSyncFailedDetail(String error);

  /// No description provided for @healthSyncTypeFull.
  ///
  /// In en, this message translates to:
  /// **'Full'**
  String get healthSyncTypeFull;

  /// No description provided for @healthSyncTypeIncremental.
  ///
  /// In en, this message translates to:
  /// **'Incremental'**
  String get healthSyncTypeIncremental;

  /// No description provided for @healthSourceData.
  ///
  /// In en, this message translates to:
  /// **'{metrics} metrics · {sessions} sessions'**
  String healthSourceData(int metrics, int sessions);

  /// No description provided for @healthBackend.
  ///
  /// In en, this message translates to:
  /// **'Backend'**
  String get healthBackend;

  /// No description provided for @healthBackendData.
  ///
  /// In en, this message translates to:
  /// **'{days} days · {sessions} sessions'**
  String healthBackendData(int days, int sessions);

  /// No description provided for @healthDateRange.
  ///
  /// In en, this message translates to:
  /// **'{from} – {to}'**
  String healthDateRange(String from, String to);

  /// No description provided for @healthDisconnect.
  ///
  /// In en, this message translates to:
  /// **'Disconnect & Delete Data'**
  String get healthDisconnect;

  /// No description provided for @healthDisconnectConfirmation.
  ///
  /// In en, this message translates to:
  /// **'Disconnect health data? All synced health data will be deleted. This cannot be undone.'**
  String get healthDisconnectConfirmation;

  /// No description provided for @healthDisconnectedSuccessfully.
  ///
  /// In en, this message translates to:
  /// **'Health data disconnected'**
  String get healthDisconnectedSuccessfully;

  /// No description provided for @failedToDisconnectHealth.
  ///
  /// In en, this message translates to:
  /// **'Failed to disconnect health data'**
  String get failedToDisconnectHealth;

  /// No description provided for @healthConnect.
  ///
  /// In en, this message translates to:
  /// **'Connect'**
  String get healthConnect;

  /// No description provided for @healthMetrics.
  ///
  /// In en, this message translates to:
  /// **'Health Metrics'**
  String get healthMetrics;

  /// No description provided for @healthAdjustment.
  ///
  /// In en, this message translates to:
  /// **'Health Adjustment'**
  String get healthAdjustment;

  /// No description provided for @healthPermissionsTitle.
  ///
  /// In en, this message translates to:
  /// **'Connect Your Wearable'**
  String get healthPermissionsTitle;

  /// No description provided for @healthPermissionsDescription.
  ///
  /// In en, this message translates to:
  /// **'Vigor reads your health data to personalize your workouts. Better sleep data, recovery metrics, and activity history mean smarter training recommendations.'**
  String get healthPermissionsDescription;

  /// No description provided for @healthPermissionsReadOnly.
  ///
  /// In en, this message translates to:
  /// **'Read-only access'**
  String get healthPermissionsReadOnly;

  /// No description provided for @healthPermissionsSleep.
  ///
  /// In en, this message translates to:
  /// **'Sleep duration and stages'**
  String get healthPermissionsSleep;

  /// No description provided for @healthPermissionsHrv.
  ///
  /// In en, this message translates to:
  /// **'Heart rate variability (HRV)'**
  String get healthPermissionsHrv;

  /// No description provided for @healthPermissionsRhr.
  ///
  /// In en, this message translates to:
  /// **'Resting heart rate'**
  String get healthPermissionsRhr;

  /// No description provided for @healthPermissionsSteps.
  ///
  /// In en, this message translates to:
  /// **'Daily steps'**
  String get healthPermissionsSteps;

  /// No description provided for @healthPermissionsWorkouts.
  ///
  /// In en, this message translates to:
  /// **'Workout sessions and heart rate'**
  String get healthPermissionsWorkouts;

  /// No description provided for @healthPermissionsGrant.
  ///
  /// In en, this message translates to:
  /// **'Connect Health Data'**
  String get healthPermissionsGrant;

  /// No description provided for @healthPermissionsSkip.
  ///
  /// In en, this message translates to:
  /// **'Not now'**
  String get healthPermissionsSkip;

  /// No description provided for @healthPermissionsGranted.
  ///
  /// In en, this message translates to:
  /// **'Health data connected'**
  String get healthPermissionsGranted;

  /// No description provided for @healthPermissionsDenied.
  ///
  /// In en, this message translates to:
  /// **'Permissions were not granted'**
  String get healthPermissionsDenied;

  /// No description provided for @healthOnboardingTitle.
  ///
  /// In en, this message translates to:
  /// **'Connect your wearable'**
  String get healthOnboardingTitle;

  /// No description provided for @healthOnboardingDescription.
  ///
  /// In en, this message translates to:
  /// **'Connect your health data and Vigor will adapt your training to how you sleep, recover, and move.'**
  String get healthOnboardingDescription;

  /// No description provided for @healthOnboardingConnect.
  ///
  /// In en, this message translates to:
  /// **'Connect'**
  String get healthOnboardingConnect;

  /// No description provided for @healthOnboardingDismiss.
  ///
  /// In en, this message translates to:
  /// **'Maybe later'**
  String get healthOnboardingDismiss;

  /// No description provided for @healthInstallHcTitle.
  ///
  /// In en, this message translates to:
  /// **'Health Connect Required'**
  String get healthInstallHcTitle;

  /// No description provided for @healthInstallHcDescription.
  ///
  /// In en, this message translates to:
  /// **'Health Connect is needed to sync your wearable data. Install it from the Play Store to continue.'**
  String get healthInstallHcDescription;

  /// No description provided for @healthInstallHc.
  ///
  /// In en, this message translates to:
  /// **'Install Health Connect'**
  String get healthInstallHc;

  /// No description provided for @heartRate.
  ///
  /// In en, this message translates to:
  /// **'Heart Rate'**
  String get heartRate;

  /// No description provided for @avgHr.
  ///
  /// In en, this message translates to:
  /// **'Avg HR'**
  String get avgHr;

  /// No description provided for @maxHr.
  ///
  /// In en, this message translates to:
  /// **'Max HR'**
  String get maxHr;

  /// No description provided for @bpm.
  ///
  /// In en, this message translates to:
  /// **'bpm'**
  String get bpm;

  /// No description provided for @hrZones.
  ///
  /// In en, this message translates to:
  /// **'HR Zones'**
  String get hrZones;

  /// No description provided for @hrZone1.
  ///
  /// In en, this message translates to:
  /// **'Zone 1'**
  String get hrZone1;

  /// No description provided for @hrZone2.
  ///
  /// In en, this message translates to:
  /// **'Zone 2'**
  String get hrZone2;

  /// No description provided for @hrZone3.
  ///
  /// In en, this message translates to:
  /// **'Zone 3'**
  String get hrZone3;

  /// No description provided for @hrZone4.
  ///
  /// In en, this message translates to:
  /// **'Zone 4'**
  String get hrZone4;

  /// No description provided for @hrZone5.
  ///
  /// In en, this message translates to:
  /// **'Zone 5'**
  String get hrZone5;

  /// No description provided for @healthDailySleep.
  ///
  /// In en, this message translates to:
  /// **'Sleep'**
  String get healthDailySleep;

  /// No description provided for @healthDailyRestingHr.
  ///
  /// In en, this message translates to:
  /// **'Resting HR'**
  String get healthDailyRestingHr;

  /// No description provided for @healthDailyHrv.
  ///
  /// In en, this message translates to:
  /// **'HRV'**
  String get healthDailyHrv;

  /// No description provided for @healthDailySteps.
  ///
  /// In en, this message translates to:
  /// **'Steps'**
  String get healthDailySteps;

  /// No description provided for @healthDailyCalories.
  ///
  /// In en, this message translates to:
  /// **'Calories'**
  String get healthDailyCalories;

  /// No description provided for @healthDailyNoData.
  ///
  /// In en, this message translates to:
  /// **'No health data yet'**
  String get healthDailyNoData;

  /// No description provided for @externalWorkout.
  ///
  /// In en, this message translates to:
  /// **'External Workout'**
  String get externalWorkout;

  /// No description provided for @exerciseTypeRunning.
  ///
  /// In en, this message translates to:
  /// **'Running'**
  String get exerciseTypeRunning;

  /// No description provided for @exerciseTypeWalking.
  ///
  /// In en, this message translates to:
  /// **'Walking'**
  String get exerciseTypeWalking;

  /// No description provided for @exerciseTypeBiking.
  ///
  /// In en, this message translates to:
  /// **'Cycling'**
  String get exerciseTypeBiking;

  /// No description provided for @exerciseTypeYoga.
  ///
  /// In en, this message translates to:
  /// **'Yoga'**
  String get exerciseTypeYoga;

  /// No description provided for @exerciseTypeSwimming.
  ///
  /// In en, this message translates to:
  /// **'Swimming'**
  String get exerciseTypeSwimming;

  /// No description provided for @exerciseTypeHiking.
  ///
  /// In en, this message translates to:
  /// **'Hiking'**
  String get exerciseTypeHiking;

  /// No description provided for @exerciseTypeStrengthTraining.
  ///
  /// In en, this message translates to:
  /// **'Strength Training'**
  String get exerciseTypeStrengthTraining;

  /// No description provided for @exerciseTypeFunctionalStrengthTraining.
  ///
  /// In en, this message translates to:
  /// **'Functional Training'**
  String get exerciseTypeFunctionalStrengthTraining;

  /// No description provided for @exerciseTypeTraditionalStrengthTraining.
  ///
  /// In en, this message translates to:
  /// **'Strength Training'**
  String get exerciseTypeTraditionalStrengthTraining;

  /// No description provided for @exerciseTypeRunningTreadmill.
  ///
  /// In en, this message translates to:
  /// **'Treadmill'**
  String get exerciseTypeRunningTreadmill;

  /// No description provided for @exerciseTypeBikingStationary.
  ///
  /// In en, this message translates to:
  /// **'Stationary Bike'**
  String get exerciseTypeBikingStationary;

  /// No description provided for @exerciseTypeWalkingTreadmill.
  ///
  /// In en, this message translates to:
  /// **'Treadmill Walk'**
  String get exerciseTypeWalkingTreadmill;

  /// No description provided for @exerciseTypeRowing.
  ///
  /// In en, this message translates to:
  /// **'Rowing'**
  String get exerciseTypeRowing;

  /// No description provided for @exerciseTypePilates.
  ///
  /// In en, this message translates to:
  /// **'Pilates'**
  String get exerciseTypePilates;

  /// No description provided for @exerciseTypeDancing.
  ///
  /// In en, this message translates to:
  /// **'Dancing'**
  String get exerciseTypeDancing;

  /// No description provided for @exerciseTypeElliptical.
  ///
  /// In en, this message translates to:
  /// **'Elliptical'**
  String get exerciseTypeElliptical;

  /// No description provided for @exerciseTypeStairClimbing.
  ///
  /// In en, this message translates to:
  /// **'Stair Climbing'**
  String get exerciseTypeStairClimbing;

  /// No description provided for @exerciseTypeOther.
  ///
  /// In en, this message translates to:
  /// **'Workout'**
  String get exerciseTypeOther;

  /// No description provided for @appLogs.
  ///
  /// In en, this message translates to:
  /// **'App Logs'**
  String get appLogs;

  /// No description provided for @viewLogs.
  ///
  /// In en, this message translates to:
  /// **'View Logs'**
  String get viewLogs;

  /// No description provided for @exportLogs.
  ///
  /// In en, this message translates to:
  /// **'Export Logs'**
  String get exportLogs;

  /// No description provided for @clearLogs.
  ///
  /// In en, this message translates to:
  /// **'Clear Logs'**
  String get clearLogs;

  /// No description provided for @noLogsYet.
  ///
  /// In en, this message translates to:
  /// **'No logs recorded yet'**
  String get noLogsYet;

  /// No description provided for @logsCleared.
  ///
  /// In en, this message translates to:
  /// **'Logs cleared'**
  String get logsCleared;

  /// No description provided for @logEntries.
  ///
  /// In en, this message translates to:
  /// **'{count} log entries'**
  String logEntries(int count);
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) => <String>[
    'de',
    'en',
    'es',
    'fr',
    'it',
    'ja',
    'ko',
    'pt',
    'ru',
    'zh',
  ].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'de':
      return AppLocalizationsDe();
    case 'en':
      return AppLocalizationsEn();
    case 'es':
      return AppLocalizationsEs();
    case 'fr':
      return AppLocalizationsFr();
    case 'it':
      return AppLocalizationsIt();
    case 'ja':
      return AppLocalizationsJa();
    case 'ko':
      return AppLocalizationsKo();
    case 'pt':
      return AppLocalizationsPt();
    case 'ru':
      return AppLocalizationsRu();
    case 'zh':
      return AppLocalizationsZh();
  }

  throw FlutterError(
    'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
    'an issue with the localizations generation tool. Please file an issue '
    'on GitHub with a reproducible sample app and the gen-l10n configuration '
    'that was used.',
  );
}
