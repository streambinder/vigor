// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Korean (`ko`).
class AppLocalizationsKo extends AppLocalizations {
  AppLocalizationsKo([String locale = 'ko']) : super(locale);

  @override
  String get appName => 'Vigor';

  @override
  String get appTagline => 'Ex Sapientia Vis';

  @override
  String get navHome => '홈';

  @override
  String get navActivity => '활동';

  @override
  String get navProfile => '프로필';

  @override
  String get storageErrorTitle => 'Vigor - 저장소 오류';

  @override
  String get storageUnavailable => '저장소를 사용할 수 없습니다';

  @override
  String get storageErrorMessage => '보안 저장소가 필요합니다. 설정을 확인하세요.';

  @override
  String get signInWithGoogle => 'Google로 로그인';

  @override
  String get signingIn => '로그인 중...';

  @override
  String get failedToInitializeGoogleSignIn => 'Google 로그인 초기화 실패';

  @override
  String signInError(String message) {
    return '로그인 오류: $message';
  }

  @override
  String get googleSignInFailed => 'Google 로그인 실패';

  @override
  String get failedToGetAuthToken => '인증 토큰을 가져올 수 없습니다';

  @override
  String errorProcessingSignIn(String message) {
    return '로그인 처리 오류: $message';
  }

  @override
  String get googleSignInInitializing => 'Google 로그인이 아직 초기화 중입니다...';

  @override
  String get readyToTrain => '운동할 준비가 되셨나요?';

  @override
  String get generateTrainingDescription => '목표에 맞는 트레이닝 생성';

  @override
  String get generateTraining => '트레이닝 생성';

  @override
  String get refresh => '새로고침';

  @override
  String get logout => '로그아웃';

  @override
  String get userDataRefreshed => '사용자 데이터가 새로고침되었습니다';

  @override
  String get editProfile => '프로필 수정';

  @override
  String get settings => '설정';

  @override
  String get other => '기타';

  @override
  String get deleteGym => '헬스장 삭제';

  @override
  String deleteGymConfirmation(String name) {
    return '\"$name\"을(를) 삭제하시겠습니까?';
  }

  @override
  String get cancel => '취소';

  @override
  String get delete => '삭제';

  @override
  String get logoutConfirmation => '로그아웃?';

  @override
  String get deleteAccount => '계정 삭제';

  @override
  String get deleteAccountConfirmation => '계정을 삭제하시겠습니까? 취소 불가.';

  @override
  String get accountDeletedSuccessfully => '계정이 성공적으로 삭제되었습니다';

  @override
  String get failedToDeleteAccount => '계정 삭제 실패';

  @override
  String get failedToLoadGyms => '헬스장 로드 실패';

  @override
  String get gymAddedSuccessfully => '헬스장이 성공적으로 추가되었습니다';

  @override
  String get failedToAddGym => '헬스장 추가 실패';

  @override
  String get gymUpdatedSuccessfully => '헬스장이 성공적으로 업데이트되었습니다';

  @override
  String get failedToUpdateGym => '헬스장 업데이트 실패';

  @override
  String get gymDeletedSuccessfully => '헬스장이 성공적으로 삭제되었습니다';

  @override
  String get failedToDeleteGym => '헬스장 삭제 실패';

  @override
  String get birthdate => '생년월일';

  @override
  String get gender => '성별';

  @override
  String get language => '언어';

  @override
  String get height => '키';

  @override
  String get weight => '체중';

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
  String get goals => '목표';

  @override
  String get injuries => '부상';

  @override
  String get limitations => '제한사항';

  @override
  String get conditions => '상태';

  @override
  String get favorites => '즐겨찾기';

  @override
  String get personalDetails => '개인 정보';

  @override
  String get healthAndGoals => '건강 및 목표';

  @override
  String get exercises => '운동';

  @override
  String get equipment => '장비';

  @override
  String startedDate(String date) {
    return '시작: $date';
  }

  @override
  String yearLabel(int year) {
    return '연도: $year';
  }

  @override
  String get myGyms => '헬스장';

  @override
  String get addGym => '헬스장 추가';

  @override
  String get noGymsAddedYet => '아직 추가된 헬스장이 없습니다';

  @override
  String get addYourFirstGym => '첫 번째 헬스장 추가';

  @override
  String get removeDefault => '기본값 제거';

  @override
  String get setAsDefault => '기본값으로 설정';

  @override
  String get edit => '수정';

  @override
  String get quickActions => '빠른 작업';

  @override
  String get dangerZone => '위험 구역';

  @override
  String get completeYourProfile => '프로필을 완성하세요';

  @override
  String get updateYourProfileInfo => '아래에서 프로필을 업데이트하세요.';

  @override
  String get pleaseCompleteProfile => '프로필을 완성하세요. * = 필수.';

  @override
  String get firstName => '이름';

  @override
  String get lastName => '성';

  @override
  String get birthDate => '생년월일';

  @override
  String get male => '남성';

  @override
  String get female => '여성';

  @override
  String get selectLanguage => '언어 선택';

  @override
  String get heightCm => '키 (cm)';

  @override
  String get weightKg => '체중 (kg)';

  @override
  String get required => '필수';

  @override
  String get invalid => '유효하지 않음';

  @override
  String get pleaseSelectBirthDate => '생년월일을 선택해 주세요';

  @override
  String get pleaseAddAtLeastOneGoal => '최소 하나의 목표를 추가해 주세요';

  @override
  String get pleaseSelectLanguage => '언어를 선택해 주세요';

  @override
  String get addAGoal => '목표 추가';

  @override
  String get injuryDescription => '부상 설명';

  @override
  String get year => '연도';

  @override
  String get addALimitation => '제한사항 추가';

  @override
  String get addACondition => '상태 추가';

  @override
  String get favoriteExercisesHint => '예: 스쿼트, 턱걸이, 달리기';

  @override
  String get favoriteEquipmentHint => '예: 덤벨, 바벨, 케틀벨';

  @override
  String get saveChanges => '변경사항 저장';

  @override
  String get saveProfile => '프로필 저장';

  @override
  String get optionalLeaveEmpty => '(선택사항)';

  @override
  String get optionalExercisesPrefer => '(선택사항)';

  @override
  String get optionalEquipmentPrefer => '(선택사항)';

  @override
  String get optionalWorkoutTypesPrefer => '(선택사항)';

  @override
  String get favoriteExercises => '좋아하는 운동';

  @override
  String get favoriteEquipment => '좋아하는 장비';

  @override
  String get favoriteWorkoutTypes => '선호하는 운동 유형';

  @override
  String get workoutTypeStrength => '근력';

  @override
  String get workoutTypeStrengthDescription => '무거운 중량과 충분한 휴식으로 최대 근력 구축';

  @override
  String get workoutTypeSupersets => '슈퍼세트';

  @override
  String get workoutTypeSupersetsDescription => '대항근을 연속으로 훈련하여 시간 효율적인 운동';

  @override
  String get workoutTypeCircuit => '서킷';

  @override
  String get workoutTypeCircuitDescription => '최소한의 휴식으로 스테이션 간 이동하며 컨디셔닝';

  @override
  String get workoutTypeEmom => 'EMOM';

  @override
  String get workoutTypeEmomDescription => '매 분: 정해진 횟수를 완료하고 다음 분까지 휴식';

  @override
  String get workoutTypeAmrap => 'AMRAP';

  @override
  String get workoutTypeAmrapDescription => '제한 시간 내 가능한 많은 라운드 완료';

  @override
  String get workoutTypeHiit => 'HIIT';

  @override
  String get workoutTypeHiitDescription => '고강도 폭발과 짧은 회복을 번갈아 수행';

  @override
  String get workoutTypeForTime => '타임 트라이얼';

  @override
  String get workoutTypeForTimeDescription => '가능한 빨리 운동 완료';

  @override
  String get workoutTypeEndurance => '지구력';

  @override
  String get workoutTypeEnduranceDescription => '중간 강도의 지속적인 운동으로 유산소 능력 향상';

  @override
  String get workoutTypeMobility => '유연성';

  @override
  String get workoutTypeMobilityDescription => '가동 범위와 관절 건강 개선';

  @override
  String get methodologyOptional => '방법론 (선택사항)';

  @override
  String get methodologyAuto => '자동';

  @override
  String get goalsOptional => '목표 (선택사항)';

  @override
  String get musclesOptional => '근육 (선택사항)';

  @override
  String get musclesAuto => '전체';

  @override
  String get advancedSettings => '고급';

  @override
  String get failedToUpdateProfile => '프로필 업데이트 실패';

  @override
  String get activity => '활동';

  @override
  String get noTrainingsYet => '아직 트레이닝이 없습니다';

  @override
  String get generateFirstTraining => '홈에서 첫 트레이닝 생성';

  @override
  String get noTrainingAvailable => '트레이닝 없음. 생성하세요.';

  @override
  String get availableTrainings => '사용 가능한 트레이닝';

  @override
  String get pastTrainings => '과거 트레이닝';

  @override
  String get stale => '만료됨';

  @override
  String get today => '오늘';

  @override
  String get yesterday => '어제';

  @override
  String daysAgo(int count) {
    return '$count일 전';
  }

  @override
  String get available => '이용 가능';

  @override
  String get completed => '완료됨';

  @override
  String get completedSingular => '완료';

  @override
  String get noPastTrainings => '완료된 훈련이 아직 없습니다';

  @override
  String get copied => '복사됨';

  @override
  String durationMin(int minutes) {
    return '$minutes분';
  }

  @override
  String durationHr(int hours) {
    return '$hours시간';
  }

  @override
  String durationHrMin(int hours, int minutes) {
    return '$hours시간 $minutes분';
  }

  @override
  String get failedToLoadTrainings => '트레이닝 로드 실패';

  @override
  String get startTraining => '트레이닝 시작';

  @override
  String get cloneTraining => '트레이닝 복제';

  @override
  String get addPartner => '파트너 추가';

  @override
  String get shareWithUser => '사용자와 공유';

  @override
  String get deleteTraining => '트레이닝 삭제';

  @override
  String get leaveTraining => '트레이닝 나가기';

  @override
  String get showAiReasoning => 'AI 추론 표시';

  @override
  String get reportIssue => '문제 보고';

  @override
  String deleteTrainingConfirmation(String name) {
    return '\"$name\" 삭제? 취소 불가.';
  }

  @override
  String leaveTrainingConfirmation(String name) {
    return '\"$name\"에서 나가기? 더 이상 표시 안 됨.';
  }

  @override
  String addPartnerConfirmation(String userName, String trainingName) {
    return '$userName을(를) \"$trainingName\"의 파트너로 추가하시겠습니까?';
  }

  @override
  String cloneTrainingConfirmation(String name) {
    return '\"$name\"을(를) 트레이닝에 복제하시겠습니까?';
  }

  @override
  String shareTrainingConfirmation(String trainingName, String userName) {
    return '\"$trainingName\"을(를) $userName과(와) 공유하시겠습니까?';
  }

  @override
  String get trainingDeletedSuccessfully => '트레이닝이 성공적으로 삭제되었습니다';

  @override
  String get failedToDeleteTraining => '트레이닝 삭제 실패';

  @override
  String get leftTrainingSuccessfully => '트레이닝에서 성공적으로 나갔습니다';

  @override
  String get partnerAddedSuccessfully => '파트너가 성공적으로 추가되었습니다';

  @override
  String get failedToAddPartner => '파트너 추가 실패';

  @override
  String get trainingSharedSuccessfully => '트레이닝이 성공적으로 공유되었습니다';

  @override
  String get failedToShareTraining => '트레이닝 공유 실패';

  @override
  String get trainingCloned => '트레이닝이 복제되었습니다';

  @override
  String get failedToCloneTraining => '트레이닝 복제 실패';

  @override
  String get trainingMarkedAsComplete => '트레이닝이 완료되었습니다';

  @override
  String get failedToCompleteTraining => '트레이닝 완료 실패';

  @override
  String get feedback => '피드백';

  @override
  String get feedbackUpdated => '피드백이 업데이트되었습니다';

  @override
  String get failedToUpdateFeedback => '피드백 업데이트 실패';

  @override
  String get reportSubmitted => '보고서가 제출되었습니다';

  @override
  String get failedToSubmitReport => '보고서 제출 실패';

  @override
  String get shuffleExercise => '운동 바꾸기';

  @override
  String get exerciseShuffled => '운동이 변경됨';

  @override
  String get failedToShuffleExercise => '운동 변경 실패';

  @override
  String get reasoning => '추론';

  @override
  String get strategy => '전략';

  @override
  String get typeSelection => '유형 선택';

  @override
  String get progression => '진행';

  @override
  String get constraints => '제약';

  @override
  String get researchApplied => '적용된 연구';

  @override
  String get targetMuscles => '타겟 근육';

  @override
  String get naming => '명명';

  @override
  String get trainingRoutines => '트레이닝';

  @override
  String get noEquipment => '장비 없음';

  @override
  String blockNumber(int number) {
    return '블록 $number';
  }

  @override
  String repeatsCount(int count) {
    return '${count}x';
  }

  @override
  String durationSeconds(int seconds) {
    return '$seconds초';
  }

  @override
  String restSeconds(int seconds) {
    return '$seconds초 휴식';
  }

  @override
  String repsCount(int count) {
    return '$count회';
  }

  @override
  String weightKgValue(double value) {
    return '$value kg';
  }

  @override
  String get markAsComplete => '완료로 표시';

  @override
  String get updateFeedback => '피드백 수정';

  @override
  String get references => '참조';

  @override
  String get literature => '문헌';

  @override
  String get request => 'Request';

  @override
  String get describeIssue => '문제를 설명해 주세요...';

  @override
  String get submit => '제출';

  @override
  String get close => '닫기';

  @override
  String get add => '추가';

  @override
  String get update => '업데이트';

  @override
  String get clone => '복제';

  @override
  String get share => '공유';

  @override
  String get leave => '나가기';

  @override
  String get tapToStart => '탭하여 시작';

  @override
  String get tapWhenDone => '완료하면 탭';

  @override
  String get trainingCompleted => '트레이닝 완료!';

  @override
  String greatJobCompleting(String name) {
    return '$name 잘하셨습니다';
  }

  @override
  String get done => '완료';

  @override
  String get complete => '완료';

  @override
  String routineCounter(int current, int total) {
    return '루틴 $current/$total';
  }

  @override
  String blockCounter(int current, int total) {
    return '블록 $current/$total';
  }

  @override
  String get exitTraining => '트레이닝을 종료하시겠습니까?';

  @override
  String get whatWouldYouLikeToDo => '다음은?';

  @override
  String get exit => '종료';

  @override
  String get continueTraining => '계속';

  @override
  String get stop => '중지';

  @override
  String get stopTraining => '트레이닝을 중지하시겠습니까?';

  @override
  String get stopTrainingConfirm => '타이머 진행 상황이 손실됩니다.';

  @override
  String get failedToMarkComplete => '트레이닝 완료 실패';

  @override
  String get durationMinutes => '시간 (분)';

  @override
  String get bodyweight => '맨몸';

  @override
  String get gym => '헬스장';

  @override
  String get custom => '기타';

  @override
  String get noEquipmentBodyweightOnly => '맨몸만';

  @override
  String get noGymsDefinedCreateOne => '헬스장 없음. 프로필에서 생성하세요.';

  @override
  String get selectAGym => '헬스장 선택';

  @override
  String get addEquipment => '장비 추가';

  @override
  String get addEquipmentAvailable => '사용 가능한 장비 추가';

  @override
  String get includeWarmupCooldown => '워밍업 및 쿨다운 포함';

  @override
  String get equipmentPlaceholder => '예: 바벨, 덤벨';

  @override
  String get customPromptOptional => '사용자 정의 프롬프트 (선택사항)';

  @override
  String get focusOnUpperBody => '예: 상체에 집중';

  @override
  String get trainingPartnersOptional => '파트너';

  @override
  String get generatingTraining => '트레이닝을 생성하는 중...';

  @override
  String get thisMayTakeAMoment => '잠시 시간이 걸릴 수 있습니다';

  @override
  String generationFailedRetrying(int attempt) {
    return '생성 실패, 재시도 #$attempt...';
  }

  @override
  String get trainingGeneratedSuccessfully => '트레이닝이 성공적으로 생성되었습니다!';

  @override
  String get failedToGenerateTraining => '트레이닝 생성 실패';

  @override
  String get generate => '생성';

  @override
  String get editGym => '헬스장 수정';

  @override
  String get gymName => '헬스장 이름';

  @override
  String get gymNamePlaceholder => '예: 홈짐, 스포애니';

  @override
  String get availableWeights => '사용 가능한 중량';

  @override
  String get availableWeightsHint => '이 헬스장의 중량 모디파이어에 사용할 중량 옵션을 설정하세요.';

  @override
  String get noEquipmentAddedYet => '아직 추가된 장비가 없습니다';

  @override
  String get pleaseEnterGymName => '헬스장 이름을 입력해 주세요';

  @override
  String get addAllEquipment => '모두 추가';

  @override
  String get failedToLoadEquipment => '장비를 불러오지 못했습니다';

  @override
  String get selectUser => '사용자 선택';

  @override
  String get searchByName => '이름으로 검색';

  @override
  String get noUsersAvailable => '사용 가능한 사용자가 없습니다';

  @override
  String get noMatchingUsers => '일치하는 사용자가 없습니다';

  @override
  String get noMatchingEquipment => '일치하는 장비가 없습니다';

  @override
  String get instructions => '지침';

  @override
  String get cues => '포인트';

  @override
  String get howWasYourTraining => '트레이닝은 어떠셨나요?';

  @override
  String get anyAdditionalComments => '추가 의견이 있으신가요?';

  @override
  String get actualDuration => '실제 소요 시간 (분)';

  @override
  String get impossible => '못 함';

  @override
  String get tooHard => '너무 어려움';

  @override
  String get ok => '적절함';

  @override
  String get easy => '쉬움';

  @override
  String get tooEasy => '너무 쉬움';

  @override
  String get flag => '표시';

  @override
  String get profile => '프로필';

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
    return '다음: $name';
  }

  @override
  String get rest => '휴식';

  @override
  String get upcoming => '다음 예정';

  @override
  String get yourProgress => '나의 진행 상황';

  @override
  String trainingsCompleted(int count) {
    return '$count회 훈련 완료';
  }

  @override
  String get completedTrainings => '완료한 훈련';

  @override
  String get partneredTrainings => '파트너 훈련';

  @override
  String get movementFamilies => '동작 카테고리';

  @override
  String get muscleActivity => '근육 활동';

  @override
  String get failedToLoadProgress => '진행 상황을 불러오지 못했습니다';

  @override
  String get noProgressYet => '훈련을 완료하여 진행 상황을 확인하세요';

  @override
  String get calibration => '보정';

  @override
  String get calibrationGlobal => '전체';

  @override
  String get calibrationNeeded =>
      '첫 번째 훈련을 완료하여 Vigor가 당신의 수준에 맞게 추천을 보정할 수 있도록 하세요';

  @override
  String get calibrationDescription =>
      '보정 기간 동안 플랫폼은 피드백 데이터를 수집하여 당신의 체력 상태와 수준에 대한 초기 평가를 합니다';

  @override
  String get calibrationInProgress => '매 세션마다 당신의 운동이 더 스마트해집니다';

  @override
  String get calibrationTrainingNote =>
      '이 운동은 목표와 완전히 일치하지 않을 수 있습니다 — 시스템이 아직 당신의 피트니스 수준을 학습 중이며, 완전한 프로필을 구축하기 위해 동작 다양성을 우선시하고 있습니다';

  @override
  String calibrationFamiliesLearned(int calibrated, int total) {
    return '$calibrated/$total 동작 패턴 학습됨';
  }

  @override
  String get capabilities => '능력';

  @override
  String get muscleHeatMap => '근육 히트맵';

  @override
  String get heatResting => '휴식 중';

  @override
  String get heatRecovered => '회복됨';

  @override
  String get heatActive => '활동 중';

  @override
  String get heatWarm => '따뜻함';

  @override
  String get heatHot => '고강도';

  @override
  String get noTrainingsCompletedYet => '훈련을 시작해서 여기에 뭔가 보여주세요';

  @override
  String get theme => '테마';

  @override
  String get themeAuto => '자동';

  @override
  String get themeLight => '라이트';

  @override
  String get themeDark => '다크';

  @override
  String get themeAutoDescription => '시스템 설정 따르기';

  @override
  String get appearance => '외관';

  @override
  String get trainingDefaults => '기본값';

  @override
  String get defaultDuration => '훈련 시간';

  @override
  String get warmupCooldown => '워밍업 및 쿨다운';

  @override
  String get timer => '타이머';

  @override
  String get intervalJingle => '인터벌 전환 시 호루라기';

  @override
  String get duckOtherAudio => '다른 오디오 줄이기';

  @override
  String get duckOtherAudioDescription => '호루라기 중 음악 및 다른 앱의 볼륨을 줄입니다';

  @override
  String get goalHypertrophy => '근육 증가';

  @override
  String get goalHypertrophyDescription => '타겟 저항 훈련으로 근육량 증가';

  @override
  String get goalFatLoss => '체지방 감소';

  @override
  String get goalFatLossDescription => '고강도 운동으로 칼로리 소모 및 체지방 감소';

  @override
  String get goalToning => '탄탄한 몸매';

  @override
  String get goalToningDescription => '선명한 근육과 탄탄한 외모 만들기';

  @override
  String get goalPosture => '자세 교정';

  @override
  String get goalPostureDescription => '등과 코어 강화로 바른 자세 유지';

  @override
  String get goalRehabilitation => '재활';

  @override
  String get goalRehabilitationDescription => '부상 회복을 위한 안전하고 통제된 운동';

  @override
  String get goalWellness => '웰니스';

  @override
  String get goalWellnessDescription => '건강과 스트레스 해소를 위한 균형 잡힌 운동';

  @override
  String get goalFlexibility => '유연성';

  @override
  String get goalFlexibilityDescription => '스트레칭과 이동성으로 가동 범위 개선';

  @override
  String get goalSports => '스포츠 퍼포먼스';

  @override
  String get goalSportsDescription => '파워와 민첩성 훈련으로 운동 능력 향상';

  @override
  String get thisWeek => '이번 주';

  @override
  String get trainingPlan => '트레이닝 플랜';

  @override
  String get sessionsPerWeek => '주당 세션';

  @override
  String get sessionDuration => '세션 시간';

  @override
  String get preferredTime => '권장 시간';

  @override
  String get recommendedTime => '추천 시간';

  @override
  String get methodologyMix => '방법론 믹스';

  @override
  String get pastWeeks => '지난 주';

  @override
  String get weeklyTarget => '주간 목표';

  @override
  String daysLeft(int count) {
    return '$count일 남음';
  }

  @override
  String get recommended => '추천';

  @override
  String get duration => '시간';

  @override
  String get familyHorizontalPush => '푸시';

  @override
  String get familyHorizontalPull => '풀';

  @override
  String get familyVerticalPush => '오버헤드';

  @override
  String get familyVerticalPull => '턱걸이';

  @override
  String get familySquat => '스쿼트';

  @override
  String get familyHinge => '힌지';

  @override
  String get familyCore => '코어';

  @override
  String get familyCarry => '캐리';

  @override
  String get familyCardio => '유산소';

  @override
  String get familyMobility => '이동성';

  @override
  String get familyBalance => '균형';

  @override
  String get trainingQuality => '이 트레이닝은 어떠셨나요?';

  @override
  String get trainingQualityHint => 'AI 트레이닝 품질 평가에 도움이 됩니다';

  @override
  String get qualityReasonHint => '개선할 점이 있나요?';

  @override
  String get good => '좋음';

  @override
  String get bad => '나쁨';

  @override
  String get loadingMsg1 => '프로필 분석 중...';

  @override
  String get loadingMsg2 => '운동 선택 중...';

  @override
  String get loadingMsg3 => '루틴 구성 중...';

  @override
  String get loadingMsg4 => '트레이닝 볼륨 계산 중...';

  @override
  String get loadingMsg5 => '휴식 간격 최적화 중...';

  @override
  String get loadingMsg6 => '운동 연구 참조 중...';

  @override
  String get loadingMsg7 => '근육군 균형 조정 중...';

  @override
  String get loadingMsg8 => '진행 경로 설계 중...';

  @override
  String get loadingMsg9 => '강도 미세 조정 중...';

  @override
  String get loadingMsg10 => '동작 패턴 검토 중...';

  @override
  String get loadingMsg11 => '회복 필요량 평가 중...';

  @override
  String get loadingMsg12 => '운동 변형 선택 중...';

  @override
  String get loadingMsg13 => '트레이닝 블록 구조화 중...';

  @override
  String get loadingMsg14 => '작업 간격 설정 중...';

  @override
  String get loadingMsg15 => '운동 과학 적용 중...';

  @override
  String get loadingMsg16 => '워밍업 설계 중...';

  @override
  String get loadingMsg17 => '동작 패밀리 매핑 중...';

  @override
  String get loadingMsg18 => '부하 분배 평가 중...';

  @override
  String get loadingMsg19 => '세션 개인화 중...';

  @override
  String get loadingMsg20 => '거의 완료...';

  @override
  String loadingMsgGoal(String goal) {
    return '$goal 최적화 중...';
  }

  @override
  String get loadingMsgInjuries => '부상에 맞춰 조정 중...';

  @override
  String get loadingMsgFavorites => '선호 운동 우선 배치 중...';

  @override
  String get loadingMsgConditions => '컨디션에 맞춰 조정 중...';

  @override
  String loadingMsgMethodology(String methodology) {
    return '$methodology 세션 설계 중...';
  }

  @override
  String get loadingMsgPartners => '파트너 운동 조정 중...';

  @override
  String loadingMsgGym(String gym) {
    return '$gym 장비 로딩 중...';
  }

  @override
  String get loadingMsgHistory => '최근 세션 분석 중...';

  @override
  String get loadingRetryMsg1 => '음, 잘 안 됐어요 — 다시 시도 중';

  @override
  String get loadingRetryMsg2 => '다시 해볼게요...';

  @override
  String get loadingRetryMsg3 => '아쉽네요 — 한 번 더';

  @override
  String get loadingRetryMsg4 => '한 번 더, 잠시만요';

  @override
  String get loadingRetryMsg5 => '이런, 재조정 중...';

  @override
  String get loadingRetryMsg6 => '거의 됐어요 — 다시 시도';

  @override
  String nSelected(int count) {
    return '$count개 선택됨';
  }

  @override
  String deleteSelectedTrainings(int count) {
    return '$count개 훈련을 삭제하시겠습니까?';
  }

  @override
  String get trainingsDeletedSuccessfully => '훈련이 삭제되었습니다';

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
  String get pendingFeedbacks => '대기 중인 피드백';

  @override
  String get pendingFeedbacksDescription =>
      '완료된 훈련 중 피드백이 없는 항목이 있습니다. 피드백은 향후 훈련 추천을 개선하는 데 도움이 됩니다.';

  @override
  String get equipmentPartner => '파트너';

  @override
  String get equipmentBalanceBoard => '밸런스 보드';

  @override
  String get equipmentBand => '밴드';

  @override
  String get equipmentBarbell => '바벨';

  @override
  String get equipmentBench => '벤치';

  @override
  String get equipmentBox => '박스';

  @override
  String get equipmentBosuBall => '보수볼';

  @override
  String get equipmentCable => '케이블 머신';

  @override
  String get equipmentDipStation => '딥 스테이션';

  @override
  String get equipmentDumbbell => '덤벨';

  @override
  String get equipmentEllipticalMachine => '일립티컬 머신';

  @override
  String get equipmentEzBarbell => 'EZ바';

  @override
  String get equipmentHammer => '해머';

  @override
  String get equipmentKettlebell => '케틀벨';

  @override
  String get equipmentLeverageMachine => '레버리지 머신';

  @override
  String get equipmentMedicineBall => '메디신볼';

  @override
  String get equipmentOlympicBarbell => '올림픽 바벨';

  @override
  String get equipmentPullUpBar => '턱걸이 바';

  @override
  String get equipmentResistanceBand => '저항 밴드';

  @override
  String get equipmentRings => '링';

  @override
  String get equipmentRoller => '폼롤러';

  @override
  String get equipmentRope => '줄';

  @override
  String get equipmentRowingMachine => '로잉 머신';

  @override
  String get equipmentSkiergMachine => '스키에르그';

  @override
  String get equipmentSledMachine => '슬레드 머신';

  @override
  String get equipmentSmithMachine => '스미스 머신';

  @override
  String get equipmentStabilityBall => '짐볼';

  @override
  String get equipmentStationaryBike => '실내 자전거';

  @override
  String get equipmentStepmillMachine => '스텝밀';

  @override
  String get equipmentTire => '타이어';

  @override
  String get equipmentTreadmill => '트레드밀';

  @override
  String get equipmentTrapBar => '트랩바';

  @override
  String get equipmentTrx => 'TRX';

  @override
  String get equipmentUpperBodyErgometer => '상체 에르고미터';

  @override
  String get equipmentWheelRoller => '복근 롤러';

  @override
  String get muscleChest => '가슴';

  @override
  String get muscleBack => '등';

  @override
  String get muscleShoulders => '어깨';

  @override
  String get muscleArms => '팔';

  @override
  String get muscleCore => '코어';

  @override
  String get muscleGlutes => '둔근';

  @override
  String get muscleLegs => '다리';

  @override
  String get modifierWeightedVest => '웨이트 조끼';

  @override
  String get modifierParallettes => '파랄렛';

  @override
  String get modifierAnkleWeights => '발목 중량';

  @override
  String get modifierDipBelt => '딥 벨트';

  @override
  String get modifierPushUpBars => '푸시업 바';

  @override
  String get modifierResistanceBands => '저항 밴드';

  @override
  String get modifierWeight => '중량';

  @override
  String get modifierWristWeights => '손목 웨이트';

  @override
  String get healthData => '건강 데이터';

  @override
  String get healthConnected => '연결됨';

  @override
  String get healthSynchronizing => '동기화 중...';

  @override
  String get healthSynchronized => '동기화 완료';

  @override
  String get healthSynchronize => '지금 동기화';

  @override
  String get healthNotConnected => '연결 안 됨';

  @override
  String get healthNativeOnly => 'iOS 및 Android에서 사용 가능';

  @override
  String healthBuildingBaselines(int days) {
    return '개인 기준값 생성 중 ($days/14일)';
  }

  @override
  String get healthSyncNoData => '기기에서 건강 데이터를 찾을 수 없습니다';

  @override
  String healthLastSyncAt(String date, String type) {
    return '마지막 동기화: $date · $type';
  }

  @override
  String get healthSyncTypeFull => '전체';

  @override
  String get healthSyncTypeIncremental => '증분';

  @override
  String healthDeviceData(int days, int sessions) {
    return '기기: $days일, $sessions세션';
  }

  @override
  String healthBackendData(int days, int sessions) {
    return '백엔드: $days일, $sessions세션';
  }

  @override
  String healthDateRange(String from, String to) {
    return '$from – $to';
  }

  @override
  String get healthDisconnect => '연결 해제 및 데이터 삭제';

  @override
  String get healthDisconnectConfirmation =>
      '건강 데이터를 연결 해제하시겠습니까? 동기화된 모든 데이터가 삭제됩니다. 이 작업은 되돌릴 수 없습니다.';

  @override
  String get healthDisconnectedSuccessfully => '건강 데이터 연결이 해제되었습니다';

  @override
  String get failedToDisconnectHealth => '건강 데이터 연결 해제 실패';

  @override
  String get healthConnect => '연결';

  @override
  String get healthMetrics => '건강 데이터';

  @override
  String get healthAdjustment => '건강 상태 조정';

  @override
  String get healthPermissionsTitle => '웨어러블 연결';

  @override
  String get healthPermissionsDescription =>
      'Vigor가 건강 데이터를 읽어 운동을 맞춤화합니다. 수면, 회복, 활동 데이터가 많을수록 더 스마트한 트레이닝 추천이 가능합니다.';

  @override
  String get healthPermissionsReadOnly => '읽기 전용';

  @override
  String get healthPermissionsSleep => '수면 시간 및 단계';

  @override
  String get healthPermissionsHrv => '심박변이도 (HRV)';

  @override
  String get healthPermissionsRhr => '안정시 심박수';

  @override
  String get healthPermissionsSteps => '일일 걸음 수';

  @override
  String get healthPermissionsWorkouts => '운동 세션 및 심박수';

  @override
  String get healthPermissionsGrant => '건강 데이터 연결';

  @override
  String get healthPermissionsSkip => '나중에';

  @override
  String get healthPermissionsGranted => '건강 데이터가 연결되었습니다';

  @override
  String get healthPermissionsDenied => '권한이 부여되지 않았습니다';

  @override
  String get healthOnboardingTitle => '웨어러블 연결';

  @override
  String get healthOnboardingDescription =>
      '건강 데이터를 연결하면 Vigor가 수면, 회복, 활동에 맞춰 트레이닝을 조정합니다.';

  @override
  String get healthOnboardingConnect => '연결';

  @override
  String get healthOnboardingDismiss => '나중에 하기';

  @override
  String get healthInstallHcTitle => 'Health Connect 필요';

  @override
  String get healthInstallHcDescription =>
      '웨어러블 데이터 동기화를 위해 Health Connect가 필요합니다. Play 스토어에서 설치해주세요.';

  @override
  String get healthInstallHc => 'Health Connect 설치';

  @override
  String get heartRate => '심박수';

  @override
  String get avgHr => '평균 심박';

  @override
  String get maxHr => '최대 심박';

  @override
  String get bpm => 'bpm';

  @override
  String get hrZones => '심박 존';

  @override
  String get hrZone1 => '존 1';

  @override
  String get hrZone2 => '존 2';

  @override
  String get hrZone3 => '존 3';

  @override
  String get hrZone4 => '존 4';

  @override
  String get hrZone5 => '존 5';

  @override
  String get healthDailySleep => '수면';

  @override
  String get healthDailyRestingHr => '안정시 심박';

  @override
  String get healthDailyHrv => 'HRV';

  @override
  String get healthDailySteps => '걸음 수';

  @override
  String get healthDailyCalories => '칼로리';

  @override
  String get healthDailyNoData => '건강 데이터 없음';

  @override
  String get externalWorkout => '외부 운동';

  @override
  String get exerciseTypeRunning => '달리기';

  @override
  String get exerciseTypeWalking => '걷기';

  @override
  String get exerciseTypeBiking => '자전거';

  @override
  String get exerciseTypeYoga => '요가';

  @override
  String get exerciseTypeSwimming => '수영';

  @override
  String get exerciseTypeHiking => '등산';

  @override
  String get exerciseTypeStrengthTraining => '근력 운동';

  @override
  String get exerciseTypeFunctionalStrengthTraining => '기능성 운동';

  @override
  String get exerciseTypeTraditionalStrengthTraining => '근력 운동';

  @override
  String get exerciseTypeRunningTreadmill => '트레드밀';

  @override
  String get exerciseTypeBikingStationary => '실내 자전거';

  @override
  String get exerciseTypeWalkingTreadmill => '트레드밀 걷기';

  @override
  String get exerciseTypeRowing => '조정';

  @override
  String get exerciseTypePilates => '필라테스';

  @override
  String get exerciseTypeDancing => '댄스';

  @override
  String get exerciseTypeElliptical => '일립티컬';

  @override
  String get exerciseTypeStairClimbing => '계단 오르기';

  @override
  String get exerciseTypeOther => '운동';
}
