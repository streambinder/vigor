// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Italian (`it`).
class AppLocalizationsIt extends AppLocalizations {
  AppLocalizationsIt([String locale = 'it']) : super(locale);

  @override
  String get appName => 'Vigor';

  @override
  String get appTagline => 'Ex Sapientia Vis';

  @override
  String get navHome => 'Home';

  @override
  String get navActivity => 'Attività';

  @override
  String get navProfile => 'Profilo';

  @override
  String get storageErrorTitle => 'Vigor - Errore di archiviazione';

  @override
  String get storageUnavailable => 'Archiviazione non disponibile';

  @override
  String get storageErrorMessage =>
      'Richiesto archivio sicuro. Controlla le impostazioni.';

  @override
  String get signInWithGoogle => 'Accedi con Google';

  @override
  String get signingIn => 'Accesso in corso...';

  @override
  String get failedToInitializeGoogleSignIn =>
      'Inizializzazione di Google Sign In fallita';

  @override
  String signInError(String message) {
    return 'Errore di accesso: $message';
  }

  @override
  String get googleSignInFailed => 'Accesso con Google fallito';

  @override
  String get failedToGetAuthToken =>
      'Impossibile ottenere il token di autenticazione';

  @override
  String errorProcessingSignIn(String message) {
    return 'Errore durante l\'elaborazione dell\'accesso: $message';
  }

  @override
  String get googleSignInInitializing =>
      'Google Sign In è ancora in fase di inizializzazione...';

  @override
  String get readyToTrain => 'Pronto ad allenarti?';

  @override
  String get generateTrainingDescription =>
      'Crea un allenamento su misura per i tuoi obiettivi';

  @override
  String get generateTraining => 'Genera Allenamento';

  @override
  String get refresh => 'Aggiorna';

  @override
  String get logout => 'Esci';

  @override
  String get userDataRefreshed => 'Dati utente aggiornati';

  @override
  String get editProfile => 'Modifica Profilo';

  @override
  String get settings => 'Impostazioni';

  @override
  String get other => 'Altro';

  @override
  String get deleteGym => 'Elimina Palestra';

  @override
  String deleteGymConfirmation(String name) {
    return 'Eliminare \"$name\"?';
  }

  @override
  String get cancel => 'Annulla';

  @override
  String get delete => 'Elimina';

  @override
  String get logoutConfirmation => 'Uscire?';

  @override
  String get deleteAccount => 'Elimina Account';

  @override
  String get deleteAccountConfirmation =>
      'Eliminare l\'account? Non può essere annullato.';

  @override
  String get accountDeletedSuccessfully => 'Account eliminato con successo';

  @override
  String get failedToDeleteAccount => 'Impossibile eliminare l\'account';

  @override
  String get failedToLoadGyms => 'Impossibile caricare le palestre';

  @override
  String get gymAddedSuccessfully => 'Palestra aggiunta con successo';

  @override
  String get failedToAddGym => 'Impossibile aggiungere la palestra';

  @override
  String get gymUpdatedSuccessfully => 'Palestra aggiornata con successo';

  @override
  String get failedToUpdateGym => 'Impossibile aggiornare la palestra';

  @override
  String get gymDeletedSuccessfully => 'Palestra eliminata con successo';

  @override
  String get failedToDeleteGym => 'Impossibile eliminare la palestra';

  @override
  String get birthdate => 'Data di nascita';

  @override
  String get gender => 'Sesso';

  @override
  String get language => 'Lingua';

  @override
  String get height => 'Altezza';

  @override
  String get weight => 'Peso';

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
  String get goals => 'Obiettivi';

  @override
  String get injuries => 'Infortuni';

  @override
  String get limitations => 'Limitazioni';

  @override
  String get favorites => 'Preferiti';

  @override
  String get personalDetails => 'Dati Personali';

  @override
  String get healthAndGoals => 'Salute e Obiettivi';

  @override
  String get exercises => 'Esercizi';

  @override
  String get equipment => 'Attrezzatura';

  @override
  String startedDate(String date) {
    return 'Inizio: $date';
  }

  @override
  String yearLabel(int year) {
    return 'Anno: $year';
  }

  @override
  String get myGyms => 'Palestre';

  @override
  String get addGym => 'Aggiungi Palestra';

  @override
  String get noGymsAddedYet => 'Nessuna palestra aggiunta';

  @override
  String get addYourFirstGym => 'Aggiungi la Tua Prima Palestra';

  @override
  String get removeDefault => 'Rimuovi Predefinita';

  @override
  String get setAsDefault => 'Imposta come Predefinita';

  @override
  String get edit => 'Modifica';

  @override
  String get quickActions => 'Azioni Rapide';

  @override
  String get dangerZone => 'Zona Pericolosa';

  @override
  String get completeYourProfile => 'Completa il Tuo Profilo';

  @override
  String get updateYourProfileInfo => 'Aggiorna le tue info qui sotto.';

  @override
  String get pleaseCompleteProfile => 'Completa il profilo. * = obbligatorio.';

  @override
  String get firstName => 'Nome';

  @override
  String get lastName => 'Cognome';

  @override
  String get birthDate => 'Data di Nascita';

  @override
  String get male => 'Maschio';

  @override
  String get female => 'Femmina';

  @override
  String get selectLanguage => 'Seleziona lingua';

  @override
  String get heightCm => 'Altezza (cm)';

  @override
  String get weightKg => 'Peso (kg)';

  @override
  String get required => 'Obbligatorio';

  @override
  String get invalid => 'Non valido';

  @override
  String get pleaseSelectBirthDate =>
      'Per favore seleziona la tua data di nascita';

  @override
  String get pleaseAddAtLeastOneGoal =>
      'Per favore aggiungi almeno un obiettivo';

  @override
  String get pleaseSelectLanguage => 'Per favore seleziona la tua lingua';

  @override
  String get addAGoal => 'Aggiungi un obiettivo';

  @override
  String get injuryDescription => 'Descrizione infortunio';

  @override
  String get year => 'Anno';

  @override
  String get addALimitation => 'Aggiungi una limitazione';

  @override
  String get favoriteExercisesHint => 'es. squat, trazioni, corsa';

  @override
  String get favoriteEquipmentHint => 'es. manubri, bilanciere, kettlebell';

  @override
  String get saveChanges => 'Salva Modifiche';

  @override
  String get saveProfile => 'Salva Profilo';

  @override
  String get optionalLeaveEmpty => '(Opzionale)';

  @override
  String get optionalExercisesPrefer => '(Opzionale)';

  @override
  String get optionalEquipmentPrefer => '(Opzionale)';

  @override
  String get optionalWorkoutTypesPrefer => '(Opzionale)';

  @override
  String get favoriteExercises => 'Esercizi Preferiti';

  @override
  String get favoriteEquipment => 'Attrezzatura Preferita';

  @override
  String get favoriteWorkoutTypes => 'Tipi di allenamento';

  @override
  String get workoutTypeStrength => 'Forza';

  @override
  String get workoutTypeCircuit => 'Circuito';

  @override
  String get workoutTypeEmom => 'EMOM';

  @override
  String get workoutTypeAmrap => 'AMRAP';

  @override
  String get workoutTypeHiit => 'HIIT';

  @override
  String get workoutTypeForTime => 'For Time';

  @override
  String get workoutTypeEndurance => 'Resistenza';

  @override
  String get workoutTypeMobility => 'Mobilità';

  @override
  String get methodologyOptional => 'Metodologia (opzionale)';

  @override
  String get methodologyAuto => 'Auto';

  @override
  String get goalsOptional => 'Obiettivi (opzionale)';

  @override
  String get musclesOptional => 'Muscoli (opzionale)';

  @override
  String get musclesAuto => 'Auto';

  @override
  String get advancedSettings => 'Avanzate';

  @override
  String get failedToUpdateProfile => 'Impossibile aggiornare il profilo';

  @override
  String get activity => 'Attività';

  @override
  String get noTrainingsYet => 'Nessun allenamento';

  @override
  String get generateFirstTraining => 'Crea il primo allenamento da Home';

  @override
  String get noTrainingAvailable => 'Nessun allenamento. Generane uno.';

  @override
  String get availableTrainings => 'Allenamenti disponibili';

  @override
  String get pastTrainings => 'Allenamenti passati';

  @override
  String get stale => 'Obsoleto';

  @override
  String get today => 'Oggi';

  @override
  String get yesterday => 'Ieri';

  @override
  String get available => 'Disponibili';

  @override
  String get completed => 'Completato';

  @override
  String get completedSingular => 'Completato';

  @override
  String get noPastTrainings => 'Nessun allenamento completato';

  @override
  String get copied => 'Copiato';

  @override
  String durationMin(int minutes) {
    return '$minutes min';
  }

  @override
  String durationHr(int hours) {
    return '$hours ora';
  }

  @override
  String durationHrMin(int hours, int minutes) {
    return '$hours ore $minutes min';
  }

  @override
  String get failedToLoadTrainings => 'Impossibile caricare gli allenamenti';

  @override
  String get startTraining => 'Inizia Allenamento';

  @override
  String get cloneTraining => 'Clona Allenamento';

  @override
  String get addPartner => 'Aggiungi Partner';

  @override
  String get shareWithUser => 'Condividi con Utente';

  @override
  String get deleteTraining => 'Elimina Allenamento';

  @override
  String get leaveTraining => 'Abbandona Allenamento';

  @override
  String get showAiReasoning => 'Mostra ragionamento IA';

  @override
  String get reportIssue => 'Segnala problema';

  @override
  String deleteTrainingConfirmation(String name) {
    return 'Eliminare \"$name\"? Non può essere annullato.';
  }

  @override
  String leaveTrainingConfirmation(String name) {
    return 'Abbandonare \"$name\"? Non lo vedrai più.';
  }

  @override
  String addPartnerConfirmation(String userName, String trainingName) {
    return 'Aggiungere $userName come partner a \"$trainingName\"?';
  }

  @override
  String cloneTrainingConfirmation(String name) {
    return 'Clonare \"$name\" nei tuoi allenamenti?';
  }

  @override
  String shareTrainingConfirmation(String trainingName, String userName) {
    return 'Condividere \"$trainingName\" con $userName?';
  }

  @override
  String get trainingDeletedSuccessfully =>
      'Allenamento eliminato con successo';

  @override
  String get failedToDeleteTraining => 'Impossibile eliminare l\'allenamento';

  @override
  String get leftTrainingSuccessfully => 'Allenamento abbandonato con successo';

  @override
  String get partnerAddedSuccessfully => 'Partner aggiunto con successo';

  @override
  String get failedToAddPartner => 'Impossibile aggiungere il partner';

  @override
  String get trainingSharedSuccessfully => 'Allenamento condiviso con successo';

  @override
  String get failedToShareTraining => 'Impossibile condividere l\'allenamento';

  @override
  String get trainingCloned => 'Allenamento clonato';

  @override
  String get failedToCloneTraining => 'Impossibile clonare l\'allenamento';

  @override
  String get trainingMarkedAsComplete => 'Allenamento completato';

  @override
  String get failedToCompleteTraining =>
      'Impossibile completare l\'allenamento';

  @override
  String get feedbackUpdated => 'Feedback aggiornato';

  @override
  String get failedToUpdateFeedback => 'Impossibile aggiornare il feedback';

  @override
  String get reportSubmitted => 'Segnalazione inviata';

  @override
  String get failedToSubmitReport => 'Impossibile inviare la segnalazione';

  @override
  String get shuffleExercise => 'Cambia esercizio';

  @override
  String get exerciseShuffled => 'Esercizio cambiato';

  @override
  String get failedToShuffleExercise => 'Impossibile cambiare esercizio';

  @override
  String get reasoning => 'Ragionamento';

  @override
  String get strategy => 'Strategia';

  @override
  String get typeSelection => 'Selezione Tipo';

  @override
  String get progression => 'Progressione';

  @override
  String get constraints => 'Vincoli';

  @override
  String get researchApplied => 'Ricerche Applicate';

  @override
  String get targetMuscles => 'Muscoli Target';

  @override
  String get naming => 'Denominazione';

  @override
  String get trainingRoutines => 'Allenamento';

  @override
  String get noEquipment => 'Nessuna attrezzatura';

  @override
  String blockNumber(int number) {
    return 'Blocco $number';
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
    return '${seconds}s riposo';
  }

  @override
  String repsCount(int count) {
    return '$count ripetizioni';
  }

  @override
  String weightKgValue(double value) {
    return '$value kg';
  }

  @override
  String get markAsComplete => 'Segna come Completato';

  @override
  String get updateFeedback => 'Aggiorna Feedback';

  @override
  String get references => 'Riferimenti';

  @override
  String get literature => 'Letteratura';

  @override
  String get describeIssue => 'Descrivi il problema...';

  @override
  String get submit => 'Invia';

  @override
  String get close => 'Chiudi';

  @override
  String get add => 'Aggiungi';

  @override
  String get update => 'Aggiorna';

  @override
  String get clone => 'Clona';

  @override
  String get share => 'Condividi';

  @override
  String get leave => 'Abbandona';

  @override
  String get tapToStart => 'Tocca per iniziare';

  @override
  String get trainingCompleted => 'Allenamento Completato!';

  @override
  String greatJobCompleting(String name) {
    return 'Ottimo lavoro con $name';
  }

  @override
  String get done => 'Fatto';

  @override
  String get complete => 'Completa';

  @override
  String routineCounter(int current, int total) {
    return 'Routine $current/$total';
  }

  @override
  String blockCounter(int current, int total) {
    return 'Blocco $current/$total';
  }

  @override
  String get exitTraining => 'Uscire dall\'Allenamento?';

  @override
  String get whatWouldYouLikeToDo => 'Cosa fare?';

  @override
  String get exit => 'Esci';

  @override
  String get continueTraining => 'Continua';

  @override
  String get failedToMarkComplete => 'Impossibile completare l\'allenamento';

  @override
  String get durationMinutes => 'Durata (minuti)';

  @override
  String get bodyweight => 'Corpo libero';

  @override
  String get gym => 'Palestra';

  @override
  String get custom => 'Altro';

  @override
  String get noEquipmentBodyweightOnly => 'Solo corpo libero';

  @override
  String get noGymsDefinedCreateOne =>
      'Nessuna palestra. Creane una nel profilo.';

  @override
  String get selectAGym => 'Seleziona una palestra';

  @override
  String get addEquipment => 'Aggiungi Attrezzatura';

  @override
  String get addEquipmentAvailable => 'Aggiungi l\'attrezzatura disponibile';

  @override
  String get includeWarmupCooldown => 'Riscaldamento e defaticamento';

  @override
  String get equipmentPlaceholder => 'es. Bilanciere, Manubri';

  @override
  String get customPromptOptional => 'Prompt Personalizzato (opzionale)';

  @override
  String get focusOnUpperBody =>
      'es. Concentrati sulla parte superiore del corpo';

  @override
  String get trainingPartnersOptional => 'Partner';

  @override
  String get generatingTraining => 'Generazione allenamento in corso...';

  @override
  String get thisMayTakeAMoment => 'Potrebbe richiedere qualche momento';

  @override
  String generationFailedRetrying(int attempt) {
    return 'Generazione fallita, nuovo tentativo #$attempt...';
  }

  @override
  String get trainingGeneratedSuccessfully =>
      'Allenamento generato con successo!';

  @override
  String get failedToGenerateTraining => 'Impossibile generare l\'allenamento';

  @override
  String get generate => 'Genera';

  @override
  String get editGym => 'Modifica Palestra';

  @override
  String get gymName => 'Nome Palestra';

  @override
  String get gymNamePlaceholder => 'es. Palestra Casa, Virgin Active';

  @override
  String get noEquipmentAddedYet => 'Nessuna attrezzatura aggiunta';

  @override
  String get pleaseEnterGymName =>
      'Per favore inserisci il nome della palestra';

  @override
  String get addAllEquipment => 'Aggiungi Tutto';

  @override
  String get failedToLoadEquipment => 'Impossibile caricare le attrezzature';

  @override
  String get selectUser => 'Seleziona Utente';

  @override
  String get searchByName => 'Cerca per nome';

  @override
  String get noUsersAvailable => 'Nessun utente disponibile';

  @override
  String get noMatchingUsers => 'Nessun utente corrispondente';

  @override
  String get instructions => 'Istruzioni';

  @override
  String get howWasYourTraining => 'Com\'è andato l\'allenamento?';

  @override
  String get anyAdditionalComments => 'Commenti aggiuntivi?';

  @override
  String get tooEasy => 'Troppo facile';

  @override
  String get tooHard => 'Troppo difficile';

  @override
  String get flag => 'Segnala';

  @override
  String get profile => 'Profilo';

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
    return 'Prossimo: $name';
  }

  @override
  String get rest => 'Riposo';

  @override
  String get upcoming => 'In arrivo';

  @override
  String get yourProgress => 'I tuoi progressi';

  @override
  String trainingsCompleted(int count) {
    return '$count allenamenti completati';
  }

  @override
  String get completedTrainings => 'allenamenti completati';

  @override
  String get partneredTrainings => 'allenamenti in coppia';

  @override
  String get movementFamilies => 'Famiglie di movimento';

  @override
  String get muscleActivity => 'Attività muscolare';

  @override
  String get failedToLoadProgress => 'Impossibile caricare i progressi';

  @override
  String get noProgressYet =>
      'Completa gli allenamenti per vedere i tuoi progressi';

  @override
  String get calibration => 'Calibrazione';

  @override
  String get calibrationGlobal => 'Globale';

  @override
  String get calibrationNeeded =>
      'Completa il tuo primo allenamento perché Vigor possa calibrare le raccomandazioni al tuo livello';

  @override
  String get calibrationDescription =>
      'Durante la calibrazione, la piattaforma raccoglie dati dai tuoi feedback per fare un primo assessment sul tuo stato di forma e il tuo livello';

  @override
  String get capabilities => 'Capacità';

  @override
  String get noTrainingsCompletedYet =>
      'Inizia ad allenarti per vedere qualcosa qui';

  @override
  String get theme => 'Tema';

  @override
  String get themeAuto => 'Auto';

  @override
  String get themeLight => 'Chiaro';

  @override
  String get themeDark => 'Scuro';

  @override
  String get themeAutoDescription => 'Segui le impostazioni di sistema';

  @override
  String get appearance => 'Aspetto';

  @override
  String get trainingDefaults => 'Predefiniti';

  @override
  String get defaultDuration => 'Durata Allenamento';
}
