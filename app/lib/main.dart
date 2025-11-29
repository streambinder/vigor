import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'providers/auth_provider.dart';
import 'screens/splash_screen.dart';
import 'screens/google_auth_screen.dart';
import 'screens/home_screen.dart';
import 'screens/profile_completion_modal.dart';
import 'utils/profile_helper.dart';
import 'utils/platform_helper.dart';
import 'theme/material_you_theme.dart';
import 'services/secure_storage_service.dart';
import 'services/app_logger.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize secure storage and fail fast if not available
  final storage = SecureStorageService();
  try {
    await storage.initialize();
  } catch (e) {
    // Show error and exit app
    runApp(StorageErrorApp(error: e.toString()));
    return;
  }

  runApp(VigorApp(storage: storage));
}

/// Error screen shown when storage initialization fails
class StorageErrorApp extends StatelessWidget {
  final String error;

  const StorageErrorApp({super.key, required this.error});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Vigor - Storage Error',
      theme: MaterialYouTheme.lightTheme,
      darkTheme: MaterialYouTheme.darkTheme,
      themeMode: ThemeMode.system,
      home: Scaffold(
        body: Center(
          child: Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(
                  Icons.error_outline,
                  size: 64,
                  color: Colors.red,
                ),
                const SizedBox(height: 24),
                const Text(
                  'Storage Unavailable',
                  style: TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 16),
                Text(
                  error,
                  style: const TextStyle(fontSize: 16),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 24),
                const Text(
                  'This app requires secure storage to protect your data. '
                  'Please check your browser settings and try again.',
                  style: TextStyle(fontSize: 14, color: Colors.grey),
                  textAlign: TextAlign.center,
                ),
              ],
            ),
          ),
        ),
      ),
      debugShowCheckedModeBanner: false,
    );
  }
}

class VigorApp extends StatelessWidget {
  final SecureStorageService storage;

  const VigorApp({super.key, required this.storage});

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        Provider<SecureStorageService>.value(value: storage),
        ChangeNotifierProvider(
          create: (_) => AuthProvider(storage: storage),
        ),
      ],
      child: MaterialApp(
        title: 'Vigor',
        // Use Material You theme for all platforms
        // iOS will use Liquid Glass widgets on top of Material base
        theme: MaterialYouTheme.lightTheme,
        darkTheme: MaterialYouTheme.darkTheme,
        themeMode: ThemeMode.system,
        home: const AuthenticationWrapper(),
        debugShowCheckedModeBanner: false,
      ),
    );
  }
}

/// Wrapper widget that handles authentication state and navigation
class AuthenticationWrapper extends StatefulWidget {
  const AuthenticationWrapper({super.key});

  @override
  State<AuthenticationWrapper> createState() => _AuthenticationWrapperState();
}

class _AuthenticationWrapperState extends State<AuthenticationWrapper>
    with WidgetsBindingObserver {
  bool _hasCheckedProfile = false;
  bool _isModalCurrentlyShown = false;
  AuthState? _previousAuthState;

  @override
  void initState() {
    super.initState();
    // Register for app lifecycle events
    WidgetsBinding.instance.addObserver(this);

    // Initialize authentication state on app startup
    // This will check for stored tokens and refresh them if available
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<AuthProvider>().initialize();
    });
  }

  @override
  void dispose() {
    // Unregister app lifecycle observer
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    super.didChangeAppLifecycleState(state);

    // Reset profile check when app comes to foreground
    // This ensures the modal appears every time the app is opened
    // BUT only if a modal isn't currently being shown
    if (state == AppLifecycleState.resumed && !_isModalCurrentlyShown) {
      setState(() {
        _hasCheckedProfile = false;
      });
    }
  }

  void _resetProfileCheckOnAuthChange(AuthState currentState) {
    // Reset profile check flag when auth state changes
    if (_previousAuthState != currentState) {
      _previousAuthState = currentState;
      if (currentState != AuthState.authenticated) {
        _hasCheckedProfile = false;
        _isModalCurrentlyShown = false;
      }
    }
  }

  void _checkAndShowProfileModal(AuthProvider authProvider) {
    if (!_hasCheckedProfile &&
        !_isModalCurrentlyShown &&
        authProvider.state == AuthState.authenticated &&
        authProvider.currentUser != null) {
      _hasCheckedProfile = true;

      final profile = authProvider.currentUser!.profile;
      final missingFields = ProfileHelper.getMissingRequiredFields(profile);

      if (missingFields.isNotEmpty) {
        // Mark modal as shown before displaying
        _isModalCurrentlyShown = true;

        // Show profile completion modal after current frame
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (mounted) {
            showDialog(
              context: context,
              barrierDismissible: false,
              builder: (context) => ProfileCompletionModal(
                profile: profile,
                missingFields: missingFields,
              ),
            ).then((_) {
              // Modal closed, reset flags to allow checking again
              if (mounted) {
                setState(() {
                  _isModalCurrentlyShown = false;
                  _hasCheckedProfile = false;
                });
              }
            });
          }
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<AuthProvider>(
      builder: (context, authProvider, child) {
        // Reset profile check when auth state changes
        _resetProfileCheckOnAuthChange(authProvider.state);

        // Show splash screen while loading
        if (authProvider.state == AuthState.initial ||
            authProvider.state == AuthState.loading) {
          return const SplashScreen();
        }

        // Show home screen if authenticated
        if (authProvider.state == AuthState.authenticated) {
          // Check profile completeness and show modal if needed
          _checkAndShowProfileModal(authProvider);
          return const HomeScreen();
        }

        // Show Google auth screen if unauthenticated
        return const GoogleAuthScreen();
      },
    );
  }
}
