import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import '../config/api_config.dart';
import '../design/tokens.dart';

/// Cached exercise image with placeholder and error handling.
/// Uses CachedNetworkImage to avoid re-fetching on rebuilds.
class CachedExerciseImage extends StatelessWidget {
  final String? imageUrl;
  final double? width;
  final double? height;
  final BoxFit fit;
  final BorderRadius? borderRadius;
  final bool isCircular;
  final IconData placeholderIcon;

  const CachedExerciseImage({
    super.key,
    required this.imageUrl,
    this.width,
    this.height,
    this.fit = BoxFit.cover,
    this.borderRadius,
    this.isCircular = false,
    this.placeholderIcon = Icons.fitness_center,
  });

  static bool isValidUrl(String? url) {
    if (url == null || url.isEmpty) return false;
    final uri = Uri.tryParse(url);
    return uri != null && uri.hasScheme && (uri.scheme == 'http' || uri.scheme == 'https');
  }

  static bool _isGif(String url) => url.toLowerCase().endsWith('.gif');

  static String proxyUrl(String url) => '${ApiConfig.baseUrl}/proxy/image?url=${Uri.encodeComponent(url)}';

  @override
  Widget build(BuildContext context) {
    if (!isValidUrl(imageUrl)) {
      return _buildPlaceholder();
    }

    // skip mem cache resizing for GIFs - resizing strips animation frames
    final isAnimated = _isGif(imageUrl!);

    final image = CachedNetworkImage(
      imageUrl: proxyUrl(imageUrl!),
      width: width,
      height: height,
      fit: fit,
      memCacheWidth: isAnimated ? null : (width != null ? (width! * 2).toInt() : null),
      memCacheHeight: isAnimated ? null : (height != null ? (height! * 2).toInt() : null),
      fadeInDuration: VigorAnimation.fast,
      fadeOutDuration: VigorAnimation.fast,
      placeholder: (_, _) => _buildPlaceholder(),
      errorWidget: (_, _, _) => _buildPlaceholder(),
    );

    if (isCircular) {
      return ClipOval(child: image);
    }

    if (borderRadius != null) {
      return ClipRRect(borderRadius: borderRadius!, child: image);
    }

    return image;
  }

  Widget _buildPlaceholder() {
    final placeholder = Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: VigorColors.stone.withValues(alpha: 0.1),
        borderRadius: isCircular ? null : borderRadius,
        shape: isCircular ? BoxShape.circle : BoxShape.rectangle,
      ),
      child: Center(
        child: Icon(
          placeholderIcon,
          size: (width ?? 72) * 0.4,
          color: VigorColors.stone.withValues(alpha: 0.5),
        ),
      ),
    );

    return placeholder;
  }
}
