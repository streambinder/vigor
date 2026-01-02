# Assets

## Brand

> **Ex Sapientia Vis** — From Wisdom, Strength

### Essence

Vigor is an AI-powered fitness platform generating personalized, science-backed training. The brand embodies **intelligent strength** — fusion of AI with physical power.

**Personality:** Intelligent, Energetic, Premium, Bold
**Philosophy:** "Controlled Intensity" — Powerful visuals restrained by clean minimalism.

### Colors

| Token | Hex | Usage |
|-------|-----|-------|
| `vigor-orange` | `#FF6B35` | Primary CTA, brand accent, energy |
| `electric-blue` | `#0EA5E9` | Secondary accent, data/stats, AI |
| `deep-charcoal` | `#0F0F0F` | Primary background (dark) |
| `soft-black` | `#1A1A1A` | Card surfaces (dark) |
| `off-white` | `#FAFAFA` | Primary background (light) |
| `success` | `#22C55E` | Completed, positive |
| `warning` | `#F59E0B` | Attention, in-progress |
| `error` | `#EF4444` | Destructive, failed |

**Gradients:**
- `gradient-brand`: `#FF6B35 → #0EA5E9` (hero elements, primary CTAs)
- `gradient-energy`: `#FF6B35 → #F59E0B` (motivation, intensity)

### Typography

**Font:** Barlow (Regular 400, Medium 500, SemiBold 600, Bold 700)
**Display:** Barlow Condensed (for headlines)

| Style | Size | Weight | Usage |
|-------|------|--------|-------|
| display | 32px | 700 | Hero headlines |
| title | 24px | 600 | Screen titles |
| headline | 20px | 600 | Section headers |
| body | 15px | 400 | Default content |
| label | 13px | 500 | Buttons, tabs |
| caption | 12px | 400 | Metadata |

### Spacing (4px grid)

| Token | Value | Token | Value |
|-------|-------|-------|-------|
| xs | 4px | lg | 24px |
| sm | 8px | xl | 32px |
| md | 16px | xxl | 48px |

### Border Radius

| Token | Value | Usage |
|-------|-------|-------|
| xs | 4px | Chips, badges |
| sm | 8px | Buttons, inputs |
| md | 12px | Cards |
| lg | 16px | Modals |
| full | 9999px | Pills, avatars |

### Platform Themes

**iOS (Liquid Glass):** Glassmorphism, 24px blur, 85% opacity, orange glow on CTAs
**Android/Web (Material You):** Material 3, seed color Vigor Orange, elevation depth

### Accessibility

- Contrast: 4.5:1 minimum (body), 3:1 (large text)
- Touch targets: 44x44px minimum, 8px spacing
- Respect `prefers-reduced-motion`

### Code Reference

```dart
import 'package:vigor/design/tokens.dart';

VigorColors.orange              // primary brand color
VigorColors.textPrimary(context) // theme-aware text
VigorSpacing.md                 // 16px
VigorRadius.card                // card border radius
VigorTypography.body            // default text style
VigorShadows.orangeGlow         // CTA glow effect
```

---

## Icons

### App

```bash
SRC="assets/vigor-app-icon.svg"

# iOS
IOS_DIR="app/ios/Runner/Assets.xcassets/AppIcon.appiconset"
magick -background none "$SRC" -resize 1024x1024 "$IOS_DIR/Icon-App-1024x1024@1x.png"
magick -background none "$SRC" -resize 20x20 "$IOS_DIR/Icon-App-20x20@1x.png"
magick -background none "$SRC" -resize 40x40 "$IOS_DIR/Icon-App-20x20@2x.png"
magick -background none "$SRC" -resize 60x60 "$IOS_DIR/Icon-App-20x20@3x.png"
magick -background none "$SRC" -resize 29x29 "$IOS_DIR/Icon-App-29x29@1x.png"
magick -background none "$SRC" -resize 58x58 "$IOS_DIR/Icon-App-29x29@2x.png"
magick -background none "$SRC" -resize 87x87 "$IOS_DIR/Icon-App-29x29@3x.png"
magick -background none "$SRC" -resize 40x40 "$IOS_DIR/Icon-App-40x40@1x.png"
magick -background none "$SRC" -resize 80x80 "$IOS_DIR/Icon-App-40x40@2x.png"
magick -background none "$SRC" -resize 120x120 "$IOS_DIR/Icon-App-40x40@3x.png"
magick -background none "$SRC" -resize 120x120 "$IOS_DIR/Icon-App-60x60@2x.png"
magick -background none "$SRC" -resize 180x180 "$IOS_DIR/Icon-App-60x60@3x.png"
magick -background none "$SRC" -resize 76x76 "$IOS_DIR/Icon-App-76x76@1x.png"
magick -background none "$SRC" -resize 152x152 "$IOS_DIR/Icon-App-76x76@2x.png"
magick -background none "$SRC" -resize 167x167 "$IOS_DIR/Icon-App-83.5x83.5@2x.png"

# Android
ANDROID_DIR="app/android/app/src/main/res"
magick -background none "$SRC" -resize 48x48 "$ANDROID_DIR/mipmap-mdpi/ic_launcher.png"
magick -background none "$SRC" -resize 72x72 "$ANDROID_DIR/mipmap-hdpi/ic_launcher.png"
magick -background none "$SRC" -resize 96x96 "$ANDROID_DIR/mipmap-xhdpi/ic_launcher.png"
magick -background none "$SRC" -resize 144x144 "$ANDROID_DIR/mipmap-xxhdpi/ic_launcher.png"
magick -background none "$SRC" -resize 192x192 "$ANDROID_DIR/mipmap-xxxhdpi/ic_launcher.png"

# Web
WEB_DIR="app/web"
magick -background none "$SRC" -resize 32x32 "$WEB_DIR/favicon.png"
magick -background none "$SRC" -resize 192x192 "$WEB_DIR/icons/Icon-192.png"
magick -background none "$SRC" -resize 512x512 "$WEB_DIR/icons/Icon-512.png"
magick -background none "$SRC" -resize 192x192 "$WEB_DIR/icons/Icon-maskable-192.png"
magick -background none "$SRC" -resize 512x512 "$WEB_DIR/icons/Icon-maskable-512.png"
```

### Cockpit

```bash
SRC="assets/vigor-cockpit-icon.svg"

magick -background none "$SRC" -resize 32x32 "cockpit/static/favicon.png"
magick -background none "$SRC" -resize 180x180 "cockpit/static/apple-touch-icon.png"
```
