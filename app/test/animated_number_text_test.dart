import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:vigor/widgets/animated_number_text.dart';

Widget _wrap(Widget child) => MaterialApp(home: Scaffold(body: child));

AnimatedNumberText _counter({double? value, double? seedValue}) =>
    AnimatedNumberText(
      value: value,
      seedValue: seedValue,
      formatValue: (v) => '${v.round()}',
    );

void main() {
  testWidgets('shows the value immediately when there is no seed', (tester) async {
    await tester.pumpWidget(_wrap(_counter(value: 58)));
    expect(find.text('58'), findsOneWidget);
  });

  testWidgets('renders statically when the seed equals the value', (tester) async {
    await tester.pumpWidget(_wrap(_counter(value: 58, seedValue: 58)));
    expect(find.text('58'), findsOneWidget);
    await tester.pump(const Duration(milliseconds: 450));
    expect(find.text('58'), findsOneWidget);
  });

  testWidgets('tweens from the seed to a changed value', (tester) async {
    await tester.pumpWidget(_wrap(_counter(value: 58, seedValue: 40)));
    // first frame starts at the seed, not at zero
    expect(find.text('40'), findsOneWidget);
    await tester.pump(const Duration(milliseconds: 450));
    final mid = find.byType(Text).evaluate().single.widget as Text;
    final midValue = int.parse(mid.data!);
    expect(midValue, greaterThan(40));
    expect(midValue, lessThan(58));
    await tester.pump(const Duration(milliseconds: 600));
    expect(find.text('58'), findsOneWidget);
  });

  testWidgets('shows the placeholder when the value is null, seed or not', (tester) async {
    await tester.pumpWidget(_wrap(_counter(value: null, seedValue: 40)));
    expect(find.text('—'), findsOneWidget);
  });

  testWidgets('tweens from the previous value when the value changes', (tester) async {
    await tester.pumpWidget(_wrap(_counter(value: 40)));
    expect(find.text('40'), findsOneWidget);
    await tester.pumpWidget(_wrap(_counter(value: 55)));
    await tester.pump(const Duration(milliseconds: 450));
    final mid = find.byType(Text).evaluate().single.widget as Text;
    final midValue = int.parse(mid.data!);
    expect(midValue, greaterThan(40));
    expect(midValue, lessThan(55));
    await tester.pump(const Duration(milliseconds: 600));
    expect(find.text('55'), findsOneWidget);
  });
}
