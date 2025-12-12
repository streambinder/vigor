import 'dart:convert';
import 'dart:io';
import 'package:http/http.dart' as http;
import '../config/api_config.dart';
import '../models/api_response.dart';
import 'app_logger.dart';

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
      AppLogger.debug('[ApiService] Headers: $headers');
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
