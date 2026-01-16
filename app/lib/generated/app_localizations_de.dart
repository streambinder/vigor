// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for German (`de`).
class AppLocalizationsDe extends AppLocalizations {
  AppLocalizationsDe([String locale = 'de']) : super(locale);

  @override
  String get appName => 'Vigor';

  @override
  String get appTagline => 'Ex Sapientia Vis';

  @override
  String get navHome => 'Startseite';

  @override
  String get navActivity => 'Aktivität';

  @override
  String get navProfile => 'Profil';

  @override
  String get storageErrorTitle => 'Vigor - Speicherfehler';

  @override
  String get storageUnavailable => 'Speicher nicht verfügbar';

  @override
  String get storageErrorMessage =>
      'Sicherer Speicher erforderlich. Prüfe die Einstellungen.';

  @override
  String get signInWithGoogle => 'Mit Google anmelden';

  @override
  String get signingIn => 'Anmeldung läuft...';

  @override
  String get failedToInitializeGoogleSignIn =>
      'Google Sign In konnte nicht initialisiert werden';

  @override
  String signInError(String message) {
    return 'Anmeldefehler: $message';
  }

  @override
  String get googleSignInFailed => 'Google-Anmeldung fehlgeschlagen';

  @override
  String get failedToGetAuthToken =>
      'Authentifizierungstoken konnte nicht abgerufen werden';

  @override
  String errorProcessingSignIn(String message) {
    return 'Fehler bei der Verarbeitung der Anmeldung: $message';
  }

  @override
  String get googleSignInInitializing =>
      'Google Sign In wird noch initialisiert...';

  @override
  String get readyToTrain => 'Bereit zum Trainieren?';

  @override
  String get generateTrainingDescription =>
      'Erstelle ein Training für deine Ziele';

  @override
  String get generateTraining => 'Training Erstellen';

  @override
  String get refresh => 'Aktualisieren';

  @override
  String get logout => 'Abmelden';

  @override
  String get userDataRefreshed => 'Benutzerdaten aktualisiert';

  @override
  String get editProfile => 'Profil Bearbeiten';

  @override
  String get settings => 'Einstellungen';

  @override
  String get deleteGym => 'Fitnessstudio Löschen';

  @override
  String deleteGymConfirmation(String name) {
    return '\"$name\" löschen?';
  }

  @override
  String get cancel => 'Abbrechen';

  @override
  String get delete => 'Löschen';

  @override
  String get logoutConfirmation => 'Abmelden?';

  @override
  String get deleteAccount => 'Konto Löschen';

  @override
  String get deleteAccountConfirmation =>
      'Konto löschen? Nicht rückgängig zu machen.';

  @override
  String get accountDeletedSuccessfully => 'Konto erfolgreich gelöscht';

  @override
  String get failedToDeleteAccount => 'Konto konnte nicht gelöscht werden';

  @override
  String get failedToLoadGyms => 'Fitnessstudios konnten nicht geladen werden';

  @override
  String get gymAddedSuccessfully => 'Fitnessstudio erfolgreich hinzugefügt';

  @override
  String get failedToAddGym => 'Fitnessstudio konnte nicht hinzugefügt werden';

  @override
  String get gymUpdatedSuccessfully => 'Fitnessstudio erfolgreich aktualisiert';

  @override
  String get failedToUpdateGym =>
      'Fitnessstudio konnte nicht aktualisiert werden';

  @override
  String get gymDeletedSuccessfully => 'Fitnessstudio erfolgreich gelöscht';

  @override
  String get failedToDeleteGym => 'Fitnessstudio konnte nicht gelöscht werden';

  @override
  String get birthdate => 'Geburtsdatum';

  @override
  String get gender => 'Geschlecht';

  @override
  String get language => 'Sprache';

  @override
  String get height => 'Größe';

  @override
  String get weight => 'Gewicht';

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
  String get goals => 'Ziele';

  @override
  String get injuries => 'Verletzungen';

  @override
  String get limitations => 'Einschränkungen';

  @override
  String get favorites => 'Favoriten';

  @override
  String get exercises => 'Übungen';

  @override
  String get equipment => 'Ausrüstung';

  @override
  String startedDate(String date) {
    return 'Beginn: $date';
  }

  @override
  String yearLabel(int year) {
    return 'Jahr: $year';
  }

  @override
  String get myGyms => 'Fitnessstudios';

  @override
  String get addGym => 'Fitnessstudio Hinzufügen';

  @override
  String get noGymsAddedYet => 'Noch keine Fitnessstudios hinzugefügt';

  @override
  String get addYourFirstGym => 'Füge Dein Erstes Fitnessstudio Hinzu';

  @override
  String get removeDefault => 'Standard Entfernen';

  @override
  String get setAsDefault => 'Als Standard Festlegen';

  @override
  String get edit => 'Bearbeiten';

  @override
  String get quickActions => 'Schnellaktionen';

  @override
  String get dangerZone => 'Gefahrenzone';

  @override
  String get completeYourProfile => 'Vervollständige Dein Profil';

  @override
  String get updateYourProfileInfo => 'Aktualisiere dein Profil unten.';

  @override
  String get pleaseCompleteProfile =>
      'Profil vervollständigen. * = Pflichtfeld.';

  @override
  String get firstName => 'Vorname';

  @override
  String get lastName => 'Nachname';

  @override
  String get birthDate => 'Geburtsdatum';

  @override
  String get male => 'Männlich';

  @override
  String get female => 'Weiblich';

  @override
  String get selectLanguage => 'Sprache auswählen';

  @override
  String get heightCm => 'Größe (cm)';

  @override
  String get weightKg => 'Gewicht (kg)';

  @override
  String get required => 'Erforderlich';

  @override
  String get invalid => 'Ungültig';

  @override
  String get pleaseSelectBirthDate => 'Bitte wähle dein Geburtsdatum';

  @override
  String get pleaseAddAtLeastOneGoal => 'Bitte füge mindestens ein Ziel hinzu';

  @override
  String get pleaseSelectLanguage => 'Bitte wähle deine Sprache';

  @override
  String get addAGoal => 'Ziel hinzufügen';

  @override
  String get injuryDescription => 'Verletzungsbeschreibung';

  @override
  String get year => 'Jahr';

  @override
  String get addALimitation => 'Einschränkung hinzufügen';

  @override
  String get favoriteExercisesHint => 'z.B. Kniebeugen, Klimmzüge, Laufen';

  @override
  String get favoriteEquipmentHint =>
      'z.B. Kurzhanteln, Langhantel, Kettlebell';

  @override
  String get saveChanges => 'Änderungen Speichern';

  @override
  String get saveProfile => 'Profil Speichern';

  @override
  String get optionalLeaveEmpty => '(Optional)';

  @override
  String get optionalExercisesPrefer => '(Optional)';

  @override
  String get optionalEquipmentPrefer => '(Optional)';

  @override
  String get optionalWorkoutTypesPrefer => '(Optional)';

  @override
  String get favoriteExercises => 'Lieblingsübungen';

  @override
  String get favoriteEquipment => 'Lieblingsausrüstung';

  @override
  String get favoriteWorkoutTypes => 'Bevorzugte Trainingsarten';

  @override
  String get workoutTypeStrength => 'Kraft';

  @override
  String get workoutTypeCircuit => 'Zirkel';

  @override
  String get workoutTypeEmom => 'EMOM';

  @override
  String get workoutTypeAmrap => 'AMRAP';

  @override
  String get workoutTypeHiit => 'HIIT';

  @override
  String get workoutTypeForTime => 'For Time';

  @override
  String get workoutTypeEndurance => 'Ausdauer';

  @override
  String get workoutTypeMobility => 'Mobilität';

  @override
  String get methodologyOptional => 'Methodik (optional)';

  @override
  String get methodologyAuto => 'Auto';

  @override
  String get goalsOptional => 'Ziele (optional)';

  @override
  String get musclesOptional => 'Muskeln (optional)';

  @override
  String get musclesAuto => 'Auto';

  @override
  String get advancedSettings => 'Erweitert';

  @override
  String get failedToUpdateProfile => 'Profil konnte nicht aktualisiert werden';

  @override
  String get activity => 'Aktivität';

  @override
  String get noTrainingsYet => 'Noch keine Trainings';

  @override
  String get generateFirstTraining =>
      'Erstelle dein erstes Training von Startseite';

  @override
  String get noTrainingAvailable => 'Kein Training. Erstelle eines.';

  @override
  String get availableTrainings => 'Verfügbare Trainings';

  @override
  String get pastTrainings => 'Vergangene Trainings';

  @override
  String get stale => 'Veraltet';

  @override
  String get today => 'Heute';

  @override
  String get yesterday => 'Gestern';

  @override
  String get available => 'Verfügbar';

  @override
  String get completed => 'Abgeschlossen';

  @override
  String get noPastTrainings => 'Noch keine abgeschlossenen Trainings';

  @override
  String get copied => 'Kopiert';

  @override
  String durationMin(int minutes) {
    return '$minutes Min';
  }

  @override
  String durationHr(int hours) {
    return '$hours Std';
  }

  @override
  String durationHrMin(int hours, int minutes) {
    return '$hours Std $minutes Min';
  }

  @override
  String get failedToLoadTrainings => 'Trainings konnten nicht geladen werden';

  @override
  String get startTraining => 'Training Starten';

  @override
  String get cloneTraining => 'Training Klonen';

  @override
  String get addPartner => 'Partner Hinzufügen';

  @override
  String get shareWithUser => 'Mit Benutzer Teilen';

  @override
  String get deleteTraining => 'Training Löschen';

  @override
  String get leaveTraining => 'Training Verlassen';

  @override
  String get showAiReasoning => 'KI-Begründung anzeigen';

  @override
  String get reportIssue => 'Problem melden';

  @override
  String deleteTrainingConfirmation(String name) {
    return '\"$name\" löschen? Nicht rückgängig zu machen.';
  }

  @override
  String leaveTrainingConfirmation(String name) {
    return '\"$name\" verlassen? Nicht mehr sichtbar.';
  }

  @override
  String addPartnerConfirmation(String userName, String trainingName) {
    return '$userName als Partner zu \"$trainingName\" hinzufügen?';
  }

  @override
  String cloneTrainingConfirmation(String name) {
    return '\"$name\" in deine Trainings klonen?';
  }

  @override
  String shareTrainingConfirmation(String trainingName, String userName) {
    return '\"$trainingName\" mit $userName teilen?';
  }

  @override
  String get trainingDeletedSuccessfully => 'Training erfolgreich gelöscht';

  @override
  String get failedToDeleteTraining => 'Training konnte nicht gelöscht werden';

  @override
  String get leftTrainingSuccessfully => 'Training erfolgreich verlassen';

  @override
  String get partnerAddedSuccessfully => 'Partner erfolgreich hinzugefügt';

  @override
  String get failedToAddPartner => 'Partner konnte nicht hinzugefügt werden';

  @override
  String get trainingSharedSuccessfully => 'Training erfolgreich geteilt';

  @override
  String get failedToShareTraining => 'Training konnte nicht geteilt werden';

  @override
  String get trainingCloned => 'Training geklont';

  @override
  String get failedToCloneTraining => 'Training konnte nicht geklont werden';

  @override
  String get trainingMarkedAsComplete => 'Training abgeschlossen';

  @override
  String get failedToCompleteTraining =>
      'Training konnte nicht abgeschlossen werden';

  @override
  String get reportSubmitted => 'Meldung gesendet';

  @override
  String get failedToSubmitReport => 'Meldung konnte nicht gesendet werden';

  @override
  String get shuffleExercise => 'Übung wechseln';

  @override
  String get exerciseShuffled => 'Übung gewechselt';

  @override
  String get failedToShuffleExercise => 'Übung konnte nicht gewechselt werden';

  @override
  String get reasoning => 'Begründung';

  @override
  String get strategy => 'Strategie';

  @override
  String get typeSelection => 'Typauswahl';

  @override
  String get progression => 'Progression';

  @override
  String get constraints => 'Einschränkungen';

  @override
  String get researchApplied => 'Angewandte Forschung';

  @override
  String get targetMuscles => 'Zielmuskeln';

  @override
  String get naming => 'Benennung';

  @override
  String get trainingRoutines => 'Trainingsroutinen';

  @override
  String get noEquipment => 'Keine Ausrüstung';

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
    return '${seconds}s Pause';
  }

  @override
  String repsCount(int count) {
    return '$count Wiederholungen';
  }

  @override
  String weightKgValue(double value) {
    return '$value kg';
  }

  @override
  String get markAsComplete => 'Als Abgeschlossen Markieren';

  @override
  String get references => 'Referenzen';

  @override
  String get describeIssue => 'Beschreibe das Problem...';

  @override
  String get submit => 'Absenden';

  @override
  String get close => 'Schließen';

  @override
  String get add => 'Hinzufügen';

  @override
  String get update => 'Aktualisieren';

  @override
  String get clone => 'Klonen';

  @override
  String get share => 'Teilen';

  @override
  String get leave => 'Verlassen';

  @override
  String get tapToStart => 'Tippe zum Starten';

  @override
  String get trainingCompleted => 'Training Abgeschlossen!';

  @override
  String greatJobCompleting(String name) {
    return 'Tolle Arbeit bei $name';
  }

  @override
  String get done => 'Fertig';

  @override
  String get complete => 'Abschließen';

  @override
  String routineCounter(int current, int total) {
    return 'Routine $current/$total';
  }

  @override
  String blockCounter(int current, int total) {
    return 'Block $current/$total';
  }

  @override
  String get exitTraining => 'Training Verlassen?';

  @override
  String get whatWouldYouLikeToDo => 'Was nun?';

  @override
  String get exit => 'Verlassen';

  @override
  String get continueTraining => 'Fortfahren';

  @override
  String get failedToMarkComplete =>
      'Training konnte nicht abgeschlossen werden';

  @override
  String get durationMinutes => 'Dauer (Minuten)';

  @override
  String get bodyweight => 'Körpergewicht';

  @override
  String get gym => 'Fitnessstudio';

  @override
  String get custom => 'Benutzerdefiniert';

  @override
  String get noEquipmentBodyweightOnly => 'Nur Körpergewicht';

  @override
  String get noGymsDefinedCreateOne =>
      'Kein Fitnessstudio. Erstelle eines im Profil.';

  @override
  String get selectAGym => 'Fitnessstudio auswählen';

  @override
  String get addEquipment => 'Ausrüstung Hinzufügen';

  @override
  String get addEquipmentAvailable => 'Füge deine Ausrüstung hinzu';

  @override
  String get includeWarmupCooldown => 'Aufwärmen und Abkühlen einschließen';

  @override
  String get equipmentPlaceholder => 'z.B. Langhantel, Kurzhanteln';

  @override
  String get customPromptOptional => 'Benutzerdefinierter Prompt (optional)';

  @override
  String get focusOnUpperBody => 'z.B. Fokus auf Oberkörper';

  @override
  String get trainingPartnersOptional => 'Trainingspartner (optional)';

  @override
  String get generatingTraining => 'Dein Training wird erstellt...';

  @override
  String get thisMayTakeAMoment => 'Dies kann einen Moment dauern';

  @override
  String generationFailedRetrying(int attempt) {
    return 'Generierung fehlgeschlagen, Versuch #$attempt...';
  }

  @override
  String get trainingGeneratedSuccessfully => 'Training erfolgreich erstellt!';

  @override
  String get failedToGenerateTraining =>
      'Training konnte nicht erstellt werden';

  @override
  String get generate => 'Erstellen';

  @override
  String get editGym => 'Fitnessstudio Bearbeiten';

  @override
  String get gymName => 'Name des Fitnessstudios';

  @override
  String get gymNamePlaceholder => 'z.B. Home Gym, McFit';

  @override
  String get noEquipmentAddedYet => 'Noch keine Ausrüstung hinzugefügt';

  @override
  String get pleaseEnterGymName => 'Bitte gib den Namen des Fitnessstudios ein';

  @override
  String get addAllEquipment => 'Alle Hinzufügen';

  @override
  String get failedToLoadEquipment => 'Ausrüstung konnte nicht geladen werden';

  @override
  String get selectUser => 'Benutzer Auswählen';

  @override
  String get searchByName => 'Nach Name suchen';

  @override
  String get noUsersAvailable => 'Keine Benutzer verfügbar';

  @override
  String get noMatchingUsers => 'Keine passenden Benutzer';

  @override
  String get instructions => 'Anleitung';

  @override
  String get howWasYourTraining => 'Wie war dein Training?';

  @override
  String get anyAdditionalComments => 'Weitere Kommentare?';

  @override
  String get tooEasy => 'Zu leicht';

  @override
  String get tooHard => 'Zu schwer';

  @override
  String get flag => 'Markieren';

  @override
  String get profile => 'Profil';

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
    return 'Nächste: $name';
  }

  @override
  String get rest => 'Pause';

  @override
  String get upcoming => 'Kommend';

  @override
  String get yourProgress => 'Dein Fortschritt';

  @override
  String trainingsCompleted(int count) {
    return '$count Trainings abgeschlossen';
  }

  @override
  String get completedTrainings => 'abgeschlossene Trainings';

  @override
  String get partneredTrainings => 'Partner-Trainings';

  @override
  String get movementFamilies => 'Bewegungsfamilien';

  @override
  String get muscleActivity => 'Muskelaktivität';

  @override
  String get failedToLoadProgress => 'Fortschritt konnte nicht geladen werden';

  @override
  String get noProgressYet =>
      'Schließe Trainings ab, um deinen Fortschritt zu sehen';

  @override
  String get calibration => 'Kalibrierung';

  @override
  String get calibrationNeeded =>
      'Absolviere dein erstes Training, damit Vigor die Empfehlungen auf dein Level abstimmen kann';

  @override
  String get calibrationDescription =>
      'Während der Kalibrierung sammelt die Plattform Daten aus deinem Feedback, um eine erste Einschätzung deines Fitnesszustands und Levels zu machen';

  @override
  String get capabilities => 'Fähigkeiten';

  @override
  String get noTrainingsCompletedYet =>
      'Starte ein Training um hier etwas zu sehen';
}
