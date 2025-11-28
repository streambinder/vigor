import 'package:flutter/material.dart';
import '../widgets/adaptive/adaptive.dart';

class HomePage extends StatelessWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context) {
    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(
        title: const Text('Vigor'),
      ),
      body: const Center(
        child: Text(
          'Home Page',
          style: TextStyle(
            fontSize: 18,
            color: Colors.grey,
          ),
        ),
      ),
    );
  }
}
