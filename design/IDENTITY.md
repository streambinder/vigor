# Vigor: Brand Identity & Design System

**Version:** 1.0
**Status:** Approved for Development
**Philosophy:** Japanese Aesthetics $\times$ Data Science

---

## 1. Executive Summary

**Vigor** is a fitness platform that rejects the hyper-aggressive, gamified aesthetic of the current market. Instead, it positions fitness as a practice of **Calibrated Capability**.

The identity bridges the gap between the app’s rigorous backend (RAG, vector embeddings, decay algorithms) and the organic reality of human movement. It treats the user’s body not as a machine to be beaten, but as a system to be optimized.

**Tagline:** _Calibrated Capability._

---

## 2. Brand Foundations

### 2.1 The Naming Strategy

- **Name:** **Vigor**
- **The Technical Bridge:** In the app’s backend, the AI analyzes "patterns" (Movement Families) and "forms" (Vector Embeddings) to calibrate the user.
- **The Promise:** By practicing with Vigor (the AI’s plan), the user masters the form.


### 2.2 Aesthetic Philosophy

The design system is grounded in four specific Japanese principles:

1. **Kanso (Radical Simplicity):** Eliminate visual clutter. If a pixel doesn't convey data (proficiency, heat, time), remove it.

2. **Fukinsei (Intentional Asymmetry):** The UI avoids rigid, equal grids. Important elements (e.g., a recovering muscle group) take up more visual weight organically.

3. **Kintsugi (Golden Repair):** The metaphor for hypertrophy. We break muscle fibers to rebuild them stronger. High-proficiency stats are rendered in gold.

4. **Seijaku (Stillness in Motion):** The interface is calm. No flashing lights, confetti, or erratic animations.


---

## 3. Visual Identity System

### 3.1 The Logo: "The Enso Vector"

The logo combines the organic _Enso_ (Zen circle) with the precision of vector geometry.

- **The Shape:** An open brush-stroke circle. The gap represents _Fukinsei_—the idea that the work is never finished; there is always room for calibration.

- **The Construction:** The "brush stroke" is actually constructed of distinct vector nodes, representing the **11 Movement Families**.

- **Dynamic Behavior:** In the app, the stroke weight thickens based on the user's "Calibration Confidence." A new user sees a thin line; a veteran sees a bold, solid mark.


### 3.2 Color Palette: "Elements & Heat"

We avoid artificial neons. The palette is derived from natural materials and thermal dynamics.

|**Color Name**|**Hex Code**|**Usage**|**Meaning**|
|---|---|---|---|
|**Sumi Ink**|`#0F1115`|Background|Deep charcoal (OLED optimized).|
|**Washi Paper**|`#F2F0EB`|Text / Light Mode|Off-white, natural texture feel.|
|**Stone**|`#888C94`|Secondary Text|Inactive elements, history.|
|**Gold (Kintsugi)**|`#D4AF37`|Achievements|High proficiency, personal records.|
|**Indigo Dye**|`#2B4C5D`|Data: Cool|Fully recovered, low intensity.|
|**Persimmon**|`#E65D38`|Data: Warm|Active muscle, building heat.|
|**Crimson**|`#8F1D21`|Data: Hot|Overload, high stress, need recovery.|
|**Byakuroku**|`#4A7C6F`|Flow / Wellness|Mineral sage-green; recovery, stillness, mobility sessions.|

### 3.3 Typography

A pairing that balances the AI's calculation with the human's reading experience.

- **Headlines:** **Space Grotesk**

    - _Role:_ Branding, Motivation, Headers.

    - _Vibe:_ Geometric but with idiosyncratic quirks.

- **Body:** **Inter (Variable)**

    - _Role:_ Content, Instructions, Long-form text.

    - _Vibe:_ Invisible, highly legible on small screens.

- **Data & Code:** **JetBrains Mono**

    - _Role:_ Calibration scores, Rep counts, Timers, RAG confidence intervals.

    - _Vibe:_ Reminds the user that a machine is calculating this.


---

## 4. User Experience (UX) Language

### 4.1 Data Visualization: The "Decay" Model

Most apps only show progress going up. VIGOR visualizes the reality of **Proficiency Decay** (the 30-day half-life).

- **The Ink Fill:** Movement Family icons act as vessels. Training fills them with "Ink."

- **The Fade:** Over 30 days of inactivity, the ink desaturates and lowers to the 30% floor.

- **Psychology:** This visually enforces the "Use it or lose it" scientific principle without shaming the user.


### 4.2 Layout: The Masonry Dashboard

Using the principle of _Fukinsei_ (Asymmetry):

- The dashboard is not a static list. It is a dynamic masonry grid.

- If **Leg Recovery** is critical today (high muscle heat), that card expands to take up 50% of the screen.

- If **Calibration** is low, the "Test Yourself" card becomes the focal point.


### 4.3 Motion: "Liquid Feedback"

- **Input:** When rating RPE (Rate of Perceived Exertion), the UI reacts like water.

- **Transition:** Screens do not "slide" or "pop." They dissolve and reshape, mimicking the rearrangement of the vector database.

- **Loading:** Never use a spinning wheel. Use "Thought Process" text: _Analyzing constraints... Merging partner profiles... Calculating load..._


---

## 5. Verbal Identity

### 5.1 Persona: "The Stoic Scientist"

The AI is not a cheerleader. It does not use exclamation points. It is an objective, trusted partner.

- **Voice:** Precise, Knowledgeable, Calm.

- **Tone:** Analytical but deeply personalized.


### 5.2 Lexicon Shift

We redefine fitness terms to align with the engineering-meets-biology backend.

| **Standard Term**        | **Vigor Term**  | **Rationale**                                                    |
| ------------------------ | --------------- | ---------------------------------------------------------------- |
| **Workout**              | **Session**     | Implies technical practice, not just sweating.                   |
| **Soreness**             | **Heat**        | Connects to the specific "3-day half-life" logic.                |
| **Level / Rank**         | **Calibration** | Emphasizes accuracy over status hierarchy.                       |
| **Injury / Restriction** | **Constraint**  | "Constraints" are engineering variables to solve, not negatives. |
| **Personal Best**        | **Capacity**    | Focuses on current capability, not past glory.                   |

### 5.3 Dialogue Examples

> **Standard App:** "Great job! You crushed that workout! Keep it up!"
>
> **Vigor:** "Session complete. Vertical Pull capacity increased by 1.2%. Deltoid heat is elevated; 48-hour recovery recommended."

> **Standard App:** "Warning: Don't do this if your knee hurts."
>
> **Vigor:** "Constraint Detected: Right Knee Meniscus (2024). Optimization: Replacing Squat Jump with Low-Impact Box Step-up."

---

## 6. Platform Implementation

### 6.1 Mobile (The Brain)

- **Role:** Deep dive, planning, and partner management.

- **Key Feature:** **Constraint Merging Visualization.** When training with a partner, two Venn diagram circles overlap. The intersection is the safe workout zone.

- **Iconography:** 11 Abstract Glyphs for movement families (Single stroke, uniform width).


### 6.2 Watch (The Pulse)

- **Role:** Execution and telemetry.

- **Design:** _Kanso_ (Radical Simplicity).

- **Display:** Black background (OLED battery saving). High-contrast **JetBrains Mono** numbers for timers.

- **Input:** Use the Digital Crown to "dial in" RPE. Haptic feedback replaces visual noise for tempo (EMOM/HIIT).


### 6.3 Web (The Archive)

- **Role:** Analytics and macro-view.

- **Key Feature:** **Vector Space Visualization.** A WebGL view of the 384-dimensional embedding space. The user sees their "dot" moving closer to the "Ideal Form" cluster over months of training.


---

## 7. Technical Implementation Guide (CSS/Tailwind)

### Color Variables

```css
:root {
  /* Structure */
  --vigor-sumi: #0F1115;
  --vigor-washi: #F2F0EB;
  --vigor-stone: #888C94;

  /* Status/Data */
  --vigor-gold: #D4AF37;
  --vigor-indigo: #2B4C5D;
  --vigor-persimmon: #E65D38;
  --vigor-crimson: #8F1D21;
  --vigor-byakuroku: #4A7C6F;
}
```

### Typography Definition

```css
/* Tailwind Config Example */
{
  theme: {
    fontFamily: {
      sans: ['Inter', 'sans-serif'],
      display: ['Space Grotesk', 'sans-serif'],
      mono: ['JetBrains Mono', 'monospace'],
    }
  }
}
```

### Accessibility Notes

- **Heat Maps:** Do not rely on color alone for muscle heat.

    - _Cool:_ Smooth fill.

    - _Warm:_ Diagonal hatch pattern.

    - _Hot:_ Stippled/Noise texture.

- **Contrast:** All "Washi" text on "Sumi" backgrounds must meet WCAG AAA standards (Contrast ratio > 7:1).
