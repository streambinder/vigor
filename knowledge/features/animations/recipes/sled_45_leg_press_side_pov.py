"""sled_45_leg_press_side_pov recipe."""

import numpy as np
from engine import ANKLE_X, PELVIS_Y, assemble, cyc


def build(ph):
    e = 1.0 - cyc(ph, 0.42, 0.52)
    points, hip_frac = [], []
    for f in range(len(ph)):
        pelvis = np.array([0.0, 0.45, -0.60 + 0.36 * e[f]])
        dir_t = np.array([0.0, 0.77, -0.64])
        dir_t = dir_t / np.linalg.norm(dir_t)
        ankles = {s: np.array([sgn * (ANKLE_X + 0.01), 0.80, 0.26]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        balls = {s: ankles[s] + np.array([0.0, 0.10, 0.145]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        hands = {s: pelvis + np.array([sgn * 0.30, -0.04, 0.06]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        points.append(
            assemble(
                pelvis,
                dir_t,
                ankles,
                hands,
                balls=balls,
                knee_poles={s: np.array([0.0, 1.0, 0.0]) for s in ("l", "r")},
                elbow_poles={s: np.array([0.0, -1.0, 0.0]) for s in ("l", "r")},
            )
        )
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "none"
