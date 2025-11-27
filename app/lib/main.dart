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

void main() {
  runApp(const VigorApp());
}

class VigorApp extends StatelessWidget {
  const VigorApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (_) => AuthProvider(),
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

class _AuthenticationWrapperState extends State<AuthenticationWrapper> {
  bool _hasCheckedProfile = false;
  AuthState? _previousAuthState;

  @override
  void initState() {
    super.initState();
    // Initialize authentication state on app startup
    // This will check for stored tokens and refresh them if available
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<AuthProvider>().initialize();
    });
  }

  void _resetProfileCheckOnAuthChange(AuthState currentState) {
    // Reset profile check flag when auth state changes
    if (_previousAuthState != currentState) {
      _previousAuthState = currentState;
      if (currentState != AuthState.authenticated) {
        _hasCheckedProfile = false;
      }
    }
  }

  void _checkAndShowProfileModal(AuthProvider authProvider) {
    if (!_hasCheckedProfile &&
        authProvider.state == AuthState.authenticated &&
        authProvider.currentUser != null) {
      _hasCheckedProfile = true;

      final profile = authProvider.currentUser!.profile;
      final missingFields = ProfileHelper.getMissingRequiredFields(profile);

      if (missingFields.isNotEmpty) {
        // Show profile completion modal after current frame
        WidgetsBinding.instance.addPostFrameCallback((_) {
          showDialog(
            context: context,
            barrierDismissible: false,
            builder: (context) => ProfileCompletionModal(
              profile: profile,
              missingFields: missingFields,
            ),
          );
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
