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
          // pure route factory — on web the cold-start url arrives as the route
          // name, so hand it down and let the wrapper own all link parsing
          onGenerateRoute: (settings) => MaterialPageRoute(
            builder: (_) => AuthenticationWrapper(initialRoute: settings.name),
          ),
          debugShowCheckedModeBanner: false,
        ),
      ),
    );
  }
}

/// `/t/<token>` → share token, any other path → null
String? _shareTokenOf(Uri uri) =>
    uri.pathSegments.length == 2 && uri.pathSegments[0] == 't' ? uri.pathSegments[1] : null;

/// Wrapper widget that handles authentication state and navigation
class AuthenticationWrapper extends StatefulWidget {
  const AuthenticationWrapper({super.key, this.initialRoute});

  /// path the app cold-started on — only carries a url on web
  final String? initialRoute;

  @override
  State<AuthenticationWrapper> createState() => _AuthenticationWrapperState();
}

class _AuthenticationWrapperState extends State<AuthenticationWrapper> {
  late final AppLinks _appLinks;
  StreamSubscription<Uri>? _linkSubscription;
  AuthProvider? _authProvider;
  bool _loadingInitialData = false;
  bool _drainingShare = false;
  String? _coldStartToken;

  @override
  void initState() {
    super.initState();
    _appLinks = AppLinks();
    _coldStartToken = _shareTokenOf(Uri.parse(widget.initialRoute ?? '/'));
    _initDeepLinks();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      // wire event bus into auth provider so profile updates emit events
      final serviceLocator = context.read<ServiceLocator>();
      context.read<AuthProvider>().emitEvent = serviceLocator.emit;
      context.read<AuthProvider>().healthDataService = serviceLocator.healthDataService;
      context.read<AuthProvider>().initialize();
      if (_coldStartToken != null) _queueShare(_coldStartToken!);
    });
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final authProvider = context.read<AuthProvider>();
    if (!identical(authProvider, _authProvider)) {
      _authProvider?.removeListener(_onAuthChanged);
      _authProvider = authProvider..addListener(_onAuthChanged);
    }
    // a fresh wrapper can mount with a share already queued (see
    // SharedTrainingScreen's "login to add" path)
    _onAuthChanged();
  }

  @override
  void dispose() {
    _authProvider?.removeListener(_onAuthChanged);
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
    // cold start — on web the same link already reached us as the route name,
    // so ignore the duplicate rather than racing the drain with it
    try {
      final initialLink = await _appLinks.getInitialLink();
      if (initialLink != null && _shareTokenOf(initialLink) != _coldStartToken) {
        _handleDeepLink(initialLink);
      }
    } catch (_) {}

    // warm start
    _linkSubscription = _appLinks.uriLinkStream.listen(_handleDeepLink);
  }

  void _handleDeepLink(Uri uri) {
    final token = _shareTokenOf(uri);
    if (token != null) _queueShare(token);
  }

  /// single-slot mailbox — the drain decides when the stack can take it, so a
  /// link that lands mid-auth is never pushed onto a navigator about to be torn down
  void _queueShare(String token) {
    if (!mounted) return;
    context.read<ServiceLocator>().pendingShareToken = token;
    _scheduleDrain();
  }

  void _scheduleDrain() =>
      WidgetsBinding.instance.addPostFrameCallback((_) => _drainPendingShare());

  Future<void> _drainPendingShare() async {
    if (!mounted || _drainingShare) return;
    final serviceLocator = context.read<ServiceLocator>();
    final token = serviceLocator.pendingShareToken;
    // leave it queued until auth settles and the home data is in
    if (token == null ||
        context.read<AuthProvider>().state != AuthState.authenticated ||
        !serviceLocator.initialDataLoaded ||
        _loadingInitialData) {
      return;
    }

    _drainingShare = true;
    try {
      Widget destination = SharedTrainingScreen(token: token);
      if (serviceLocator.pendingShareAutoClaim) {
        // user tapped "login to add" on the share screen — claim before showing it
        try {
          final response = await serviceLocator.trainingService.claimSharedTraining(token);
          if (response.isSuccess && response.data != null) {
            destination = TrainingDetailsScreen(training: response.data!);
          }
        } catch (e) {
          AppLogger.error('[Share] claim failed for token $token', e);
        }
      }
      if (!mounted) return; // token stays queued — the next trigger retries
      Navigator.of(context).push(MaterialPageRoute(builder: (_) => destination));
      // consumed only once the route is actually on the stack
      serviceLocator.pendingShareToken = null;
      serviceLocator.pendingShareAutoClaim = false;
    } finally {
      _drainingShare = false;
    }
  }

  /// post-frame so a provider notification mid-build can't trigger a nested
  /// setState or a push
  void _onAuthChanged() =>
      WidgetsBinding.instance.addPostFrameCallback((_) => _syncFromAuth());

  /// reactions to being authenticated: locale, initial data, queued share.
  /// ordered — the drain waits for the data load to finish.
  void _syncFromAuth() {
    if (!mounted || context.read<AuthProvider>().state != AuthState.authenticated) return;

    final user = context.read<AuthProvider>().currentUser;
    if (user != null) {
      context.read<LocaleProvider>().setFromProfileLanguage(user.profile.language);
    }

    final serviceLocator = context.read<ServiceLocator>();
    if (!serviceLocator.initialDataLoaded && !_loadingInitialData) {
      _loadInitialData(serviceLocator);
    } else {
      _drainPendingShare();
    }
  }

  /// pre-load homepage data before leaving splash
  Future<void> _loadInitialData(ServiceLocator serviceLocator) async {
    _loadingInitialData = true;
    await serviceLocator.loadInitialData();
    // trigger health sync in background after initial data loads
    _triggerHealthSync(serviceLocator);
    if (!mounted) return;
    setState(() => _loadingInitialData = false);
    _drainPendingShare();
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<AuthProvider>(
      builder: (context, authProvider, child) {
        if (authProvider.state == AuthState.authenticated) {
          // stay on splash until _loadInitialData has filled the caches
          return context.read<ServiceLocator>().initialDataLoaded && !_loadingInitialData
              ? const HomeScreen()
              : const SplashScreen();
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
