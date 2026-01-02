# Vigor Brand Identity System

> **Ex Sapientia Vis** — From Wisdom, Strength

---

## Brand Essence

Vigor is an AI-powered fitness platform that generates personalized, science-backed training programs. The brand embodies **intelligent strength** — the fusion of cutting-edge AI with raw physical power.

### Brand Personality
- **Intelligent** — Data-driven, precise, research-backed
- **Energetic** — Dynamic, motivating, high-intensity
- **Premium** — Sophisticated, professional, trustworthy
- **Bold** — Confident, impactful, unapologetic

### Design Philosophy
**"Controlled Intensity"** — Powerful visuals restrained by clean minimalism. Energy channeled through precision.

---

## Color System

### Primary Palette

| Token | Hex | RGB | Usage |
|-------|-----|-----|-------|
| `vigor-orange` | `#FF6B35` | 255, 107, 53 | Primary CTA, brand accent, energy |
| `electric-blue` | `#0EA5E9` | 14, 165, 233 | Secondary accent, data/stats, trust |

### Gradient System

| Gradient | Colors | Usage |
|----------|--------|-------|
| `gradient-brand` | `#FF6B35 → #0EA5E9` | Hero elements, primary CTAs |
| `gradient-energy` | `#FF6B35 → #F59E0B` | Motivation, intensity |
| `gradient-calm` | `#0EA5E9 → #22C55E` | Success states, rest periods |
| `gradient-dark` | `#1A1A1A → #0F0F0F` | Background depth |

### Dark Mode (Primary)

| Token | Hex | Usage |
|-------|-----|-------|
| `deep-charcoal` | `#0F0F0F` | Primary background |
| `soft-black` | `#1A1A1A` | Card surfaces |
| `elevated-surface` | `#262626` | Modals, elevated cards |
| `border-subtle` | `#333333` | Subtle borders |
| `border-accent` | `#404040` | Active borders |

### Light Mode (Secondary)

| Token | Hex | Usage |
|-------|-----|-------|
| `off-white` | `#FAFAFA` | Primary background |
| `pure-white` | `#FFFFFF` | Card surfaces |
| `border-light` | `#E5E5E5` | Borders |

### Semantic Colors

| Token | Hex | Usage |
|-------|-----|-------|
| `success` | `#22C55E` | Completed, positive |
| `warning` | `#F59E0B` | Attention, in-progress |
| `error` | `#EF4444` | Destructive, failed |
| `info` | `#3B82F6` | Information |

### Text Colors

| Context | Dark Mode | Light Mode |
|---------|-----------|------------|
| Primary | `#FFFFFF` | `#0F0F0F` |
| Secondary | `#A3A3A3` | `#525252` |
| Muted | `#737373` | `#737373` |
| Disabled | `#525252` | `#A3A3A3` |

---

## Typography

### Font Stack
- **Display**: Barlow Condensed (700 Bold)
- **Headings**: Barlow (600 SemiBold)
- **Body**: Barlow (400 Regular, 500 Medium)
- **Mono**: JetBrains Mono (for data/stats)

### Type Scale

| Style | Size | Weight | Tracking | Line Height | Use |
|-------|------|--------|----------|-------------|-----|
| `display-xl` | 48px | 700 | -1.5px | 1.1 | Hero numbers |
| `display` | 32px | 700 | -0.5px | 1.2 | Section titles |
| `title` | 24px | 600 | -0.3px | 1.3 | Screen titles |
| `headline` | 20px | 600 | -0.2px | 1.4 | Card headers |
| `body-lg` | 17px | 400 | -0.4px | 1.5 | Featured text |
| `body` | 15px | 400 | -0.2px | 1.5 | Default body |
| `label` | 13px | 500 | 0px | 1.4 | Labels, buttons |
| `caption` | 12px | 400 | 0.1px | 1.4 | Metadata |
| `overline` | 11px | 600 | 0.5px | 1.2 | Tags (uppercase) |

### Typography Rules
1. **Headlines**: Always use Barlow Condensed for maximum impact
2. **Gradient text**: Use ShaderMask for hero stats only (1 per screen max)
3. **Uppercase**: Reserved for overlines and short labels (max 3 words)
4. **Contrast**: 4.5:1 minimum for body, 3:1 for large text

---

## Spacing System

Based on 4px grid:

| Token | Value | Usage |
|-------|-------|-------|
| `xs` | 4px | Tight spacing, icon gaps |
| `sm` | 8px | Chip padding, small gaps |
| `md` | 16px | Card padding, section gaps |
| `lg` | 24px | Screen padding, large gaps |
| `xl` | 32px | Section separation |
| `xxl` | 48px | Major sections |
| `3xl` | 64px | Page margins |

---

## Border Radius

| Token | Value | Usage |
|-------|-------|-------|
| `xs` | 4px | Chips, small tags |
| `sm` | 8px | Buttons, inputs |
| `md` | 12px | Cards, containers |
| `lg` | 16px | Modals, large cards |
| `xl` | 24px | Feature cards |
| `full` | 9999px | Pills, avatars |

---

## Elevation & Depth

### Shadow System (Dark Mode)

| Level | Shadow | Usage |
|-------|--------|-------|
| 0 | None | Flat surfaces |
| 1 | `0 2px 8px rgba(0,0,0,0.3)` | Cards |
| 2 | `0 4px 16px rgba(0,0,0,0.4)` | Modals |
| 3 | `0 8px 32px rgba(0,0,0,0.5)` | Popovers |

### Glow Effects (CTAs only)

| Type | Shadow | Usage |
|------|--------|-------|
| Orange glow | `0 4px 20px rgba(255,107,53,0.3)` | Primary CTA hover |
| Blue glow | `0 4px 20px rgba(14,165,233,0.3)` | Secondary CTA hover |
| Success glow | `0 4px 20px rgba(34,197,94,0.3)` | Complete buttons |

---

## Component Patterns

### Cards

```
┌─────────────────────────────────────┐
│ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ Gradient header    │
├─────────────────────────────────────┤
│                                     │
│  Content with md padding            │
│                                     │
└─────────────────────────────────────┘
```

**Card hierarchy:**
1. **Hero card**: Gradient border, gradient header
2. **Feature card**: Solid surface, icon accent
3. **Data card**: Minimal, stats-focused
4. **Action card**: CTA-focused, gradient button

### Buttons

| Type | Background | Border | Text |
|------|------------|--------|------|
| Primary | Gradient (orange→blue) | None | White |
| Secondary | Transparent | orange @ 40% | orange |
| Ghost | Transparent | None | textSecondary |
| Destructive | error @ 10% | error @ 30% | error |

### Pills & Chips

```
┌──────────────────┐
│ ● Label text     │  ← icon + label, colored by semantic
└──────────────────┘
```

Colors: Use semantic colors with 15% opacity background, 30% border

### Stats Display

```
┌─────────────────┐
│      127        │  ← Large gradient number (ShaderMask)
│  completed      │  ← Muted label
└─────────────────┘
```

---

## Animation & Motion

### Timing

| Token | Duration | Easing | Usage |
|-------|----------|--------|-------|
| `instant` | 100ms | ease-out | Hover feedback |
| `fast` | 150ms | ease-out | Button press |
| `medium` | 300ms | ease-out | Card transitions |
| `slow` | 500ms | ease-out-cubic | Page transitions |

### Curves

| Curve | Usage |
|-------|-------|
| `ease-out` | Entering elements |
| `ease-in` | Exiting elements |
| `ease-out-cubic` | Natural feel |
| `spring` | Playful bounces (use sparingly) |

### Animation Rules
1. **Max 1-2 animated elements per view**
2. **Always respect prefers-reduced-motion**
3. **Never use infinite animations except loaders**
4. **No scale transforms on hover (causes layout shift)**

---

## Iconography

### Icon Set
Use **Lucide Icons** (or Material Icons as fallback)
- Size: 24px default, 20px compact, 16px inline
- Stroke: 2px
- Color: Inherit from text or use semantic colors

### Icon Rules
1. **Never use emojis as UI icons**
2. **Consistent size within same context**
3. **Always use semantic colors for status icons**
4. **Add visual weight to important actions**

---

## Visual Patterns

### Gradient Header Pattern
Use at top of hero sections:
```
LinearGradient(
  colors: [orange.opacity(0.15), blue.opacity(0.15)],
  begin: topLeft,
  end: bottomRight,
)
```

### Gradient Text (Hero Stats)
```dart
ShaderMask(
  shaderCallback: (bounds) => LinearGradient(
    colors: [VigorColors.orange, VigorColors.electricBlue],
  ).createShader(bounds),
  child: Text(..., style: TextStyle(color: Colors.white)),
)
```

### Accent Ring (Avatars)
```dart
Container(
  padding: EdgeInsets.all(3),
  decoration: BoxDecoration(
    shape: BoxShape.circle,
    gradient: LinearGradient(colors: [orange, blue]),
  ),
  child: CircleAvatar(...),
)
```

### Section Divider
```dart
Container(
  height: 1,
  decoration: BoxDecoration(
    gradient: LinearGradient(
      colors: [transparent, border, transparent],
    ),
  ),
)
```

---

## Screen Layout Structure

### Standard Screen

```
┌─────────────────────────────────────┐
│ App Bar (transparent/blur on iOS)   │
├─────────────────────────────────────┤
│                                     │
│ ┌─────────────────────────────────┐ │
│ │ Hero Section (gradient bg)      │ │
│ │ Main stat / Welcome message     │ │
│ └─────────────────────────────────┘ │
│                                     │
│ ┌─────── Section Header ──────────┐ │
│ │ ● Icon  Title            Action │ │
│ └─────────────────────────────────┘ │
│                                     │
│ ┌─────────────────────────────────┐ │
│ │ Content Cards                   │ │
│ │ ...                             │ │
│ └─────────────────────────────────┘ │
│                                     │
│ [FAB if applicable]                 │
│                                     │
└─────────────────────────────────────┘
│ Bottom Nav (if main screen)         │
└─────────────────────────────────────┘
```

---

## Accessibility

### Color Contrast
- Body text: 4.5:1 minimum
- Large text (18px+): 3:1 minimum
- Interactive elements: Clear focus states

### Motion
- Respect `MediaQuery.of(context).disableAnimations`
- Provide static alternatives for animated content

### Touch Targets
- Minimum 44x44px for all interactive elements
- 8px minimum spacing between targets

---

## Implementation Checklist

Before shipping any screen:

- [ ] Uses design tokens, not hardcoded values
- [ ] Hero section has gradient background
- [ ] Maximum 1 gradient text element per screen
- [ ] All buttons have proper states (normal, hover, pressed, disabled)
- [ ] Cards have consistent border radius
- [ ] Proper spacing using token scale
- [ ] Dark/light mode tested
- [ ] No emojis as icons
- [ ] Touch targets ≥44px
- [ ] Text contrast verified

---

## File Structure

```
app/lib/design/
├── tokens.dart         # All design tokens
├── vigor_theme.dart    # ThemeData builders
├── gradients.dart      # Gradient definitions
└── animations.dart     # Motion constants

assets/brand/
├── BRAND_IDENTITY.md   # This document
├── colors.json         # Color tokens (for tooling)
└── logos/              # Brand assets
```
