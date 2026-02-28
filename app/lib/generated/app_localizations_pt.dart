// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Portuguese (`pt`).
class AppLocalizationsPt extends AppLocalizations {
  AppLocalizationsPt([String locale = 'pt']) : super(locale);

  @override
  String get appName => 'Vigor';

  @override
  String get appTagline => 'Ex Sapientia Vis';

  @override
  String get navHome => 'Início';

  @override
  String get navActivity => 'Atividade';

  @override
  String get navProfile => 'Perfil';

  @override
  String get storageErrorTitle => 'Vigor - Erro de Armazenamento';

  @override
  String get storageUnavailable => 'Armazenamento Indisponível';

  @override
  String get storageErrorMessage =>
      'Requer armazenamento seguro. Verifica as configurações.';

  @override
  String get signInWithGoogle => 'Entrar com Google';

  @override
  String get signingIn => 'A entrar...';

  @override
  String get failedToInitializeGoogleSignIn =>
      'Falha ao inicializar o Google Sign In';

  @override
  String signInError(String message) {
    return 'Erro de login: $message';
  }

  @override
  String get googleSignInFailed => 'Falha no login do Google';

  @override
  String get failedToGetAuthToken =>
      'Não foi possível obter o token de autenticação';

  @override
  String errorProcessingSignIn(String message) {
    return 'Erro ao processar o login: $message';
  }

  @override
  String get googleSignInInitializing =>
      'O Google Sign In ainda está a inicializar...';

  @override
  String get readyToTrain => 'Pronto para treinar?';

  @override
  String get generateTrainingDescription =>
      'Cria um treino adaptado aos teus objetivos';

  @override
  String get generateTraining => 'Gerar Treino';

  @override
  String get refresh => 'Atualizar';

  @override
  String get logout => 'Sair';

  @override
  String get userDataRefreshed => 'Dados do utilizador atualizados';

  @override
  String get editProfile => 'Editar Perfil';

  @override
  String get settings => 'Configurações';

  @override
  String get other => 'Outros';

  @override
  String get deleteGym => 'Eliminar Ginásio';

  @override
  String deleteGymConfirmation(String name) {
    return 'Eliminar \"$name\"?';
  }

  @override
  String get cancel => 'Cancelar';

  @override
  String get delete => 'Eliminar';

  @override
  String get logoutConfirmation => 'Sair?';

  @override
  String get deleteAccount => 'Eliminar Conta';

  @override
  String get deleteAccountConfirmation =>
      'Eliminar a conta? Não pode ser desfeito.';

  @override
  String get accountDeletedSuccessfully => 'Conta eliminada com sucesso';

  @override
  String get failedToDeleteAccount => 'Falha ao eliminar a conta';

  @override
  String get failedToLoadGyms => 'Falha ao carregar ginásios';

  @override
  String get gymAddedSuccessfully => 'Ginásio adicionado com sucesso';

  @override
  String get failedToAddGym => 'Falha ao adicionar ginásio';

  @override
  String get gymUpdatedSuccessfully => 'Ginásio atualizado com sucesso';

  @override
  String get failedToUpdateGym => 'Falha ao atualizar ginásio';

  @override
  String get gymDeletedSuccessfully => 'Ginásio eliminado com sucesso';

  @override
  String get failedToDeleteGym => 'Falha ao eliminar ginásio';

  @override
  String get birthdate => 'Data de nascimento';

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
  String get injuries => 'Lesões';

  @override
  String get limitations => 'Limitações';

  @override
  String get conditions => 'Condições';

  @override
  String get favorites => 'Favoritos';

  @override
  String get personalDetails => 'Dados Pessoais';

  @override
  String get healthAndGoals => 'Saúde e Objetivos';

  @override
  String get exercises => 'Exercícios';

  @override
  String get equipment => 'Equipamento';

  @override
  String startedDate(String date) {
    return 'Início: $date';
  }

  @override
  String yearLabel(int year) {
    return 'Ano: $year';
  }

  @override
  String get myGyms => 'Ginásios';

  @override
  String get addGym => 'Adicionar Ginásio';

  @override
  String get noGymsAddedYet => 'Nenhum ginásio adicionado';

  @override
  String get addYourFirstGym => 'Adiciona o Teu Primeiro Ginásio';

  @override
  String get removeDefault => 'Remover Padrão';

  @override
  String get setAsDefault => 'Definir como Padrão';

  @override
  String get edit => 'Editar';

  @override
  String get quickActions => 'Ações Rápidas';

  @override
  String get dangerZone => 'Zona de Perigo';

  @override
  String get completeYourProfile => 'Completa o Teu Perfil';

  @override
  String get updateYourProfileInfo => 'Atualiza o teu perfil abaixo.';

  @override
  String get pleaseCompleteProfile => 'Completa o perfil. * = obrigatório.';

  @override
  String get firstName => 'Nome';

  @override
  String get lastName => 'Apelido';

  @override
  String get birthDate => 'Data de Nascimento';

  @override
  String get male => 'Masculino';

  @override
  String get female => 'Feminino';

  @override
  String get selectLanguage => 'Selecionar idioma';

  @override
  String get heightCm => 'Altura (cm)';

  @override
  String get weightKg => 'Peso (kg)';

  @override
  String get required => 'Obrigatório';

  @override
  String get invalid => 'Inválido';

  @override
  String get pleaseSelectBirthDate =>
      'Por favor, seleciona a tua data de nascimento';

  @override
  String get pleaseAddAtLeastOneGoal =>
      'Por favor, adiciona pelo menos um objetivo';

  @override
  String get pleaseSelectLanguage => 'Por favor, seleciona o teu idioma';

  @override
  String get addAGoal => 'Adicionar um objetivo';

  @override
  String get injuryDescription => 'Descrição da lesão';

  @override
  String get year => 'Ano';

  @override
  String get addALimitation => 'Adicionar uma limitação';

  @override
  String get addACondition => 'Adicionar uma condição';

  @override
  String get favoriteExercisesHint => 'ex. agachamentos, elevações, corrida';

  @override
  String get favoriteEquipmentHint => 'ex. halteres, barra, kettlebell';

  @override
  String get saveChanges => 'Guardar Alterações';

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
  String get favoriteExercises => 'Exercícios Favoritos';

  @override
  String get favoriteEquipment => 'Equipamento Favorito';

  @override
  String get favoriteWorkoutTypes => 'Tipos de Treino Preferidos';

  @override
  String get workoutTypeStrength => 'Força';

  @override
  String get workoutTypeStrengthDescription =>
      'Desenvolva força máxima com cargas pesadas e descanso completo';

  @override
  String get workoutTypeCircuit => 'Circuito';

  @override
  String get workoutTypeCircuitDescription =>
      'Passe pelas estações com descanso mínimo para condicionamento';

  @override
  String get workoutTypeEmom => 'EMOM';

  @override
  String get workoutTypeEmomDescription =>
      'Cada minuto: complete as repetições e descanse até o próximo minuto';

  @override
  String get workoutTypeAmrap => 'AMRAP';

  @override
  String get workoutTypeAmrapDescription =>
      'Quantas rodadas possíveis dentro do limite de tempo';

  @override
  String get workoutTypeHiit => 'HIIT';

  @override
  String get workoutTypeHiitDescription =>
      'Alterne explosões de alta intensidade com recuperação curta';

  @override
  String get workoutTypeForTime => 'For Time';

  @override
  String get workoutTypeForTimeDescription =>
      'Complete o treino o mais rápido possível';

  @override
  String get workoutTypeEndurance => 'Resistência';

  @override
  String get workoutTypeEnduranceDescription =>
      'Esforço sustentado em intensidade moderada para capacidade aeróbica';

  @override
  String get workoutTypeMobility => 'Mobilidade';

  @override
  String get workoutTypeMobilityDescription =>
      'Melhore a amplitude de movimento e a saúde articular';

  @override
  String get methodologyOptional => 'Metodologia (opcional)';

  @override
  String get methodologyAuto => 'Auto';

  @override
  String get goalsOptional => 'Objetivos (opcional)';

  @override
  String get musclesOptional => 'Músculos (opcional)';

  @override
  String get musclesAuto => 'Todos';

  @override
  String get advancedSettings => 'Avançado';

  @override
  String get failedToUpdateProfile => 'Falha ao atualizar o perfil';

  @override
  String get activity => 'Atividade';

  @override
  String get noTrainingsYet => 'Nenhum treino ainda';

  @override
  String get generateFirstTraining =>
      'Cria o primeiro treino a partir de Início';

  @override
  String get noTrainingAvailable => 'Sem treinos. Gera um.';

  @override
  String get availableTrainings => 'Treinos disponíveis';

  @override
  String get pastTrainings => 'Treinos anteriores';

  @override
  String get stale => 'Desatualizado';

  @override
  String get today => 'Hoje';

  @override
  String get yesterday => 'Ontem';

  @override
  String get available => 'Disponível';

  @override
  String get completed => 'Concluído';

  @override
  String get completedSingular => 'Concluído';

  @override
  String get noPastTrainings => 'Nenhum treino concluído ainda';

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
  String get failedToLoadTrainings => 'Falha ao carregar treinos';

  @override
  String get startTraining => 'Iniciar Treino';

  @override
  String get cloneTraining => 'Clonar Treino';

  @override
  String get addPartner => 'Adicionar Parceiro';

  @override
  String get shareWithUser => 'Partilhar com Utilizador';

  @override
  String get deleteTraining => 'Eliminar Treino';

  @override
  String get leaveTraining => 'Abandonar Treino';

  @override
  String get showAiReasoning => 'Mostrar raciocínio da IA';

  @override
  String get reportIssue => 'Reportar problema';

  @override
  String deleteTrainingConfirmation(String name) {
    return 'Eliminar \"$name\"? Não pode ser desfeito.';
  }

  @override
  String leaveTrainingConfirmation(String name) {
    return 'Abandonar \"$name\"? Não o verás mais.';
  }

  @override
  String addPartnerConfirmation(String userName, String trainingName) {
    return 'Adicionar $userName como parceiro a \"$trainingName\"?';
  }

  @override
  String cloneTrainingConfirmation(String name) {
    return 'Clonar \"$name\" para os teus treinos?';
  }

  @override
  String shareTrainingConfirmation(String trainingName, String userName) {
    return 'Partilhar \"$trainingName\" com $userName?';
  }

  @override
  String get trainingDeletedSuccessfully => 'Treino eliminado com sucesso';

  @override
  String get failedToDeleteTraining => 'Falha ao eliminar o treino';

  @override
  String get leftTrainingSuccessfully => 'Treino abandonado com sucesso';

  @override
  String get partnerAddedSuccessfully => 'Parceiro adicionado com sucesso';

  @override
  String get failedToAddPartner => 'Falha ao adicionar parceiro';

  @override
  String get trainingSharedSuccessfully => 'Treino partilhado com sucesso';

  @override
  String get failedToShareTraining => 'Falha ao partilhar o treino';

  @override
  String get trainingCloned => 'Treino clonado';

  @override
  String get failedToCloneTraining => 'Falha ao clonar o treino';

  @override
  String get trainingMarkedAsComplete => 'Treino concluído';

  @override
  String get failedToCompleteTraining => 'Falha ao concluir o treino';

  @override
  String get feedback => 'Feedback';

  @override
  String get feedbackUpdated => 'Feedback atualizado';

  @override
  String get failedToUpdateFeedback => 'Falha ao atualizar o feedback';

  @override
  String get reportSubmitted => 'Relatório enviado';

  @override
  String get failedToSubmitReport => 'Falha ao enviar o relatório';

  @override
  String get shuffleExercise => 'Trocar exercício';

  @override
  String get exerciseShuffled => 'Exercício trocado';

  @override
  String get failedToShuffleExercise => 'Falha ao trocar exercício';

  @override
  String get reasoning => 'Raciocínio';

  @override
  String get strategy => 'Estratégia';

  @override
  String get typeSelection => 'Seleção de Tipo';

  @override
  String get progression => 'Progressão';

  @override
  String get constraints => 'Restrições';

  @override
  String get researchApplied => 'Pesquisa Aplicada';

  @override
  String get targetMuscles => 'Músculos Alvo';

  @override
  String get naming => 'Denominação';

  @override
  String get trainingRoutines => 'Treino';

  @override
  String get noEquipment => 'Sem equipamento';

  @override
  String blockNumber(int number) {
    return 'Bloco $number';
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
    return '$count repetições';
  }

  @override
  String weightKgValue(double value) {
    return '$value kg';
  }

  @override
  String get markAsComplete => 'Marcar como Concluído';

  @override
  String get updateFeedback => 'Atualizar Feedback';

  @override
  String get references => 'Referências';

  @override
  String get literature => 'Literatura';

  @override
  String get request => 'Request';

  @override
  String get describeIssue => 'Descreve o problema...';

  @override
  String get submit => 'Enviar';

  @override
  String get close => 'Fechar';

  @override
  String get add => 'Adicionar';

  @override
  String get update => 'Atualizar';

  @override
  String get clone => 'Clonar';

  @override
  String get share => 'Partilhar';

  @override
  String get leave => 'Abandonar';

  @override
  String get tapToStart => 'Toca para começar';

  @override
  String get tapWhenDone => 'Toca quando terminar';

  @override
  String get trainingCompleted => 'Treino Concluído!';

  @override
  String greatJobCompleting(String name) {
    return 'Excelente trabalho com $name';
  }

  @override
  String get done => 'Concluído';

  @override
  String get complete => 'Concluir';

  @override
  String routineCounter(int current, int total) {
    return 'Rotina $current/$total';
  }

  @override
  String blockCounter(int current, int total) {
    return 'Bloco $current/$total';
  }

  @override
  String get exitTraining => 'Sair do Treino?';

  @override
  String get whatWouldYouLikeToDo => 'O que fazer?';

  @override
  String get exit => 'Sair';

  @override
  String get continueTraining => 'Continuar';

  @override
  String get failedToMarkComplete => 'Falha ao concluir o treino';

  @override
  String get durationMinutes => 'Duração (minutos)';

  @override
  String get bodyweight => 'Peso corporal';

  @override
  String get gym => 'Ginásio';

  @override
  String get custom => 'Outro';

  @override
  String get noEquipmentBodyweightOnly => 'Apenas peso corporal';

  @override
  String get noGymsDefinedCreateOne => 'Sem ginásios. Cria um no perfil.';

  @override
  String get selectAGym => 'Selecionar um ginásio';

  @override
  String get addEquipment => 'Adicionar Equipamento';

  @override
  String get addEquipmentAvailable => 'Adiciona o teu equipamento disponível';

  @override
  String get includeWarmupCooldown => 'Incluir aquecimento e arrefecimento';

  @override
  String get equipmentPlaceholder => 'ex. Barra, Halteres';

  @override
  String get customPromptOptional => 'Prompt Personalizado (opcional)';

  @override
  String get focusOnUpperBody => 'ex. Focar na parte superior do corpo';

  @override
  String get trainingPartnersOptional => 'Partner';

  @override
  String get generatingTraining => 'A gerar o teu treino...';

  @override
  String get thisMayTakeAMoment => 'Isto pode demorar um momento';

  @override
  String generationFailedRetrying(int attempt) {
    return 'Geração falhou, tentativa #$attempt...';
  }

  @override
  String get trainingGeneratedSuccessfully => 'Treino gerado com sucesso!';

  @override
  String get failedToGenerateTraining => 'Falha ao gerar o treino';

  @override
  String get generate => 'Gerar';

  @override
  String get editGym => 'Editar Ginásio';

  @override
  String get gymName => 'Nome do Ginásio';

  @override
  String get gymNamePlaceholder => 'ex. Ginásio Casa, Holmes Place';

  @override
  String get noEquipmentAddedYet => 'Nenhum equipamento adicionado';

  @override
  String get pleaseEnterGymName => 'Por favor, insere o nome do ginásio';

  @override
  String get addAllEquipment => 'Adicionar Tudo';

  @override
  String get failedToLoadEquipment => 'Falha ao carregar equipamentos';

  @override
  String get selectUser => 'Selecionar Utilizador';

  @override
  String get searchByName => 'Pesquisar por nome';

  @override
  String get noUsersAvailable => 'Nenhum utilizador disponível';

  @override
  String get noMatchingUsers => 'Nenhum utilizador correspondente';

  @override
  String get instructions => 'Instruções';

  @override
  String get cues => 'Dicas';

  @override
  String get howWasYourTraining => 'Como foi o teu treino?';

  @override
  String get anyAdditionalComments => 'Algum comentário adicional?';

  @override
  String get actualDuration => 'Duração real (min)';

  @override
  String get impossible => 'Não consigo';

  @override
  String get tooHard => 'Muito difícil';

  @override
  String get ok => 'OK';

  @override
  String get easy => 'Fácil';

  @override
  String get tooEasy => 'Muito fácil';

  @override
  String get flag => 'Sinalizar';

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
    return 'Próximo: $name';
  }

  @override
  String get rest => 'Descanso';

  @override
  String get upcoming => 'Próximos';

  @override
  String get yourProgress => 'Seu progresso';

  @override
  String trainingsCompleted(int count) {
    return '$count treinos concluídos';
  }

  @override
  String get completedTrainings => 'treinos concluídos';

  @override
  String get partneredTrainings => 'treinos em parceria';

  @override
  String get movementFamilies => 'Famílias de movimento';

  @override
  String get muscleActivity => 'Atividade muscular';

  @override
  String get failedToLoadProgress => 'Falha ao carregar progresso';

  @override
  String get noProgressYet => 'Complete treinos para ver seu progresso';

  @override
  String get calibration => 'Calibração';

  @override
  String get calibrationGlobal => 'Global';

  @override
  String get calibrationNeeded =>
      'Complete seu primeiro treino para que o Vigor calibre as recomendações ao seu nível';

  @override
  String get calibrationDescription =>
      'Durante a calibração, a plataforma coleta dados dos seus feedbacks para fazer uma avaliação inicial do seu estado físico e nível';

  @override
  String get calibrationInProgress =>
      'Seus treinos ficam mais inteligentes a cada sessão';

  @override
  String calibrationFamiliesLearned(int calibrated, int total) {
    return '$calibrated/$total padrões de movimento aprendidos';
  }

  @override
  String get capabilities => 'Capacidades';

  @override
  String get muscleHeatMap => 'Mapa de Calor Muscular';

  @override
  String get heatResting => 'Em repouso';

  @override
  String get heatRecovered => 'Recuperado';

  @override
  String get heatActive => 'Ativo';

  @override
  String get heatWarm => 'Quente';

  @override
  String get heatHot => 'Intenso';

  @override
  String get noTrainingsCompletedYet => 'Comece a treinar para ver algo aqui';

  @override
  String get theme => 'Tema';

  @override
  String get themeAuto => 'Auto';

  @override
  String get themeLight => 'Claro';

  @override
  String get themeDark => 'Escuro';

  @override
  String get themeAutoDescription => 'Seguir configurações do sistema';

  @override
  String get appearance => 'Aparência';

  @override
  String get trainingDefaults => 'Padrões';

  @override
  String get defaultDuration => 'Duração do Treino';

  @override
  String get warmupCooldown => 'Aquecimento e arrefecimento';

  @override
  String get timer => 'Temporizador';

  @override
  String get intervalJingle => 'Apito na troca de intervalo';

  @override
  String get duckOtherAudio => 'Baixar outro áudio';

  @override
  String get duckOtherAudioDescription =>
      'Reduz o volume de música e outros apps durante o apito';

  @override
  String get goalHypertrophy => 'Hipertrofia';

  @override
  String get goalHypertrophyDescription =>
      'Aumente a massa muscular com treino de resistência';

  @override
  String get goalFatLoss => 'Perda de Gordura';

  @override
  String get goalFatLossDescription =>
      'Queime calorias e reduza gordura com treinos intensos';

  @override
  String get goalToning => 'Tonificação';

  @override
  String get goalToningDescription =>
      'Desenvolva músculos definidos e uma aparência atlética';

  @override
  String get goalPosture => 'Postura';

  @override
  String get goalPostureDescription =>
      'Fortaleça costas e core para melhor alinhamento';

  @override
  String get goalRehabilitation => 'Reabilitação';

  @override
  String get goalRehabilitationDescription =>
      'Exercícios seguros e controlados para recuperação de lesões';

  @override
  String get goalWellness => 'Bem-estar';

  @override
  String get goalWellnessDescription =>
      'Treinos equilibrados para saúde e alívio do stress';

  @override
  String get goalFlexibility => 'Flexibilidade';

  @override
  String get goalFlexibilityDescription =>
      'Melhore a amplitude de movimento com alongamentos e mobilidade';

  @override
  String get goalSports => 'Desempenho Desportivo';

  @override
  String get goalSportsDescription =>
      'Melhore a capacidade atlética com treino de potência e agilidade';

  @override
  String get thisWeek => 'Esta Semana';

  @override
  String get trainingPlan => 'Plano de Treino';

  @override
  String get sessionsPerWeek => 'Sessões por semana';

  @override
  String get sessionDuration => 'Duração da sessão';

  @override
  String get preferredTime => 'Horário preferido';

  @override
  String get recommendedTime => 'Horário recomendado';

  @override
  String get methodologyMix => 'Mix de metodologias';

  @override
  String get pastWeeks => 'Semanas anteriores';

  @override
  String get weeklyTarget => 'Meta Semanal';

  @override
  String daysLeft(int count) {
    return '$count dias restantes';
  }

  @override
  String get recommended => 'Recomendado';

  @override
  String get duration => 'Duração';

  @override
  String get familyHorizontalPush => 'Empurrar';

  @override
  String get familyHorizontalPull => 'Puxar';

  @override
  String get familyVerticalPush => 'Press vertical';

  @override
  String get familyVerticalPull => 'Barra fixa';

  @override
  String get familySquat => 'Agachamento';

  @override
  String get familyHinge => 'Dobradiça';

  @override
  String get familyCore => 'Core';

  @override
  String get familyCarry => 'Transporte';

  @override
  String get familyCardio => 'Cardio';

  @override
  String get familyMobility => 'Mobilidade';

  @override
  String get familyBalance => 'Equilíbrio';

  @override
  String get trainingQuality => 'O que achou deste treino?';

  @override
  String get trainingQualityHint =>
      'Nos ajuda a avaliar a qualidade do treino gerado';

  @override
  String get qualityReasonHint => 'O que poderia ser melhorado?';

  @override
  String get good => 'Bom';

  @override
  String get bad => 'Ruim';

  @override
  String get loadingMsg1 => 'Analisando seu perfil...';

  @override
  String get loadingMsg2 => 'Selecionando exercícios...';

  @override
  String get loadingMsg3 => 'Montando sua rotina...';

  @override
  String get loadingMsg4 => 'Calculando volume de treino...';

  @override
  String get loadingMsg5 => 'Otimizando intervalos de descanso...';

  @override
  String get loadingMsg6 => 'Consultando pesquisas...';

  @override
  String get loadingMsg7 => 'Equilibrando grupos musculares...';

  @override
  String get loadingMsg8 => 'Criando caminho de progressão...';

  @override
  String get loadingMsg9 => 'Ajustando a intensidade...';

  @override
  String get loadingMsg10 => 'Revisando padrões de movimento...';

  @override
  String get loadingMsg11 => 'Avaliando necessidades de recuperação...';

  @override
  String get loadingMsg12 => 'Escolhendo variações de exercício...';

  @override
  String get loadingMsg13 => 'Estruturando blocos de treino...';

  @override
  String get loadingMsg14 => 'Definindo intervalos de trabalho...';

  @override
  String get loadingMsg15 => 'Aplicando ciência do exercício...';

  @override
  String get loadingMsg16 => 'Planejando aquecimento...';

  @override
  String get loadingMsg17 => 'Mapeando famílias de movimento...';

  @override
  String get loadingMsg18 => 'Avaliando distribuição de carga...';

  @override
  String get loadingMsg19 => 'Personalizando sua sessão...';

  @override
  String get loadingMsg20 => 'Quase pronto...';

  @override
  String loadingMsgGoal(String goal) {
    return 'Otimizando para $goal...';
  }

  @override
  String get loadingMsgInjuries => 'Adaptando às suas lesões...';

  @override
  String get loadingMsgFavorites => 'Priorizando seus exercícios favoritos...';

  @override
  String get loadingMsgConditions => 'Adaptando às suas condições...';

  @override
  String loadingMsgMethodology(String methodology) {
    return 'Criando sessão de $methodology...';
  }

  @override
  String get loadingMsgPartners => 'Coordenando treino em dupla...';

  @override
  String loadingMsgGym(String gym) {
    return 'Carregando equipamentos de $gym...';
  }

  @override
  String get loadingMsgHistory => 'Analisando suas sessões recentes...';

  @override
  String get loadingRetryMsg1 => 'hmm, não saiu bem — tentando de novo';

  @override
  String get loadingRetryMsg2 => 'deixa eu tentar de novo...';

  @override
  String get loadingRetryMsg3 => 'não rolou — mais uma tentativa';

  @override
  String get loadingRetryMsg4 => 'mais uma tentativa, um momento';

  @override
  String get loadingRetryMsg5 => 'oops, recalibrando...';

  @override
  String get loadingRetryMsg6 => 'quase lá — tentando de novo';

  @override
  String nSelected(int count) {
    return '$count selecionados';
  }

  @override
  String deleteSelectedTrainings(int count) {
    return 'Excluir $count treinos?';
  }

  @override
  String get trainingsDeletedSuccessfully => 'Treinos excluídos com sucesso';
}
