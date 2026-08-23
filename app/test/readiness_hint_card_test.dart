import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:vigor/generated/app_localizations.dart';
import 'package:vigor/widgets/readiness_hint_card.dart';

Widget _wrap(Widget child, {Locale locale = const Locale('en')}) {
  return MaterialApp(
    locale: locale,
    localizationsDelegates: const [
      AppLocalizations.delegate,
      GlobalMaterialLocalizations.delegate,
      GlobalWidgetsLocalizations.delegate,
      GlobalCupertinoLocalizations.delegate,
    ],
    supportedLocales: AppLocalizations.supportedLocales,
    home: Scaffold(body: child),
  );
}

void main() {
  testWidgets('shows the verdict label and score for each level', (
    tester,
  ) async {
    final l10n = await AppLocalizations.delegate.load(const Locale('en'));
    for (final (level, label) in [
      ('green', l10n.readinessGreen),
      ('yellow', l10n.readinessYellow),
      ('red', l10n.readinessRed),
    ]) {
      await tester.pumpWidget(
        _wrap(const ReadinessHintCard(score: 42, level: 'green', summary: 's')),
      );
      // override level per iteration
      await tester.pumpWidget(
        _wrap(ReadinessHintCard(score: 42, level: level, summary: 's')),
      );
      await tester.pumpAndSettle();
      expect(find.text(label), findsOneWidget);
      expect(find.text('42'), findsOneWidget);
    }
  });

  testWidgets('tap opens dialog with title, score and summary', (tester) async {
    final l10n = await AppLocalizations.delegate.load(const Locale('en'));
    await tester.pumpWidget(
      _wrap(
        const ReadinessHintCard(
          score: 78,
          level: 'green',
          summary: 'Sleep was solid, legs are fresh',
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text(l10n.readinessGreen));
    await tester.pumpAndSettle();

    expect(find.text(l10n.readinessTitle), findsOneWidget);
    expect(find.textContaining('78/100'), findsOneWidget);
    expect(
      find.textContaining('Sleep was solid, legs are fresh'),
      findsOneWidget,
    );

    await tester.tap(find.text(l10n.close));
    await tester.pumpAndSettle();
    expect(find.text(l10n.readinessTitle), findsNothing);
  });

  testWidgets('renders localized labels', (tester) async {
    final l10n = await AppLocalizations.delegate.load(const Locale('it'));
    await tester.pumpWidget(
      _wrap(
        const ReadinessHintCard(score: 10, level: 'red', summary: 'x'),
        locale: const Locale('it'),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text(l10n.readinessRed), findsOneWidget);
  });
}
