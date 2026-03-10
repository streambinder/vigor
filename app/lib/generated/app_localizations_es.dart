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
  String get other => 'Otros';

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
  String get conditions => 'Condiciones';

  @override
  String get favorites => 'Favoritos';

  @override
  String get personalDetails => 'Datos Personales';

  @override
  String get healthAndGoals => 'Salud y Objetivos';

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
  String get addACondition => 'Agregar una condición';

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
  String get workoutTypeStrengthDescription =>
      'Desarrolla fuerza máxima con cargas pesadas y descanso completo';

  @override
  String get workoutTypeSupersets => 'Superseries';

  @override
  String get workoutTypeSupersetsDescription =>
      'Combina músculos opuestos consecutivamente para un entrenamiento eficiente';

  @override
  String get workoutTypeCircuit => 'Circuito';

  @override
  String get workoutTypeCircuitDescription =>
      'Pasa por estaciones con mínimo descanso para acondicionamiento';

  @override
  String get workoutTypeEmom => 'EMOM';

  @override
  String get workoutTypeEmomDescription =>
      'Cada minuto: completa las repeticiones y descansa hasta el siguiente minuto';

  @override
  String get workoutTypeAmrap => 'AMRAP';

  @override
  String get workoutTypeAmrapDescription =>
      'Tantas rondas como sea posible dentro del tiempo límite';

  @override
  String get workoutTypeHiit => 'HIIT';

  @override
  String get workoutTypeHiitDescription =>
      'Alterna ráfagas de alta intensidad con recuperación corta';

  @override
  String get workoutTypeForTime => 'For Time';

  @override
  String get workoutTypeForTimeDescription =>
      'Completa el entrenamiento lo más rápido posible';

  @override
  String get workoutTypeEndurance => 'Resistencia';

  @override
  String get workoutTypeEnduranceDescription =>
      'Esfuerzo sostenido a intensidad moderada para capacidad aeróbica';

  @override
  String get workoutTypeMobility => 'Movilidad';

  @override
  String get workoutTypeMobilityDescription =>
      'Mejora el rango de movimiento y la salud articular';

  @override
  String get methodologyOptional => 'Metodología (opcional)';

  @override
  String get methodologyAuto => 'Auto';

  @override
  String get goalsOptional => 'Objetivos (opcional)';

  @override
  String get musclesOptional => 'Músculos (opcional)';

  @override
  String get musclesAuto => 'Todos';

  @override
  String get advancedSettings => 'Avanzado';

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
  String get today => 'Hoy';

  @override
  String get yesterday => 'Ayer';

  @override
  String daysAgo(int count) {
    return 'hace ${count}d';
  }

  @override
  String get available => 'Disponible';

  @override
  String get completed => 'Completado';

  @override
  String get completedSingular => 'Completado';

  @override
  String get noPastTrainings => 'Aún no hay entrenamientos completados';

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
  String get feedback => 'Feedback';

  @override
  String get feedbackUpdated => 'Feedback actualizado';

  @override
  String get failedToUpdateFeedback => 'Error al actualizar el feedback';

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
  String get trainingRoutines => 'Entrenamiento';

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
  String get updateFeedback => 'Actualizar Feedback';

  @override
  String get references => 'Referencias';

  @override
  String get literature => 'Literatura';

  @override
  String get request => 'Request';

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
  String get tapWhenDone => 'Toca cuando termines';

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
  String get stop => 'Detener';

  @override
  String get stopTraining => '¿Detener Entrenamiento?';

  @override
  String get stopTrainingConfirm => 'Se perderá el progreso del temporizador.';

  @override
  String get failedToMarkComplete => 'Error al completar el entrenamiento';

  @override
  String get durationMinutes => 'Duración (minutos)';

  @override
  String get bodyweight => 'Peso corporal';

  @override
  String get gym => 'Gimnasio';

  @override
  String get custom => 'Otro';

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
  String get trainingPartnersOptional => 'Partner';

  @override
  String get generatingTraining => 'Generando tu entrenamiento...';

  @override
  String get thisMayTakeAMoment => 'Esto puede tomar un momento';

  @override
  String generationFailedRetrying(int attempt) {
    return 'Generación fallida, reintentando #$attempt...';
  }

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
  String get availableWeights => 'Pesos Disponibles';

  @override
  String get availableWeightsHint =>
      'Configura las opciones de peso para los modificadores con peso en este gimnasio.';

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
  String get noMatchingEquipment => 'No hay equipamiento coincidente';

  @override
  String get instructions => 'Instrucciones';

  @override
  String get cues => 'Indicaciones';

  @override
  String get howWasYourTraining => '¿Cómo estuvo tu entrenamiento?';

  @override
  String get anyAdditionalComments => '¿Algún comentario adicional?';

  @override
  String get actualDuration => 'Duración real (min)';

  @override
  String get impossible => 'No puedo';

  @override
  String get tooHard => 'Muy difícil';

  @override
  String get ok => 'OK';

  @override
  String get easy => 'Fácil';

  @override
  String get tooEasy => 'Muy fácil';

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
  String get partneredTrainings => 'entrenamientos en pareja';

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
  String get calibrationGlobal => 'Global';

  @override
  String get calibrationNeeded =>
      'Completa tu primer entrenamiento para que Vigor calibre las recomendaciones a tu nivel';

  @override
  String get calibrationDescription =>
      'Durante la calibración, la plataforma recopila datos de tus comentarios para hacer una evaluación inicial de tu estado físico y nivel';

  @override
  String get calibrationInProgress =>
      'Tus entrenamientos se vuelven más inteligentes con cada sesión';

  @override
  String get calibrationTrainingNote =>
      'Este entrenamiento puede no coincidir completamente con tus objetivos — el sistema aún está aprendiendo tu nivel y prioriza la variedad de movimientos para construir un perfil completo';

  @override
  String calibrationFamiliesLearned(int calibrated, int total) {
    return '$calibrated/$total patrones de movimiento aprendidos';
  }

  @override
  String get capabilities => 'Capacidades';

  @override
  String get muscleHeatMap => 'Mapa de Calor Muscular';

  @override
  String get heatResting => 'En reposo';

  @override
  String get heatRecovered => 'Recuperado';

  @override
  String get heatActive => 'Activo';

  @override
  String get heatWarm => 'Caliente';

  @override
  String get heatHot => 'Intenso';

  @override
  String get noTrainingsCompletedYet => 'Empieza a entrenar para ver algo aquí';

  @override
  String get theme => 'Tema';

  @override
  String get themeAuto => 'Auto';

  @override
  String get themeLight => 'Claro';

  @override
  String get themeDark => 'Oscuro';

  @override
  String get themeAutoDescription => 'Seguir configuración del sistema';

  @override
  String get appearance => 'Apariencia';

  @override
  String get trainingDefaults => 'Predeterminados';

  @override
  String get defaultDuration => 'Duración del Entrenamiento';

  @override
  String get warmupCooldown => 'Calentamiento y enfriamiento';

  @override
  String get timer => 'Temporizador';

  @override
  String get intervalJingle => 'Silbato al cambio de intervalo';

  @override
  String get duckOtherAudio => 'Bajar otro audio';

  @override
  String get duckOtherAudioDescription =>
      'Reduce el volumen de música y otras apps durante el silbato';

  @override
  String get goalHypertrophy => 'Desarrollo Muscular';

  @override
  String get goalHypertrophyDescription =>
      'Aumenta el tamaño muscular con entrenamiento de resistencia';

  @override
  String get goalFatLoss => 'Pérdida de Grasa';

  @override
  String get goalFatLossDescription =>
      'Quema calorías y reduce grasa corporal con entrenamientos intensos';

  @override
  String get goalToning => 'Tonificación';

  @override
  String get goalToningDescription =>
      'Desarrolla músculos definidos y una apariencia atlética';

  @override
  String get goalPosture => 'Postura';

  @override
  String get goalPostureDescription =>
      'Fortalece espalda y core para mejor alineación';

  @override
  String get goalRehabilitation => 'Rehabilitación';

  @override
  String get goalRehabilitationDescription =>
      'Ejercicios seguros y controlados para recuperación de lesiones';

  @override
  String get goalWellness => 'Bienestar';

  @override
  String get goalWellnessDescription =>
      'Entrenamientos equilibrados para salud y alivio del estrés';

  @override
  String get goalFlexibility => 'Flexibilidad';

  @override
  String get goalFlexibilityDescription =>
      'Mejora el rango de movimiento con estiramientos y movilidad';

  @override
  String get goalSports => 'Rendimiento Deportivo';

  @override
  String get goalSportsDescription =>
      'Mejora la capacidad atlética con entrenamiento de potencia y agilidad';

  @override
  String get thisWeek => 'Esta Semana';

  @override
  String get trainingPlan => 'Plan de Entrenamiento';

  @override
  String get sessionsPerWeek => 'Sesiones por semana';

  @override
  String get sessionDuration => 'Duración de sesión';

  @override
  String get preferredTime => 'Horario preferido';

  @override
  String get recommendedTime => 'Horario recomendado';

  @override
  String get methodologyMix => 'Mezcla de metodologías';

  @override
  String get pastWeeks => 'Semanas anteriores';

  @override
  String get weeklyTarget => 'Objetivo Semanal';

  @override
  String daysLeft(int count) {
    return '$count días restantes';
  }

  @override
  String get recommended => 'Recomendado';

  @override
  String get duration => 'Duración';

  @override
  String get familyHorizontalPush => 'Empuje';

  @override
  String get familyHorizontalPull => 'Tirón';

  @override
  String get familyVerticalPush => 'Press vertical';

  @override
  String get familyVerticalPull => 'Dominada';

  @override
  String get familySquat => 'Sentadilla';

  @override
  String get familyHinge => 'Bisagra';

  @override
  String get familyCore => 'Core';

  @override
  String get familyCarry => 'Acarreo';

  @override
  String get familyCardio => 'Cardio';

  @override
  String get familyMobility => 'Movilidad';

  @override
  String get familyBalance => 'Equilibrio';

  @override
  String get trainingQuality => '¿Qué te pareció este entrenamiento?';

  @override
  String get trainingQualityHint =>
      'Nos ayuda a evaluar la calidad del entrenamiento generado';

  @override
  String get qualityReasonHint => '¿Qué podría mejorar?';

  @override
  String get good => 'Bueno';

  @override
  String get bad => 'Malo';

  @override
  String get loadingMsg1 => 'Analizando tu perfil...';

  @override
  String get loadingMsg2 => 'Seleccionando ejercicios...';

  @override
  String get loadingMsg3 => 'Construyendo tu rutina...';

  @override
  String get loadingMsg4 => 'Calculando volumen de entrenamiento...';

  @override
  String get loadingMsg5 => 'Optimizando intervalos de descanso...';

  @override
  String get loadingMsg6 => 'Consultando investigaciones...';

  @override
  String get loadingMsg7 => 'Equilibrando grupos musculares...';

  @override
  String get loadingMsg8 => 'Diseñando ruta de progresión...';

  @override
  String get loadingMsg9 => 'Ajustando la intensidad...';

  @override
  String get loadingMsg10 => 'Revisando patrones de movimiento...';

  @override
  String get loadingMsg11 => 'Evaluando necesidades de recuperación...';

  @override
  String get loadingMsg12 => 'Eligiendo variantes de ejercicio...';

  @override
  String get loadingMsg13 => 'Estructurando bloques de trabajo...';

  @override
  String get loadingMsg14 => 'Definiendo intervalos de trabajo...';

  @override
  String get loadingMsg15 => 'Aplicando ciencia del ejercicio...';

  @override
  String get loadingMsg16 => 'Diseñando calentamiento...';

  @override
  String get loadingMsg17 => 'Mapeando familias de movimiento...';

  @override
  String get loadingMsg18 => 'Evaluando distribución de carga...';

  @override
  String get loadingMsg19 => 'Personalizando tu sesión...';

  @override
  String get loadingMsg20 => 'Casi listo...';

  @override
  String loadingMsgGoal(String goal) {
    return 'Optimizando para $goal...';
  }

  @override
  String get loadingMsgInjuries => 'Adaptando según tus lesiones...';

  @override
  String get loadingMsgFavorites => 'Priorizando tus ejercicios favoritos...';

  @override
  String get loadingMsgConditions => 'Adaptando a tus condiciones...';

  @override
  String loadingMsgMethodology(String methodology) {
    return 'Diseñando sesión de $methodology...';
  }

  @override
  String get loadingMsgPartners => 'Coordinando entrenamiento en pareja...';

  @override
  String loadingMsgGym(String gym) {
    return 'Cargando equipamiento de $gym...';
  }

  @override
  String get loadingMsgHistory => 'Analizando tus sesiones recientes...';

  @override
  String get loadingRetryMsg1 => 'hmm, no salió bien — reintentando';

  @override
  String get loadingRetryMsg2 => 'déjame intentar de nuevo...';

  @override
  String get loadingRetryMsg3 => 'no del todo — otro intento';

  @override
  String get loadingRetryMsg4 => 'un intento más, un momento';

  @override
  String get loadingRetryMsg5 => 'oops, recalibrando...';

  @override
  String get loadingRetryMsg6 => 'casi lo tenía — otra vez';

  @override
  String nSelected(int count) {
    return '$count seleccionados';
  }

  @override
  String deleteSelectedTrainings(int count) {
    return '¿Eliminar $count entrenamientos?';
  }

  @override
  String get trainingsDeletedSuccessfully =>
      'Entrenamientos eliminados correctamente';

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
  String get pendingFeedbacks => 'Feedbacks pendientes';

  @override
  String get pendingFeedbacksDescription =>
      'Algunos de tus entrenamientos están marcados como completados pero no tienen tu feedback. El feedback ayuda a mejorar las recomendaciones para tus futuros entrenamientos.';

  @override
  String get equipmentPartner => 'Compañero';

  @override
  String get equipmentBalanceBoard => 'Tabla de Equilibrio';

  @override
  String get equipmentBand => 'Banda Elástica';

  @override
  String get equipmentBarbell => 'Barra';

  @override
  String get equipmentBench => 'Banco';

  @override
  String get equipmentBox => 'Cajón';

  @override
  String get equipmentBosuBall => 'Bosu Ball';

  @override
  String get equipmentCable => 'Polea';

  @override
  String get equipmentDipStation => 'Estación de Fondos';

  @override
  String get equipmentDumbbell => 'Mancuerna';

  @override
  String get equipmentEllipticalMachine => 'Elíptica';

  @override
  String get equipmentEzBarbell => 'Barra EZ';

  @override
  String get equipmentHammer => 'Martillo';

  @override
  String get equipmentKettlebell => 'Kettlebell';

  @override
  String get equipmentLeverageMachine => 'Máquina de Palanca';

  @override
  String get equipmentMedicineBall => 'Balón Medicinal';

  @override
  String get equipmentOlympicBarbell => 'Barra Olímpica';

  @override
  String get equipmentPullUpBar => 'Barra de Dominadas';

  @override
  String get equipmentResistanceBand => 'Banda de Resistencia';

  @override
  String get equipmentRings => 'Anillas';

  @override
  String get equipmentRoller => 'Rodillo';

  @override
  String get equipmentRope => 'Cuerda';

  @override
  String get equipmentRowingMachine => 'Máquina de Remo';

  @override
  String get equipmentSkiergMachine => 'Máquina de Esquí';

  @override
  String get equipmentSledMachine => 'Trineo';

  @override
  String get equipmentSmithMachine => 'Máquina Smith';

  @override
  String get equipmentStabilityBall => 'Pelota Suiza';

  @override
  String get equipmentStationaryBike => 'Bicicleta Estática';

  @override
  String get equipmentStepmillMachine => 'Escalera Infinita';

  @override
  String get equipmentTire => 'Neumático';

  @override
  String get equipmentTreadmill => 'Cinta de Correr';

  @override
  String get equipmentTrapBar => 'Barra Hexagonal';

  @override
  String get equipmentTrx => 'TRX';

  @override
  String get equipmentUpperBodyErgometer => 'Ergómetro de Brazos';

  @override
  String get equipmentWheelRoller => 'Rueda Abdominal';

  @override
  String get muscleChest => 'Pecho';

  @override
  String get muscleBack => 'Espalda';

  @override
  String get muscleShoulders => 'Hombros';

  @override
  String get muscleArms => 'Brazos';

  @override
  String get muscleCore => 'Core';

  @override
  String get muscleGlutes => 'Glúteos';

  @override
  String get muscleLegs => 'Piernas';

  @override
  String get modifierWeightedVest => 'Chaleco Lastrado';

  @override
  String get modifierParallettes => 'Paralelas Bajas';

  @override
  String get modifierAnkleWeights => 'Pesas de Tobillo';

  @override
  String get modifierDipBelt => 'Cinturón de Lastre';

  @override
  String get modifierPushUpBars => 'Barras para Flexiones';

  @override
  String get modifierResistanceBands => 'Bandas de Resistencia';

  @override
  String get modifierWeight => 'Peso';

  @override
  String get modifierWristWeights => 'Pesas de muñeca';

  @override
  String get healthData => 'Datos de salud';

  @override
  String get healthConnected => 'Conectado';

  @override
  String get healthSynchronizing => 'Sincronizando...';

  @override
  String get healthSynchronized => 'Sincronizado';

  @override
  String get healthSynchronize => 'Sincronizar ahora';

  @override
  String get healthNotConnected => 'No conectado';

  @override
  String healthBuildingBaselines(int days) {
    return 'Creando tus referencias personales ($days/14 días)';
  }

  @override
  String get healthDisconnect => 'Desconectar y eliminar datos';

  @override
  String get healthDisconnectConfirmation =>
      '¿Desconectar datos de salud? Todos los datos sincronizados serán eliminados. Esta acción no se puede deshacer.';

  @override
  String get healthDisconnectedSuccessfully => 'Datos de salud desconectados';

  @override
  String get failedToDisconnectHealth =>
      'No se pudieron desconectar los datos de salud';

  @override
  String get healthConnect => 'Conectar';

  @override
  String get healthMetrics => 'Datos de Salud';

  @override
  String get healthAdjustment => 'Ajuste de salud';

  @override
  String get healthPermissionsTitle => 'Conecta tu wearable';

  @override
  String get healthPermissionsDescription =>
      'Vigor lee tus datos de salud para personalizar tus entrenamientos. Mejores datos de sueño, recuperación y actividad significan recomendaciones más inteligentes.';

  @override
  String get healthPermissionsReadOnly => 'Acceso de solo lectura';

  @override
  String get healthPermissionsSleep => 'Duración y fases del sueño';

  @override
  String get healthPermissionsHrv =>
      'Variabilidad de la frecuencia cardíaca (HRV)';

  @override
  String get healthPermissionsRhr => 'Frecuencia cardíaca en reposo';

  @override
  String get healthPermissionsSteps => 'Pasos diarios';

  @override
  String get healthPermissionsWorkouts =>
      'Sesiones de entrenamiento y frecuencia cardíaca';

  @override
  String get healthPermissionsGrant => 'Conectar datos de salud';

  @override
  String get healthPermissionsSkip => 'Ahora no';

  @override
  String get healthPermissionsGranted => 'Datos de salud conectados';

  @override
  String get healthPermissionsDenied => 'Los permisos no fueron otorgados';

  @override
  String get healthOnboardingTitle => 'Conecta tu wearable';

  @override
  String get healthOnboardingDescription =>
      'Conecta tus datos de salud y Vigor adaptará tu entrenamiento a cómo duermes, te recuperas y te mueves.';

  @override
  String get healthOnboardingConnect => 'Conectar';

  @override
  String get healthOnboardingDismiss => 'Quizás después';

  @override
  String get healthInstallHcTitle => 'Health Connect requerido';

  @override
  String get healthInstallHcDescription =>
      'Health Connect es necesario para sincronizar los datos de tu wearable. Instálalo desde Play Store.';

  @override
  String get healthInstallHc => 'Instalar Health Connect';

  @override
  String get heartRate => 'Frecuencia cardíaca';

  @override
  String get avgHr => 'FC media';

  @override
  String get maxHr => 'FC máx';

  @override
  String get bpm => 'lpm';

  @override
  String get hrZones => 'Zonas FC';

  @override
  String get hrZone1 => 'Zona 1';

  @override
  String get hrZone2 => 'Zona 2';

  @override
  String get hrZone3 => 'Zona 3';

  @override
  String get hrZone4 => 'Zona 4';

  @override
  String get hrZone5 => 'Zona 5';

  @override
  String get healthDailySleep => 'Sueño';

  @override
  String get healthDailyRestingHr => 'FC en reposo';

  @override
  String get healthDailyHrv => 'HRV';

  @override
  String get healthDailySteps => 'Pasos';

  @override
  String get healthDailyCalories => 'Calorías';

  @override
  String get healthDailyNoData => 'Sin datos de salud';

  @override
  String get externalWorkout => 'Entrenamiento externo';

  @override
  String get exerciseTypeRunning => 'Carrera';

  @override
  String get exerciseTypeWalking => 'Caminata';

  @override
  String get exerciseTypeBiking => 'Ciclismo';

  @override
  String get exerciseTypeYoga => 'Yoga';

  @override
  String get exerciseTypeSwimming => 'Natación';

  @override
  String get exerciseTypeHiking => 'Senderismo';

  @override
  String get exerciseTypeStrengthTraining => 'Entrenamiento de fuerza';

  @override
  String get exerciseTypeFunctionalStrengthTraining =>
      'Entrenamiento funcional';

  @override
  String get exerciseTypeTraditionalStrengthTraining =>
      'Entrenamiento de fuerza';

  @override
  String get exerciseTypeRunningTreadmill => 'Cinta de correr';

  @override
  String get exerciseTypeBikingStationary => 'Bicicleta estática';

  @override
  String get exerciseTypeWalkingTreadmill => 'Caminata en cinta';

  @override
  String get exerciseTypeRowing => 'Remo';

  @override
  String get exerciseTypePilates => 'Pilates';

  @override
  String get exerciseTypeDancing => 'Baile';

  @override
  String get exerciseTypeElliptical => 'Elíptica';

  @override
  String get exerciseTypeStairClimbing => 'Escaleras';

  @override
  String get exerciseTypeOther => 'Entrenamiento';
}
