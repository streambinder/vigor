import 'package:flutter/material.dart';
import '../widgets/adaptive/adaptive.dart';

class ActivityScreen extends StatelessWidget {
  const ActivityScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(
        title: const Text('Activity'),
      ),
      body: const Center(
        child: Text(
          'Activity Screen',
          style: TextStyle(
            fontSize: 18,
            color: Colors.grey,
          ),
        ),
      ),
    );
  }
}
