import '../models/activity.dart';
import '../models/exercise.dart';

enum IntervalType { work, rest }

class TrainingInterval {
  final IntervalType type;
  final int duration;
  final String routineName;
  final String? activityName;
  final Activity? activity;
  final Exercise? exercise;
  final int activityNumber;
  final int totalActivities;
  final int blockNumber;
  final int totalBlocks;
  final int routineNumber;
  final int totalRoutines;

  TrainingInterval({
    required this.type,
    required this.duration,
    required this.routineName,
    this.activityName,
    this.activity,
    this.exercise,
    required this.activityNumber,
    required this.totalActivities,
    required this.blockNumber,
    required this.totalBlocks,
    required this.routineNumber,
    required this.totalRoutines,
  });
}
