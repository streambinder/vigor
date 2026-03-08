// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for French (`fr`).
class AppLocalizationsFr extends AppLocalizations {
  AppLocalizationsFr([String locale = 'fr']) : super(locale);

  @override
  String get appName => 'Vigor';

  @override
  String get appTagline => 'Ex Sapientia Vis';

  @override
  String get navHome => 'Accueil';

  @override
  String get navActivity => 'Activité';

  @override
  String get navProfile => 'Profil';

  @override
  String get storageErrorTitle => 'Vigor - Erreur de stockage';

  @override
  String get storageUnavailable => 'Stockage non disponible';

  @override
  String get storageErrorMessage =>
      'Stockage sécurisé requis. Vérifie les paramètres.';

  @override
  String get signInWithGoogle => 'Se connecter avec Google';

  @override
  String get signingIn => 'Connexion en cours...';

  @override
  String get failedToInitializeGoogleSignIn =>
      'Échec de l\'initialisation de Google Sign In';

  @override
  String signInError(String message) {
    return 'Erreur de connexion : $message';
  }

  @override
  String get googleSignInFailed => 'Échec de la connexion Google';

  @override
  String get failedToGetAuthToken =>
      'Impossible d\'obtenir le jeton d\'authentification';

  @override
  String errorProcessingSignIn(String message) {
    return 'Erreur lors du traitement de la connexion : $message';
  }

  @override
  String get googleSignInInitializing =>
      'Google Sign In est encore en cours d\'initialisation...';

  @override
  String get readyToTrain => 'Prêt à t\'entraîner ?';

  @override
  String get generateTrainingDescription =>
      'Crée un entraînement adapté à tes objectifs';

  @override
  String get generateTraining => 'Générer un Entraînement';

  @override
  String get refresh => 'Actualiser';

  @override
  String get logout => 'Déconnexion';

  @override
  String get userDataRefreshed => 'Données utilisateur actualisées';

  @override
  String get editProfile => 'Modifier le Profil';

  @override
  String get settings => 'Paramètres';

  @override
  String get other => 'Autre';

  @override
  String get deleteGym => 'Supprimer la Salle';

  @override
  String deleteGymConfirmation(String name) {
    return 'Supprimer \"$name\" ?';
  }

  @override
  String get cancel => 'Annuler';

  @override
  String get delete => 'Supprimer';

  @override
  String get logoutConfirmation => 'Se déconnecter ?';

  @override
  String get deleteAccount => 'Supprimer le Compte';

  @override
  String get deleteAccountConfirmation =>
      'Supprimer ton compte ? Irréversible.';

  @override
  String get accountDeletedSuccessfully => 'Compte supprimé avec succès';

  @override
  String get failedToDeleteAccount => 'Échec de la suppression du compte';

  @override
  String get failedToLoadGyms => 'Échec du chargement des salles';

  @override
  String get gymAddedSuccessfully => 'Salle ajoutée avec succès';

  @override
  String get failedToAddGym => 'Échec de l\'ajout de la salle';

  @override
  String get gymUpdatedSuccessfully => 'Salle mise à jour avec succès';

  @override
  String get failedToUpdateGym => 'Échec de la mise à jour de la salle';

  @override
  String get gymDeletedSuccessfully => 'Salle supprimée avec succès';

  @override
  String get failedToDeleteGym => 'Échec de la suppression de la salle';

  @override
  String get birthdate => 'Date de naissance';

  @override
  String get gender => 'Sexe';

  @override
  String get language => 'Langue';

  @override
  String get height => 'Taille';

  @override
  String get weight => 'Poids';

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
  String get goals => 'Objectifs';

  @override
  String get injuries => 'Blessures';

  @override
  String get limitations => 'Limitations';

  @override
  String get conditions => 'Conditions';

  @override
  String get favorites => 'Favoris';

  @override
  String get personalDetails => 'Informations Personnelles';

  @override
  String get healthAndGoals => 'Santé et Objectifs';

  @override
  String get exercises => 'Exercices';

  @override
  String get equipment => 'Équipement';

  @override
  String startedDate(String date) {
    return 'Début : $date';
  }

  @override
  String yearLabel(int year) {
    return 'Année : $year';
  }

  @override
  String get myGyms => 'Salles';

  @override
  String get addGym => 'Ajouter une Salle';

  @override
  String get noGymsAddedYet => 'Aucune salle ajoutée';

  @override
  String get addYourFirstGym => 'Ajoute Ta Première Salle';

  @override
  String get removeDefault => 'Retirer par Défaut';

  @override
  String get setAsDefault => 'Définir par Défaut';

  @override
  String get edit => 'Modifier';

  @override
  String get quickActions => 'Actions Rapides';

  @override
  String get dangerZone => 'Zone Dangereuse';

  @override
  String get completeYourProfile => 'Complète Ton Profil';

  @override
  String get updateYourProfileInfo => 'Mets à jour ton profil ci-dessous.';

  @override
  String get pleaseCompleteProfile => 'Complète ton profil. * = obligatoire.';

  @override
  String get firstName => 'Prénom';

  @override
  String get lastName => 'Nom';

  @override
  String get birthDate => 'Date de Naissance';

  @override
  String get male => 'Homme';

  @override
  String get female => 'Femme';

  @override
  String get selectLanguage => 'Sélectionner la langue';

  @override
  String get heightCm => 'Taille (cm)';

  @override
  String get weightKg => 'Poids (kg)';

  @override
  String get required => 'Obligatoire';

  @override
  String get invalid => 'Non valide';

  @override
  String get pleaseSelectBirthDate =>
      'Merci de sélectionner ta date de naissance';

  @override
  String get pleaseAddAtLeastOneGoal => 'Merci d\'ajouter au moins un objectif';

  @override
  String get pleaseSelectLanguage => 'Merci de sélectionner ta langue';

  @override
  String get addAGoal => 'Ajouter un objectif';

  @override
  String get injuryDescription => 'Description de la blessure';

  @override
  String get year => 'Année';

  @override
  String get addALimitation => 'Ajouter une limitation';

  @override
  String get addACondition => 'Ajouter une condition';

  @override
  String get favoriteExercisesHint => 'ex. squats, tractions, course';

  @override
  String get favoriteEquipmentHint => 'ex. haltères, barre, kettlebell';

  @override
  String get saveChanges => 'Enregistrer les Modifications';

  @override
  String get saveProfile => 'Enregistrer le Profil';

  @override
  String get optionalLeaveEmpty => '(Optionnel)';

  @override
  String get optionalExercisesPrefer => '(Optionnel)';

  @override
  String get optionalEquipmentPrefer => '(Optionnel)';

  @override
  String get optionalWorkoutTypesPrefer => '(Optionnel)';

  @override
  String get favoriteExercises => 'Exercices Favoris';

  @override
  String get favoriteEquipment => 'Équipement Favori';

  @override
  String get favoriteWorkoutTypes => 'Types d\'Entraînement Préférés';

  @override
  String get workoutTypeStrength => 'Force';

  @override
  String get workoutTypeStrengthDescription =>
      'Développez la force maximale avec des charges lourdes et repos complet';

  @override
  String get workoutTypeSupersets => 'Supersets';

  @override
  String get workoutTypeSupersetsDescription =>
      'Enchaînez muscles opposés pour un entraînement rapide et efficace';

  @override
  String get workoutTypeCircuit => 'Circuit';

  @override
  String get workoutTypeCircuitDescription =>
      'Passez d\'une station à l\'autre avec un repos minimal pour le conditionnement';

  @override
  String get workoutTypeEmom => 'EMOM';

  @override
  String get workoutTypeEmomDescription =>
      'Chaque minute: complétez les répétitions puis reposez jusqu\'à la minute suivante';

  @override
  String get workoutTypeAmrap => 'AMRAP';

  @override
  String get workoutTypeAmrapDescription =>
      'Autant de tours que possible dans le temps imparti';

  @override
  String get workoutTypeHiit => 'HIIT';

  @override
  String get workoutTypeHiitDescription =>
      'Alternez des rafales de haute intensité avec une récupération courte';

  @override
  String get workoutTypeForTime => 'For Time';

  @override
  String get workoutTypeForTimeDescription =>
      'Terminez l\'entraînement le plus rapidement possible';

  @override
  String get workoutTypeEndurance => 'Endurance';

  @override
  String get workoutTypeEnduranceDescription =>
      'Effort soutenu à intensité modérée pour la capacité aérobie';

  @override
  String get workoutTypeMobility => 'Mobilité';

  @override
  String get workoutTypeMobilityDescription =>
      'Améliorez l\'amplitude de mouvement et la santé articulaire';

  @override
  String get methodologyOptional => 'Méthodologie (optionnel)';

  @override
  String get methodologyAuto => 'Auto';

  @override
  String get goalsOptional => 'Objectifs (optionnel)';

  @override
  String get musclesOptional => 'Muscles (optionnel)';

  @override
  String get musclesAuto => 'Tous';

  @override
  String get advancedSettings => 'Avancé';

  @override
  String get failedToUpdateProfile => 'Échec de la mise à jour du profil';

  @override
  String get activity => 'Activité';

  @override
  String get noTrainingsYet => 'Aucun entraînement';

  @override
  String get generateFirstTraining =>
      'Crée ton premier entraînement depuis Accueil';

  @override
  String get noTrainingAvailable => 'Aucun entraînement. Génère-en un.';

  @override
  String get availableTrainings => 'Entraînements disponibles';

  @override
  String get pastTrainings => 'Entraînements passés';

  @override
  String get stale => 'Obsolète';

  @override
  String get today => 'Aujourd\'hui';

  @override
  String get yesterday => 'Hier';

  @override
  String daysAgo(int count) {
    return 'il y a ${count}j';
  }

  @override
  String get available => 'Disponible';

  @override
  String get completed => 'Terminé';

  @override
  String get completedSingular => 'Terminé';

  @override
  String get noPastTrainings => 'Aucun entraînement terminé pour l\'instant';

  @override
  String get copied => 'Copié';

  @override
  String durationMin(int minutes) {
    return '$minutes min';
  }

  @override
  String durationHr(int hours) {
    return '$hours h';
  }

  @override
  String durationHrMin(int hours, int minutes) {
    return '$hours h $minutes min';
  }

  @override
  String get failedToLoadTrainings => 'Échec du chargement des entraînements';

  @override
  String get startTraining => 'Démarrer l\'Entraînement';

  @override
  String get cloneTraining => 'Cloner l\'Entraînement';

  @override
  String get addPartner => 'Ajouter un Partenaire';

  @override
  String get shareWithUser => 'Partager avec un Utilisateur';

  @override
  String get deleteTraining => 'Supprimer l\'Entraînement';

  @override
  String get leaveTraining => 'Quitter l\'Entraînement';

  @override
  String get showAiReasoning => 'Afficher le raisonnement IA';

  @override
  String get reportIssue => 'Signaler un problème';

  @override
  String deleteTrainingConfirmation(String name) {
    return 'Supprimer \"$name\" ? Irréversible.';
  }

  @override
  String leaveTrainingConfirmation(String name) {
    return 'Quitter \"$name\" ? Tu ne le verras plus.';
  }

  @override
  String addPartnerConfirmation(String userName, String trainingName) {
    return 'Ajouter $userName comme partenaire à \"$trainingName\" ?';
  }

  @override
  String cloneTrainingConfirmation(String name) {
    return 'Cloner \"$name\" dans vos entraînements ?';
  }

  @override
  String shareTrainingConfirmation(String trainingName, String userName) {
    return 'Partager \"$trainingName\" avec $userName ?';
  }

  @override
  String get trainingDeletedSuccessfully => 'Entraînement supprimé avec succès';

  @override
  String get failedToDeleteTraining =>
      'Échec de la suppression de l\'entraînement';

  @override
  String get leftTrainingSuccessfully => 'Entraînement quitté avec succès';

  @override
  String get partnerAddedSuccessfully => 'Partenaire ajouté avec succès';

  @override
  String get failedToAddPartner => 'Échec de l\'ajout du partenaire';

  @override
  String get trainingSharedSuccessfully => 'Entraînement partagé avec succès';

  @override
  String get failedToShareTraining => 'Échec du partage de l\'entraînement';

  @override
  String get trainingCloned => 'Entraînement cloné';

  @override
  String get failedToCloneTraining => 'Échec du clonage de l\'entraînement';

  @override
  String get trainingMarkedAsComplete => 'Entraînement terminé';

  @override
  String get failedToCompleteTraining =>
      'Échec de la complétion de l\'entraînement';

  @override
  String get feedback => 'Feedback';

  @override
  String get feedbackUpdated => 'Feedback mis à jour';

  @override
  String get failedToUpdateFeedback => 'Échec de la mise à jour du feedback';

  @override
  String get reportSubmitted => 'Signalement envoyé';

  @override
  String get failedToSubmitReport => 'Échec de l\'envoi du signalement';

  @override
  String get shuffleExercise => 'Changer d\'exercice';

  @override
  String get exerciseShuffled => 'Exercice changé';

  @override
  String get failedToShuffleExercise => 'Impossible de changer d\'exercice';

  @override
  String get reasoning => 'Raisonnement';

  @override
  String get strategy => 'Stratégie';

  @override
  String get typeSelection => 'Sélection du Type';

  @override
  String get progression => 'Progression';

  @override
  String get constraints => 'Contraintes';

  @override
  String get researchApplied => 'Recherche Appliquée';

  @override
  String get targetMuscles => 'Muscles Ciblés';

  @override
  String get naming => 'Dénomination';

  @override
  String get trainingRoutines => 'Entraînement';

  @override
  String get noEquipment => 'Sans équipement';

  @override
  String blockNumber(int number) {
    return 'Bloc $number';
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
    return '${seconds}s repos';
  }

  @override
  String repsCount(int count) {
    return '$count répétitions';
  }

  @override
  String weightKgValue(double value) {
    return '$value kg';
  }

  @override
  String get markAsComplete => 'Marquer comme Terminé';

  @override
  String get updateFeedback => 'Modifier le Feedback';

  @override
  String get references => 'Références';

  @override
  String get literature => 'Littérature';

  @override
  String get request => 'Request';

  @override
  String get describeIssue => 'Décris le problème...';

  @override
  String get submit => 'Envoyer';

  @override
  String get close => 'Fermer';

  @override
  String get add => 'Ajouter';

  @override
  String get update => 'Mettre à jour';

  @override
  String get clone => 'Cloner';

  @override
  String get share => 'Partager';

  @override
  String get leave => 'Quitter';

  @override
  String get tapToStart => 'Appuie pour commencer';

  @override
  String get tapWhenDone => 'Appuie quand c\'est fait';

  @override
  String get trainingCompleted => 'Entraînement Terminé !';

  @override
  String greatJobCompleting(String name) {
    return 'Excellent travail pour $name';
  }

  @override
  String get done => 'Terminé';

  @override
  String get complete => 'Terminer';

  @override
  String routineCounter(int current, int total) {
    return 'Routine $current/$total';
  }

  @override
  String blockCounter(int current, int total) {
    return 'Bloc $current/$total';
  }

  @override
  String get exitTraining => 'Quitter l\'Entraînement ?';

  @override
  String get whatWouldYouLikeToDo => 'Que faire ?';

  @override
  String get exit => 'Quitter';

  @override
  String get continueTraining => 'Continuer';

  @override
  String get failedToMarkComplete =>
      'Échec de la complétion de l\'entraînement';

  @override
  String get durationMinutes => 'Durée (minutes)';

  @override
  String get bodyweight => 'Poids du corps';

  @override
  String get gym => 'Salle';

  @override
  String get custom => 'Autre';

  @override
  String get noEquipmentBodyweightOnly => 'Poids du corps seulement';

  @override
  String get noGymsDefinedCreateOne =>
      'Aucune salle. Crée-en une dans le profil.';

  @override
  String get selectAGym => 'Sélectionner une salle';

  @override
  String get addEquipment => 'Ajouter de l\'Équipement';

  @override
  String get addEquipmentAvailable => 'Ajoute ton équipement disponible';

  @override
  String get includeWarmupCooldown => 'Inclure échauffement et récupération';

  @override
  String get equipmentPlaceholder => 'ex. Barre, Haltères';

  @override
  String get customPromptOptional => 'Prompt Personnalisé (optionnel)';

  @override
  String get focusOnUpperBody => 'ex. Se concentrer sur le haut du corps';

  @override
  String get trainingPartnersOptional => 'Partner';

  @override
  String get generatingTraining => 'Génération de ton entraînement...';

  @override
  String get thisMayTakeAMoment => 'Cela peut prendre un moment';

  @override
  String generationFailedRetrying(int attempt) {
    return 'Génération échouée, nouvel essai #$attempt...';
  }

  @override
  String get trainingGeneratedSuccessfully =>
      'Entraînement généré avec succès !';

  @override
  String get failedToGenerateTraining =>
      'Échec de la génération de l\'entraînement';

  @override
  String get generate => 'Générer';

  @override
  String get editGym => 'Modifier la Salle';

  @override
  String get gymName => 'Nom de la Salle';

  @override
  String get gymNamePlaceholder => 'ex. Salle Maison, Basic Fit';

  @override
  String get availableWeights => 'Poids Disponibles';

  @override
  String get availableWeightsHint =>
      'Configurez les options de poids pour les modificateurs lestés dans cette salle.';

  @override
  String get noEquipmentAddedYet => 'Aucun équipement ajouté';

  @override
  String get pleaseEnterGymName => 'Merci d\'entrer le nom de la salle';

  @override
  String get addAllEquipment => 'Tout Ajouter';

  @override
  String get failedToLoadEquipment => 'Impossible de charger l\'équipement';

  @override
  String get selectUser => 'Sélectionner un Utilisateur';

  @override
  String get searchByName => 'Rechercher par nom';

  @override
  String get noUsersAvailable => 'Aucun utilisateur disponible';

  @override
  String get noMatchingUsers => 'Aucun utilisateur correspondant';

  @override
  String get noMatchingEquipment => 'Aucun équipement correspondant';

  @override
  String get instructions => 'Instructions';

  @override
  String get cues => 'Conseils';

  @override
  String get howWasYourTraining => 'Comment s\'est passé ton entraînement ?';

  @override
  String get anyAdditionalComments => 'Des commentaires supplémentaires ?';

  @override
  String get actualDuration => 'Durée réelle (min)';

  @override
  String get impossible => 'Impossible';

  @override
  String get tooHard => 'Trop difficile';

  @override
  String get ok => 'OK';

  @override
  String get easy => 'Facile';

  @override
  String get tooEasy => 'Trop facile';

  @override
  String get flag => 'Signaler';

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
    return 'Suivant: $name';
  }

  @override
  String get rest => 'Repos';

  @override
  String get upcoming => 'À venir';

  @override
  String get yourProgress => 'Vos progrès';

  @override
  String trainingsCompleted(int count) {
    return '$count entraînements terminés';
  }

  @override
  String get completedTrainings => 'entraînements terminés';

  @override
  String get partneredTrainings => 'entraînements en duo';

  @override
  String get movementFamilies => 'Familles de mouvements';

  @override
  String get muscleActivity => 'Activité musculaire';

  @override
  String get failedToLoadProgress => 'Impossible de charger les progrès';

  @override
  String get noProgressYet =>
      'Terminez des entraînements pour voir vos progrès';

  @override
  String get calibration => 'Calibrage';

  @override
  String get calibrationGlobal => 'Global';

  @override
  String get calibrationNeeded =>
      'Terminez votre premier entraînement pour que Vigor calibre les recommandations à votre niveau';

  @override
  String get calibrationDescription =>
      'Pendant le calibrage, la plateforme collecte des données de vos retours pour faire une première évaluation de votre condition physique et niveau';

  @override
  String get calibrationInProgress =>
      'Vos entraînements deviennent plus intelligents à chaque séance';

  @override
  String get calibrationTrainingNote =>
      'Cet entraînement peut ne pas correspondre entièrement à vos objectifs — le système apprend encore votre niveau et privilégie la variété des mouvements pour construire un profil complet';

  @override
  String calibrationFamiliesLearned(int calibrated, int total) {
    return '$calibrated/$total schémas moteurs appris';
  }

  @override
  String get capabilities => 'Capacités';

  @override
  String get muscleHeatMap => 'Carte de Chaleur Musculaire';

  @override
  String get heatResting => 'Au repos';

  @override
  String get heatRecovered => 'Récupéré';

  @override
  String get heatActive => 'Actif';

  @override
  String get heatWarm => 'Chaud';

  @override
  String get heatHot => 'Intense';

  @override
  String get noTrainingsCompletedYet =>
      'Commence à t\'entraîner pour voir quelque chose';

  @override
  String get theme => 'Thème';

  @override
  String get themeAuto => 'Auto';

  @override
  String get themeLight => 'Clair';

  @override
  String get themeDark => 'Sombre';

  @override
  String get themeAutoDescription => 'Suivre les paramètres système';

  @override
  String get appearance => 'Apparence';

  @override
  String get trainingDefaults => 'Paramètres par défaut';

  @override
  String get defaultDuration => 'Durée de l\'Entraînement';

  @override
  String get warmupCooldown => 'Échauffement et récupération';

  @override
  String get timer => 'Minuteur';

  @override
  String get intervalJingle => 'Sifflet au changement d\'intervalle';

  @override
  String get duckOtherAudio => 'Baisser l\'autre audio';

  @override
  String get duckOtherAudioDescription =>
      'Réduit le volume de la musique et des autres apps pendant le sifflet';

  @override
  String get goalHypertrophy => 'Prise de Masse';

  @override
  String get goalHypertrophyDescription =>
      'Développe la masse musculaire avec un entraînement ciblé';

  @override
  String get goalFatLoss => 'Perte de Graisse';

  @override
  String get goalFatLossDescription =>
      'Brûle des calories et réduis la graisse avec des séances intensives';

  @override
  String get goalToning => 'Tonification';

  @override
  String get goalToningDescription =>
      'Développe des muscles définis et une apparence athlétique';

  @override
  String get goalPosture => 'Posture';

  @override
  String get goalPostureDescription =>
      'Renforce le dos et le core pour un meilleur alignement';

  @override
  String get goalRehabilitation => 'Rééducation';

  @override
  String get goalRehabilitationDescription =>
      'Exercices sûrs et contrôlés pour la récupération';

  @override
  String get goalWellness => 'Bien-être';

  @override
  String get goalWellnessDescription =>
      'Entraînements équilibrés pour la santé et la détente';

  @override
  String get goalFlexibility => 'Souplesse';

  @override
  String get goalFlexibilityDescription =>
      'Améliore l\'amplitude de mouvement avec étirements et mobilité';

  @override
  String get goalSports => 'Performance Sportive';

  @override
  String get goalSportsDescription =>
      'Améliore les capacités athlétiques avec puissance et agilité';

  @override
  String get thisWeek => 'Cette Semaine';

  @override
  String get trainingPlan => 'Plan d\'Entraînement';

  @override
  String get sessionsPerWeek => 'Séances par semaine';

  @override
  String get sessionDuration => 'Durée de séance';

  @override
  String get preferredTime => 'Horaire préféré';

  @override
  String get recommendedTime => 'Horaire recommandé';

  @override
  String get methodologyMix => 'Mix de méthodologies';

  @override
  String get pastWeeks => 'Semaines passées';

  @override
  String get weeklyTarget => 'Objectif Hebdomadaire';

  @override
  String daysLeft(int count) {
    return '$count jours restants';
  }

  @override
  String get recommended => 'Recommandé';

  @override
  String get duration => 'Durée';

  @override
  String get familyHorizontalPush => 'Poussée';

  @override
  String get familyHorizontalPull => 'Tirage';

  @override
  String get familyVerticalPush => 'Développé';

  @override
  String get familyVerticalPull => 'Traction';

  @override
  String get familySquat => 'Squat';

  @override
  String get familyHinge => 'Charnière';

  @override
  String get familyCore => 'Core';

  @override
  String get familyCarry => 'Portage';

  @override
  String get familyCardio => 'Cardio';

  @override
  String get familyMobility => 'Mobilité';

  @override
  String get familyBalance => 'Équilibre';

  @override
  String get trainingQuality => 'Comment avez-vous trouvé cet entraînement ?';

  @override
  String get trainingQualityHint =>
      'Nous aide à évaluer la qualité de l\'entraînement généré';

  @override
  String get qualityReasonHint => 'Qu\'est-ce qui pourrait être amélioré ?';

  @override
  String get good => 'Bon';

  @override
  String get bad => 'Mauvais';

  @override
  String get loadingMsg1 => 'Analyse de votre profil...';

  @override
  String get loadingMsg2 => 'Sélection des exercices...';

  @override
  String get loadingMsg3 => 'Construction de votre routine...';

  @override
  String get loadingMsg4 => 'Calcul du volume d\'entraînement...';

  @override
  String get loadingMsg5 => 'Optimisation des temps de repos...';

  @override
  String get loadingMsg6 => 'Consultation des études...';

  @override
  String get loadingMsg7 => 'Équilibrage des groupes musculaires...';

  @override
  String get loadingMsg8 => 'Création du parcours de progression...';

  @override
  String get loadingMsg9 => 'Réglage de l\'intensité...';

  @override
  String get loadingMsg10 => 'Révision des schémas de mouvement...';

  @override
  String get loadingMsg11 => 'Évaluation des besoins de récupération...';

  @override
  String get loadingMsg12 => 'Choix des variantes d\'exercice...';

  @override
  String get loadingMsg13 => 'Structuration des blocs...';

  @override
  String get loadingMsg14 => 'Définition des intervalles...';

  @override
  String get loadingMsg15 => 'Application des sciences du sport...';

  @override
  String get loadingMsg16 => 'Conception de l\'échauffement...';

  @override
  String get loadingMsg17 => 'Mapping des familles de mouvement...';

  @override
  String get loadingMsg18 => 'Évaluation de la répartition des charges...';

  @override
  String get loadingMsg19 => 'Personnalisation de votre séance...';

  @override
  String get loadingMsg20 => 'Presque prêt...';

  @override
  String loadingMsgGoal(String goal) {
    return 'Optimisation pour $goal...';
  }

  @override
  String get loadingMsgInjuries => 'Adaptation à vos blessures...';

  @override
  String get loadingMsgFavorites => 'Priorisation de vos exercices préférés...';

  @override
  String get loadingMsgConditions => 'Adaptation à vos conditions...';

  @override
  String loadingMsgMethodology(String methodology) {
    return 'Conception d\'une séance $methodology...';
  }

  @override
  String get loadingMsgPartners => 'Coordination de l\'entraînement à deux...';

  @override
  String loadingMsgGym(String gym) {
    return 'Chargement de l\'équipement de $gym...';
  }

  @override
  String get loadingMsgHistory => 'Analyse de vos séances récentes...';

  @override
  String get loadingRetryMsg1 =>
      'hmm, ce n\'est pas sorti comme prévu — on réessaie';

  @override
  String get loadingRetryMsg2 => 'laissez-moi réessayer...';

  @override
  String get loadingRetryMsg3 => 'pas tout à fait — encore un essai';

  @override
  String get loadingRetryMsg4 => 'encore un essai, un instant';

  @override
  String get loadingRetryMsg5 => 'oops, recalibrage...';

  @override
  String get loadingRetryMsg6 => 'j\'y étais presque — on réessaie';

  @override
  String nSelected(int count) {
    return '$count sélectionnés';
  }

  @override
  String deleteSelectedTrainings(int count) {
    return 'Supprimer $count entraînements ?';
  }

  @override
  String get trainingsDeletedSuccessfully =>
      'Entraînements supprimés avec succès';

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
  String get pendingFeedbacks => 'Feedbacks en attente';

  @override
  String get pendingFeedbacksDescription =>
      'Certains de vos entraînements sont marqués comme terminés mais n\'ont pas encore votre feedback. Le feedback aide à améliorer les recommandations pour vos futurs entraînements.';

  @override
  String get equipmentPartner => 'Partenaire';

  @override
  String get equipmentBalanceBoard => 'Planche d\'Équilibre';

  @override
  String get equipmentBand => 'Bande Élastique';

  @override
  String get equipmentBarbell => 'Barre';

  @override
  String get equipmentBench => 'Banc';

  @override
  String get equipmentBox => 'Box';

  @override
  String get equipmentBosuBall => 'Bosu Ball';

  @override
  String get equipmentCable => 'Poulie';

  @override
  String get equipmentDipStation => 'Barres à Dips';

  @override
  String get equipmentDumbbell => 'Haltère';

  @override
  String get equipmentEllipticalMachine => 'Vélo Elliptique';

  @override
  String get equipmentEzBarbell => 'Barre EZ';

  @override
  String get equipmentHammer => 'Masse';

  @override
  String get equipmentKettlebell => 'Kettlebell';

  @override
  String get equipmentLeverageMachine => 'Machine à Levier';

  @override
  String get equipmentMedicineBall => 'Médecine Ball';

  @override
  String get equipmentOlympicBarbell => 'Barre Olympique';

  @override
  String get equipmentPullUpBar => 'Barre de Traction';

  @override
  String get equipmentResistanceBand => 'Bande de Résistance';

  @override
  String get equipmentRings => 'Anneaux';

  @override
  String get equipmentRoller => 'Rouleau';

  @override
  String get equipmentRope => 'Corde';

  @override
  String get equipmentRowingMachine => 'Rameur';

  @override
  String get equipmentSkiergMachine => 'SkiErg';

  @override
  String get equipmentSledMachine => 'Traîneau';

  @override
  String get equipmentSmithMachine => 'Machine Smith';

  @override
  String get equipmentStabilityBall => 'Ballon Suisse';

  @override
  String get equipmentStationaryBike => 'Vélo d\'Appartement';

  @override
  String get equipmentStepmillMachine => 'Escalier Infini';

  @override
  String get equipmentTire => 'Pneu';

  @override
  String get equipmentTreadmill => 'Tapis de Course';

  @override
  String get equipmentTrapBar => 'Barre Hexagonale';

  @override
  String get equipmentTrx => 'TRX';

  @override
  String get equipmentUpperBodyErgometer => 'Ergomètre à Bras';

  @override
  String get equipmentWheelRoller => 'Roue Abdominale';

  @override
  String get muscleChest => 'Poitrine';

  @override
  String get muscleBack => 'Dos';

  @override
  String get muscleShoulders => 'Épaules';

  @override
  String get muscleArms => 'Bras';

  @override
  String get muscleCore => 'Core';

  @override
  String get muscleGlutes => 'Fessiers';

  @override
  String get muscleLegs => 'Jambes';

  @override
  String get modifierWeightedVest => 'Gilet Lesté';

  @override
  String get modifierParallettes => 'Parallettes';

  @override
  String get modifierAnkleWeights => 'Poids de Cheville';

  @override
  String get modifierDipBelt => 'Ceinture de Lest';

  @override
  String get modifierPushUpBars => 'Poignées de Pompes';

  @override
  String get modifierResistanceBands => 'Bandes de Résistance';

  @override
  String get modifierWeight => 'Poids';

  @override
  String get modifierWristWeights => 'Lests de poignet';

  @override
  String get healthData => 'Données de santé';

  @override
  String get healthConnected => 'Connecté';

  @override
  String get healthSynchronizing => 'Synchronisation...';

  @override
  String get healthSynchronized => 'Synchronisé';

  @override
  String get healthSynchronize => 'Synchroniser maintenant';

  @override
  String get healthNotConnected => 'Non connecté';

  @override
  String healthBuildingBaselines(int days) {
    return 'Création de vos références personnelles ($days/14 jours)';
  }

  @override
  String get healthDisconnect => 'Déconnecter et supprimer les données';

  @override
  String get healthDisconnectConfirmation =>
      'Déconnecter les données de santé ? Toutes les données synchronisées seront supprimées. Cette action est irréversible.';

  @override
  String get healthDisconnectedSuccessfully => 'Données de santé déconnectées';

  @override
  String get failedToDisconnectHealth =>
      'Échec de la déconnexion des données de santé';

  @override
  String get healthConnect => 'Connecter';

  @override
  String get healthMetrics => 'Données Santé';

  @override
  String get healthAdjustment => 'Ajustement santé';

  @override
  String get healthPermissionsTitle => 'Connectez votre wearable';

  @override
  String get healthPermissionsDescription =>
      'Vigor lit vos données de santé pour personnaliser vos entraînements. De meilleures données de sommeil, récupération et activité permettent des recommandations plus intelligentes.';

  @override
  String get healthPermissionsReadOnly => 'Accès en lecture seule';

  @override
  String get healthPermissionsSleep => 'Durée et phases du sommeil';

  @override
  String get healthPermissionsHrv =>
      'Variabilité de la fréquence cardiaque (VFC)';

  @override
  String get healthPermissionsRhr => 'Fréquence cardiaque au repos';

  @override
  String get healthPermissionsSteps => 'Pas quotidiens';

  @override
  String get healthPermissionsWorkouts =>
      'Séances d\'entraînement et fréquence cardiaque';

  @override
  String get healthPermissionsGrant => 'Connecter les données de santé';

  @override
  String get healthPermissionsSkip => 'Pas maintenant';

  @override
  String get healthPermissionsGranted => 'Données de santé connectées';

  @override
  String get healthPermissionsDenied =>
      'Les autorisations n\'ont pas été accordées';

  @override
  String get healthOnboardingTitle => 'Connectez votre wearable';

  @override
  String get healthOnboardingDescription =>
      'Connectez vos données de santé et Vigor adaptera votre entraînement à votre sommeil, récupération et activité.';

  @override
  String get healthOnboardingConnect => 'Connecter';

  @override
  String get healthOnboardingDismiss => 'Peut-être plus tard';

  @override
  String get healthInstallHcTitle => 'Health Connect requis';

  @override
  String get healthInstallHcDescription =>
      'Health Connect est nécessaire pour synchroniser les données de votre wearable. Installez-le depuis le Play Store.';

  @override
  String get healthInstallHc => 'Installer Health Connect';

  @override
  String get heartRate => 'Fréquence cardiaque';

  @override
  String get avgHr => 'FC moy';

  @override
  String get maxHr => 'FC max';

  @override
  String get bpm => 'bpm';

  @override
  String get hrZones => 'Zones FC';

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
  String get healthDailySleep => 'Sommeil';

  @override
  String get healthDailyRestingHr => 'FC au repos';

  @override
  String get healthDailyHrv => 'VFC';

  @override
  String get healthDailySteps => 'Pas';

  @override
  String get healthDailyCalories => 'Calories';

  @override
  String get healthDailyNoData => 'Pas encore de données santé';

  @override
  String get externalWorkout => 'Entraînement externe';
}
