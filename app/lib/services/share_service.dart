import '../models/api_response.dart';
import 'api_service.dart';

/// Unauthenticated service for fetching shared training data.
class ShareService {
  final ApiService _apiService;

  ShareService({ApiService? apiService}) : _apiService = apiService ?? ApiService();

  Future<ApiResponse<Map<String, dynamic>>> getSharedTraining(String token) async {
    return _apiService.get('/training/shared/$token');
  }
}
