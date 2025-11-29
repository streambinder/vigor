import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/training_generation_modal.dart';
import '../screens/training_details_screen.dart';
import '../services/gym_service.dart';
import '../services/secure_storage_service.dart';
import '../models/gym.dart';
import '../models/training.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  GymService? _gymService;
  List<Gym>? _gyms;
  bool _isLoadingGyms = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final storage = context.read<SecureStorageService>();
      _gymService = GymService(storageService: storage);
      _loadGyms();
    });
  }

  Future<void> _loadGyms() async {
    if (_gymService == null) return;

    setState(() {
      _isLoadingGyms = true;
    });

    final response = await _gymService!.getGyms();
    if (response.isSuccess && mounted) {
      setState(() {
        _gyms = response.data;
        _isLoadingGyms = false;
      });
    } else if (mounted) {
      setState(() {
        _isLoadingGyms = false;
      });
    }
  }

  void _showTrainingGenerationModal() {
    if (_gyms == null || _gyms!.isEmpty) {
      AdaptiveNotification.showError(
        context: context,
        message: 'Please add a gym first from your Profile',
      );
      return;
    }

    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => TrainingGenerationModal(
        gyms: _gyms!,
        onSuccess: (training) {
          // Navigate to the generated training details
          Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => TrainingDetailsScreen(training: training),
            ),
          );
        },
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(
        title: const Text('Vigor'),
      ),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.fitness_center,
                size: 80,
                color: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.primaryColor
                    : Theme.of(context).colorScheme.primary,
              ),
              const SizedBox(height: 24),
              Text(
                'Ready to train?',
                style: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 28)
                    : Theme.of(context).textTheme.headlineMedium,
              ),
              const SizedBox(height: 12),
              Text(
                'Generate a personalized workout based on your profile and goals',
                textAlign: TextAlign.center,
                style: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.bodyStyle
                    : Theme.of(context).textTheme.bodyLarge?.copyWith(
                          color: Colors.grey[600],
                        ),
              ),
              const SizedBox(height: 32),
              AdaptiveButton(
                onPressed: _isLoadingGyms ? null : _showTrainingGenerationModal,
                useGradient: true,
                child: const Text('Generate Workout'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
