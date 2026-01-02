import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:provider/provider.dart';
import 'design/tokens.dart';
import 'design/vigor_theme.dart';
import 'generated/app_localizations.dart';
import 'providers/auth_provider.dart';
import 'providers/locale_provider.dart';
import 'screens/splash_screen.dart';
import 'screens/google_auth_screen.dart';
import 'screens/home_screen.dart';
import 'services/secure_storage_service.dart';
import 'services/preferences_service.dart';
import 'services/service_locator.dart';
import 'services/app_logger.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  AppLogger.info('vigor app starting');

  // Initialize secure storage and fail fast if not available
  final storage = SecureStorageService();
  try {
    await storage.initialize();
  } catch (e) {
    AppLogger.error('Secure storage initialization failed', e);
    runApp(StorageErrorApp(error: e.toString()));
    return;
  }

  final prefs = PreferencesService();
  await prefs.initialize();

  runApp(VigorApp(storage: storage, prefs: prefs));
}

/// Error screen shown when storage initialization fails
class StorageErrorApp extends StatelessWidget {
  final String error;

  const StorageErrorApp({super.key, required this.error});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Vigor - Storage Error',
      theme: VigorTheme.light,
      darkTheme: VigorTheme.dark,
      themeMode: ThemeMode.system,
      home: Scaffold(
        body: Center(
          child: Padding(
            padding: VigorSpacing.paddingLg,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  Icons.error_outline,
                  size: 64,
                  color: VigorColors.error,
                ),
                SizedBox(height: VigorSpacing.lg),
                Text(
                  'Storage Unavailable',
                  style: VigorTypography.title.copyWith(
                    color: VigorColors.lightTextPrimary,
                  ),
                  textAlign: TextAlign.center,
                ),
                SizedBox(height: VigorSpacing.md),
                Text(
                  error,
                  style: VigorTypography.body.copyWith(
                    color: VigorColors.lightTextPrimary,
                  ),
                  textAlign: TextAlign.center,
                ),
                SizedBox(height: VigorSpacing.lg),
                Text(
                  'This app requires secure storage to protect your data. '
                  'Please check your browser settings and try again.',
                  style: VigorTypography.body.copyWith(
                    color: VigorColors.lightTextSecondary,
                  ),
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
  final PreferencesService prefs;

  const VigorApp({super.key, required this.storage, required this.prefs});

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        Provider<SecureStorageService>.value(value: storage),
        Provider<PreferencesService>.value(value: prefs),
        Provider<ServiceLocator>(create: (_) => ServiceLocator(storage)),
        ChangeNotifierProvider(
          create: (_) => AuthProvider(storage: storage),
        ),
        ChangeNotifierProvider(
          create: (_) => LocaleProvider(prefs),
        ),
      ],
      child: Consumer<LocaleProvider>(
        builder: (context, localeProvider, _) => MaterialApp(
          title: 'Vigor',
          locale: localeProvider.locale,
          supportedLocales: supportedLocales,
          localizationsDelegates: const [
            AppLocalizations.delegate,
            GlobalMaterialLocalizations.delegate,
            GlobalWidgetsLocalizations.delegate,
            GlobalCupertinoLocalizations.delegate,
          ],
          theme: VigorTheme.light,
          darkTheme: VigorTheme.dark,
          themeMode: ThemeMode.system,
          home: const AuthenticationWrapper(),
          debugShowCheckedModeBanner: false,
        ),
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
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<AuthProvider>().initialize();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<AuthProvider>(
      builder: (context, authProvider, child) {
        // sync locale with user profile language when authenticated
        if (authProvider.state == AuthState.authenticated) {
          final user = authProvider.currentUser;
          if (user != null) {
            WidgetsBinding.instance.addPostFrameCallback((_) {
              context.read<LocaleProvider>().setFromProfileLanguage(user.profile.language);
            });
          }
        }

        if (authProvider.state == AuthState.initial ||
            authProvider.state == AuthState.loading) {
          return const SplashScreen();
        }

        if (authProvider.state == AuthState.authenticated) {
          return const HomeScreen();
        }

        return const GoogleAuthScreen();
      },
    );
  }
}
