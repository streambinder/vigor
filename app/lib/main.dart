import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_web_plugins/url_strategy.dart';
import 'package:provider/provider.dart';
import 'package:app_links/app_links.dart';
import 'design/tokens.dart';
import 'design/vigor_theme.dart';
import 'generated/app_localizations.dart';
import 'providers/auth_provider.dart';
import 'providers/locale_provider.dart';
import 'providers/theme_provider.dart';
import 'screens/splash_screen.dart';
import 'screens/google_auth_screen.dart';
import 'screens/home_screen.dart';
import 'screens/shared_training_screen.dart';
import 'screens/training_details_screen.dart';
import 'services/secure_storage_service.dart';
import 'services/preferences_service.dart';
import 'services/service_locator.dart';
import 'services/app_logger.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  usePathUrlStrategy();
  AppLogger.captureAllLogs();
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
                const Icon(
                  Icons.error_outline,
                  size: 64,
                  color: VigorColors.error,
                ),
                const SizedBox(height: VigorSpacing.lg),
                Text(
                  'Storage Unavailable',
                  style: VigorTypography.title.copyWith(
                    color: VigorColors.lightTextPrimary,
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: VigorSpacing.md),
                Text(
                  error,
                  style: VigorTypography.body.copyWith(
                    color: VigorColors.lightTextPrimary,
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: VigorSpacing.lg),
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
        ChangeNotifierProvider<ServiceLocator>(create: (_) => ServiceLocator(storage, prefs)),
        ChangeNotifierProvider(
          create: (_) => AuthProvider(storage: storage),
        ),
        ChangeNotifierProvider(
          create: (_) => LocaleProvider(prefs),
        ),
        ChangeNotifierProvider(
          create: (_) => ThemeProvider(prefs),
        ),
      ],
      child: Consumer2<LocaleProvider, ThemeProvider>(
        builder: (context, localeProvider, themeProvider, _) => MaterialApp(
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
          themeMode: themeProvider.themeMode,
          onGenerateRoute: (settings) {
            final uri = Uri.parse(settings.name ?? '/');
            // /t/<token> → store token for post-auth processing
            if (uri.pathSegments.length == 2 && uri.pathSegments[0] == 't') {
              final serviceLocator = context.read<ServiceLocator>();
              serviceLocator.pendingShareToken = uri.pathSegments[1];
            }
            return MaterialPageRoute(builder: (_) => const AuthenticationWrapper());
          },
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
  late final AppLinks _appLinks;
  StreamSubscription<Uri>? _linkSubscription;
  bool _loadingInitialData = false;
  bool _processingPendingToken = false;

  @override
  void initState() {
    super.initState();
    _appLinks = AppLinks();
    _initDeepLinks();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      // wire event bus into auth provider so profile updates emit events
      final serviceLocator = context.read<ServiceLocator>();
      context.read<AuthProvider>().emitEvent = serviceLocator.emit;
      context.read<AuthProvider>().healthDataService = serviceLocator.healthDataService;
      context.read<AuthProvider>().initialize();
    });
  }

  @override
  void dispose() {
    _linkSubscription?.cancel();
    super.dispose();
  }

  /// fire-and-forget health sync — runs async in background, does not block UI
  void _triggerHealthSync(ServiceLocator serviceLocator) {
    final healthService = serviceLocator.healthDataService;
    if (healthService == null) return;
    healthService.syncToBackend().then((success) {
      if (success) AppLogger.debug('[HealthSync] background sync completed');
    });
  }

  Future<void> _initDeepLinks() async {
    // cold start
    try {
      final initialLink = await _appLinks.getInitialLink();
      if (initialLink != null) _handleDeepLink(initialLink);
    } catch (_) {}

    // warm start
    _linkSubscription = _appLinks.uriLinkStream.listen(_handleDeepLink);
  }

  void _handleDeepLink(Uri uri) {
    if (uri.pathSegments.length == 2 && uri.pathSegments[0] == 't') {
      final token = uri.pathSegments[1];
      if (!mounted) return;

      final serviceLocator = context.read<ServiceLocator>();
      // skip if already queued (e.g. onGenerateRoute already stored it)
      if (serviceLocator.pendingShareToken == token) return;

      final authState = context.read<AuthProvider>().state;
      if (authState == AuthState.authenticated) {
        Navigator.of(context).push(
          MaterialPageRoute(builder: (_) => SharedTrainingScreen(token: token)),
        );
      } else {
        // store for post-login processing — avoids pushing onto a navigator
        // that gets rebuilt during auth state transitions
        serviceLocator.pendingShareToken = token;
      }
    }
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

          // pre-load homepage data before leaving splash
          final serviceLocator = context.read<ServiceLocator>();
          if (!serviceLocator.initialDataLoaded && !_loadingInitialData) {
            _loadingInitialData = true;
            WidgetsBinding.instance.addPostFrameCallback((_) async {
              await serviceLocator.loadInitialData();
              // trigger health sync in background after initial data loads
              _triggerHealthSync(serviceLocator);
              if (mounted) setState(() => _loadingInitialData = false);
            });
            return const SplashScreen();
          }
          if (_loadingInitialData) return const SplashScreen();

          // process pending share token after auth + initial data
          if (!_processingPendingToken && serviceLocator.pendingShareToken != null) {
            _processingPendingToken = true;
            final token = serviceLocator.pendingShareToken!;
            final autoClaim = serviceLocator.pendingShareAutoClaim;
            serviceLocator.pendingShareToken = null;
            serviceLocator.pendingShareAutoClaim = false;
            WidgetsBinding.instance.addPostFrameCallback((_) async {
              if (autoClaim) {
                // user explicitly tapped "login to add" — claim directly
                final navigator = Navigator.of(context);
                final response = await serviceLocator.trainingService.claimSharedTraining(token);
                if (!mounted) return;
                if (response.isSuccess && response.data != null) {
                  navigator.push(
                    MaterialPageRoute(builder: (_) => TrainingDetailsScreen(training: response.data!)),
                  );
                } else {
                  navigator.push(
                    MaterialPageRoute(builder: (_) => SharedTrainingScreen(token: token)),
                  );
                }
              } else {
                // came from a link — just show the shared training screen
                if (!mounted) return;
                Navigator.of(context).push(
                  MaterialPageRoute(builder: (_) => SharedTrainingScreen(token: token)),
                );
              }
              _processingPendingToken = false;
            });
          }

          return const HomeScreen();
        }

        if (authProvider.state == AuthState.initial ||
            authProvider.state == AuthState.loading) {
          return const SplashScreen();
        }

        return const GoogleAuthScreen();
      },
    );
  }
}
