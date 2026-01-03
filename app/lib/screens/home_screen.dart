import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/auth_provider.dart';
import 'main_navigation.dart';
import 'profile_edit_screen.dart';
import '../utils/profile_helper.dart';

/// Main entry screen after authentication.
/// Handles profile completion screen on app open/resume.
class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> with WidgetsBindingObserver {
  bool _hasCheckedProfile = false;
  bool _isCompletionScreenShown = false;
  AuthState? _previousAuthState;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    super.didChangeAppLifecycleState(state);
    if (state == AppLifecycleState.resumed && !_isCompletionScreenShown) {
      setState(() {
        _hasCheckedProfile = false;
      });
    }
  }

  void _resetProfileCheckOnAuthChange(AuthState currentState) {
    if (_previousAuthState != currentState) {
      _previousAuthState = currentState;
      if (currentState != AuthState.authenticated) {
        _hasCheckedProfile = false;
        _isCompletionScreenShown = false;
      }
    }
  }

  void _checkAndShowProfileCompletion(AuthProvider authProvider) {
    if (!_hasCheckedProfile &&
        !_isCompletionScreenShown &&
        authProvider.state == AuthState.authenticated &&
        authProvider.currentUser != null) {
      _hasCheckedProfile = true;

      final profile = authProvider.currentUser!.profile;
      final missingFields = ProfileHelper.getMissingRequiredFields(profile);

      if (missingFields.isNotEmpty) {
        _isCompletionScreenShown = true;

        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (mounted) {
            Navigator.of(context).push(
              MaterialPageRoute(
                builder: (_) => ProfileEditScreen(
                  profile: profile,
                  missingFields: missingFields,
                ),
              ),
            ).then((_) {
              if (mounted) {
                setState(() {
                  _isCompletionScreenShown = false;
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
        _resetProfileCheckOnAuthChange(authProvider.state);
        _checkAndShowProfileCompletion(authProvider);

        return MainNavigation(key: MainNavigation.navKey);
      },
    );
  }
}
