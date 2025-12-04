import 'package:flutter/material.dart';
import '../config/api_config.dart';
import '../models/exercise.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';
import '../widgets/adaptive/adaptive.dart';

class ExerciseModal {
  static bool _isValidImageUrl(String? url) {
    if (url == null || url.isEmpty) return false;
    final uri = Uri.tryParse(url);
    if (uri == null) return false;
    return uri.hasScheme && (uri.scheme == 'http' || uri.scheme == 'https');
  }

  // proxy external image URLs through backend to avoid CORS issues on web
  static String _proxyImageUrl(String url) {
    return '${ApiConfig.baseUrl}/proxy/image?url=${Uri.encodeComponent(url)}';
  }

  static void show(BuildContext context, Exercise exercise) {
    showDialog(
      context: context,
      barrierDismissible: true,
      builder: (context) => Dialog(
        backgroundColor: Colors.transparent,
        child: Container(
          constraints: const BoxConstraints(maxWidth: 600),
          decoration: PlatformHelper.useLiquidGlass
              ? LiquidGlassTheme.glassDecoration(
                  borderRadius: 20,
                  border: Border.all(
                    color: Colors.white.withOpacity(0.2),
                    width: 1.5,
                  ),
                )
              : BoxDecoration(
                  color: Theme.of(context).colorScheme.surface,
                  borderRadius: BorderRadius.circular(20),
                ),
          child: ClipRRect(
            borderRadius: BorderRadius.circular(20),
            child: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // Image
                  if (_isValidImageUrl(exercise.reference))
                    Image.network(
                      _proxyImageUrl(exercise.reference),
                      fit: BoxFit.contain,
                      errorBuilder: (context, error, stackTrace) {
                        final isDark = Theme.of(context).brightness == Brightness.dark;
                        return Container(
                          height: 200,
                          color: isDark
                              ? Colors.white.withOpacity(0.1)
                              : Colors.grey[300],
                          child: Center(
                            child: Icon(
                              Icons.broken_image,
                              size: 64,
                              color: isDark
                                  ? Colors.white.withOpacity(0.5)
                                  : Colors.grey.shade600,
                            ),
                          ),
                        );
                      },
                    ),
                  // Exercise details
                  Padding(
                    padding: const EdgeInsets.all(24.0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          exercise.name,
                          style: PlatformHelper.useLiquidGlass
                              ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 22)
                              : Theme.of(context).textTheme.headlineSmall,
                        ),
                        if (exercise.instructions.isNotEmpty) ...[
                          const SizedBox(height: 16),
                          Text(
                            'Instructions',
                            style: PlatformHelper.useLiquidGlass
                                ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 16)
                                : Theme.of(context).textTheme.titleMedium?.copyWith(
                                      fontWeight: FontWeight.bold,
                                    ),
                          ),
                          const SizedBox(height: 8),
                          ...exercise.instructions.asMap().entries.map((entry) {
                            final index = entry.key;
                            final instruction = entry.value;
                            return Padding(
                              padding: const EdgeInsets.only(bottom: 8.0),
                              child: Row(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Container(
                                    width: 24,
                                    height: 24,
                                    decoration: BoxDecoration(
                                      color: PlatformHelper.useLiquidGlass
                                          ? LiquidGlassTheme.primaryColor.withOpacity(0.2)
                                          : Theme.of(context).colorScheme.primaryContainer,
                                      shape: BoxShape.circle,
                                    ),
                                    child: Center(
                                      child: Text(
                                        '${index + 1}',
                                        style: TextStyle(
                                          fontSize: 12,
                                          fontWeight: FontWeight.bold,
                                          color: PlatformHelper.useLiquidGlass
                                              ? LiquidGlassTheme.primaryColor
                                              : Theme.of(context).colorScheme.primary,
                                        ),
                                      ),
                                    ),
                                  ),
                                  const SizedBox(width: 12),
                                  Expanded(
                                    child: Text(
                                      instruction,
                                      style: PlatformHelper.useLiquidGlass
                                          ? LiquidGlassTheme.bodyStyle
                                          : Theme.of(context).textTheme.bodyMedium,
                                    ),
                                  ),
                                ],
                              ),
                            );
                          }),
                        ],
                        const SizedBox(height: 16),
                        Align(
                          alignment: Alignment.centerRight,
                          child: AdaptiveTextButton(
                            onPressed: () => Navigator.of(context).pop(),
                            child: const Text('Close'),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
