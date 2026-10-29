"""cable_pulldown recipe."""

import numpy as np
from engine import ANKLE_H, ANKLE_X, PELVIS_Y, SHO_HALF, TRUNK_LEN, assemble, cyc


def build(ph):
    w = cyc(ph, 0.42, 0.52)
    points, hip_frac = [], []
    for f in range(len(ph)):
        pelvis = np.array([0.0, 0.50, -0.28])
        dir_t = np.array([0.0, np.cos(np.radians(-10)), np.sin(np.radians(-10))])
        dir_t = dir_t / np.linalg.norm(dir_t)
        sho_c = pelvis + dir_t * TRUNK_LEN
        ankles = {s: np.array([sgn * ANKLE_X, ANKLE_H, 0.30]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        hands = {}
        for s, sgn in (("l", 1.0), ("r", -1.0)):
            sho = sho_c + np.array([sgn * SHO_HALF, 0.0, 0.0])
            top = sho + np.array([sgn * 0.30, 0.44, 0.18])
            pulled = sho + np.array([sgn * 0.26, -0.04, 0.30])
            hands[s] = top + (pulled - top) * w[f]
        points.append(
            assemble(
                pelvis,
                dir_t,
                ankles,
                hands,
                knee_poles={s: np.array([0.0, 0.2, 1.0]) for s in ("l", "r")},
                elbow_poles={s: np.array([sgn * 0.5, -0.8, -0.4]) for s, sgn in (("l", 1.0), ("r", -1.0))},
            )
        )
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "latbar"
