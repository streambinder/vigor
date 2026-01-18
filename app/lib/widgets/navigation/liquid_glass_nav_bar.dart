import 'dart:ui';
import 'package:flutter/material.dart';
import '../../design/tokens.dart';

class LiquidGlassNavItem {
  final IconData icon;
  final String label;

  const LiquidGlassNavItem({
    required this.icon,
    required this.label,
  });
}

class LiquidGlassNavBar extends StatelessWidget {
  final int currentIndex;
  final Function(int) onTap;
  final List<LiquidGlassNavItem> items;

  const LiquidGlassNavBar({
    super.key,
    required this.currentIndex,
    required this.onTap,
    required this.items,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    // RepaintBoundary isolates the expensive blur effect
    return RepaintBoundary(
      child: ClipRRect(
        borderRadius: VigorRadius.navigationBar,
        child: BackdropFilter(
          filter: ImageFilter.blur(
            sigmaX: VigorColors.glassBlur,
            sigmaY: VigorColors.glassBlur,
          ),
          child: Container(
            height: 70,
            decoration: BoxDecoration(
              color: isDark
                  ? Colors.black.withValues(alpha: 0.3)
                  : Colors.white.withValues(alpha: 0.3),
              borderRadius: VigorRadius.navigationBar,
              border: Border.all(
                color: isDark
                    ? Colors.white.withValues(alpha: 0.1)
                    : Colors.white.withValues(alpha: 0.3),
                width: 1.5,
              ),
              boxShadow: VigorShadows.elevation2(context),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: List.generate(
                items.length,
                (index) => _buildNavItem(
                  context,
                  items[index],
                  index,
                  isDark,
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildNavItem(
    BuildContext context,
    LiquidGlassNavItem item,
    int index,
    bool isDark,
  ) {
    final isSelected = currentIndex == index;
    final unselectedColor = isDark
        ? VigorColors.darkTextSecondary
        : VigorColors.lightTextSecondary;

    return Expanded(
      child: GestureDetector(
        onTap: () => onTap(index),
        behavior: HitTestBehavior.opaque,
        child: Container(
          color: Colors.transparent,
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              AnimatedContainer(
                duration: VigorAnimation.medium,
                curve: VigorAnimation.defaultCurve,
                padding: const EdgeInsets.symmetric(
                  horizontal: 20,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: isSelected
                      ? VigorColors.indigo.withValues(alpha: 0.15)
                      : Colors.transparent,
                  borderRadius: VigorRadius.radiusFull,
                ),
                child: Icon(
                  item.icon,
                  color: isSelected ? VigorColors.indigo : unselectedColor,
                  size: 24,
                ),
              ),
              const SizedBox(height: VigorSpacing.xs),
              AnimatedDefaultTextStyle(
                duration: VigorAnimation.medium,
                curve: VigorAnimation.defaultCurve,
                style: VigorTypography.caption.copyWith(
                  fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
                  color: isSelected ? VigorColors.indigo : unselectedColor,
                ),
                child: Text(item.label),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
