// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Spanish Castilian (`es`).
class AppLocalizationsEs extends AppLocalizations {
  AppLocalizationsEs([String locale = 'es']) : super(locale);

  @override
  String get appName => 'Vigor';

  @override
  String get appTagline => 'Ex Sapientia Vis';

  @override
  String get navHome => 'Inicio';

  @override
  String get navActivity => 'Actividad';

  @override
  String get navProfile => 'Perfil';

  @override
  String get storageErrorTitle => 'Vigor - Error de almacenamiento';

  @override
  String get storageUnavailable => 'Almacenamiento no disponible';

  @override
  String get storageErrorMessage =>
      'Se requiere almacenamiento seguro. Verifica la configuración.';

  @override
  String get signInWithGoogle => 'Iniciar sesión con Google';

  @override
  String get signingIn => 'Iniciando sesión...';

  @override
  String get failedToInitializeGoogleSignIn =>
      'Error al inicializar Google Sign In';

  @override
  String signInError(String message) {
    return 'Error de inicio de sesión: $message';
  }

  @override
  String get googleSignInFailed => 'Inicio de sesión con Google fallido';

  @override
  String get failedToGetAuthToken =>
      'No se pudo obtener el token de autenticación';

  @override
  String errorProcessingSignIn(String message) {
    return 'Error al procesar el inicio de sesión: $message';
  }

  @override
  String get googleSignInInitializing =>
      'Google Sign In todavía se está inicializando...';

  @override
  String get readyToTrain => '¿Listo para entrenar?';

  @override
  String get generateTrainingDescription =>
      'Crea un entrenamiento adaptado a tus objetivos';

  @override
  String get generateTraining => 'Generar Entrenamiento';

  @override
  String get refresh => 'Actualizar';

  @override
  String get logout => 'Cerrar sesión';

  @override
  String get userDataRefreshed => 'Datos de usuario actualizados';

  @override
  String get editProfile => 'Editar Perfil';

  @override
  String get settings => 'Configuración';

  @override
  String get deleteGym => 'Eliminar Gimnasio';

  @override
  String deleteGymConfirmation(String name) {
    return '¿Eliminar \"$name\"?';
  }

  @override
  String get cancel => 'Cancelar';

  @override
  String get delete => 'Eliminar';

  @override
  String get logoutConfirmation => '¿Cerrar sesión?';

  @override
  String get deleteAccount => 'Eliminar Cuenta';

  @override
  String get deleteAccountConfirmation =>
      '¿Eliminar tu cuenta? No se puede deshacer.';

  @override
  String get accountDeletedSuccessfully => 'Cuenta eliminada exitosamente';

  @override
  String get failedToDeleteAccount => 'Error al eliminar la cuenta';

  @override
  String get failedToLoadGyms => 'Error al cargar gimnasios';

  @override
  String get gymAddedSuccessfully => 'Gimnasio agregado exitosamente';

  @override
  String get failedToAddGym => 'Error al agregar gimnasio';

  @override
  String get gymUpdatedSuccessfully => 'Gimnasio actualizado exitosamente';

  @override
  String get failedToUpdateGym => 'Error al actualizar gimnasio';

  @override
  String get gymDeletedSuccessfully => 'Gimnasio eliminado exitosamente';

  @override
  String get failedToDeleteGym => 'Error al eliminar gimnasio';

  @override
  String get birthdate => 'Fecha de nacimiento';

  @override
  String get gender => 'Género';

  @override
  String get language => 'Idioma';

  @override
  String get height => 'Altura';

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
  String get goals => 'Objetivos';

  @override
  String get injuries => 'Lesiones';

  @override
  String get limitations => 'Limitaciones';

  @override
  String get favorites => 'Favoritos';

  @override
  String get exercises => 'Ejercicios';

  @override
  String get equipment => 'Equipamiento';

  @override
  String startedDate(String date) {
    return 'Inicio: $date';
  }

  @override
  String yearLabel(int year) {
    return 'Año: $year';
  }

  @override
  String get myGyms => 'Gimnasios';

  @override
  String get addGym => 'Agregar Gimnasio';

  @override
  String get noGymsAddedYet => 'No hay gimnasios agregados';

  @override
  String get addYourFirstGym => 'Agrega Tu Primer Gimnasio';

  @override
  String get removeDefault => 'Quitar Predeterminado';

  @override
  String get setAsDefault => 'Establecer como Predeterminado';

  @override
  String get edit => 'Editar';

  @override
  String get quickActions => 'Acciones Rápidas';

  @override
  String get dangerZone => 'Zona de Peligro';

  @override
  String get completeYourProfile => 'Completa Tu Perfil';

  @override
  String get updateYourProfileInfo => 'Actualiza tu info a continuación.';

  @override
  String get pleaseCompleteProfile => 'Completa tu perfil. * = obligatorio.';

  @override
  String get firstName => 'Nombre';

  @override
  String get lastName => 'Apellido';

  @override
  String get birthDate => 'Fecha de Nacimiento';

  @override
  String get male => 'Masculino';

  @override
  String get female => 'Femenino';

  @override
  String get selectLanguage => 'Seleccionar idioma';

  @override
  String get heightCm => 'Altura (cm)';

  @override
  String get weightKg => 'Peso (kg)';

  @override
  String get required => 'Obligatorio';

  @override
  String get invalid => 'No válido';

  @override
  String get pleaseSelectBirthDate =>
      'Por favor selecciona tu fecha de nacimiento';

  @override
  String get pleaseAddAtLeastOneGoal => 'Por favor agrega al menos un objetivo';

  @override
  String get pleaseSelectLanguage => 'Por favor selecciona tu idioma';

  @override
  String get addAGoal => 'Agregar un objetivo';

  @override
  String get injuryDescription => 'Descripción de la lesión';

  @override
  String get year => 'Año';

  @override
  String get addALimitation => 'Agregar una limitación';

  @override
  String get favoriteExercisesHint => 'ej. sentadillas, dominadas, correr';

  @override
  String get favoriteEquipmentHint => 'ej. mancuernas, barra, pesa rusa';

  @override
  String get saveChanges => 'Guardar Cambios';

  @override
  String get saveProfile => 'Guardar Perfil';

  @override
  String get optionalLeaveEmpty => '(Opcional)';

  @override
  String get optionalExercisesPrefer => '(Opcional)';

  @override
  String get optionalEquipmentPrefer => '(Opcional)';

  @override
  String get optionalWorkoutTypesPrefer => '(Opcional)';

  @override
  String get favoriteExercises => 'Ejercicios Favoritos';

  @override
  String get favoriteEquipment => 'Equipamiento Favorito';

  @override
  String get favoriteWorkoutTypes => 'Tipos de Entrenamiento Preferidos';

  @override
  String get workoutTypeStrength => 'Fuerza';

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
  String get workoutTypeEndurance => 'Resistencia';

  @override
  String get workoutTypeMobility => 'Movilidad';

  @override
  String get methodologyOptional => 'Metodología (opcional)';

  @override
  String get methodologyAuto => 'Auto';

  @override
  String get failedToUpdateProfile => 'Error al actualizar el perfil';

  @override
  String get activity => 'Actividad';

  @override
  String get noTrainingsYet => 'No hay entrenamientos';

  @override
  String get generateFirstTraining =>
      'Crea tu primer entrenamiento desde Inicio';

  @override
  String get noTrainingAvailable => 'Sin entrenamientos. Genera uno.';

  @override
  String get availableTrainings => 'Entrenamientos disponibles';

  @override
  String get pastTrainings => 'Entrenamientos pasados';

  @override
  String get stale => 'Obsoleto';

  @override
  String get copied => 'Copiado';

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
  String get failedToLoadTrainings => 'Error al cargar entrenamientos';

  @override
  String get startTraining => 'Iniciar Entrenamiento';

  @override
  String get cloneTraining => 'Clonar Entrenamiento';

  @override
  String get addPartner => 'Agregar Compañero';

  @override
  String get shareWithUser => 'Compartir con Usuario';

  @override
  String get deleteTraining => 'Eliminar Entrenamiento';

  @override
  String get leaveTraining => 'Abandonar Entrenamiento';

  @override
  String get showAiReasoning => 'Mostrar razonamiento IA';

  @override
  String get reportIssue => 'Reportar problema';

  @override
  String deleteTrainingConfirmation(String name) {
    return '¿Eliminar \"$name\"? No se puede deshacer.';
  }

  @override
  String leaveTrainingConfirmation(String name) {
    return '¿Abandonar \"$name\"? Ya no lo verás.';
  }

  @override
  String addPartnerConfirmation(String userName, String trainingName) {
    return '¿Agregar a $userName como compañero a \"$trainingName\"?';
  }

  @override
  String cloneTrainingConfirmation(String name) {
    return '¿Clonar \"$name\" a tus entrenamientos?';
  }

  @override
  String shareTrainingConfirmation(String trainingName, String userName) {
    return '¿Compartir \"$trainingName\" con $userName?';
  }

  @override
  String get trainingDeletedSuccessfully =>
      'Entrenamiento eliminado exitosamente';

  @override
  String get failedToDeleteTraining => 'Error al eliminar el entrenamiento';

  @override
  String get leftTrainingSuccessfully =>
      'Entrenamiento abandonado exitosamente';

  @override
  String get partnerAddedSuccessfully => 'Compañero agregado exitosamente';

  @override
  String get failedToAddPartner => 'Error al agregar compañero';

  @override
  String get trainingSharedSuccessfully =>
      'Entrenamiento compartido exitosamente';

  @override
  String get failedToShareTraining => 'Error al compartir el entrenamiento';

  @override
  String get trainingCloned => 'Entrenamiento clonado';

  @override
  String get failedToCloneTraining => 'Error al clonar el entrenamiento';

  @override
  String get trainingMarkedAsComplete => 'Entrenamiento completado';

  @override
  String get failedToCompleteTraining => 'Error al completar el entrenamiento';

  @override
  String get reportSubmitted => 'Reporte enviado';

  @override
  String get failedToSubmitReport => 'Error al enviar el reporte';

  @override
  String get shuffleExercise => 'Cambiar ejercicio';

  @override
  String get exerciseShuffled => 'Ejercicio cambiado';

  @override
  String get failedToShuffleExercise => 'Error al cambiar ejercicio';

  @override
  String get reasoning => 'Razonamiento';

  @override
  String get strategy => 'Estrategia';

  @override
  String get typeSelection => 'Selección de Tipo';

  @override
  String get progression => 'Progresión';

  @override
  String get constraints => 'Restricciones';

  @override
  String get researchApplied => 'Investigación Aplicada';

  @override
  String get targetMuscles => 'Músculos Objetivo';

  @override
  String get naming => 'Denominación';

  @override
  String get trainingRoutines => 'Rutinas de Entrenamiento';

  @override
  String get noEquipment => 'Sin equipamiento';

  @override
  String blockNumber(int number) {
    return 'Bloque $number';
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
    return '${seconds}s descanso';
  }

  @override
  String repsCount(int count) {
    return '$count repeticiones';
  }

  @override
  String weightKgValue(double value) {
    return '$value kg';
  }

  @override
  String get markAsComplete => 'Marcar como Completado';

  @override
  String get references => 'Referencias';

  @override
  String get describeIssue => 'Describe el problema...';

  @override
  String get submit => 'Enviar';

  @override
  String get close => 'Cerrar';

  @override
  String get add => 'Agregar';

  @override
  String get update => 'Actualizar';

  @override
  String get clone => 'Clonar';

  @override
  String get share => 'Compartir';

  @override
  String get leave => 'Abandonar';

  @override
  String get tapToStart => 'Toca para comenzar';

  @override
  String get trainingCompleted => '¡Entrenamiento Completado!';

  @override
  String greatJobCompleting(String name) {
    return '¡Excelente trabajo completando $name';
  }

  @override
  String get done => 'Hecho';

  @override
  String get complete => 'Completar';

  @override
  String routineCounter(int current, int total) {
    return 'Rutina $current/$total';
  }

  @override
  String blockCounter(int current, int total) {
    return 'Bloque $current/$total';
  }

  @override
  String get exitTraining => '¿Salir del Entrenamiento?';

  @override
  String get whatWouldYouLikeToDo => '¿Qué hacer?';

  @override
  String get exit => 'Salir';

  @override
  String get continueTraining => 'Continuar';

  @override
  String get failedToMarkComplete => 'Error al completar el entrenamiento';

  @override
  String get durationMinutes => 'Duración (minutos)';

  @override
  String get bodyweight => 'Peso corporal';

  @override
  String get gym => 'Gimnasio';

  @override
  String get custom => 'Personalizado';

  @override
  String get noEquipmentBodyweightOnly => 'Solo peso corporal';

  @override
  String get noGymsDefinedCreateOne => 'Sin gimnasios. Crea uno en el perfil.';

  @override
  String get selectAGym => 'Seleccionar un gimnasio';

  @override
  String get addEquipment => 'Agregar Equipamiento';

  @override
  String get addEquipmentAvailable => 'Agrega tu equipamiento disponible';

  @override
  String get includeWarmupCooldown => 'Incluir calentamiento y enfriamiento';

  @override
  String get equipmentPlaceholder => 'ej. Barra, Mancuernas';

  @override
  String get customPromptOptional => 'Prompt Personalizado (opcional)';

  @override
  String get focusOnUpperBody =>
      'ej. Enfocarse en la parte superior del cuerpo';

  @override
  String get trainingPartnersOptional =>
      'Compañeros de Entrenamiento (opcional)';

  @override
  String get generatingTraining => 'Generando tu entrenamiento...';

  @override
  String get thisMayTakeAMoment => 'Esto puede tomar un momento';

  @override
  String get trainingGeneratedSuccessfully =>
      '¡Entrenamiento generado exitosamente!';

  @override
  String get failedToGenerateTraining => 'Error al generar el entrenamiento';

  @override
  String get generate => 'Generar';

  @override
  String get editGym => 'Editar Gimnasio';

  @override
  String get gymName => 'Nombre del Gimnasio';

  @override
  String get gymNamePlaceholder => 'ej. Gimnasio Casa, McFit';

  @override
  String get noEquipmentAddedYet => 'No se ha agregado equipamiento';

  @override
  String get pleaseEnterGymName => 'Por favor ingresa el nombre del gimnasio';

  @override
  String get addAllEquipment => 'Agregar Todo';

  @override
  String get failedToLoadEquipment => 'Error al cargar equipamiento';

  @override
  String get selectUser => 'Seleccionar Usuario';

  @override
  String get searchByName => 'Buscar por nombre';

  @override
  String get noUsersAvailable => 'No hay usuarios disponibles';

  @override
  String get noMatchingUsers => 'No hay usuarios coincidentes';

  @override
  String get instructions => 'Instrucciones';

  @override
  String get howWasYourTraining => '¿Cómo estuvo tu entrenamiento?';

  @override
  String get anyAdditionalComments => '¿Algún comentario adicional?';

  @override
  String get tooEasy => 'Muy fácil';

  @override
  String get tooHard => 'Muy difícil';

  @override
  String get flag => 'Marcar';

  @override
  String get profile => 'Perfil';

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
    return 'Siguiente: $name';
  }

  @override
  String get rest => 'Descanso';

  @override
  String get upcoming => 'Próximos';

  @override
  String get yourProgress => 'Tu progreso';

  @override
  String trainingsCompleted(int count) {
    return '$count entrenamientos completados';
  }

  @override
  String get completedTrainings => 'entrenamientos completados';

  @override
  String get movementFamilies => 'Familias de movimiento';

  @override
  String get muscleActivity => 'Actividad muscular';

  @override
  String get failedToLoadProgress => 'No se pudo cargar el progreso';

  @override
  String get noProgressYet => 'Completa entrenamientos para ver tu progreso';

  @override
  String get calibration => 'Calibración';

  @override
  String get calibrationNeeded =>
      'Completa tu primer entrenamiento para que Vigor calibre las recomendaciones a tu nivel';

  @override
  String get calibrationDescription =>
      'Durante la calibración, la plataforma recopila datos de tus comentarios para alinear mejor los entrenamientos con tus capacidades';

  @override
  String get capabilities => 'Capacidades';

  @override
  String get noTrainingsCompletedYet => 'Empieza a entrenar para ver algo aquí';
}
