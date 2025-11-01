import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:logger/logger.dart';
import 'package:provider/provider.dart';

import '../providers/auth_provider.dart';
import '../services/app_logger.dart';

class GoogleAuthScreen extends StatefulWidget {
  const GoogleAuthScreen({super.key});

  @override
  State<GoogleAuthScreen> createState() => _GoogleAuthScreenState();
}

class _GoogleAuthScreenState extends State<GoogleAuthScreen> {
  // Configure GoogleSignIn differently for web vs mobile
  late final GoogleSignIn _googleSignIn;
  final Logger _log = AppLogger.getLogger('GoogleAuthScreen');

  @override
  void initState() {
    super.initState();

    if (kIsWeb) {
      // Web configuration - need to specify clientId to get ID tokens
      _googleSignIn = GoogleSignIn(
        clientId:
            '559332153701-0auu2d1c1q43u7kf5k12akdllo060flh.apps.googleusercontent.com',
        scopes: ['email', 'profile'],
      );
    } else {
      // Mobile configuration - use serverClientId for ID tokens
      _googleSignIn = GoogleSignIn(
        scopes: ['email', 'profile', 'openid'],
        serverClientId:
            '559332153701-0auu2d1c1q43u7kf5k12akdllo060flh.apps.googleusercontent.com',
      );
    }
  }

  bool _isLoading = false;
  String? _errorMessage;

  Future<void> _handleGoogleSignIn() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    // Capture AuthProvider reference before any async operations
    final authProvider = context.read<AuthProvider>();

    try {
      // Trigger Google sign-in flow
      final GoogleSignInAccount? googleUser = await _googleSignIn.signIn();

      if (googleUser == null) {
        // User canceled the sign-in
        setState(() {
          _isLoading = false;
        });
        return;
      }

      // Get authentication details
      final GoogleSignInAuthentication googleAuth =
          await googleUser.authentication;

      final String? idToken = googleAuth.idToken;
      final String? accessToken = googleAuth.accessToken;

      _log.d(
          'Google auth received: idToken=${idToken != null ? "${idToken.length}b" : "none"} '
          'accessToken=${accessToken != null ? "${accessToken.length}b" : "none"}');

      // Prefer ID token over access token (more reliable across platforms)
      // Fall back to access token for web if ID token unavailable
      String? tokenToSend = idToken ?? accessToken;

      if (tokenToSend == null) {
        _log.w('No token received from Google sign-in');
        setState(() {
          _isLoading = false;
          _errorMessage = 'Failed to get authentication token';
        });
        return;
      }

      final tokenType = tokenToSend == idToken ? 'ID' : 'Access';
      _log.i('Authenticating with $tokenType token (${tokenToSend.length}b)');

      // Send token to backend
      final success = await authProvider.loginWithGoogle(idToken: tokenToSend);

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
        await _googleSignIn.signOut();
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _isLoading = false;
          _errorMessage = 'Sign-in error: ${e.toString()}';
        });
      }
      // Sign out from Google on error
      await _googleSignIn.signOut();
    }
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

                // Google Sign-In Button
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
    // Don't disconnect here - it revokes tokens that might be in use
    // The user can sign out explicitly if needed
    super.dispose();
  }
}
