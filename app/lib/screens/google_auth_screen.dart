import 'dart:async';

import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:logger/logger.dart';
import 'package:provider/provider.dart';

import '../providers/auth_provider.dart';
import '../services/app_logger.dart';

// Import web-only methods when on web
import 'package:google_sign_in_web/web_only.dart' as web_only
    if (dart.library.io) '../stubs/web_only_stub.dart';

class GoogleAuthScreen extends StatefulWidget {
  const GoogleAuthScreen({super.key});

  @override
  State<GoogleAuthScreen> createState() => _GoogleAuthScreenState();
}

class _GoogleAuthScreenState extends State<GoogleAuthScreen> {
  final Logger _log = AppLogger.getLogger('GoogleAuthScreen');
  bool _initialized = false;
  bool _isLoading = false;
  String? _errorMessage;
  StreamSubscription<GoogleSignInAuthenticationEvent>? _authEventsSubscription;

  @override
  void initState() {
    super.initState();
    _initializeGoogleSignIn();
  }

  Future<void> _initializeGoogleSignIn() async {
    try {
      if (kIsWeb) {
        // Web configuration - need to specify clientId to get ID tokens
        await GoogleSignIn.instance.initialize(
          clientId:
              '559332153701-0auu2d1c1q43u7kf5k12akdllo060flh.apps.googleusercontent.com',
        );

        // On web, listen to authentication events
        _authEventsSubscription =
            GoogleSignIn.instance.authenticationEvents.listen(
          _handleAuthenticationEvent,
          onError: _handleAuthenticationError,
        );
      } else {
        // Mobile configuration - use serverClientId for ID tokens
        await GoogleSignIn.instance.initialize(
          serverClientId:
              '559332153701-0auu2d1c1q43u7kf5k12akdllo060flh.apps.googleusercontent.com',
        );
      }
      setState(() {
        _initialized = true;
      });
    } catch (e) {
      _log.e('Failed to initialize Google Sign In: $e');
      setState(() {
        _errorMessage = 'Failed to initialize Google Sign In';
      });
    }
  }

  void _handleAuthenticationEvent(GoogleSignInAuthenticationEvent event) {
    _log.d('Authentication event received: ${event.runtimeType}');

    if (event is GoogleSignInAuthenticationEventSignIn) {
      _handleSignInSuccess(event.user);
    } else if (event is GoogleSignInAuthenticationEventSignOut) {
      _log.i('User signed out');
    }
  }

  void _handleAuthenticationError(Object error, StackTrace stackTrace) {
    _log.e('Authentication error: $error', error: error, stackTrace: stackTrace);

    if (error is GoogleSignInException) {
      // Don't show error for user cancellation
      if (error.code == GoogleSignInExceptionCode.canceled) {
        setState(() {
          _isLoading = false;
        });
        return;
      }

      setState(() {
        _isLoading = false;
        _errorMessage = 'Sign-in error: ${error.description ?? error.code.toString()}';
      });
    } else {
      setState(() {
        _isLoading = false;
        _errorMessage = 'Sign-in error: ${error.toString()}';
      });
    }
  }

  Future<void> _handleSignInSuccess(GoogleSignInAccount user) async {
    final authProvider = context.read<AuthProvider>();

    try {
      // Get authentication details (ID token only in v7.x)
      final GoogleSignInAuthentication googleAuth = user.authentication;
      final String? idToken = googleAuth.idToken;

      _log.d('Google auth received: idToken=${idToken != null ? "${idToken.length}b" : "none"}');

      if (idToken == null) {
        _log.w('No ID token received from Google sign-in');
        setState(() {
          _isLoading = false;
          _errorMessage = 'Failed to get authentication token';
        });
        return;
      }

      _log.i('Authenticating with ID token (${idToken.length}b)');

      // Send token to backend
      final success = await authProvider.loginWithGoogle(idToken: idToken);

      if (success) {
        _log.i('Google sign-in completed successfully');
      } else {
        _log.e('Google sign-in failed: ${authProvider.errorMessage}');
      }

      if (!success && mounted) {
        setState(() {
          _isLoading = false;
          _errorMessage = authProvider.errorMessage ?? 'Google sign-in failed';
        });

        // Sign out from Google on failure
        await GoogleSignIn.instance.signOut();
      }
    } catch (e) {
      _log.e('Error processing sign-in: $e');
      if (mounted) {
        setState(() {
          _isLoading = false;
          _errorMessage = 'Error processing sign-in: ${e.toString()}';
        });
      }
      await GoogleSignIn.instance.signOut();
    }
  }

  // Mobile-only: trigger authentication flow
  Future<void> _handleGoogleSignIn() async {
    if (!_initialized) {
      _log.w('Google Sign In not initialized yet');
      setState(() {
        _errorMessage = 'Google Sign In is still initializing...';
      });
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      // Trigger Google sign-in flow with scope hint
      final GoogleSignInAccount googleUser =
          await GoogleSignIn.instance.authenticate(
            scopeHint: ['email', 'profile'],
          );

      await _handleSignInSuccess(googleUser);
    } on GoogleSignInException catch (e) {
      _log.w('Google sign-in exception: ${e.code} - ${e.description}');

      // User canceled - don't show error
      if (e.code == GoogleSignInExceptionCode.canceled) {
        if (mounted) {
          setState(() {
            _isLoading = false;
          });
        }
        return;
      }

      // Show error for other exceptions
      if (mounted) {
        setState(() {
          _isLoading = false;
          _errorMessage = 'Sign-in error: ${e.description ?? e.code.toString()}';
        });
      }
      await GoogleSignIn.instance.signOut();
    } catch (e) {
      _log.e('Unexpected error during Google sign-in: $e');
      if (mounted) {
        setState(() {
          _isLoading = false;
          _errorMessage = 'Sign-in error: ${e.toString()}';
        });
      }
      await GoogleSignIn.instance.signOut();
    }
  }

  // Web-only: user clicked the Google sign-in button (loading state management)
  void _onWebSignInButtonPressed() {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Logo or App Name
                const Icon(
                  Icons.fitness_center,
                  size: 80,
                  color: Colors.blue,
                ),
                const SizedBox(height: 16),
                const Text(
                  'Vigor',
                  style: TextStyle(
                    fontSize: 32,
                    fontWeight: FontWeight.bold,
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 8),
                const Text(
                  'Your personal training assistant',
                  style: TextStyle(
                    fontSize: 16,
                    color: Colors.grey,
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 48),

                // Google Sign-In Button (platform-specific)
                if (!_initialized)
                  const Center(child: CircularProgressIndicator())
                else if (kIsWeb)
                  // Web: Use Google's renderButton widget
                  Center(
                    child: web_only.renderButton(
                      configuration: web_only.GSIButtonConfiguration(
                        type: web_only.GSIButtonType.standard,
                        theme: web_only.GSIButtonTheme.outline,
                        size: web_only.GSIButtonSize.large,
                        text: web_only.GSIButtonText.signinWith,
                        shape: web_only.GSIButtonShape.rectangular,
                      ),
                    ),
                  )
                else
                  // Mobile: Use custom button with authenticate()
                  ElevatedButton.icon(
                    onPressed: _isLoading ? null : _handleGoogleSignIn,
                    icon: _isLoading
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                            ),
                          )
                        : const Icon(Icons.login, size: 24),
                    label: Text(
                      _isLoading ? 'Signing in...' : 'Sign in with Google',
                      style: const TextStyle(fontSize: 16),
                    ),
                    style: ElevatedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(
                        vertical: 16,
                        horizontal: 24,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                  ),

                // Loading indicator for web (separate from button)
                if (kIsWeb && _isLoading) ...[
                  const SizedBox(height: 16),
                  const Center(
                    child: CircularProgressIndicator(),
                  ),
                ],

                // Error message
                if (_errorMessage != null) ...[
                  const SizedBox(height: 16),
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.red.shade50,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: Colors.red.shade200),
                    ),
                    child: Text(
                      _errorMessage!,
                      style: TextStyle(
                        color: Colors.red.shade900,
                        fontSize: 14,
                      ),
                      textAlign: TextAlign.center,
                    ),
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }

  @override
  void dispose() {
    _authEventsSubscription?.cancel();
    // Don't disconnect here - it revokes tokens that might be in use
    // The user can sign out explicitly if needed
    super.dispose();
  }
}
