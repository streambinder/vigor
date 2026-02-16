// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Chinese (`zh`).
class AppLocalizationsZh extends AppLocalizations {
  AppLocalizationsZh([String locale = 'zh']) : super(locale);

  @override
  String get appName => 'Vigor';

  @override
  String get appTagline => 'Ex Sapientia Vis';

  @override
  String get navHome => '主页';

  @override
  String get navActivity => '活动';

  @override
  String get navProfile => '个人资料';

  @override
  String get storageErrorTitle => 'Vigor - 存储错误';

  @override
  String get storageUnavailable => '存储不可用';

  @override
  String get storageErrorMessage => '需要安全存储。请检查设置。';

  @override
  String get signInWithGoogle => '使用 Google 登录';

  @override
  String get signingIn => '正在登录...';

  @override
  String get failedToInitializeGoogleSignIn => '无法初始化 Google 登录';

  @override
  String signInError(String message) {
    return '登录错误：$message';
  }

  @override
  String get googleSignInFailed => 'Google 登录失败';

  @override
  String get failedToGetAuthToken => '无法获取身份验证令牌';

  @override
  String errorProcessingSignIn(String message) {
    return '处理登录时出错：$message';
  }

  @override
  String get googleSignInInitializing => 'Google 登录仍在初始化中...';

  @override
  String get readyToTrain => '准备好训练了吗？';

  @override
  String get generateTrainingDescription => '创建符合您目标的训练';

  @override
  String get generateTraining => '生成训练';

  @override
  String get refresh => '刷新';

  @override
  String get logout => '退出登录';

  @override
  String get userDataRefreshed => '用户数据已刷新';

  @override
  String get editProfile => '编辑个人资料';

  @override
  String get settings => '设置';

  @override
  String get other => '其他';

  @override
  String get deleteGym => '删除健身房';

  @override
  String deleteGymConfirmation(String name) {
    return '删除「$name」？';
  }

  @override
  String get cancel => '取消';

  @override
  String get delete => '删除';

  @override
  String get logoutConfirmation => '退出登录？';

  @override
  String get deleteAccount => '删除账户';

  @override
  String get deleteAccountConfirmation => '删除账户？不可撤销。';

  @override
  String get accountDeletedSuccessfully => '账户已成功删除';

  @override
  String get failedToDeleteAccount => '删除账户失败';

  @override
  String get failedToLoadGyms => '加载健身房失败';

  @override
  String get gymAddedSuccessfully => '健身房添加成功';

  @override
  String get failedToAddGym => '添加健身房失败';

  @override
  String get gymUpdatedSuccessfully => '健身房更新成功';

  @override
  String get failedToUpdateGym => '更新健身房失败';

  @override
  String get gymDeletedSuccessfully => '健身房删除成功';

  @override
  String get failedToDeleteGym => '删除健身房失败';

  @override
  String get birthdate => '出生日期';

  @override
  String get gender => '性别';

  @override
  String get language => '语言';

  @override
  String get height => '身高';

  @override
  String get weight => '体重';

  @override
  String get heightUnit => '厘米';

  @override
  String get weightUnit => '公斤';

  @override
  String heightWithUnit(double value) {
    return '$value 厘米';
  }

  @override
  String weightWithUnit(double value) {
    return '$value 公斤';
  }

  @override
  String get goals => '目标';

  @override
  String get injuries => '伤病';

  @override
  String get limitations => '限制';

  @override
  String get conditions => '身体状况';

  @override
  String get favorites => '收藏';

  @override
  String get personalDetails => '个人信息';

  @override
  String get healthAndGoals => '健康与目标';

  @override
  String get exercises => '练习';

  @override
  String get equipment => '设备';

  @override
  String startedDate(String date) {
    return '开始：$date';
  }

  @override
  String yearLabel(int year) {
    return '年份：$year';
  }

  @override
  String get myGyms => '健身房';

  @override
  String get addGym => '添加健身房';

  @override
  String get noGymsAddedYet => '尚未添加健身房';

  @override
  String get addYourFirstGym => '添加您的第一个健身房';

  @override
  String get removeDefault => '移除默认';

  @override
  String get setAsDefault => '设为默认';

  @override
  String get edit => '编辑';

  @override
  String get quickActions => '快速操作';

  @override
  String get dangerZone => '危险区域';

  @override
  String get completeYourProfile => '完善您的个人资料';

  @override
  String get updateYourProfileInfo => '在下方更新您的资料。';

  @override
  String get pleaseCompleteProfile => '完善资料。* = 必填。';

  @override
  String get firstName => '名';

  @override
  String get lastName => '姓';

  @override
  String get birthDate => '出生日期';

  @override
  String get male => '男';

  @override
  String get female => '女';

  @override
  String get selectLanguage => '选择语言';

  @override
  String get heightCm => '身高（厘米）';

  @override
  String get weightKg => '体重（公斤）';

  @override
  String get required => '必填';

  @override
  String get invalid => '无效';

  @override
  String get pleaseSelectBirthDate => '请选择您的出生日期';

  @override
  String get pleaseAddAtLeastOneGoal => '请至少添加一个目标';

  @override
  String get pleaseSelectLanguage => '请选择您的语言';

  @override
  String get addAGoal => '添加目标';

  @override
  String get injuryDescription => '伤病描述';

  @override
  String get year => '年份';

  @override
  String get addALimitation => '添加限制';

  @override
  String get addACondition => '添加身体状况';

  @override
  String get favoriteExercisesHint => '例如：深蹲、引体向上、跑步';

  @override
  String get favoriteEquipmentHint => '例如：哑铃、杠铃、壶铃';

  @override
  String get saveChanges => '保存更改';

  @override
  String get saveProfile => '保存个人资料';

  @override
  String get optionalLeaveEmpty => '（可选）';

  @override
  String get optionalExercisesPrefer => '（可选）';

  @override
  String get optionalEquipmentPrefer => '（可选）';

  @override
  String get optionalWorkoutTypesPrefer => '（可选）';

  @override
  String get favoriteExercises => '喜爱的练习';

  @override
  String get favoriteEquipment => '喜爱的设备';

  @override
  String get favoriteWorkoutTypes => '偏好的训练类型';

  @override
  String get workoutTypeStrength => '力量';

  @override
  String get workoutTypeStrengthDescription => '使用重负荷和充分休息来增强最大力量';

  @override
  String get workoutTypeCircuit => '循环';

  @override
  String get workoutTypeCircuitDescription => '在各站之间以最少休息进行体能训练';

  @override
  String get workoutTypeEmom => 'EMOM';

  @override
  String get workoutTypeEmomDescription => '每分钟：完成规定次数后休息至下一分钟';

  @override
  String get workoutTypeAmrap => 'AMRAP';

  @override
  String get workoutTypeAmrapDescription => '在规定时间内完成尽可能多的轮次';

  @override
  String get workoutTypeHiit => 'HIIT';

  @override
  String get workoutTypeHiitDescription => '高强度爆发与短暂恢复交替进行';

  @override
  String get workoutTypeForTime => '计时';

  @override
  String get workoutTypeForTimeDescription => '尽快完成训练';

  @override
  String get workoutTypeEndurance => '耐力';

  @override
  String get workoutTypeEnduranceDescription => '中等强度的持续运动以提高有氧能力';

  @override
  String get workoutTypeMobility => '灵活性';

  @override
  String get workoutTypeMobilityDescription => '改善活动范围和关节健康';

  @override
  String get methodologyOptional => '方法（可选）';

  @override
  String get methodologyAuto => '自动';

  @override
  String get goalsOptional => '目标（可选）';

  @override
  String get musclesOptional => '肌肉（可选）';

  @override
  String get musclesAuto => '全部';

  @override
  String get advancedSettings => '高级';

  @override
  String get failedToUpdateProfile => '更新个人资料失败';

  @override
  String get activity => '活动';

  @override
  String get noTrainingsYet => '暂无训练';

  @override
  String get generateFirstTraining => '从主页创建首个训练';

  @override
  String get noTrainingAvailable => '暂无训练。生成一个。';

  @override
  String get availableTrainings => '可用训练';

  @override
  String get pastTrainings => '过去的训练';

  @override
  String get stale => '已过期';

  @override
  String get today => '今天';

  @override
  String get yesterday => '昨天';

  @override
  String get available => '可用';

  @override
  String get completed => '已完成';

  @override
  String get completedSingular => '完成';

  @override
  String get noPastTrainings => '暂无已完成的训练';

  @override
  String get copied => '已复制';

  @override
  String durationMin(int minutes) {
    return '$minutes 分钟';
  }

  @override
  String durationHr(int hours) {
    return '$hours 小时';
  }

  @override
  String durationHrMin(int hours, int minutes) {
    return '$hours 小时 $minutes 分钟';
  }

  @override
  String get failedToLoadTrainings => '加载训练失败';

  @override
  String get startTraining => '开始训练';

  @override
  String get cloneTraining => '克隆训练';

  @override
  String get addPartner => '添加伙伴';

  @override
  String get shareWithUser => '与用户分享';

  @override
  String get deleteTraining => '删除训练';

  @override
  String get leaveTraining => '离开训练';

  @override
  String get showAiReasoning => '显示 AI 推理';

  @override
  String get reportIssue => '报告问题';

  @override
  String deleteTrainingConfirmation(String name) {
    return '删除「$name」？不可撤销。';
  }

  @override
  String leaveTrainingConfirmation(String name) {
    return '离开「$name」？将不再显示。';
  }

  @override
  String addPartnerConfirmation(String userName, String trainingName) {
    return '将 $userName 添加为「$trainingName」的伙伴？';
  }

  @override
  String cloneTrainingConfirmation(String name) {
    return '将「$name」克隆到您的训练中？';
  }

  @override
  String shareTrainingConfirmation(String trainingName, String userName) {
    return '与 $userName 分享「$trainingName」？';
  }

  @override
  String get trainingDeletedSuccessfully => '训练删除成功';

  @override
  String get failedToDeleteTraining => '删除训练失败';

  @override
  String get leftTrainingSuccessfully => '成功离开训练';

  @override
  String get partnerAddedSuccessfully => '伙伴添加成功';

  @override
  String get failedToAddPartner => '添加伙伴失败';

  @override
  String get trainingSharedSuccessfully => '训练分享成功';

  @override
  String get failedToShareTraining => '分享训练失败';

  @override
  String get trainingCloned => '训练已克隆';

  @override
  String get failedToCloneTraining => '克隆训练失败';

  @override
  String get trainingMarkedAsComplete => '训练已完成';

  @override
  String get failedToCompleteTraining => '完成训练失败';

  @override
  String get feedbackUpdated => '反馈已更新';

  @override
  String get failedToUpdateFeedback => '更新反馈失败';

  @override
  String get reportSubmitted => '报告已提交';

  @override
  String get failedToSubmitReport => '提交报告失败';

  @override
  String get shuffleExercise => '更换动作';

  @override
  String get exerciseShuffled => '动作已更换';

  @override
  String get failedToShuffleExercise => '更换动作失败';

  @override
  String get reasoning => '推理';

  @override
  String get strategy => '策略';

  @override
  String get typeSelection => '类型选择';

  @override
  String get progression => '进阶';

  @override
  String get constraints => '约束';

  @override
  String get researchApplied => '应用研究';

  @override
  String get targetMuscles => '目标肌肉';

  @override
  String get naming => '命名';

  @override
  String get trainingRoutines => '训练';

  @override
  String get noEquipment => '无设备';

  @override
  String blockNumber(int number) {
    return '区块 $number';
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
    return '$seconds秒休息';
  }

  @override
  String repsCount(int count) {
    return '$count 次';
  }

  @override
  String weightKgValue(double value) {
    return '$value 公斤';
  }

  @override
  String get markAsComplete => '标记为完成';

  @override
  String get updateFeedback => '更新反馈';

  @override
  String get references => '参考';

  @override
  String get literature => '文献';

  @override
  String get request => 'Request';

  @override
  String get describeIssue => '描述问题...';

  @override
  String get submit => '提交';

  @override
  String get close => '关闭';

  @override
  String get add => '添加';

  @override
  String get update => '更新';

  @override
  String get clone => '克隆';

  @override
  String get share => '分享';

  @override
  String get leave => '离开';

  @override
  String get tapToStart => '点击开始';

  @override
  String get tapWhenDone => '完成后点击';

  @override
  String get trainingCompleted => '训练完成！';

  @override
  String greatJobCompleting(String name) {
    return '完成 $name 做得好';
  }

  @override
  String get done => '完成';

  @override
  String get complete => '完成';

  @override
  String routineCounter(int current, int total) {
    return '程序 $current/$total';
  }

  @override
  String blockCounter(int current, int total) {
    return '区块 $current/$total';
  }

  @override
  String get exitTraining => '退出训练？';

  @override
  String get whatWouldYouLikeToDo => '下一步？';

  @override
  String get exit => '退出';

  @override
  String get continueTraining => '继续';

  @override
  String get failedToMarkComplete => '完成训练失败';

  @override
  String get durationMinutes => '时长（分钟）';

  @override
  String get bodyweight => '自重';

  @override
  String get gym => '健身房';

  @override
  String get custom => '其他';

  @override
  String get noEquipmentBodyweightOnly => '仅自重';

  @override
  String get noGymsDefinedCreateOne => '无健身房。在资料中创建。';

  @override
  String get selectAGym => '选择健身房';

  @override
  String get addEquipment => '添加设备';

  @override
  String get addEquipmentAvailable => '添加可用设备';

  @override
  String get includeWarmupCooldown => '包含热身和放松';

  @override
  String get equipmentPlaceholder => '例如：杠铃、哑铃';

  @override
  String get customPromptOptional => '自定义提示（可选）';

  @override
  String get focusOnUpperBody => '例如：专注于上半身';

  @override
  String get trainingPartnersOptional => '伙伴';

  @override
  String get generatingTraining => '正在生成您的训练...';

  @override
  String get thisMayTakeAMoment => '这可能需要一点时间';

  @override
  String generationFailedRetrying(int attempt) {
    return '生成失败，重试 #$attempt...';
  }

  @override
  String get trainingGeneratedSuccessfully => '训练生成成功！';

  @override
  String get failedToGenerateTraining => '生成训练失败';

  @override
  String get generate => '生成';

  @override
  String get editGym => '编辑健身房';

  @override
  String get gymName => '健身房名称';

  @override
  String get gymNamePlaceholder => '例如：家庭健身房、威尔士';

  @override
  String get noEquipmentAddedYet => '尚未添加设备';

  @override
  String get pleaseEnterGymName => '请输入健身房名称';

  @override
  String get addAllEquipment => '全部添加';

  @override
  String get failedToLoadEquipment => '加载设备失败';

  @override
  String get selectUser => '选择用户';

  @override
  String get searchByName => '按名称搜索';

  @override
  String get noUsersAvailable => '没有可用的用户';

  @override
  String get noMatchingUsers => '没有匹配的用户';

  @override
  String get instructions => '说明';

  @override
  String get cues => '提示';

  @override
  String get howWasYourTraining => '您的训练感觉如何？';

  @override
  String get anyAdditionalComments => '有其他评论吗？';

  @override
  String get actualDuration => '实际时长（分钟）';

  @override
  String get impossible => '做不到';

  @override
  String get tooHard => '太难';

  @override
  String get ok => '合适';

  @override
  String get easy => '简单';

  @override
  String get tooEasy => '太简单';

  @override
  String get flag => '标记';

  @override
  String get profile => '个人资料';

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
    return '下一个: $name';
  }

  @override
  String get rest => '休息';

  @override
  String get upcoming => '即将进行';

  @override
  String get yourProgress => '您的进度';

  @override
  String trainingsCompleted(int count) {
    return '已完成 $count 次训练';
  }

  @override
  String get completedTrainings => '已完成训练';

  @override
  String get partneredTrainings => '合作训练';

  @override
  String get movementFamilies => '动作类别';

  @override
  String get muscleActivity => '肌肉活动';

  @override
  String get failedToLoadProgress => '加载进度失败';

  @override
  String get noProgressYet => '完成训练以查看您的进度';

  @override
  String get calibration => '校准';

  @override
  String get calibrationGlobal => '整体';

  @override
  String get calibrationNeeded => '完成您的第一次训练，让Vigor根据您的水平校准推荐';

  @override
  String get calibrationDescription => '在校准期间，平台会收集您的反馈数据，对您的健身状态和水平进行初步评估';

  @override
  String get calibrationInProgress => '每次训练都让你的计划更智能';

  @override
  String calibrationFamiliesLearned(int calibrated, int total) {
    return '已学习 $calibrated/$total 个运动模式';
  }

  @override
  String get capabilities => '能力';

  @override
  String get muscleHeatMap => '肌肉热力图';

  @override
  String get heatResting => '休息中';

  @override
  String get heatRecovered => '已恢复';

  @override
  String get heatActive => '活跃';

  @override
  String get heatWarm => '温热';

  @override
  String get heatHot => '高强度';

  @override
  String get noTrainingsCompletedYet => '开始训练来看点什么吧';

  @override
  String get theme => '主题';

  @override
  String get themeAuto => '自动';

  @override
  String get themeLight => '浅色';

  @override
  String get themeDark => '深色';

  @override
  String get themeAutoDescription => '跟随系统设置';

  @override
  String get appearance => '外观';

  @override
  String get trainingDefaults => '默认值';

  @override
  String get defaultDuration => '训练时长';

  @override
  String get warmupCooldown => '热身和放松';

  @override
  String get timer => '计时器';

  @override
  String get intervalJingle => '间隔完成时播放提示音';

  @override
  String get goalHypertrophy => '增肌';

  @override
  String get goalHypertrophyDescription => '通过针对性力量训练增加肌肉量';

  @override
  String get goalFatLoss => '减脂';

  @override
  String get goalFatLossDescription => '通过高强度训练燃烧卡路里、减少体脂';

  @override
  String get goalToning => '塑形';

  @override
  String get goalToningDescription => '塑造线条分明的肌肉和健美体态';

  @override
  String get goalPosture => '体态';

  @override
  String get goalPostureDescription => '强化背部和核心以改善体态';

  @override
  String get goalRehabilitation => '康复';

  @override
  String get goalRehabilitationDescription => '安全可控的伤后恢复训练';

  @override
  String get goalWellness => '健康';

  @override
  String get goalWellnessDescription => '均衡训练促进整体健康和缓解压力';

  @override
  String get goalFlexibility => '柔韧性';

  @override
  String get goalFlexibilityDescription => '通过拉伸和灵活性训练提高活动范围';

  @override
  String get goalSports => '运动表现';

  @override
  String get goalSportsDescription => '通过力量和敏捷训练提升运动能力';

  @override
  String get thisWeek => '本周';

  @override
  String get trainingPlan => '训练计划';

  @override
  String get sessionsPerWeek => '每周训练次数';

  @override
  String get sessionDuration => '训练时长';

  @override
  String get preferredTime => '建议时间';

  @override
  String get recommendedTime => '推荐时间';

  @override
  String get methodologyMix => '方法组合';

  @override
  String get pastWeeks => '过去几周';

  @override
  String get weeklyTarget => '每周目标';

  @override
  String daysLeft(int count) {
    return '还剩 $count 天';
  }

  @override
  String get recommended => '推荐';

  @override
  String get duration => '时长';

  @override
  String get familyHorizontalPush => '推';

  @override
  String get familyHorizontalPull => '拉';

  @override
  String get familyVerticalPush => '过顶推';

  @override
  String get familyVerticalPull => '引体向上';

  @override
  String get familySquat => '深蹲';

  @override
  String get familyHinge => '铰链';

  @override
  String get familyCore => '核心';

  @override
  String get familyCarry => '负重行走';

  @override
  String get familyCardio => '有氧';

  @override
  String get familyMobility => '灵活性';

  @override
  String get familyBalance => '平衡';

  @override
  String get trainingQuality => '你觉得这次训练怎么样？';

  @override
  String get trainingQualityHint => '帮助我们评估AI生成训练的质量';

  @override
  String get qualityReasonHint => '有什么可以改进的？';

  @override
  String get good => '好';

  @override
  String get bad => '差';

  @override
  String get loadingMsg1 => '分析你的个人资料...';

  @override
  String get loadingMsg2 => '选择练习动作...';

  @override
  String get loadingMsg3 => '构建你的训练计划...';

  @override
  String get loadingMsg4 => '计算训练量...';

  @override
  String get loadingMsg5 => '优化休息间隔...';

  @override
  String get loadingMsg6 => '对照运动研究...';

  @override
  String get loadingMsg7 => '平衡肌肉群...';

  @override
  String get loadingMsg8 => '制定进阶路径...';

  @override
  String get loadingMsg9 => '微调训练强度...';

  @override
  String get loadingMsg10 => '审查动作模式...';

  @override
  String get loadingMsg11 => '评估恢复需求...';

  @override
  String get loadingMsg12 => '挑选动作变式...';

  @override
  String get loadingMsg13 => '组织训练模块...';

  @override
  String get loadingMsg14 => '安排工作间隔...';

  @override
  String get loadingMsg15 => '应用运动科学...';

  @override
  String get loadingMsg16 => '设计热身方案...';

  @override
  String get loadingMsg17 => '映射动作家族...';

  @override
  String get loadingMsg18 => '评估负荷分布...';

  @override
  String get loadingMsg19 => '个性化你的训练...';

  @override
  String get loadingMsg20 => '即将完成...';

  @override
  String loadingMsgGoal(String goal) {
    return '为$goal优化...';
  }

  @override
  String get loadingMsgInjuries => '根据伤病情况调整...';

  @override
  String get loadingMsgFavorites => '优先安排你喜欢的动作...';

  @override
  String get loadingMsgConditions => '适应你的身体状况...';

  @override
  String loadingMsgMethodology(String methodology) {
    return '设计$methodology训练...';
  }

  @override
  String get loadingMsgPartners => '协调搭档训练...';

  @override
  String loadingMsgGym(String gym) {
    return '加载$gym的设备...';
  }

  @override
  String get loadingMsgHistory => '分析你最近的训练...';

  @override
  String get loadingRetryMsg1 => '嗯，没出来对——重试中';

  @override
  String get loadingRetryMsg2 => '让我再试一次...';

  @override
  String get loadingRetryMsg3 => '差一点——再来一次';

  @override
  String get loadingRetryMsg4 => '再试一次，稍等';

  @override
  String get loadingRetryMsg5 => '哎呀，重新校准...';

  @override
  String get loadingRetryMsg6 => '差点就好了——再试';
}
