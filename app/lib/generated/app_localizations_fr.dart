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
  String get favorites => 'Favoris';

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
  String get workoutTypeMobility => 'Mobilité';

  @override
  String get methodologyOptional => 'Méthodologie (optionnel)';

  @override
  String get methodologyAuto => 'Auto';

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
  String get trainingRoutines => 'Routines d\'Entraînement';

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
  String get references => 'Références';

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
  String get custom => 'Personnalisé';

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
  String get trainingPartnersOptional =>
      'Partenaires d\'Entraînement (optionnel)';

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
  String get instructions => 'Instructions';

  @override
  String get howWasYourTraining => 'Comment s\'est passé ton entraînement ?';

  @override
  String get anyAdditionalComments => 'Des commentaires supplémentaires ?';

  @override
  String get tooEasy => 'Trop facile';

  @override
  String get tooHard => 'Trop difficile';

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
  String get calibrationNeeded =>
      'Terminez votre premier entraînement pour que Vigor calibre les recommandations à votre niveau';

  @override
  String get calibrationDescription =>
      'Pendant le calibrage, la plateforme collecte des données de vos retours pour mieux adapter les entraînements à vos capacités';

  @override
  String get capabilities => 'Capacités';

  @override
  String get noTrainingsCompletedYet =>
      'Commence à t\'entraîner pour voir quelque chose';
}
