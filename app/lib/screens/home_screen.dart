import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/auth_provider.dart';
import 'main_navigation.dart';
import 'profile_completion_modal.dart';
import '../utils/profile_helper.dart';

/// This is the main entry screen after authentication
/// It wraps the MainNavigation with profile completion logic
class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> with WidgetsBindingObserver {
  bool _hasCheckedProfile = false;
  bool _isModalCurrentlyShown = false;
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
    if (state == AppLifecycleState.resumed && !_isModalCurrentlyShown) {
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
        _isModalCurrentlyShown = true;

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
        _resetProfileCheckOnAuthChange(authProvider.state);
        _checkAndShowProfileModal(authProvider);

        return MainNavigation(key: MainNavigation.navKey);
      },
    );
  }
}
