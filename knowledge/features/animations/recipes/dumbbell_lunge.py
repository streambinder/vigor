"""dumbbell_lunge recipe."""

import numpy as np
from engine import (
    ANKLE_H,
    ANKLE_X,
    FA,
    PELVIS_Y,
    SHO_HALF,
    TRUNK_LEN,
    UA,
    assemble,
    cyc,
)


def build(ph):
    u = cyc(ph, 0.42, 0.52)
    points, hip_frac = [], []
    for f in range(len(ph)):
        pelvis = np.array([0.0, 0.88 - 0.34 * u[f], 0.02])
        t = np.radians(8)
        dir_t = np.array([0.0, np.cos(t), np.sin(t)])
        ankles = {
            "l": np.array([-ANKLE_X, ANKLE_H, 0.47]),
            "r": np.array([ANKLE_X, 0.17, -0.46]),
        }
        balls = {
            "l": ankles["l"] + np.array([0.0, -ANKLE_H + 0.035, 0.159]),
            "r": np.array([ANKLE_X, 0.035, -0.34]),
        }
        hands = {
            s: np.array([sgn * (SHO_HALF + 0.05), pelvis[1] + TRUNK_LEN - UA - FA + 0.03, 0.0])
            for s, sgn in (("l", 1.0), ("r", -1.0))
        }
        points.append(
            assemble(
                pelvis,
                dir_t,
                ankles,
                hands,
                balls=balls,
                knee_poles={
                    "l": np.array([0.0, 0.15, 1.0]),
                    "r": np.array([0.0, -0.6, -0.8]),
                },
                elbow_poles={s: np.array([0.0, -1.0, -0.2]) for s in ("l", "r")},
            )
        )
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "dumbbell"
