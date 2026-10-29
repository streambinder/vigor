"""barbell_bent_over_row recipe."""

import numpy as np
from engine import D_HAND, FA, UA, cyc, trunk_R
from hinge import build_hinge
from rig import ik_2bone


def build(ph):
    w = cyc(ph, 0.42, 0.52)

    points, hf, _ = build_hinge("bor", 148.0, 58.0, 0.22, "hang_bar", "bar", u_const=1.0)
    # trunk stays hinged for the whole cycle (u constant);
    # only the arms row, on w

    for f in range(len(ph)):
        pts = points[f]
        dt = pts["spine_03"] - pts["pelvis"]
        R = trunk_R(dt / np.linalg.norm(dt))
        for s in ("l", "r"):
            sgn = 1.0 if s == "l" else -1.0
            start = np.array([sgn * 0.33, 0.60, 0.26])
            end = np.array([sgn * 0.30, 0.84, 0.02])
            pts[f"hand_{s}"] = start + (end - start) * w[f] + R @ D_HAND[s]
            pts[f"lowerarm_{s}"] = ik_2bone(
                pts[f"upperarm_{s}"],
                pts[f"hand_{s}"],
                UA,
                FA,
                np.array([sgn * 0.4, -0.5, -0.9]),
            )
    return points, hf, "bar"
