import 'dart:async';

import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:provider/provider.dart';

import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../providers/auth_provider.dart';
import '../services/app_logger.dart';
import '../widgets/adaptive/adaptive.dart';

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
        await GoogleSignIn.instance.initialize(
          clientId: '559332153701-0auu2d1c1q43u7kf5k12akdllo060flh.apps.googleusercontent.com',
        );
        _authEventsSubscription = GoogleSignIn.instance.authenticationEvents.listen(
          _handleAuthenticationEvent,
          onError: _handleAuthenticationError,
        );
      } else {
        await GoogleSignIn.instance.initialize(
          serverClientId: '559332153701-0auu2d1c1q43u7kf5k12akdllo060flh.apps.googleusercontent.com',
        );
      }
      setState(() {
        _initialized = true;
      });
    } catch (e) {
      AppLogger.error('[GoogleAuthScreen] failed to initialize: $e');
      if (mounted) {
        setState(() {
          _errorMessage = AppLocalizations.of(context).failedToInitializeGoogleSignIn;
        });
      }
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
      if (error.code == GoogleSignInExceptionCode.canceled) {
        setState(() {
          _isLoading = false;
        });
        return;
      }
      setState(() {
        _isLoading = false;
        _errorMessage = AppLocalizations.of(context).signInError(error.description ?? error.code.toString());
      });
    } else {
      setState(() {
        _isLoading = false;
        _errorMessage = AppLocalizations.of(context).signInError(error.toString());
      });
    }
  }

  Future<void> _handleSignInSuccess(GoogleSignInAccount user) async {
    final authProvider = context.read<AuthProvider>();
    final l10n = AppLocalizations.of(context);

    try {
      final GoogleSignInAuthentication googleAuth = user.authentication;
      final String? idToken = googleAuth.idToken;

      AppLogger.debug('[GoogleAuthScreen] google auth received: idToken=${idToken != null ? "${idToken.length}b" : "none"}');

      if (idToken == null) {
        AppLogger.warning('[GoogleAuthScreen] no ID token received');
        setState(() {
          _isLoading = false;
          _errorMessage = l10n.failedToGetAuthToken;
        });
        return;
      }

      AppLogger.info('[GoogleAuthScreen] authenticating with ID token (${idToken.length}b)');
      final success = await authProvider.loginWithGoogle(idToken: idToken);

      if (success) {
        AppLogger.info('[GoogleAuthScreen] sign-in completed successfully');
      } else {
        AppLogger.error('[GoogleAuthScreen] sign-in failed: ${authProvider.errorMessage}');
      }

      if (!success && mounted) {
        setState(() {
          _isLoading = false;
          _errorMessage = authProvider.errorMessage ?? l10n.googleSignInFailed;
        });
        await GoogleSignIn.instance.signOut();
      }
    } catch (e) {
      AppLogger.error('[GoogleAuthScreen] error processing sign-in: $e');
      if (mounted) {
        setState(() {
          _isLoading = false;
          _errorMessage = l10n.errorProcessingSignIn(e.toString());
        });
      }
      await GoogleSignIn.instance.signOut();
    }
  }

  Future<void> _handleGoogleSignIn() async {
    if (!_initialized) {
      AppLogger.warning('[GoogleAuthScreen] not initialized yet');
      setState(() {
        _errorMessage = AppLocalizations.of(context).googleSignInInitializing;
      });
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final GoogleSignInAccount googleUser = await GoogleSignIn.instance.authenticate(
        scopeHint: ['email', 'profile'],
      );
      await _handleSignInSuccess(googleUser);
    } on GoogleSignInException catch (e) {
      AppLogger.warning('[GoogleAuthScreen] sign-in exception: ${e.code} - ${e.description}');
      if (e.code == GoogleSignInExceptionCode.canceled) {
        if (mounted) {
          setState(() {
            _isLoading = false;
          });
        }
        return;
      }
      if (mounted) {
        setState(() {
          _isLoading = false;
          _errorMessage = AppLocalizations.of(context).signInError(e.description ?? e.code.toString());
        });
      }
      await GoogleSignIn.instance.signOut();
    } catch (e) {
      AppLogger.error('[GoogleAuthScreen] unexpected error: $e');
      if (mounted) {
        setState(() {
          _isLoading = false;
          _errorMessage = AppLocalizations.of(context).signInError(e.toString());
        });
      }
      await GoogleSignIn.instance.signOut();
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final textColor = VigorColors.textPrimary(context);
    final secondaryColor = VigorColors.textSecondary(context);

    return AdaptiveScaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: VigorSpacing.paddingLg,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // lightning bolt with gradient
                ShaderMask(
                  shaderCallback: (bounds) => LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: [VigorColors.orange, VigorColors.electricBlue],
                  ).createShader(bounds),
                  child: const Icon(
                    Icons.bolt,
                    size: 80,
                    color: Colors.white,
                  ),
                ),
                SizedBox(height: VigorSpacing.md),
                Text(
                  l10n.appName.toUpperCase(),
                  style: VigorTypography.display.copyWith(
                    color: textColor,
                    letterSpacing: 4,
                  ),
                  textAlign: TextAlign.center,
                ),
                SizedBox(height: VigorSpacing.sm),
                Text(
                  l10n.appTagline,
                  style: VigorTypography.body.copyWith(color: secondaryColor),
                  textAlign: TextAlign.center,
                ),
                SizedBox(height: VigorSpacing.xxl),

                // Google Sign-In Button (platform-specific)
                if (!_initialized)
                  const Center(child: AdaptiveLoadingIndicator())
                else if (kIsWeb)
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
                  AdaptiveButton(
                    onPressed: _isLoading ? null : _handleGoogleSignIn,
                    useGradient: true,
                    child: _isLoading
                        ? Row(
                            mainAxisSize: MainAxisSize.min,
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              SizedBox(
                                height: 20,
                                width: 20,
                                child: AdaptiveLoadingIndicator(color: Colors.white),
                              ),
                              SizedBox(width: VigorSpacing.sm),
                              Text(l10n.signingIn),
                            ],
                          )
                        : Row(
                            mainAxisSize: MainAxisSize.min,
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              const Icon(Icons.login, size: 24, color: Colors.white),
                              SizedBox(width: VigorSpacing.sm),
                              Text(l10n.signInWithGoogle),
                            ],
                          ),
                  ),

                // Loading indicator for web
                if (kIsWeb && _isLoading) ...[
                  SizedBox(height: VigorSpacing.md),
                  const Center(child: AdaptiveLoadingIndicator()),
                ],

                // Error message
                if (_errorMessage != null) ...[
                  SizedBox(height: VigorSpacing.md),
                  Container(
                    padding: VigorSpacing.paddingMd,
                    decoration: BoxDecoration(
                      color: VigorColors.error.withValues(alpha: 0.1),
                      borderRadius: VigorRadius.radiusMd,
                      border: Border.all(
                        color: VigorColors.error.withValues(alpha: 0.3),
                      ),
                    ),
                    child: Text(
                      _errorMessage!,
                      style: VigorTypography.body.copyWith(color: VigorColors.error),
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
    super.dispose();
  }
}
