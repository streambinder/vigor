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
  String get other => 'Sonstiges';

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
  String get conditions => 'Beschwerden';

  @override
  String get favorites => 'Favoriten';

  @override
  String get personalDetails => 'Persönliche Daten';

  @override
  String get healthAndGoals => 'Gesundheit & Ziele';

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
  String get addACondition => 'Beschwerde hinzufügen';

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
  String get workoutTypeStrengthDescription =>
      'Maximalkraft mit schweren Lasten und voller Erholung aufbauen';

  @override
  String get workoutTypeSupersets => 'Supersätze';

  @override
  String get workoutTypeSupersetsDescription =>
      'Gegenüberliegende Muskeln direkt nacheinander für zeiteffizientes Training';

  @override
  String get workoutTypeCircuit => 'Zirkel';

  @override
  String get workoutTypeCircuitDescription =>
      'Wechseln Sie zwischen Stationen mit minimaler Pause für Konditionierung';

  @override
  String get workoutTypeEmom => 'EMOM';

  @override
  String get workoutTypeEmomDescription =>
      'Jede Minute: Wiederholungen abschließen und bis zur nächsten Minute ruhen';

  @override
  String get workoutTypeAmrap => 'AMRAP';

  @override
  String get workoutTypeAmrapDescription =>
      'So viele Runden wie möglich innerhalb des Zeitlimits';

  @override
  String get workoutTypeHiit => 'HIIT';

  @override
  String get workoutTypeHiitDescription =>
      'Hochintensive Intervalle mit kurzer Erholung abwechseln';

  @override
  String get workoutTypeForTime => 'For Time';

  @override
  String get workoutTypeForTimeDescription =>
      'Das Training so schnell wie möglich abschließen';

  @override
  String get workoutTypeEndurance => 'Ausdauer';

  @override
  String get workoutTypeEnduranceDescription =>
      'Anhaltende Anstrengung bei moderater Intensität für aerobe Kapazität';

  @override
  String get workoutTypeMobility => 'Mobilität';

  @override
  String get workoutTypeMobilityDescription =>
      'Bewegungsumfang und Gelenkgesundheit verbessern';

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
  String daysAgo(int count) {
    return 'vor ${count}T';
  }

  @override
  String get available => 'Verfügbar';

  @override
  String get completed => 'Abgeschlossen';

  @override
  String get completedSingular => 'Abgeschlossen';

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
  String get feedback => 'Feedback';

  @override
  String get feedbackUpdated => 'Feedback aktualisiert';

  @override
  String get failedToUpdateFeedback =>
      'Feedback konnte nicht aktualisiert werden';

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
  String get trainingRoutines => 'Training';

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
  String get updateFeedback => 'Feedback Aktualisieren';

  @override
  String get references => 'Referenzen';

  @override
  String get literature => 'Literatur';

  @override
  String get request => 'Anfrage';

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
  String get tapWhenDone => 'Tippe wenn fertig';

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
  String get stop => 'Stopp';

  @override
  String get stopTraining => 'Training Stoppen?';

  @override
  String get stopTrainingConfirm => 'Der Timer-Fortschritt geht verloren.';

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
  String get custom => 'Sonstiges';

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
  String get trainingPartnersOptional => 'Partner';

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
  String get availableWeights => 'Verfügbare Gewichte';

  @override
  String get availableWeightsHint =>
      'Gewichtsoptionen für gewichtete Modifikatoren in diesem Fitnessstudio konfigurieren.';

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
  String get noMatchingEquipment => 'Keine passende Ausrüstung';

  @override
  String get instructions => 'Anleitung';

  @override
  String get cues => 'Hinweise';

  @override
  String get howWasYourTraining => 'Wie war dein Training?';

  @override
  String get anyAdditionalComments => 'Weitere Kommentare?';

  @override
  String get actualDuration => 'Tatsächliche Dauer (Min)';

  @override
  String get impossible => 'Unmöglich';

  @override
  String get tooHard => 'Zu schwer';

  @override
  String get ok => 'OK';

  @override
  String get easy => 'Leicht';

  @override
  String get tooEasy => 'Zu leicht';

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
  String get calibrationGlobal => 'Gesamt';

  @override
  String get calibrationNeeded =>
      'Absolviere dein erstes Training, damit Vigor die Empfehlungen auf dein Level abstimmen kann';

  @override
  String get calibrationDescription =>
      'Während der Kalibrierung sammelt die Plattform Daten aus deinem Feedback, um eine erste Einschätzung deines Fitnesszustands und Levels zu machen';

  @override
  String get calibrationInProgress =>
      'Deine Trainings werden mit jeder Einheit intelligenter';

  @override
  String get calibrationTrainingNote =>
      'Dieses Training stimmt möglicherweise nicht vollständig mit deinen Zielen überein — das System lernt noch dein Fitnesslevel und priorisiert Bewegungsvielfalt, um ein vollständiges Profil aufzubauen';

  @override
  String calibrationFamiliesLearned(int calibrated, int total) {
    return '$calibrated/$total Bewegungsmuster gelernt';
  }

  @override
  String get capabilities => 'Fähigkeiten';

  @override
  String get muscleHeatMap => 'Muskel-Heatmap';

  @override
  String get heatResting => 'Ruhend';

  @override
  String get heatRecovered => 'Erholt';

  @override
  String get heatActive => 'Aktiv';

  @override
  String get heatWarm => 'Warm';

  @override
  String get heatHot => 'Intensiv';

  @override
  String get noTrainingsCompletedYet =>
      'Starte ein Training um hier etwas zu sehen';

  @override
  String get theme => 'Thema';

  @override
  String get themeAuto => 'Auto';

  @override
  String get themeLight => 'Hell';

  @override
  String get themeDark => 'Dunkel';

  @override
  String get themeAutoDescription => 'Systemeinstellungen folgen';

  @override
  String get appearance => 'Erscheinungsbild';

  @override
  String get trainingDefaults => 'Standardwerte';

  @override
  String get defaultDuration => 'Trainingsdauer';

  @override
  String get warmupCooldown => 'Aufwärmen und Abkühlen';

  @override
  String get timer => 'Timer';

  @override
  String get intervalJingle => 'Pfeife bei Intervallwechsel';

  @override
  String get duckOtherAudio => 'Andere Audiowiedergabe leiser';

  @override
  String get duckOtherAudioDescription =>
      'Reduziert die Lautstärke von Musik und anderen Apps während des Pfiffs';

  @override
  String get liveTimerNotification => 'Live-Timer-Benachrichtigung';

  @override
  String get liveTimerNotificationDescription =>
      'Zeigt die verstrichene Zeit und Steuerungen in der Statusleiste während des Trainings an';

  @override
  String get goalHypertrophy => 'Muskelaufbau';

  @override
  String get goalHypertrophyDescription =>
      'Baue Muskelmasse mit gezieltem Krafttraining auf';

  @override
  String get goalFatLoss => 'Fettabbau';

  @override
  String get goalFatLossDescription =>
      'Verbrenne Kalorien und reduziere Körperfett mit intensiven Workouts';

  @override
  String get goalToning => 'Straffung';

  @override
  String get goalToningDescription =>
      'Entwickle definierte Muskeln und ein athletisches Erscheinungsbild';

  @override
  String get goalPosture => 'Haltung';

  @override
  String get goalPostureDescription =>
      'Stärke Rücken und Core für bessere Ausrichtung';

  @override
  String get goalRehabilitation => 'Rehabilitation';

  @override
  String get goalRehabilitationDescription =>
      'Sichere, kontrollierte Übungen zur Verletzungserholung';

  @override
  String get goalWellness => 'Wohlbefinden';

  @override
  String get goalWellnessDescription =>
      'Ausgewogene Workouts für Gesundheit und Stressabbau';

  @override
  String get goalFlexibility => 'Flexibilität';

  @override
  String get goalFlexibilityDescription =>
      'Verbessere die Beweglichkeit mit Dehnung und Mobilität';

  @override
  String get goalSports => 'Sportliche Leistung';

  @override
  String get goalSportsDescription =>
      'Steigere die athletische Fähigkeit mit Kraft- und Agilitätstraining';

  @override
  String get thisWeek => 'Diese Woche';

  @override
  String get trainingPlan => 'Trainingsplan';

  @override
  String get sessionsPerWeek => 'Einheiten pro Woche';

  @override
  String get sessionDuration => 'Einheitsdauer';

  @override
  String get preferredTime => 'Bevorzugte Zeit';

  @override
  String get recommendedTime => 'Empfohlene Zeit';

  @override
  String get methodologyMix => 'Methodenmix';

  @override
  String get pastWeeks => 'Vergangene Wochen';

  @override
  String get weeklyTarget => 'Wochenziel';

  @override
  String daysLeft(int count) {
    return '$count Tage übrig';
  }

  @override
  String get recommended => 'Empfohlen';

  @override
  String get duration => 'Dauer';

  @override
  String get familyHorizontalPush => 'Drücken';

  @override
  String get familyHorizontalPull => 'Ziehen';

  @override
  String get familyVerticalPush => 'Überkopf';

  @override
  String get familyVerticalPull => 'Klimmzug';

  @override
  String get familySquat => 'Kniebeuge';

  @override
  String get familyHinge => 'Hüftbeugung';

  @override
  String get familyCore => 'Core';

  @override
  String get familyCarry => 'Tragen';

  @override
  String get familyCardio => 'Cardio';

  @override
  String get familyMobility => 'Mobilität';

  @override
  String get familyBalance => 'Gleichgewicht';

  @override
  String get trainingQuality => 'Wie hat dir das Training gefallen?';

  @override
  String get trainingQualityHint =>
      'Hilft uns, die Qualität des generierten Trainings zu bewerten';

  @override
  String get qualityReasonHint => 'Was könnte verbessert werden?';

  @override
  String get good => 'Gut';

  @override
  String get bad => 'Schlecht';

  @override
  String get loadingMsg1 => 'Analyse deines Profils...';

  @override
  String get loadingMsg2 => 'Übungen auswählen...';

  @override
  String get loadingMsg3 => 'Routine wird erstellt...';

  @override
  String get loadingMsg4 => 'Trainingsvolumen berechnen...';

  @override
  String get loadingMsg5 => 'Pausenzeiten optimieren...';

  @override
  String get loadingMsg6 => 'Studien abgleichen...';

  @override
  String get loadingMsg7 => 'Muskelgruppen ausbalancieren...';

  @override
  String get loadingMsg8 => 'Progressionspfad erstellen...';

  @override
  String get loadingMsg9 => 'Intensität feinabstimmen...';

  @override
  String get loadingMsg10 => 'Bewegungsmuster prüfen...';

  @override
  String get loadingMsg11 => 'Erholungsbedarf einschätzen...';

  @override
  String get loadingMsg12 => 'Übungsvarianten auswählen...';

  @override
  String get loadingMsg13 => 'Trainingsblöcke strukturieren...';

  @override
  String get loadingMsg14 => 'Arbeitsintervalle festlegen...';

  @override
  String get loadingMsg15 => 'Sportwissenschaft anwenden...';

  @override
  String get loadingMsg16 => 'Aufwärmphase gestalten...';

  @override
  String get loadingMsg17 => 'Bewegungsfamilien zuordnen...';

  @override
  String get loadingMsg18 => 'Lastverteilung bewerten...';

  @override
  String get loadingMsg19 => 'Deine Session personalisieren...';

  @override
  String get loadingMsg20 => 'Fast fertig...';

  @override
  String loadingMsgGoal(String goal) {
    return 'Optimierung für $goal...';
  }

  @override
  String get loadingMsgInjuries => 'Anpassung an deine Verletzungen...';

  @override
  String get loadingMsgFavorites => 'Deine Lieblingsübungen priorisieren...';

  @override
  String get loadingMsgConditions => 'Anpassung an deine Bedingungen...';

  @override
  String loadingMsgMethodology(String methodology) {
    return '$methodology-Session gestalten...';
  }

  @override
  String get loadingMsgPartners => 'Partnertraining koordinieren...';

  @override
  String loadingMsgGym(String gym) {
    return '$gym-Geräte laden...';
  }

  @override
  String get loadingMsgHistory => 'Letzte Trainings analysieren...';

  @override
  String get loadingRetryMsg1 => 'hmm, das war nicht ganz richtig — nochmal';

  @override
  String get loadingRetryMsg2 => 'ich versuch\'s nochmal...';

  @override
  String get loadingRetryMsg3 => 'nicht ganz — noch ein Versuch';

  @override
  String get loadingRetryMsg4 => 'noch ein Versuch, Moment';

  @override
  String get loadingRetryMsg5 => 'oops, wird neu kalibriert...';

  @override
  String get loadingRetryMsg6 => 'fast gehabt — nochmal';

  @override
  String nSelected(int count) {
    return '$count ausgewählt';
  }

  @override
  String deleteSelectedTrainings(int count) {
    return '$count Trainings löschen?';
  }

  @override
  String get trainingsDeletedSuccessfully => 'Trainings erfolgreich gelöscht';

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
  String get pendingFeedbacks => 'Ausstehende Feedbacks';

  @override
  String get pendingFeedbacksDescription =>
      'Einige deiner Trainings sind als abgeschlossen markiert, aber du hast noch kein Feedback gegeben. Feedback hilft, deine zukünftigen Trainingsempfehlungen zu verbessern.';

  @override
  String get equipmentPartner => 'Trainingspartner';

  @override
  String get equipmentBalanceBoard => 'Balance Board';

  @override
  String get equipmentBand => 'Gummiband';

  @override
  String get equipmentBarbell => 'Langhantel';

  @override
  String get equipmentBench => 'Hantelbank';

  @override
  String get equipmentBox => 'Sprungbox';

  @override
  String get equipmentBosuBall => 'Bosu Ball';

  @override
  String get equipmentCable => 'Kabelzug';

  @override
  String get equipmentDipStation => 'Dipstation';

  @override
  String get equipmentDumbbell => 'Kurzhantel';

  @override
  String get equipmentEllipticalMachine => 'Crosstrainer';

  @override
  String get equipmentEzBarbell => 'EZ-Stange';

  @override
  String get equipmentHammer => 'Vorschlaghammer';

  @override
  String get equipmentKettlebell => 'Kettlebell';

  @override
  String get equipmentLeverageMachine => 'Hebelmaschine';

  @override
  String get equipmentMedicineBall => 'Medizinball';

  @override
  String get equipmentOlympicBarbell => 'Olympische Langhantel';

  @override
  String get equipmentPullUpBar => 'Klimmzugstange';

  @override
  String get equipmentResistanceBand => 'Widerstandsband';

  @override
  String get equipmentRings => 'Turnringe';

  @override
  String get equipmentRoller => 'Faszienrolle';

  @override
  String get equipmentRope => 'Seil';

  @override
  String get equipmentRowingMachine => 'Rudergerät';

  @override
  String get equipmentSkiergMachine => 'Skiergometer';

  @override
  String get equipmentSledMachine => 'Schlitten';

  @override
  String get equipmentSmithMachine => 'Multipresse';

  @override
  String get equipmentStabilityBall => 'Gymnastikball';

  @override
  String get equipmentStationaryBike => 'Heimtrainer';

  @override
  String get equipmentStepmillMachine => 'Treppensteiger';

  @override
  String get equipmentTire => 'Reifen';

  @override
  String get equipmentTreadmill => 'Laufband';

  @override
  String get equipmentTrapBar => 'Trap Bar';

  @override
  String get equipmentTrx => 'TRX';

  @override
  String get equipmentUpperBodyErgometer => 'Armergometer';

  @override
  String get equipmentWheelRoller => 'Bauchroller';

  @override
  String get muscleChest => 'Brust';

  @override
  String get muscleBack => 'Rücken';

  @override
  String get muscleShoulders => 'Schultern';

  @override
  String get muscleArms => 'Arme';

  @override
  String get muscleCore => 'Core';

  @override
  String get muscleGlutes => 'Gesäß';

  @override
  String get muscleLegs => 'Beine';

  @override
  String get modifierWeightedVest => 'Gewichtsweste';

  @override
  String get modifierParallettes => 'Mini-Barren';

  @override
  String get modifierAnkleWeights => 'Fußgewichte';

  @override
  String get modifierDipBelt => 'Dip-Gürtel';

  @override
  String get modifierPushUpBars => 'Liegestützgriffe';

  @override
  String get modifierResistanceBands => 'Widerstandsbänder';

  @override
  String get modifierWeight => 'Gewicht';

  @override
  String get modifierWristWeights => 'Handgelenkgewichte';

  @override
  String get healthData => 'Gesundheitsdaten';

  @override
  String get healthConnected => 'Verbunden';

  @override
  String get healthSynchronizing => 'Synchronisierung...';

  @override
  String get healthSynchronized => 'Synchronisiert';

  @override
  String get healthSynchronize => 'Jetzt synchronisieren';

  @override
  String get healthNotConnected => 'Nicht verbunden';

  @override
  String get healthNativeOnly => 'Verfügbar auf iOS und Android';

  @override
  String healthBuildingBaselines(int days) {
    return 'Erstelle deine persönlichen Referenzwerte ($days/14 Tage)';
  }

  @override
  String get healthSyncNoData =>
      'Keine Gesundheitsdaten auf dem Gerät gefunden';

  @override
  String get healthSyncFailed => 'Sync failed';

  @override
  String healthSyncFailedDetail(String error) {
    return 'Data read from device but upload failed: $error';
  }

  @override
  String get healthSyncTypeFull => 'Vollständig';

  @override
  String get healthSyncTypeIncremental => 'Inkrementell';

  @override
  String healthSourceData(int metrics, int sessions) {
    return '$metrics Metriken · $sessions Sitzungen';
  }

  @override
  String get healthBackend => 'Backend';

  @override
  String healthBackendData(int days, int sessions) {
    return '$days Tage · $sessions Sitzungen';
  }

  @override
  String healthDateRange(String from, String to) {
    return '$from – $to';
  }

  @override
  String get healthDisconnect => 'Trennen & Daten löschen';

  @override
  String get healthDisconnectConfirmation =>
      'Gesundheitsdaten trennen? Alle synchronisierten Daten werden gelöscht. Dies kann nicht rückgängig gemacht werden.';

  @override
  String get healthDisconnectedSuccessfully => 'Gesundheitsdaten getrennt';

  @override
  String get failedToDisconnectHealth =>
      'Gesundheitsdaten konnten nicht getrennt werden';

  @override
  String get healthConnect => 'Verbinden';

  @override
  String get healthMetrics => 'Gesundheitsdaten';

  @override
  String get healthAdjustment => 'Gesundheitsanpassung';

  @override
  String get healthPermissionsTitle => 'Wearable verbinden';

  @override
  String get healthPermissionsDescription =>
      'Vigor liest deine Gesundheitsdaten, um dein Training zu personalisieren. Bessere Schlaf-, Erholungs- und Aktivitätsdaten bedeuten intelligentere Trainingsempfehlungen.';

  @override
  String get healthPermissionsReadOnly =>
      'Größe & Gewicht mit Gesundheit synchronisiert';

  @override
  String get healthPermissionsSleep => 'Schlafdauer und -phasen';

  @override
  String get healthPermissionsHrv => 'Herzfrequenzvariabilität (HRV)';

  @override
  String get healthPermissionsRhr => 'Ruheherzfrequenz';

  @override
  String get healthPermissionsSteps => 'Tägliche Schritte';

  @override
  String get healthPermissionsWorkouts => 'Trainingseinheiten und Herzfrequenz';

  @override
  String get healthPermissionsGrant => 'Gesundheitsdaten verbinden';

  @override
  String get healthPermissionsSkip => 'Jetzt nicht';

  @override
  String get healthPermissionsGranted => 'Gesundheitsdaten verbunden';

  @override
  String get healthPermissionsDenied => 'Berechtigungen wurden nicht erteilt';

  @override
  String get healthOnboardingTitle => 'Wearable verbinden';

  @override
  String get healthOnboardingDescription =>
      'Verbinde deine Gesundheitsdaten und Vigor passt dein Training an Schlaf, Erholung und Bewegung an.';

  @override
  String get healthOnboardingConnect => 'Verbinden';

  @override
  String get healthOnboardingDismiss => 'Vielleicht später';

  @override
  String get healthInstallHcTitle => 'Health Connect erforderlich';

  @override
  String get healthInstallHcDescription =>
      'Health Connect wird benötigt, um deine Wearable-Daten zu synchronisieren. Installiere es aus dem Play Store.';

  @override
  String get healthInstallHc => 'Health Connect installieren';

  @override
  String get heartRate => 'Herzfrequenz';

  @override
  String get avgHr => 'Ø HF';

  @override
  String get maxHr => 'Max HF';

  @override
  String get bpm => 'bpm';

  @override
  String get hrZones => 'HF-Zonen';

  @override
  String get hrZone1 => 'Zone 1';

  @override
  String get hrZone2 => 'Zone 2';

  @override
  String get hrZone3 => 'Zone 3';

  @override
  String get hrZone4 => 'Zone 4';

  @override
  String get hrZone5 => 'Zone 5';

  @override
  String get healthDailySleep => 'Schlaf';

  @override
  String get healthDailyRestingHr => 'Ruhe-HF';

  @override
  String get healthDailyHrv => 'HRV';

  @override
  String get healthDailySteps => 'Schritte';

  @override
  String get healthDailyCalories => 'Kalorien';

  @override
  String get healthDailyNoData => 'Noch keine Gesundheitsdaten';

  @override
  String get externalWorkout => 'Externes Training';

  @override
  String get exerciseTypeRunning => 'Laufen';

  @override
  String get exerciseTypeWalking => 'Gehen';

  @override
  String get exerciseTypeBiking => 'Radfahren';

  @override
  String get exerciseTypeYoga => 'Yoga';

  @override
  String get exerciseTypeSwimming => 'Schwimmen';

  @override
  String get exerciseTypeHiking => 'Wandern';

  @override
  String get exerciseTypeStrengthTraining => 'Krafttraining';

  @override
  String get exerciseTypeFunctionalStrengthTraining => 'Funktionelles Training';

  @override
  String get exerciseTypeTraditionalStrengthTraining => 'Krafttraining';

  @override
  String get exerciseTypeRunningTreadmill => 'Laufband';

  @override
  String get exerciseTypeBikingStationary => 'Ergometer';

  @override
  String get exerciseTypeWalkingTreadmill => 'Laufband gehen';

  @override
  String get exerciseTypeRowing => 'Rudern';

  @override
  String get exerciseTypePilates => 'Pilates';

  @override
  String get exerciseTypeDancing => 'Tanzen';

  @override
  String get exerciseTypeElliptical => 'Crosstrainer';

  @override
  String get exerciseTypeStairClimbing => 'Treppensteigen';

  @override
  String get exerciseTypeOther => 'Training';

  @override
  String get appLogs => 'App-Protokolle';

  @override
  String get viewLogs => 'Protokolle anzeigen';

  @override
  String get exportLogs => 'Protokolle exportieren';

  @override
  String get clearLogs => 'Protokolle löschen';

  @override
  String get noLogsYet => 'Noch keine Protokolle vorhanden';

  @override
  String get logsCleared => 'Protokolle gelöscht';

  @override
  String logEntries(int count) {
    return '$count Protokolleinträge';
  }

  @override
  String get generateSession => 'Sitzung Erstellen';

  @override
  String get flowSession => 'Flow-Sitzung';

  @override
  String get flowSessionDescription =>
      'Yoga, Dehnung und Mobilität für Erholung und Wohlbefinden';

  @override
  String get loadingMsgFlow1 => 'Kürzlich trainierte Muskeln erfassen...';

  @override
  String get loadingMsgFlow2 => 'Deinen Erholungsflow gestalten...';

  @override
  String get loadingMsgFlow3 => 'Haltungen für deine Mobilität auswählen...';

  @override
  String get loadingMsgFlow4 => 'Verletzungssichere Bewegungen prüfen...';

  @override
  String get loadingMsgFlow5 =>
      'Deine achtsame Bewegungssequenz zusammenstellen...';

  @override
  String get noFlowSessionsYet => 'Noch keine Flow-Sitzungen';

  @override
  String get generateFirstFlow =>
      'Erstelle deine erste Flow-Sitzung, um mit Erholung und Dehnung zu beginnen.';
}
