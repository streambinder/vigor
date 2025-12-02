import 'dart:async';

import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:provider/provider.dart';

import '../providers/auth_provider.dart';
import '../services/app_logger.dart';
import '../widgets/adaptive/adaptive.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';

// Import web-only methods when on web
import 'package:google_sign_in_web/web_only.dart' as web_only
    if (dart.library.io) '../stubs/web_only_stub.dart';

class GoogleAuthScreen extends StatefulWidget {
  const GoogleAuthScreen({super.key});

  @override
  State<GoogleAuthScreen> createState() => _GoogleAuthScreenState();
}

class _GoogleAuthScreenState extends State<GoogleAuthScreen> {
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
      AppLogger.error('[GoogleAuthScreen] failed to initialize: $e');
      setState(() {
        _errorMessage = 'Failed to initialize Google Sign In';
      });
    }
  }

  void _handleAuthenticationEvent(GoogleSignInAuthenticationEvent event) {
    AppLogger.debug('[GoogleAuthScreen] authentication event: ${event.runtimeType}');

    if (event is GoogleSignInAuthenticationEventSignIn) {
      _handleSignInSuccess(event.user);
    } else if (event is GoogleSignInAuthenticationEventSignOut) {
      AppLogger.info('[GoogleAuthScreen] user signed out');
    }
  }

  void _handleAuthenticationError(Object error, StackTrace stackTrace) {
    AppLogger.error('[GoogleAuthScreen] authentication error', error, stackTrace);

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

      AppLogger.debug('[GoogleAuthScreen] google auth received: idToken=${idToken != null ? "${idToken.length}b" : "none"}');

      if (idToken == null) {
        AppLogger.warning('[GoogleAuthScreen] no ID token received');
        setState(() {
          _isLoading = false;
          _errorMessage = 'Failed to get authentication token';
        });
        return;
      }

      AppLogger.info('[GoogleAuthScreen] authenticating with ID token (${idToken.length}b)');

      // Send token to backend
      final success = await authProvider.loginWithGoogle(idToken: idToken);

      if (success) {
        AppLogger.info('[GoogleAuthScreen] sign-in completed successfully');
      } else {
        AppLogger.error('[GoogleAuthScreen] sign-in failed: ${authProvider.errorMessage}');
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
      AppLogger.error('[GoogleAuthScreen] error processing sign-in: $e');
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
      AppLogger.warning('[GoogleAuthScreen] not initialized yet');
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
      AppLogger.warning('[GoogleAuthScreen] sign-in exception: ${e.code} - ${e.description}');

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
      AppLogger.error('[GoogleAuthScreen] unexpected error: $e');
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
    return AdaptiveScaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Logo or App Name
                Icon(
                  Icons.fitness_center,
                  size: 80,
                  color: PlatformHelper.useLiquidGlass
                      ? LiquidGlassTheme.primaryColor
                      : Theme.of(context).colorScheme.primary,
                ),
                const SizedBox(height: 16),
                Text(
                  'Vigor',
                  style: PlatformHelper.useLiquidGlass
                      ? LiquidGlassTheme.titleStyle
                      : const TextStyle(
                          fontSize: 32,
                          fontWeight: FontWeight.bold,
                        ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 8),
                Text(
                  'Your personal training assistant',
                  style: PlatformHelper.useLiquidGlass
                      ? LiquidGlassTheme.bodyStyle.copyWith(
                          color: LiquidGlassTheme.captionStyle.color,
                        )
                      : const TextStyle(
                          fontSize: 16,
                          color: Colors.grey,
                        ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 48),

                // Google Sign-In Button (platform-specific)
                if (!_initialized)
                  const Center(child: AdaptiveLoadingIndicator())
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
                  // Mobile: Use adaptive button with authenticate()
                  AdaptiveButton(
                    onPressed: _isLoading ? null : _handleGoogleSignIn,
                    child: _isLoading
                        ? const Row(
                            mainAxisSize: MainAxisSize.min,
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              SizedBox(
                                height: 20,
                                width: 20,
                                child: AdaptiveLoadingIndicator(),
                              ),
                              SizedBox(width: 12),
                              Text('Signing in...'),
                            ],
                          )
                        : const Row(
                            mainAxisSize: MainAxisSize.min,
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(Icons.login, size: 24, color: Colors.white),
                              SizedBox(width: 12),
                              Text('Sign in with Google'),
                            ],
                          ),
                  ),

                // Loading indicator for web (separate from button)
                if (kIsWeb && _isLoading) ...[
                  const SizedBox(height: 16),
                  const Center(
                    child: AdaptiveLoadingIndicator(),
                  ),
                ],

                // Error message
                if (_errorMessage != null) ...[
                  const SizedBox(height: 16),
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.errorColor.withOpacity(0.1)
                          : Colors.red.shade50,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(
                        color: PlatformHelper.useLiquidGlass
                            ? LiquidGlassTheme.errorColor.withOpacity(0.3)
                            : Colors.red.shade200,
                      ),
                    ),
                    child: Text(
                      _errorMessage!,
                      style: TextStyle(
                        color: PlatformHelper.useLiquidGlass
                            ? LiquidGlassTheme.errorColor
                            : Colors.red.shade900,
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
