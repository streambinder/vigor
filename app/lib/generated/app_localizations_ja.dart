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
  String get musclesAuto => 'すべて';

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
  String get instructions => '手順';

  @override
  String get howWasYourTraining => 'トレーニングはいかがでしたか？';

  @override
  String get anyAdditionalComments => '追加のコメントはありますか？';

  @override
  String get actualDuration => '実際の所要時間（分）';

  @override
  String get tooEasy => '簡単すぎる';

  @override
  String get tooHard => '難しすぎる';

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
  String get intervalJingle => 'インターバル完了時にサウンド';

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
}
