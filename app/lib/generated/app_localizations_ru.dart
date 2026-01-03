// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Russian (`ru`).
class AppLocalizationsRu extends AppLocalizations {
  AppLocalizationsRu([String locale = 'ru']) : super(locale);

  @override
  String get appName => 'Vigor';

  @override
  String get appTagline => 'Ex Sapientia Vis';

  @override
  String get navHome => 'Главная';

  @override
  String get navActivity => 'Активность';

  @override
  String get navProfile => 'Профиль';

  @override
  String get storageErrorTitle => 'Vigor - Ошибка хранилища';

  @override
  String get storageUnavailable => 'Хранилище недоступно';

  @override
  String get storageErrorMessage =>
      'Требуется безопасное хранилище. Проверьте настройки.';

  @override
  String get signInWithGoogle => 'Войти через Google';

  @override
  String get signingIn => 'Вход...';

  @override
  String get failedToInitializeGoogleSignIn =>
      'Не удалось инициализировать Google Sign In';

  @override
  String signInError(String message) {
    return 'Ошибка входа: $message';
  }

  @override
  String get googleSignInFailed => 'Ошибка входа через Google';

  @override
  String get failedToGetAuthToken => 'Не удалось получить токен аутентификации';

  @override
  String errorProcessingSignIn(String message) {
    return 'Ошибка обработки входа: $message';
  }

  @override
  String get googleSignInInitializing =>
      'Google Sign In всё ещё инициализируется...';

  @override
  String get readyToTrain => 'Готовы тренироваться?';

  @override
  String get generateTrainingDescription => 'Создайте тренировку под ваши цели';

  @override
  String get generateTraining => 'Создать Тренировку';

  @override
  String get refresh => 'Обновить';

  @override
  String get logout => 'Выйти';

  @override
  String get userDataRefreshed => 'Данные пользователя обновлены';

  @override
  String get editProfile => 'Редактировать Профиль';

  @override
  String get settings => 'Настройки';

  @override
  String get deleteGym => 'Удалить Зал';

  @override
  String deleteGymConfirmation(String name) {
    return 'Удалить \"$name\"?';
  }

  @override
  String get cancel => 'Отмена';

  @override
  String get delete => 'Удалить';

  @override
  String get logoutConfirmation => 'Выйти?';

  @override
  String get deleteAccount => 'Удалить Аккаунт';

  @override
  String get deleteAccountConfirmation => 'Удалить аккаунт? Это необратимо.';

  @override
  String get accountDeletedSuccessfully => 'Аккаунт успешно удалён';

  @override
  String get failedToDeleteAccount => 'Не удалось удалить аккаунт';

  @override
  String get failedToLoadGyms => 'Не удалось загрузить залы';

  @override
  String get gymAddedSuccessfully => 'Зал успешно добавлен';

  @override
  String get failedToAddGym => 'Не удалось добавить зал';

  @override
  String get gymUpdatedSuccessfully => 'Зал успешно обновлён';

  @override
  String get failedToUpdateGym => 'Не удалось обновить зал';

  @override
  String get gymDeletedSuccessfully => 'Зал успешно удалён';

  @override
  String get failedToDeleteGym => 'Не удалось удалить зал';

  @override
  String get birthdate => 'Дата рождения';

  @override
  String get gender => 'Пол';

  @override
  String get language => 'Язык';

  @override
  String get height => 'Рост';

  @override
  String get weight => 'Вес';

  @override
  String get heightUnit => 'см';

  @override
  String get weightUnit => 'кг';

  @override
  String heightWithUnit(double value) {
    return '$value см';
  }

  @override
  String weightWithUnit(double value) {
    return '$value кг';
  }

  @override
  String get goals => 'Цели';

  @override
  String get injuries => 'Травмы';

  @override
  String get limitations => 'Ограничения';

  @override
  String get favorites => 'Избранное';

  @override
  String get exercises => 'Упражнения';

  @override
  String get equipment => 'Оборудование';

  @override
  String startedDate(String date) {
    return 'Начало: $date';
  }

  @override
  String yearLabel(int year) {
    return 'Год: $year';
  }

  @override
  String get myGyms => 'Залы';

  @override
  String get addGym => 'Добавить Зал';

  @override
  String get noGymsAddedYet => 'Залы ещё не добавлены';

  @override
  String get addYourFirstGym => 'Добавьте Свой Первый Зал';

  @override
  String get removeDefault => 'Убрать по Умолчанию';

  @override
  String get setAsDefault => 'Установить по Умолчанию';

  @override
  String get edit => 'Редактировать';

  @override
  String get quickActions => 'Быстрые Действия';

  @override
  String get dangerZone => 'Опасная Зона';

  @override
  String get completeYourProfile => 'Заполните Свой Профиль';

  @override
  String get updateYourProfileInfo => 'Обновите свой профиль ниже.';

  @override
  String get pleaseCompleteProfile => 'Заполните профиль. * = обязательно.';

  @override
  String get firstName => 'Имя';

  @override
  String get lastName => 'Фамилия';

  @override
  String get birthDate => 'Дата Рождения';

  @override
  String get male => 'Мужской';

  @override
  String get female => 'Женский';

  @override
  String get selectLanguage => 'Выберите язык';

  @override
  String get heightCm => 'Рост (см)';

  @override
  String get weightKg => 'Вес (кг)';

  @override
  String get required => 'Обязательно';

  @override
  String get invalid => 'Недействительно';

  @override
  String get pleaseSelectBirthDate => 'Пожалуйста, выберите дату рождения';

  @override
  String get pleaseAddAtLeastOneGoal =>
      'Пожалуйста, добавьте хотя бы одну цель';

  @override
  String get pleaseSelectLanguage => 'Пожалуйста, выберите язык';

  @override
  String get addAGoal => 'Добавить цель';

  @override
  String get injuryDescription => 'Описание травмы';

  @override
  String get year => 'Год';

  @override
  String get addALimitation => 'Добавить ограничение';

  @override
  String get favoriteExercisesHint => 'напр. приседания, подтягивания, бег';

  @override
  String get favoriteEquipmentHint => 'напр. гантели, штанга, гиря';

  @override
  String get saveChanges => 'Сохранить Изменения';

  @override
  String get saveProfile => 'Сохранить Профиль';

  @override
  String get optionalLeaveEmpty => '(Необязательно)';

  @override
  String get optionalExercisesPrefer => '(Необязательно)';

  @override
  String get optionalEquipmentPrefer => '(Необязательно)';

  @override
  String get optionalWorkoutTypesPrefer => '(Необязательно)';

  @override
  String get favoriteExercises => 'Любимые Упражнения';

  @override
  String get favoriteEquipment => 'Любимое Оборудование';

  @override
  String get favoriteWorkoutTypes => 'Предпочитаемые Типы Тренировок';

  @override
  String get workoutTypeStrength => 'Сила';

  @override
  String get workoutTypeCircuit => 'Круговая';

  @override
  String get workoutTypeEmom => 'EMOM';

  @override
  String get workoutTypeAmrap => 'AMRAP';

  @override
  String get workoutTypeHiit => 'HIIT';

  @override
  String get workoutTypeForTime => 'На время';

  @override
  String get workoutTypeEndurance => 'Выносливость';

  @override
  String get workoutTypeMobility => 'Мобильность';

  @override
  String get methodologyOptional => 'Методология (необязательно)';

  @override
  String get methodologyAuto => 'Авто';

  @override
  String get goalsOptional => 'Цели (необязательно)';

  @override
  String get musclesOptional => 'Мышцы (необязательно)';

  @override
  String get musclesAuto => 'Авто';

  @override
  String get failedToUpdateProfile => 'Не удалось обновить профиль';

  @override
  String get activity => 'Активность';

  @override
  String get noTrainingsYet => 'Тренировок пока нет';

  @override
  String get generateFirstTraining => 'Создайте первую тренировку из Главной';

  @override
  String get noTrainingAvailable => 'Нет тренировок. Создайте.';

  @override
  String get availableTrainings => 'Доступные тренировки';

  @override
  String get pastTrainings => 'Прошлые тренировки';

  @override
  String get stale => 'Устарело';

  @override
  String get today => 'Сегодня';

  @override
  String get yesterday => 'Вчера';

  @override
  String get available => 'Доступно';

  @override
  String get completed => 'Завершено';

  @override
  String get noPastTrainings => 'Пока нет завершённых тренировок';

  @override
  String get copied => 'Скопировано';

  @override
  String durationMin(int minutes) {
    return '$minutes мин';
  }

  @override
  String durationHr(int hours) {
    return '$hours ч';
  }

  @override
  String durationHrMin(int hours, int minutes) {
    return '$hours ч $minutes мин';
  }

  @override
  String get failedToLoadTrainings => 'Не удалось загрузить тренировки';

  @override
  String get startTraining => 'Начать Тренировку';

  @override
  String get cloneTraining => 'Клонировать Тренировку';

  @override
  String get addPartner => 'Добавить Партнёра';

  @override
  String get shareWithUser => 'Поделиться с Пользователем';

  @override
  String get deleteTraining => 'Удалить Тренировку';

  @override
  String get leaveTraining => 'Покинуть Тренировку';

  @override
  String get showAiReasoning => 'Показать рассуждения ИИ';

  @override
  String get reportIssue => 'Сообщить о проблеме';

  @override
  String deleteTrainingConfirmation(String name) {
    return 'Удалить \"$name\"? Это необратимо.';
  }

  @override
  String leaveTrainingConfirmation(String name) {
    return 'Покинуть \"$name\"? Вы больше не увидите её.';
  }

  @override
  String addPartnerConfirmation(String userName, String trainingName) {
    return 'Добавить $userName как партнёра к \"$trainingName\"?';
  }

  @override
  String cloneTrainingConfirmation(String name) {
    return 'Клонировать \"$name\" в ваши тренировки?';
  }

  @override
  String shareTrainingConfirmation(String trainingName, String userName) {
    return 'Поделиться \"$trainingName\" с $userName?';
  }

  @override
  String get trainingDeletedSuccessfully => 'Тренировка успешно удалена';

  @override
  String get failedToDeleteTraining => 'Не удалось удалить тренировку';

  @override
  String get leftTrainingSuccessfully => 'Тренировка успешно покинута';

  @override
  String get partnerAddedSuccessfully => 'Партнёр успешно добавлен';

  @override
  String get failedToAddPartner => 'Не удалось добавить партнёра';

  @override
  String get trainingSharedSuccessfully => 'Тренировка успешно отправлена';

  @override
  String get failedToShareTraining => 'Не удалось поделиться тренировкой';

  @override
  String get trainingCloned => 'Тренировка клонирована';

  @override
  String get failedToCloneTraining => 'Не удалось клонировать тренировку';

  @override
  String get trainingMarkedAsComplete => 'Тренировка завершена';

  @override
  String get failedToCompleteTraining => 'Не удалось завершить тренировку';

  @override
  String get reportSubmitted => 'Отчёт отправлен';

  @override
  String get failedToSubmitReport => 'Не удалось отправить отчёт';

  @override
  String get shuffleExercise => 'Сменить упражнение';

  @override
  String get exerciseShuffled => 'Упражнение изменено';

  @override
  String get failedToShuffleExercise => 'Не удалось сменить упражнение';

  @override
  String get reasoning => 'Рассуждение';

  @override
  String get strategy => 'Стратегия';

  @override
  String get typeSelection => 'Выбор Типа';

  @override
  String get progression => 'Прогрессия';

  @override
  String get constraints => 'Ограничения';

  @override
  String get researchApplied => 'Применённые Исследования';

  @override
  String get targetMuscles => 'Целевые Мышцы';

  @override
  String get naming => 'Наименование';

  @override
  String get trainingRoutines => 'Тренировочные Программы';

  @override
  String get noEquipment => 'Без оборудования';

  @override
  String blockNumber(int number) {
    return 'Блок $number';
  }

  @override
  String repeatsCount(int count) {
    return '${count}x';
  }

  @override
  String durationSeconds(int seconds) {
    return '$secondsс';
  }

  @override
  String restSeconds(int seconds) {
    return '$secondsс отдых';
  }

  @override
  String repsCount(int count) {
    return '$count повторений';
  }

  @override
  String weightKgValue(double value) {
    return '$value кг';
  }

  @override
  String get markAsComplete => 'Отметить как Завершённое';

  @override
  String get references => 'Ссылки';

  @override
  String get describeIssue => 'Опишите проблему...';

  @override
  String get submit => 'Отправить';

  @override
  String get close => 'Закрыть';

  @override
  String get add => 'Добавить';

  @override
  String get update => 'Обновить';

  @override
  String get clone => 'Клонировать';

  @override
  String get share => 'Поделиться';

  @override
  String get leave => 'Покинуть';

  @override
  String get tapToStart => 'Нажмите, чтобы начать';

  @override
  String get trainingCompleted => 'Тренировка Завершена!';

  @override
  String greatJobCompleting(String name) {
    return 'Отличная работа с $name';
  }

  @override
  String get done => 'Готово';

  @override
  String get complete => 'Завершить';

  @override
  String routineCounter(int current, int total) {
    return 'Программа $current/$total';
  }

  @override
  String blockCounter(int current, int total) {
    return 'Блок $current/$total';
  }

  @override
  String get exitTraining => 'Выйти из Тренировки?';

  @override
  String get whatWouldYouLikeToDo => 'Что делать?';

  @override
  String get exit => 'Выйти';

  @override
  String get continueTraining => 'Продолжить';

  @override
  String get failedToMarkComplete => 'Не удалось завершить тренировку';

  @override
  String get durationMinutes => 'Продолжительность (минуты)';

  @override
  String get bodyweight => 'С собственным весом';

  @override
  String get gym => 'Зал';

  @override
  String get custom => 'Настраиваемое';

  @override
  String get noEquipmentBodyweightOnly => 'Только с собственным весом';

  @override
  String get noGymsDefinedCreateOne => 'Нет залов. Создайте в профиле.';

  @override
  String get selectAGym => 'Выбрать зал';

  @override
  String get addEquipment => 'Добавить Оборудование';

  @override
  String get addEquipmentAvailable => 'Добавьте доступное оборудование';

  @override
  String get includeWarmupCooldown => 'Включить разминку и заминку';

  @override
  String get equipmentPlaceholder => 'напр. Штанга, Гантели';

  @override
  String get customPromptOptional => 'Пользовательский Запрос (необязательно)';

  @override
  String get focusOnUpperBody => 'напр. Сосредоточиться на верхней части тела';

  @override
  String get trainingPartnersOptional =>
      'Партнёры по Тренировке (необязательно)';

  @override
  String get generatingTraining => 'Создаём вашу тренировку...';

  @override
  String get thisMayTakeAMoment => 'Это может занять некоторое время';

  @override
  String generationFailedRetrying(int attempt) {
    return 'Генерация не удалась, попытка #$attempt...';
  }

  @override
  String get trainingGeneratedSuccessfully => 'Тренировка успешно создана!';

  @override
  String get failedToGenerateTraining => 'Не удалось создать тренировку';

  @override
  String get generate => 'Создать';

  @override
  String get editGym => 'Редактировать Зал';

  @override
  String get gymName => 'Название Зала';

  @override
  String get gymNamePlaceholder => 'напр. Домашний Зал, World Class';

  @override
  String get noEquipmentAddedYet => 'Оборудование ещё не добавлено';

  @override
  String get pleaseEnterGymName => 'Пожалуйста, введите название зала';

  @override
  String get addAllEquipment => 'Добавить Все';

  @override
  String get failedToLoadEquipment => 'Не удалось загрузить оборудование';

  @override
  String get selectUser => 'Выбрать Пользователя';

  @override
  String get searchByName => 'Поиск по имени';

  @override
  String get noUsersAvailable => 'Пользователи недоступны';

  @override
  String get noMatchingUsers => 'Подходящих пользователей нет';

  @override
  String get instructions => 'Инструкции';

  @override
  String get howWasYourTraining => 'Как прошла тренировка?';

  @override
  String get anyAdditionalComments => 'Дополнительные комментарии?';

  @override
  String get tooEasy => 'Слишком легко';

  @override
  String get tooHard => 'Слишком сложно';

  @override
  String get flag => 'Отметить';

  @override
  String get profile => 'Профиль';

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
    return 'Далее: $name';
  }

  @override
  String get rest => 'Отдых';

  @override
  String get upcoming => 'Предстоящие';

  @override
  String get yourProgress => 'Ваш прогресс';

  @override
  String trainingsCompleted(int count) {
    return '$count тренировок завершено';
  }

  @override
  String get completedTrainings => 'завершённых тренировок';

  @override
  String get partneredTrainings => 'совместных тренировок';

  @override
  String get movementFamilies => 'Семейства движений';

  @override
  String get muscleActivity => 'Мышечная активность';

  @override
  String get failedToLoadProgress => 'Не удалось загрузить прогресс';

  @override
  String get noProgressYet => 'Завершите тренировки, чтобы увидеть прогресс';

  @override
  String get calibration => 'Калибровка';

  @override
  String get calibrationNeeded =>
      'Завершите первую тренировку, чтобы Vigor откалибровал рекомендации под ваш уровень';

  @override
  String get calibrationDescription =>
      'Во время калибровки платформа собирает данные из ваших отзывов, чтобы лучше подстраивать тренировки под ваши способности';

  @override
  String get capabilities => 'Возможности';

  @override
  String get noTrainingsCompletedYet =>
      'Начни тренироваться, чтобы увидеть что-то здесь';
}
