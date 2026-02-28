/// Timer modes for different training methodologies
enum TimerMode {
  /// Linear interval sequencing - countdown each activity, auto-advance
  /// Used for: strength, circuit, hiit, mobility, endurance
  interval,

  /// Minute-synchronized timer - user taps when work done, rest fills remainder
  /// Used for: emom
  emom,

  /// Global countdown with infinite activity loop - user taps to advance
  /// Used for: amrap
  amrap,

  /// Count-up timer with fixed rounds - user taps to advance
  /// Used for: for_time
  forTime,
}

extension TimerModeExt on TimerMode {
  /// Maps a methodology string to the appropriate timer mode
  static TimerMode fromMethodology(String methodology) {
    switch (methodology) {
      case 'emom':
        return TimerMode.emom;
      case 'amrap':
        return TimerMode.amrap;
      case 'for_time':
        return TimerMode.forTime;
      default:
        // strength, supersets, circuit, hiit, mobility, endurance all use interval mode
        return TimerMode.interval;
    }
  }

  /// Whether this mode uses user-paced progression (tap to advance)
  bool get isUserPaced => this != TimerMode.interval;

  /// Whether this mode shows a global/total timer
  bool get hasGlobalTimer => this == TimerMode.amrap || this == TimerMode.forTime;

  /// Whether the timer counts up (true) or down (false)
  bool get countsUp => this == TimerMode.forTime;
}
