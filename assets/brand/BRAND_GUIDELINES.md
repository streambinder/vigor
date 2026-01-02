# Vigor Brand Guidelines

> **"Ex Sapientia Vis"** — From Knowledge, Strength

## Brand Identity

Vigor is an AI-powered fitness training platform that generates personalized workout plans using intelligent analysis of your profile, goals, and available equipment.

### Brand Personality
- **Intelligent**: AI-driven, data-backed, research-grounded
- **Energetic**: Action-oriented, motivating, dynamic
- **Inclusive**: All fitness levels, all workout types
- **Premium**: High-quality design, attention to detail

---

## Color Palette

### Primary Colors

| Color | Hex | RGB | Usage |
|-------|-----|-----|-------|
| **Vigor Orange** | `#FF6B35` | 255, 107, 53 | Primary actions, CTAs, energy states |
| **Electric Blue** | `#0EA5E9` | 14, 165, 233 | Secondary actions, AI/intelligence, links |

### Neutral Colors (Dark Mode)

| Color | Hex | RGB | Usage |
|-------|-----|-----|-------|
| **Deep Charcoal** | `#0F0F0F` | 15, 15, 15 | Primary background (OLED-friendly) |
| **Soft Black** | `#1A1A1A` | 26, 26, 26 | Cards, elevated surfaces |
| **Elevated** | `#262626` | 38, 38, 38 | Modals, dialogs |
| **Border** | `#333333` | 51, 51, 51 | Subtle separators |

### Neutral Colors (Light Mode)

| Color | Hex | RGB | Usage |
|-------|-----|-----|-------|
| **Background** | `#FAFAFA` | 250, 250, 250 | Primary background |
| **Surface** | `#FFFFFF` | 255, 255, 255 | Cards |
| **Border** | `#E5E5E5` | 229, 229, 229 | Separators |

### Semantic Colors

| Color | Hex | Usage |
|-------|-----|-------|
| **Success** | `#22C55E` | Completed workouts, positive feedback |
| **Warning** | `#F59E0B` | Caution states, rest timers |
| **Error** | `#EF4444` | Failed states, "too hard" feedback |
| **Info** | `#3B82F6` | Tips, AI reasoning blocks |

### Text Colors

| Context | Dark Mode | Light Mode |
|---------|-----------|------------|
| Primary | `#FFFFFF` | `#0F0F0F` |
| Secondary | `#A3A3A3` | `#525252` |
| Muted | `#737373` | `#737373` |

---

## Typography

### Font Family: Barlow

Barlow is a slightly rounded, low-contrast typeface with a modern feel. It works well for both headlines and body text.

**Barlow Condensed** is used for display headlines to create impact.

### Type Scale

| Style | Size | Weight | Letter Spacing | Usage |
|-------|------|--------|----------------|-------|
| Display | 32px | Bold (700) | -0.5px | Hero headlines |
| Title | 24px | SemiBold (600) | -0.3px | Screen titles |
| Headline | 20px | SemiBold (600) | -0.2px | Section headers |
| Body Large | 17px | Regular (400) | -0.4px | Primary content |
| Body | 15px | Regular (400) | -0.2px | Secondary content |
| Label | 13px | Medium (500) | 0 | Buttons, tabs |
| Caption | 12px | Regular (400) | +0.1px | Timestamps, hints |

### Font Files

```
assets/fonts/
├── Barlow-Regular.ttf      (400)
├── Barlow-Medium.ttf       (500)
├── Barlow-SemiBold.ttf     (600)
├── Barlow-Bold.ttf         (700)
├── BarlowCondensed-SemiBold.ttf (600)
└── BarlowCondensed-Bold.ttf    (700)
```

---

## Spacing

| Token | Value | Usage |
|-------|-------|-------|
| xs | 4px | Tight gaps, icon padding |
| sm | 8px | Related items, compact lists |
| md | 16px | Standard padding, card content |
| lg | 24px | Section spacing |
| xl | 32px | Large gaps, hero sections |
| xxl | 48px | Screen margins |

---

## Border Radius

| Token | Value | Usage |
|-------|-------|-------|
| xs | 4px | Chips, badges |
| sm | 8px | Buttons, inputs |
| md | 12px | Cards |
| lg | 16px | Modals, sheets |
| xl | 24px | Navigation bars |
| full | 9999px | Pills, avatars |

---

## Platform Design

### iOS (Liquid Glass)
- Glassmorphism with frosted blur effects
- Blur: 24px
- Glass opacity: 85% (dark) / 75% (light)
- Soft borders at 10% opacity
- Orange glow on primary buttons

### Android/Web (Material You)
- Material Design 3
- Seed color: Vigor Orange
- Dynamic color system
- Elevation-based depth
- Standard Material components

---

## Usage Guidelines

### Color Application

1. **Orange** for primary actions (Generate Training, Save, Start)
2. **Blue** for secondary/informational actions (Learn More, AI Insights)
3. **Never use orange and blue together** in the same button
4. **Error states** always use red, never orange

### Typography Application

1. **Condensed** fonts only for display headlines
2. **SemiBold (600)** for interactive elements
3. **Regular (400)** for body text
4. **Never** use Bold (700) for body text

### Accessibility

- Minimum contrast ratio: 4.5:1 (WCAG AA)
- Touch targets: minimum 44x44px
- Support `prefers-reduced-motion`
- All interactive elements must have focus states

---

## App Icon

The Vigor icon features a lightning bolt symbolizing energy and power.

### Icon Colors
- Background: Deep Charcoal (`#0F0F0F`) or gradient
- Lightning bolt: Vigor Orange (`#FF6B35`) with Electric Blue (`#0EA5E9`) accent

### Icon Files
```
assets/
├── vigor-app-icon.svg       (1024x1024 source)
└── vigor-cockpit-icon.svg   (admin variant)
```

---

## Code Reference

All design tokens are defined in:
- `app/lib/design/tokens.dart` — Colors, spacing, typography, shadows
- `app/lib/design/vigor_theme.dart` — Material theme builder

Import tokens:
```dart
import 'package:vigor/design/tokens.dart';

// Use colors
VigorColors.orange
VigorColors.textPrimary(context)

// Use spacing
VigorSpacing.md
VigorSpacing.paddingMd

// Use radius
VigorRadius.card
VigorRadius.radiusSm

// Use typography
VigorTypography.body
VigorTypography.bodyColored(context)

// Use shadows
VigorShadows.elevation1(context)
VigorShadows.orangeGlow
```
