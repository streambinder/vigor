"""cable_seated_row recipe."""

import numpy as np
from engine import ANKLE_X, PELVIS_Y, assemble, cyc


def build(ph):
    w = cyc(ph, 0.42, 0.52)
    points, hip_frac = [], []
    for f in range(len(ph)):
        pelvis = np.array([0.0, 0.50, -0.35])
        t = np.radians(10 - 6 * w[f])
        dir_t = np.array([0.0, np.cos(t), np.sin(t)])
        ankles = {s: np.array([sgn * ANKLE_X, 0.26, 0.44]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        balls = {s: ankles[s] + np.array([0.0, 0.10, 0.14]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        hands = {
            s: np.array([sgn * (0.16 + 0.05 * w[f]), 0.74 - 0.12 * w[f], 0.55 - 0.52 * w[f]])
            for s, sgn in (("l", 1.0), ("r", -1.0))
        }
        points.append(
            assemble(
                pelvis,
                dir_t,
                ankles,
                hands,
                balls=balls,
                knee_poles={s: np.array([0.0, 1.0, 0.2]) for s in ("l", "r")},
                elbow_poles={s: np.array([0.0, -1.0, -0.3]) for s in ("l", "r")},
            )
        )
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "none"
