// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Japanese (`ja`).
class AppLocalizationsJa extends AppLocalizations {
  AppLocalizationsJa([String locale = 'ja']) : super(locale);

  @override
  String get appName => 'Vigor';

  @override
  String get appTagline => 'Ex Sapientia Vis';

  @override
  String get navHome => 'ホーム';

  @override
  String get navActivity => 'アクティビティ';

  @override
  String get navProfile => 'プロフィール';

  @override
  String get storageErrorTitle => 'Vigor - ストレージエラー';

  @override
  String get storageUnavailable => 'ストレージが利用できません';

  @override
  String get storageErrorMessage => '安全なストレージが必要です。設定を確認してください。';

  @override
  String get signInWithGoogle => 'Googleでサインイン';

  @override
  String get signingIn => 'サインイン中...';

  @override
  String get failedToInitializeGoogleSignIn => 'Google サインインの初期化に失敗しました';

  @override
  String signInError(String message) {
    return 'サインインエラー：$message';
  }

  @override
  String get googleSignInFailed => 'Googleサインインに失敗しました';

  @override
  String get failedToGetAuthToken => '認証トークンを取得できませんでした';

  @override
  String errorProcessingSignIn(String message) {
    return 'サインイン処理エラー：$message';
  }

  @override
  String get googleSignInInitializing => 'Google サインインはまだ初期化中です...';

  @override
  String get readyToTrain => 'トレーニングの準備はできましたか？';

  @override
  String get generateTrainingDescription => '目標に合わせたトレーニングを作成';

  @override
  String get generateTraining => 'トレーニングを生成';

  @override
  String get refresh => '更新';

  @override
  String get logout => 'ログアウト';

  @override
  String get userDataRefreshed => 'ユーザーデータを更新しました';

  @override
  String get editProfile => 'プロフィールを編集';

  @override
  String get settings => '設定';

  @override
  String get other => 'その他';

  @override
  String get deleteGym => 'ジムを削除';

  @override
  String deleteGymConfirmation(String name) {
    return '「$name」を削除しますか？';
  }

  @override
  String get cancel => 'キャンセル';

  @override
  String get delete => '削除';

  @override
  String get logoutConfirmation => 'ログアウトしますか？';

  @override
  String get deleteAccount => 'アカウントを削除';

  @override
  String get deleteAccountConfirmation => 'アカウントを削除しますか？元に戻せません。';

  @override
  String get accountDeletedSuccessfully => 'アカウントが正常に削除されました';

  @override
  String get failedToDeleteAccount => 'アカウントの削除に失敗しました';

  @override
  String get failedToLoadGyms => 'ジムの読み込みに失敗しました';

  @override
  String get gymAddedSuccessfully => 'ジムが正常に追加されました';

  @override
  String get failedToAddGym => 'ジムの追加に失敗しました';

  @override
  String get gymUpdatedSuccessfully => 'ジムが正常に更新されました';

  @override
  String get failedToUpdateGym => 'ジムの更新に失敗しました';

  @override
  String get gymDeletedSuccessfully => 'ジムが正常に削除されました';

  @override
  String get failedToDeleteGym => 'ジムの削除に失敗しました';

  @override
  String get birthdate => '生年月日';

  @override
  String get gender => '性別';

  @override
  String get language => '言語';

  @override
  String get height => '身長';

  @override
  String get weight => '体重';

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
  String get goals => '目標';

  @override
  String get injuries => '怪我';

  @override
  String get limitations => '制限';

  @override
  String get conditions => '体調';

  @override
  String get favorites => 'お気に入り';

  @override
  String get personalDetails => '個人情報';

  @override
  String get healthAndGoals => '健康と目標';

  @override
  String get exercises => 'エクササイズ';

  @override
  String get equipment => '器具';

  @override
  String startedDate(String date) {
    return '開始：$date';
  }

  @override
  String yearLabel(int year) {
    return '年：$year';
  }

  @override
  String get myGyms => 'ジム';

  @override
  String get addGym => 'ジムを追加';

  @override
  String get noGymsAddedYet => 'まだジムが追加されていません';

  @override
  String get addYourFirstGym => '最初のジムを追加';

  @override
  String get removeDefault => 'デフォルトを解除';

  @override
  String get setAsDefault => 'デフォルトに設定';

  @override
  String get edit => '編集';

  @override
  String get quickActions => 'クイックアクション';

  @override
  String get dangerZone => '危険ゾーン';

  @override
  String get completeYourProfile => 'プロフィールを完成させてください';

  @override
  String get updateYourProfileInfo => '以下でプロフィールを更新。';

  @override
  String get pleaseCompleteProfile => 'プロフィールを完成させてください。* = 必須。';

  @override
  String get firstName => '名';

  @override
  String get lastName => '姓';

  @override
  String get birthDate => '生年月日';

  @override
  String get male => '男性';

  @override
  String get female => '女性';

  @override
  String get selectLanguage => '言語を選択';

  @override
  String get heightCm => '身長（cm）';

  @override
  String get weightKg => '体重（kg）';

  @override
  String get required => '必須';

  @override
  String get invalid => '無効';

  @override
  String get pleaseSelectBirthDate => '生年月日を選択してください';

  @override
  String get pleaseAddAtLeastOneGoal => '少なくとも1つの目標を追加してください';

  @override
  String get pleaseSelectLanguage => '言語を選択してください';

  @override
  String get addAGoal => '目標を追加';

  @override
  String get injuryDescription => '怪我の説明';

  @override
  String get year => '年';

  @override
  String get addALimitation => '制限を追加';

  @override
  String get addACondition => '体調を追加';

  @override
  String get favoriteExercisesHint => '例：スクワット、懸垂、ランニング';

  @override
  String get favoriteEquipmentHint => '例：ダンベル、バーベル、ケトルベル';

  @override
  String get saveChanges => '変更を保存';

  @override
  String get saveProfile => 'プロフィールを保存';

  @override
  String get optionalLeaveEmpty => '（任意）';

  @override
  String get optionalExercisesPrefer => '（任意）';

  @override
  String get optionalEquipmentPrefer => '（任意）';

  @override
  String get optionalWorkoutTypesPrefer => '（任意）';

  @override
  String get favoriteExercises => 'お気に入りのエクササイズ';

  @override
  String get favoriteEquipment => 'お気に入りの器具';

  @override
  String get favoriteWorkoutTypes => '好みのワークアウトタイプ';

  @override
  String get workoutTypeStrength => '筋力';

  @override
  String get workoutTypeStrengthDescription => '重い負荷と十分な休息で最大筋力を構築';

  @override
  String get workoutTypeSupersets => 'スーパーセット';

  @override
  String get workoutTypeSupersetsDescription => '拮抗筋を連続して鍛え、時間効率の良いトレーニング';

  @override
  String get workoutTypeCircuit => 'サーキット';

  @override
  String get workoutTypeCircuitDescription => '最小限の休息でステーション間を移動してコンディショニング';

  @override
  String get workoutTypeEmom => 'EMOM';

  @override
  String get workoutTypeEmomDescription => '毎分：回数を完了し次の分まで休息';

  @override
  String get workoutTypeAmrap => 'AMRAP';

  @override
  String get workoutTypeAmrapDescription => '制限時間内にできるだけ多くのラウンドを完了';

  @override
  String get workoutTypeHiit => 'HIIT';

  @override
  String get workoutTypeHiitDescription => '高強度バーストと短い回復を交互に実施';

  @override
  String get workoutTypeForTime => 'タイムトライアル';

  @override
  String get workoutTypeForTimeDescription => 'できるだけ速くワークアウトを完了';

  @override
  String get workoutTypeEndurance => '持久力';

  @override
  String get workoutTypeEnduranceDescription => '中程度の強度で持続的な運動により有酸素能力を向上';

  @override
  String get workoutTypeMobility => 'モビリティ';

  @override
  String get workoutTypeMobilityDescription => '可動域と関節の健康を改善';

  @override
  String get methodologyOptional => 'メソッド（任意）';

  @override
  String get methodologyAuto => '自動';

  @override
  String get goalsOptional => '目標（任意）';

  @override
  String get musclesOptional => '筋肉（任意）';

  @override
  String get musclesAuto => 'Auto';

  @override
  String get advancedSettings => '詳細設定';

  @override
  String get failedToUpdateProfile => 'プロフィールの更新に失敗しました';

  @override
  String get activity => 'アクティビティ';

  @override
  String get noTrainingsYet => 'まだトレーニングがありません';

  @override
  String get generateFirstTraining => 'ホームから最初のトレーニングを作成';

  @override
  String get noTrainingAvailable => 'トレーニングがありません。生成してください。';

  @override
  String get availableTrainings => '利用可能なトレーニング';

  @override
  String get pastTrainings => '過去のトレーニング';

  @override
  String get stale => '期限切れ';

  @override
  String get today => '今日';

  @override
  String get yesterday => '昨日';

  @override
  String daysAgo(int count) {
    return '$count日前';
  }

  @override
  String get available => '利用可能';

  @override
  String get completed => '完了';

  @override
  String get completedSingular => '完了';

  @override
  String get noPastTrainings => '完了したトレーニングはまだありません';

  @override
  String get copied => 'コピー済み';

  @override
  String durationMin(int minutes) {
    return '$minutes分';
  }

  @override
  String durationHr(int hours) {
    return '$hours時間';
  }

  @override
  String durationHrMin(int hours, int minutes) {
    return '$hours時間$minutes分';
  }

  @override
  String get failedToLoadTrainings => 'トレーニングの読み込みに失敗しました';

  @override
  String get startTraining => 'トレーニングを開始';

  @override
  String get cloneTraining => 'トレーニングを複製';

  @override
  String get addPartner => 'パートナーを追加';

  @override
  String get shareWithUser => 'ユーザーと共有';

  @override
  String get deleteTraining => 'トレーニングを削除';

  @override
  String get leaveTraining => 'トレーニングから退出';

  @override
  String get showAiReasoning => 'AI推論を表示';

  @override
  String get reportIssue => '問題を報告';

  @override
  String deleteTrainingConfirmation(String name) {
    return '「$name」を削除しますか？元に戻せません。';
  }

  @override
  String leaveTrainingConfirmation(String name) {
    return '「$name」から退出しますか？表示されなくなります。';
  }

  @override
  String addPartnerConfirmation(String userName, String trainingName) {
    return '$userNameを「$trainingName」のパートナーとして追加しますか？';
  }

  @override
  String cloneTrainingConfirmation(String name) {
    return '「$name」をトレーニングに複製しますか？';
  }

  @override
  String shareTrainingConfirmation(String trainingName, String userName) {
    return '「$trainingName」を$userNameと共有しますか？';
  }

  @override
  String get trainingDeletedSuccessfully => 'トレーニングが正常に削除されました';

  @override
  String get failedToDeleteTraining => 'トレーニングの削除に失敗しました';

  @override
  String get leftTrainingSuccessfully => 'トレーニングから正常に退出しました';

  @override
  String get partnerAddedSuccessfully => 'パートナーが正常に追加されました';

  @override
  String get failedToAddPartner => 'パートナーの追加に失敗しました';

  @override
  String get trainingSharedSuccessfully => 'トレーニングが正常に共有されました';

  @override
  String get failedToShareTraining => 'トレーニングの共有に失敗しました';

  @override
  String get trainingCloned => 'トレーニングを複製しました';

  @override
  String get failedToCloneTraining => 'トレーニングの複製に失敗しました';

  @override
  String get trainingMarkedAsComplete => 'トレーニングが完了しました';

  @override
  String get failedToCompleteTraining => 'トレーニングの完了に失敗しました';

  @override
  String get feedback => 'フィードバック';

  @override
  String get feedbackUpdated => 'フィードバックを更新しました';

  @override
  String get failedToUpdateFeedback => 'フィードバックの更新に失敗しました';

  @override
  String get reportSubmitted => 'レポートが送信されました';

  @override
  String get failedToSubmitReport => 'レポートの送信に失敗しました';

  @override
  String get shuffleExercise => 'エクササイズを変更';

  @override
  String get exerciseShuffled => 'エクササイズを変更しました';

  @override
  String get failedToShuffleExercise => 'エクササイズの変更に失敗しました';

  @override
  String get reasoning => '推論';

  @override
  String get strategy => '戦略';

  @override
  String get typeSelection => 'タイプ選択';

  @override
  String get progression => '進行';

  @override
  String get constraints => '制約';

  @override
  String get researchApplied => '適用された研究';

  @override
  String get targetMuscles => 'ターゲット筋肉';

  @override
  String get naming => '命名';

  @override
  String get trainingRoutines => 'トレーニング';

  @override
  String get noEquipment => '器具なし';

  @override
  String blockNumber(int number) {
    return 'ブロック $number';
  }

  @override
  String repeatsCount(int count) {
    return '${count}x';
  }

  @override
  String durationSeconds(int seconds) {
    return '$seconds秒';
  }

  @override
  String restSeconds(int seconds) {
    return '$seconds秒休憩';
  }

  @override
  String repsCount(int count) {
    return '$count回';
  }

  @override
  String weightKgValue(double value) {
    return '$value kg';
  }

  @override
  String get markAsComplete => '完了としてマーク';

  @override
  String get updateFeedback => 'フィードバックを更新';

  @override
  String get references => '参考資料';

  @override
  String get literature => '文献';

  @override
  String get request => 'Request';

  @override
  String get describeIssue => '問題を説明してください...';

  @override
  String get submit => '送信';

  @override
  String get close => '閉じる';

  @override
  String get add => '追加';

  @override
  String get update => '更新';

  @override
  String get clone => '複製';

  @override
  String get share => '共有';

  @override
  String get leave => '退出';

  @override
  String get tapToStart => 'タップして開始';

  @override
  String get tapWhenDone => '完了したらタップ';

  @override
  String get trainingCompleted => 'トレーニング完了！';

  @override
  String greatJobCompleting(String name) {
    return '$nameをよく頑張りました';
  }

  @override
  String get done => '完了';

  @override
  String get complete => '完了';

  @override
  String routineCounter(int current, int total) {
    return 'ルーティン $current/$total';
  }

  @override
  String blockCounter(int current, int total) {
    return 'ブロック $current/$total';
  }

  @override
  String get exitTraining => 'トレーニングを終了しますか？';

  @override
  String get whatWouldYouLikeToDo => '次は？';

  @override
  String get exit => '終了';

  @override
  String get continueTraining => '続行';

  @override
  String get stop => '停止';

  @override
  String get stopTraining => 'トレーニングを停止しますか？';

  @override
  String get stopTrainingConfirm => 'タイマーの進行状況は失われます。';

  @override
  String get failedToMarkComplete => 'トレーニングの完了に失敗しました';

  @override
  String get durationMinutes => '時間（分）';

  @override
  String get bodyweight => '自重';

  @override
  String get gym => 'ジム';

  @override
  String get custom => 'その他';

  @override
  String get noEquipmentBodyweightOnly => '自重のみ';

  @override
  String get noGymsDefinedCreateOne => 'ジムがありません。プロフィールで作成してください。';

  @override
  String get selectAGym => 'ジムを選択';

  @override
  String get addEquipment => '器具を追加';

  @override
  String get addEquipmentAvailable => '利用可能な器具を追加';

  @override
  String get includeWarmupCooldown => 'ウォームアップとクールダウンを含める';

  @override
  String get equipmentPlaceholder => '例：バーベル、ダンベル';

  @override
  String get customPromptOptional => 'カスタムプロンプト（任意）';

  @override
  String get focusOnUpperBody => '例：上半身に集中';

  @override
  String get trainingPartnersOptional => 'パートナー';

  @override
  String get generatingTraining => 'トレーニングを生成中...';

  @override
  String get thisMayTakeAMoment => '少々お待ちください';

  @override
  String generationFailedRetrying(int attempt) {
    return '生成に失敗しました、リトライ #$attempt...';
  }

  @override
  String get trainingGeneratedSuccessfully => 'トレーニングが正常に生成されました！';

  @override
  String get failedToGenerateTraining => 'トレーニングの生成に失敗しました';

  @override
  String get generate => '生成';

  @override
  String get editGym => 'ジムを編集';

  @override
  String get gymName => 'ジム名';

  @override
  String get gymNamePlaceholder => '例：ホームジム、エニタイム';

  @override
  String get availableWeights => '利用可能なウェイト';

  @override
  String get availableWeightsHint => 'このジムの加重モディファイアで使用するウェイトオプションを設定します。';

  @override
  String get noEquipmentAddedYet => 'まだ器具が追加されていません';

  @override
  String get pleaseEnterGymName => 'ジム名を入力してください';

  @override
  String get addAllEquipment => 'すべて追加';

  @override
  String get failedToLoadEquipment => '器具の読み込みに失敗しました';

  @override
  String get selectUser => 'ユーザーを選択';

  @override
  String get searchByName => '名前で検索';

  @override
  String get noUsersAvailable => '利用可能なユーザーがいません';

  @override
  String get noMatchingUsers => '一致するユーザーがいません';

  @override
  String get noMatchingEquipment => '一致する器具がありません';

  @override
  String get instructions => '手順';

  @override
  String get cues => 'ポイント';

  @override
  String get howWasYourTraining => 'トレーニングはいかがでしたか？';

  @override
  String get anyAdditionalComments => '追加のコメントはありますか？';

  @override
  String get actualDuration => '実際の所要時間（分）';

  @override
  String get impossible => 'できない';

  @override
  String get tooHard => '難しすぎる';

  @override
  String get ok => '適切';

  @override
  String get easy => '簡単';

  @override
  String get tooEasy => '簡単すぎる';

  @override
  String get flag => 'フラグ';

  @override
  String get profile => 'プロフィール';

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
    return '次: $name';
  }

  @override
  String get rest => '休憩';

  @override
  String get upcoming => '今後';

  @override
  String get yourProgress => '進捗状況';

  @override
  String trainingsCompleted(int count) {
    return '$count回のトレーニング完了';
  }

  @override
  String get completedTrainings => '完了したトレーニング';

  @override
  String get partneredTrainings => 'パートナートレーニング';

  @override
  String get movementFamilies => '動作カテゴリ';

  @override
  String get muscleActivity => '筋肉活動';

  @override
  String get failedToLoadProgress => '進捗の読み込みに失敗しました';

  @override
  String get noProgressYet => 'トレーニングを完了して進捗を確認';

  @override
  String get calibration => 'キャリブレーション';

  @override
  String get calibrationGlobal => '全体';

  @override
  String get calibrationNeeded =>
      '最初のトレーニングを完了して、Vigorがあなたのレベルに合わせた推奨を調整できるようにしましょう';

  @override
  String get calibrationDescription =>
      'キャリブレーション中、プラットフォームはフィードバックからデータを収集し、あなたのフィットネス状態とレベルの初期評価を行います';

  @override
  String get calibrationInProgress => 'トレーニングを重ねるごとにプランが賢くなります';

  @override
  String get calibrationTrainingNote =>
      'このトレーニングは目標と完全に一致しない場合があります — システムはまだあなたのフィットネスレベルを学習中で、完全なプロフィールを構築するために動作の多様性を優先しています';

  @override
  String calibrationFamiliesLearned(int calibrated, int total) {
    return '$calibrated/$total 動作パターンを学習済み';
  }

  @override
  String get capabilities => '能力';

  @override
  String get muscleHeatMap => '筋肉ヒートマップ';

  @override
  String get heatResting => '休息中';

  @override
  String get heatRecovered => '回復済み';

  @override
  String get heatActive => '活動中';

  @override
  String get heatWarm => 'ウォーム';

  @override
  String get heatHot => '高強度';

  @override
  String get noTrainingsCompletedYet => 'トレーニングを始めて何か表示しよう';

  @override
  String get theme => 'テーマ';

  @override
  String get themeAuto => '自動';

  @override
  String get themeLight => 'ライト';

  @override
  String get themeDark => 'ダーク';

  @override
  String get themeAutoDescription => 'システム設定に従う';

  @override
  String get appearance => '外観';

  @override
  String get trainingDefaults => 'デフォルト';

  @override
  String get defaultDuration => 'トレーニング時間';

  @override
  String get warmupCooldown => 'ウォームアップとクールダウン';

  @override
  String get timer => 'タイマー';

  @override
  String get intervalJingle => 'インターバル切替時にホイッスル';

  @override
  String get duckOtherAudio => '他のオーディオを下げる';

  @override
  String get duckOtherAudioDescription => 'ホイッスル中に音楽や他のアプリの音量を下げます';

  @override
  String get liveTimerNotification => 'ライブタイマー通知';

  @override
  String get liveTimerNotificationDescription =>
      'トレーニング中にステータスバーに経過時間とコントロールを表示します';

  @override
  String get goalHypertrophy => '筋肥大';

  @override
  String get goalHypertrophyDescription => 'ターゲットを絞った筋トレで筋肉量を増やす';

  @override
  String get goalFatLoss => '脂肪燃焼';

  @override
  String get goalFatLossDescription => '高強度ワークアウトでカロリーを消費し体脂肪を減らす';

  @override
  String get goalToning => '引き締め';

  @override
  String get goalToningDescription => '引き締まった筋肉とアスリート体型を作る';

  @override
  String get goalPosture => '姿勢';

  @override
  String get goalPostureDescription => '背中とコアを強化して姿勢を改善';

  @override
  String get goalRehabilitation => 'リハビリ';

  @override
  String get goalRehabilitationDescription => '怪我からの回復のための安全で管理されたエクササイズ';

  @override
  String get goalWellness => 'ウェルネス';

  @override
  String get goalWellnessDescription => '健康とストレス解消のためのバランスの取れたワークアウト';

  @override
  String get goalFlexibility => '柔軟性';

  @override
  String get goalFlexibilityDescription => 'ストレッチとモビリティで可動域を改善';

  @override
  String get goalSports => 'スポーツパフォーマンス';

  @override
  String get goalSportsDescription => 'パワーと俊敏性のトレーニングで運動能力を向上';

  @override
  String get thisWeek => '今週';

  @override
  String get trainingPlan => 'トレーニングプラン';

  @override
  String get sessionsPerWeek => '週あたりのセッション数';

  @override
  String get sessionDuration => 'セッション時間';

  @override
  String get preferredTime => '推奨時間帯';

  @override
  String get recommendedTime => 'おすすめ時間帯';

  @override
  String get methodologyMix => 'メソッドミックス';

  @override
  String get pastWeeks => '過去の週';

  @override
  String get weeklyTarget => '週間目標';

  @override
  String daysLeft(int count) {
    return '残り$count日';
  }

  @override
  String get recommended => 'おすすめ';

  @override
  String get duration => '時間';

  @override
  String get familyHorizontalPush => 'プッシュ';

  @override
  String get familyHorizontalPull => 'プル';

  @override
  String get familyVerticalPush => 'オーバーヘッド';

  @override
  String get familyVerticalPull => '懸垂';

  @override
  String get familySquat => 'スクワット';

  @override
  String get familyHinge => 'ヒンジ';

  @override
  String get familyCore => 'コア';

  @override
  String get familyCarry => 'キャリー';

  @override
  String get familyCardio => 'カーディオ';

  @override
  String get familyMobility => 'モビリティ';

  @override
  String get familyBalance => 'バランス';

  @override
  String get trainingQuality => 'このトレーニングはいかがでしたか？';

  @override
  String get trainingQualityHint => 'AIトレーニングの品質評価に役立ちます';

  @override
  String get qualityReasonHint => '改善できる点は？';

  @override
  String get good => '良い';

  @override
  String get bad => '悪い';

  @override
  String get loadingMsg1 => 'プロフィールを分析中...';

  @override
  String get loadingMsg2 => 'エクササイズを選択中...';

  @override
  String get loadingMsg3 => 'ルーティンを構築中...';

  @override
  String get loadingMsg4 => 'トレーニング量を計算中...';

  @override
  String get loadingMsg5 => '休憩時間を最適化中...';

  @override
  String get loadingMsg6 => '運動研究を参照中...';

  @override
  String get loadingMsg7 => '筋肉群をバランス調整中...';

  @override
  String get loadingMsg8 => '進行パスを作成中...';

  @override
  String get loadingMsg9 => '強度を微調整中...';

  @override
  String get loadingMsg10 => '動作パターンを確認中...';

  @override
  String get loadingMsg11 => '回復ニーズを評価中...';

  @override
  String get loadingMsg12 => 'エクササイズの変化形を選択中...';

  @override
  String get loadingMsg13 => 'トレーニングブロックを構成中...';

  @override
  String get loadingMsg14 => 'ワーク間隔を設定中...';

  @override
  String get loadingMsg15 => '運動科学を適用中...';

  @override
  String get loadingMsg16 => 'ウォームアップを設計中...';

  @override
  String get loadingMsg17 => '動作ファミリーをマッピング中...';

  @override
  String get loadingMsg18 => '負荷分配を評価中...';

  @override
  String get loadingMsg19 => 'セッションをパーソナライズ中...';

  @override
  String get loadingMsg20 => 'もうすぐ完了...';

  @override
  String loadingMsgGoal(String goal) {
    return '$goalを最適化中...';
  }

  @override
  String get loadingMsgInjuries => '怪我に合わせて調整中...';

  @override
  String get loadingMsgFavorites => 'お気に入りの種目を優先中...';

  @override
  String get loadingMsgConditions => '体調に合わせて調整中...';

  @override
  String loadingMsgMethodology(String methodology) {
    return '$methodologyセッションを設計中...';
  }

  @override
  String get loadingMsgPartners => 'パートナートレーニングを調整中...';

  @override
  String loadingMsgGym(String gym) {
    return '$gymの器具を読み込み中...';
  }

  @override
  String get loadingMsgHistory => '最近のセッションを分析中...';

  @override
  String get loadingRetryMsg1 => 'うーん、うまくいかなかった — リトライ中';

  @override
  String get loadingRetryMsg2 => 'もう一度やってみます...';

  @override
  String get loadingRetryMsg3 => 'もう少し — もう一度';

  @override
  String get loadingRetryMsg4 => 'もう一回、お待ちください';

  @override
  String get loadingRetryMsg5 => 'おっと、再調整中...';

  @override
  String get loadingRetryMsg6 => 'あと少しだった — もう一度';

  @override
  String nSelected(int count) {
    return '$count件選択中';
  }

  @override
  String deleteSelectedTrainings(int count) {
    return '$count件のトレーニングを削除しますか？';
  }

  @override
  String get trainingsDeletedSuccessfully => 'トレーニングを削除しました';

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
  String get pendingFeedbacks => '未提出のフィードバック';

  @override
  String get pendingFeedbacksDescription =>
      '完了済みのトレーニングにフィードバックが未提出です。フィードバックは今後のトレーニング提案の改善に役立ちます。';

  @override
  String get equipmentPartner => 'パートナー';

  @override
  String get equipmentBalanceBoard => 'バランスボード';

  @override
  String get equipmentBand => 'バンド';

  @override
  String get equipmentBarbell => 'バーベル';

  @override
  String get equipmentBench => 'ベンチ';

  @override
  String get equipmentBox => 'ボックス';

  @override
  String get equipmentBosuBall => 'ボスボール';

  @override
  String get equipmentCable => 'ケーブルマシン';

  @override
  String get equipmentDipStation => 'ディップスタンド';

  @override
  String get equipmentDumbbell => 'ダンベル';

  @override
  String get equipmentEllipticalMachine => 'エリプティカルマシン';

  @override
  String get equipmentEzBarbell => 'EZバー';

  @override
  String get equipmentHammer => 'ハンマー';

  @override
  String get equipmentKettlebell => 'ケトルベル';

  @override
  String get equipmentLeverageMachine => 'レバレッジマシン';

  @override
  String get equipmentMedicineBall => 'メディシンボール';

  @override
  String get equipmentOlympicBarbell => 'オリンピックバーベル';

  @override
  String get equipmentPullUpBar => '懸垂バー';

  @override
  String get equipmentResistanceBand => 'レジスタンスバンド';

  @override
  String get equipmentRings => '吊り輪';

  @override
  String get equipmentRoller => 'フォームローラー';

  @override
  String get equipmentRope => 'ロープ';

  @override
  String get equipmentRowingMachine => 'ローイングマシン';

  @override
  String get equipmentSkiergMachine => 'スキーエルゴマシン';

  @override
  String get equipmentSledMachine => 'スレッドマシン';

  @override
  String get equipmentSmithMachine => 'スミスマシン';

  @override
  String get equipmentStabilityBall => 'バランスボール';

  @override
  String get equipmentStationaryBike => 'エアロバイク';

  @override
  String get equipmentStepmillMachine => 'ステップマシン';

  @override
  String get equipmentTire => 'タイヤ';

  @override
  String get equipmentTreadmill => 'トレッドミル';

  @override
  String get equipmentTrapBar => 'トラップバー';

  @override
  String get equipmentTrx => 'TRX';

  @override
  String get equipmentUpperBodyErgometer => '上半身エルゴメーター';

  @override
  String get equipmentWheelRoller => '腹筋ローラー';

  @override
  String get muscleChest => '胸';

  @override
  String get muscleBack => '背中';

  @override
  String get muscleShoulders => '肩';

  @override
  String get muscleArms => '腕';

  @override
  String get muscleCore => '体幹';

  @override
  String get muscleGlutes => '臀部';

  @override
  String get muscleLegs => '脚';

  @override
  String get modifierWeightedVest => 'ウエイトベスト';

  @override
  String get modifierParallettes => 'パラレット';

  @override
  String get modifierAnkleWeights => 'アンクルウエイト';

  @override
  String get modifierDipBelt => 'ディップベルト';

  @override
  String get modifierPushUpBars => 'プッシュアップバー';

  @override
  String get modifierResistanceBands => 'レジスタンスバンド';

  @override
  String get modifierWeight => 'ウエイト';

  @override
  String get modifierWristWeights => 'リストウェイト';

  @override
  String get healthData => 'ヘルスデータ';

  @override
  String get healthConnected => '接続済み';

  @override
  String get healthSynchronizing => '同期中...';

  @override
  String get healthSynchronized => '同期済み';

  @override
  String get healthSynchronize => '今すぐ同期';

  @override
  String get healthNotConnected => '未接続';

  @override
  String get healthNativeOnly => 'iOSとAndroidで利用可能';

  @override
  String healthBuildingBaselines(int days) {
    return 'パーソナル基準値を作成中（$days/14日）';
  }

  @override
  String get healthSyncNoData => 'デバイスにヘルスデータが見つかりません';

  @override
  String get healthSyncFailed => 'Sync failed';

  @override
  String healthSyncFailedDetail(String error) {
    return 'Data read from device but upload failed: $error';
  }

  @override
  String get healthSyncTypeFull => 'フル';

  @override
  String get healthSyncTypeIncremental => '増分';

  @override
  String healthSourceData(int metrics, int sessions) {
    return '$metrics指標 · $sessionsセッション';
  }

  @override
  String get healthBackend => 'バックエンド';

  @override
  String healthBackendData(int days, int sessions) {
    return '$days日間 · $sessionsセッション';
  }

  @override
  String healthDateRange(String from, String to) {
    return '$from – $to';
  }

  @override
  String get healthDisconnect => '切断してデータを削除';

  @override
  String get healthDisconnectConfirmation =>
      'ヘルスデータを切断しますか？同期されたすべてのデータが削除されます。この操作は取り消せません。';

  @override
  String get healthDisconnectedSuccessfully => 'ヘルスデータが切断されました';

  @override
  String get failedToDisconnectHealth => 'ヘルスデータの切断に失敗しました';

  @override
  String get healthConnect => '接続';

  @override
  String get healthMetrics => '健康データ';

  @override
  String get healthAdjustment => '健康状態に基づく調整';

  @override
  String get healthPermissionsTitle => 'ウェアラブルを接続';

  @override
  String get healthPermissionsDescription =>
      'Vigorはヘルスデータを読み取り、ワークアウトをパーソナライズします。睡眠・回復・活動データが充実するほど、より的確なトレーニング提案が可能になります。';

  @override
  String get healthPermissionsReadOnly => '身長・体重をヘルスと同期';

  @override
  String get healthPermissionsSleep => '睡眠時間とステージ';

  @override
  String get healthPermissionsHrv => '心拍変動（HRV）';

  @override
  String get healthPermissionsRhr => '安静時心拍数';

  @override
  String get healthPermissionsSteps => '1日の歩数';

  @override
  String get healthPermissionsWorkouts => 'ワークアウトと心拍数';

  @override
  String get healthPermissionsGrant => 'ヘルスデータを接続';

  @override
  String get healthPermissionsSkip => '今はしない';

  @override
  String get healthPermissionsGranted => 'ヘルスデータが接続されました';

  @override
  String get healthPermissionsDenied => '権限が許可されませんでした';

  @override
  String get healthOnboardingTitle => 'ウェアラブルを接続';

  @override
  String get healthOnboardingDescription =>
      'ヘルスデータを接続すると、Vigorが睡眠・回復・活動に合わせてトレーニングを調整します。';

  @override
  String get healthOnboardingConnect => '接続';

  @override
  String get healthOnboardingDismiss => 'あとで';

  @override
  String get healthInstallHcTitle => 'Health Connectが必要です';

  @override
  String get healthInstallHcDescription =>
      'ウェアラブルデータの同期にはHealth Connectが必要です。Play Storeからインストールしてください。';

  @override
  String get healthInstallHc => 'Health Connectをインストール';

  @override
  String get heartRate => '心拍数';

  @override
  String get avgHr => '平均心拍';

  @override
  String get maxHr => '最大心拍';

  @override
  String get bpm => 'bpm';

  @override
  String get hrZones => '心拍ゾーン';

  @override
  String get hrZone1 => 'ゾーン1';

  @override
  String get hrZone2 => 'ゾーン2';

  @override
  String get hrZone3 => 'ゾーン3';

  @override
  String get hrZone4 => 'ゾーン4';

  @override
  String get hrZone5 => 'ゾーン5';

  @override
  String get healthDailySleep => '睡眠';

  @override
  String get healthDailyRestingHr => '安静時心拍';

  @override
  String get healthDailyHrv => 'HRV';

  @override
  String get healthDailySteps => '歩数';

  @override
  String get healthDailyCalories => 'カロリー';

  @override
  String get healthDailyNoData => '健康データなし';

  @override
  String get externalWorkout => '外部ワークアウト';

  @override
  String get exerciseTypeRunning => 'ランニング';

  @override
  String get exerciseTypeWalking => 'ウォーキング';

  @override
  String get exerciseTypeBiking => 'サイクリング';

  @override
  String get exerciseTypeYoga => 'ヨガ';

  @override
  String get exerciseTypeSwimming => '水泳';

  @override
  String get exerciseTypeHiking => 'ハイキング';

  @override
  String get exerciseTypeStrengthTraining => '筋力トレーニング';

  @override
  String get exerciseTypeFunctionalStrengthTraining => 'ファンクショナルトレーニング';

  @override
  String get exerciseTypeTraditionalStrengthTraining => '筋力トレーニング';

  @override
  String get exerciseTypeRunningTreadmill => 'トレッドミル';

  @override
  String get exerciseTypeBikingStationary => 'エアロバイク';

  @override
  String get exerciseTypeWalkingTreadmill => 'トレッドミルウォーク';

  @override
  String get exerciseTypeRowing => 'ローイング';

  @override
  String get exerciseTypePilates => 'ピラティス';

  @override
  String get exerciseTypeDancing => 'ダンス';

  @override
  String get exerciseTypeElliptical => 'エリプティカル';

  @override
  String get exerciseTypeStairClimbing => '階段昇降';

  @override
  String get exerciseTypeOther => 'ワークアウト';

  @override
  String get appLogs => 'アプリログ';

  @override
  String get viewLogs => 'ログを表示';

  @override
  String get exportLogs => 'ログをエクスポート';

  @override
  String get clearLogs => 'ログを消去';

  @override
  String get noLogsYet => 'ログはまだありません';

  @override
  String get logsCleared => 'ログを消去しました';

  @override
  String logEntries(int count) {
    return '$count 件のログ';
  }

  @override
  String get generateSession => 'セッションを生成';

  @override
  String get flowSession => 'フローセッション';

  @override
  String get flowSessionDescription => 'リカバリーとウェルネスのためのヨガ、ストレッチ、モビリティ';

  @override
  String get loadingMsgFlow1 => '最近トレーニングした筋肉を特定中...';

  @override
  String get loadingMsgFlow2 => 'リカバリーフローを設計中...';

  @override
  String get loadingMsgFlow3 => 'モビリティに合わせたポーズを選択中...';

  @override
  String get loadingMsgFlow4 => '怪我に安全な動きを確認中...';

  @override
  String get loadingMsgFlow5 => 'マインドフルな動きのシーケンスを作成中...';

  @override
  String get noFlowSessionsYet => 'フローセッションはまだありません';

  @override
  String get generateFirstFlow => '最初のフローセッションを生成して、リカバリーとストレッチを始めましょう。';
}
