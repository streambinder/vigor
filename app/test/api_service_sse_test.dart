import 'dart:async';
import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:vigor/services/api_service.dart';

/// An http client whose response body is a stream the test controls:
/// nothing arrives unless the test adds it, so a subscription that is
/// still attached to the body stream is observable via hasListener.
class _ControlledClient extends http.BaseClient {
  final StreamController<List<int>> body = StreamController<List<int>>();

  @override
  Future<http.StreamedResponse> send(http.BaseRequest request) async {
    return http.StreamedResponse(
      body.stream,
      200,
      headers: const {'content-type': 'text/event-stream'},
    );
  }
}

void main() {
  test('postSSE parses step and done events', () async {
    final client = _ControlledClient();
    final api = ApiService(client: client);

    final eventsFuture = api.postSSE('/training').toList();
    await Future<void>.delayed(Duration.zero);
    expect(client.body.hasListener, isTrue);

    client.body.add(utf8.encode('event: step\ndata: {"step":"history"}\n\n'));
    client.body.add(utf8.encode('event: done\ndata: {"id":"t1"}\n\n'));
    await client.body.close();

    final events = await eventsFuture;
    expect(events, hasLength(2));
    expect(events[0].event, 'step');
    expect(events[0].data['step'], 'history');
    expect(events[1].event, 'done');
    expect(events[1].data['id'], 't1');
  });

  test('cancelling postSSE detaches from the response stream at once',
      () async {
    final client = _ControlledClient();
    final api = ApiService(client: client);

    final sub = api.postSSE('/training').listen((_) {});
    await Future<void>.delayed(Duration.zero);
    expect(client.body.hasListener, isTrue);

    // no chunk ever arrives: the cancellation must still propagate to
    // the response stream immediately, closing the connection, instead
    // of waiting for the next chunk the way an async* generator does
    await sub.cancel();
    await Future<void>.delayed(Duration.zero);
    expect(client.body.hasListener, isFalse);

    await client.body.close();
  });
}
