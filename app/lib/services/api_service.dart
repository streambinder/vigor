import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';
import 'package:http/http.dart' as http;
import '../config/api_config.dart';
import '../models/api_response.dart';
import 'app_logger.dart';

/// A single SSE event with an event type and parsed JSON data.
class SSEEvent {
  final String event;
  final Map<String, dynamic> data;
  SSEEvent(this.event, this.data);
}

class ApiService {
  final http.Client _client;

  ApiService({http.Client? client}) : _client = client ?? http.Client();

  /// Make a GET request
  Future<ApiResponse<Map<String, dynamic>>> get(
    String endpoint, {
    Map<String, String>? headers,
  }) async {
    try {
      final url = Uri.parse('${ApiConfig.baseUrl}$endpoint');
      AppLogger.debug('[ApiService] GET ${url.toString()}');
      final response = await _client.get(
        url,
        headers: _buildHeaders(headers),
      );

      return _handleResponse(response);
    } on SocketException {
      return ApiResponse.networkError('No internet connection');
    } on HttpException {
      return ApiResponse.networkError('HTTP error occurred');
    } on FormatException {
      return ApiResponse.networkError('Bad response format');
    } catch (e) {
      AppLogger.error('[ApiService] GET request failed', e);
      return ApiResponse.networkError('Something went wrong');
    }
  }

  /// Make a GET request returning raw bytes, for payloads the Map-based
  /// get cannot hold: JSON arrays and binary files such as models.
  Future<ApiResponse<Uint8List>> getBytes(
    String endpoint, {
    Map<String, String>? headers,
  }) async {
    try {
      final url = Uri.parse('${ApiConfig.baseUrl}$endpoint');
      AppLogger.debug('[ApiService] GET bytes ${url.toString()}');
      final response = await _client.get(
        url,
        headers: _buildHeaders(headers),
      );
      if (response.statusCode >= 200 && response.statusCode < 300) {
        return ApiResponse.success(response.bodyBytes, response.statusCode);
      }
      return ApiResponse.error('Request failed', response.statusCode);
    } on SocketException {
      return ApiResponse.networkError('No internet connection');
    } on HttpException {
      return ApiResponse.networkError('HTTP error occurred');
    } on FormatException {
      return ApiResponse.networkError('Bad response format');
    } catch (e) {
      AppLogger.error('[ApiService] GET bytes request failed', e);
      return ApiResponse.networkError('Something went wrong');
    }
  }

  /// Make a POST request
  Future<ApiResponse<Map<String, dynamic>>> post(
    String endpoint, {
    Map<String, dynamic>? body,
    Map<String, String>? headers,
  }) async {
    try {
      final url = Uri.parse('${ApiConfig.baseUrl}$endpoint');
      final response = await _client.post(
        url,
        headers: _buildHeaders(headers),
        body: body != null ? jsonEncode(body) : null,
      );

      return _handleResponse(response);
    } on SocketException {
      return ApiResponse.networkError('No internet connection');
    } on HttpException {
      return ApiResponse.networkError('HTTP error occurred');
    } on FormatException {
      return ApiResponse.networkError('Bad response format');
    } catch (e) {
      AppLogger.error('[ApiService] POST request failed', e);
      return ApiResponse.networkError('Something went wrong');
    }
  }

  /// Make a PUT request
  Future<ApiResponse<Map<String, dynamic>>> put(
    String endpoint, {
    Map<String, dynamic>? body,
    Map<String, String>? headers,
  }) async {
    try {
      final url = Uri.parse('${ApiConfig.baseUrl}$endpoint');
      final response = await _client.put(
        url,
        headers: _buildHeaders(headers),
        body: body != null ? jsonEncode(body) : null,
      );

      return _handleResponse(response);
    } on SocketException {
      return ApiResponse.networkError('No internet connection');
    } on HttpException {
      return ApiResponse.networkError('HTTP error occurred');
    } on FormatException {
      return ApiResponse.networkError('Bad response format');
    } catch (e) {
      AppLogger.error('[ApiService] PUT request failed', e);
      return ApiResponse.networkError('Something went wrong');
    }
  }

  /// Make a DELETE request
  Future<ApiResponse<Map<String, dynamic>>> delete(
    String endpoint, {
    Map<String, String>? headers,
  }) async {
    try {
      final url = Uri.parse('${ApiConfig.baseUrl}$endpoint');
      final response = await _client.delete(
        url,
        headers: _buildHeaders(headers),
      );

      return _handleResponse(response);
    } on SocketException {
      return ApiResponse.networkError('No internet connection');
    } on HttpException {
      return ApiResponse.networkError('HTTP error occurred');
    } on FormatException {
      return ApiResponse.networkError('Bad response format');
    } catch (e) {
      AppLogger.error('[ApiService] DELETE request failed', e);
      return ApiResponse.networkError('Something went wrong');
    }
  }

  /// Make a multipart POST request (platform-agnostic, no dart:io file access)
  Future<ApiResponse<Map<String, dynamic>>> postMultipart(
    String endpoint, {
    required List<int> bytes,
    required String fieldName,
    required String filename,
    Map<String, String>? headers,
  }) async {
    try {
      final url = Uri.parse('${ApiConfig.baseUrl}$endpoint');
      final request = http.MultipartRequest('POST', url);
      if (headers != null) request.headers.addAll(headers);
      request.files.add(http.MultipartFile.fromBytes(fieldName, bytes, filename: filename));
      final streamed = await request.send();
      final response = await http.Response.fromStream(streamed);
      return _handleResponse(response);
    } on SocketException {
      return ApiResponse.networkError('No internet connection');
    } on HttpException {
      return ApiResponse.networkError('HTTP error occurred');
    } on FormatException {
      return ApiResponse.networkError('Bad response format');
    } catch (e) {
      AppLogger.error('[ApiService] POST multipart request failed', e);
      return ApiResponse.networkError('Something went wrong');
    }
  }

  /// POST request that returns an SSE event stream.
  ///
  /// The stream is a transformer chain over the response byte stream,
  /// never an async* generator: cancelling the subscription must reach
  /// the socket at once, so the server sees the disconnect and abandons
  /// the request. An async* generator parked on its next chunk applies
  /// a cancellation only when that chunk arrives, which kept cancelled
  /// generations alive server-side until they completed.
  Stream<SSEEvent> postSSE(
    String endpoint, {
    Map<String, dynamic>? body,
    Map<String, String>? headers,
  }) {
    final url = Uri.parse('${ApiConfig.baseUrl}$endpoint');
    final request = http.Request('POST', url);
    request.headers.addAll(_buildHeaders(headers));
    request.headers['Accept'] = 'text/event-stream';
    if (body != null) request.body = jsonEncode(body);

    var eventType = '';
    final parser = StreamTransformer<String, SSEEvent>.fromHandlers(
      handleData: (line, sink) {
        if (line.startsWith('event: ')) {
          eventType = line.substring(7).trim();
        } else if (line.startsWith('data: ')) {
          try {
            sink.add(
              SSEEvent(
                eventType,
                jsonDecode(line.substring(6).trim()) as Map<String, dynamic>,
              ),
            );
          } catch (_) {}
        }
      },
    );

    return _client
        .send(request)
        .asStream()
        .asyncExpand(
          (response) => response.stream
              .transform(utf8.decoder)
              .transform(const LineSplitter())
              .transform(parser),
        );
  }

  /// Build headers with default values
  Map<String, String> _buildHeaders(Map<String, String>? customHeaders) {
    final headers = {
      ApiConfig.contentTypeHeader: ApiConfig.contentTypeJson,
    };

    if (customHeaders != null) {
      headers.addAll(customHeaders);
    }

    return headers;
  }

  /// Handle HTTP response
  ApiResponse<Map<String, dynamic>> _handleResponse(http.Response response) {
    final statusCode = response.statusCode;

    try {
      final jsonData = jsonDecode(response.body) as Map<String, dynamic>;

      if (statusCode >= 200 && statusCode < 300) {
        return ApiResponse.success(jsonData, statusCode);
      } else {
        final error = jsonData['error'] as String? ?? 'Unknown error';
        return ApiResponse.error(error, statusCode);
      }
    } catch (e) {
      if (statusCode >= 200 && statusCode < 300) {
        // Empty response body on success
        return ApiResponse.success({}, statusCode);
      } else {
        return ApiResponse.error('Failed to parse response', statusCode);
      }
    }
  }

  /// Close the client
  void dispose() {
    _client.close();
  }
}
