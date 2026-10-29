"""kettlebell_swing recipe."""

import numpy as np
from engine import D_HAND, FA, UA, cyc, trunk_R
from hinge import build_hinge
from rig import ik_2bone


def build(ph):
    u = cyc(ph, 0.35, 0.45) ** 0.85

    points, hf, _ = build_hinge("swing", 128.0, 42.0, 0.30, "hang_bar", "kb", tempo=(0.35, 0.45))
    # override hands: kettlebell path, low-back at the bottom,
    # chest-high forward at the top

    for f in range(len(ph)):
        pts = points[f]
        dt = pts["spine_03"] - pts["pelvis"]
        R = trunk_R(dt / np.linalg.norm(dt))
        for s in ("l", "r"):
            sgn = 1.0 if s == "l" else -1.0
            pts[f"hand_{s}"] = np.array([sgn * 0.09, 1.40 - 0.94 * u[f], 0.52 - 0.66 * u[f]]) + R @ D_HAND[s]
            pts[f"lowerarm_{s}"] = ik_2bone(
                pts[f"upperarm_{s}"],
                pts[f"hand_{s}"],
                UA,
                FA,
                np.array([0.0, -1.0, -0.2]),
            )
    return points, hf, "kb"
