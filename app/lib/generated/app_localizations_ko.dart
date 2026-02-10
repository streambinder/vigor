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
  String get workoutTypeCircuit => '서킷';

  @override
  String get workoutTypeEmom => 'EMOM';

  @override
  String get workoutTypeAmrap => 'AMRAP';

  @override
  String get workoutTypeHiit => 'HIIT';

  @override
  String get workoutTypeForTime => '타임 트라이얼';

  @override
  String get workoutTypeEndurance => '지구력';

  @override
  String get workoutTypeMobility => '유연성';

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
  String get instructions => '지침';

  @override
  String get howWasYourTraining => '트레이닝은 어떠셨나요?';

  @override
  String get anyAdditionalComments => '추가 의견이 있으신가요?';

  @override
  String get tooEasy => '너무 쉬움';

  @override
  String get tooHard => '너무 어려움';

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
  String get capabilities => '능력';

  @override
  String get muscleHeatMap => 'Muscle Heat Map';

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
  String get intervalJingle => '인터벌 완료 시 소리';

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
}
