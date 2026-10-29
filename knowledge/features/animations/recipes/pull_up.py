"""pull_up recipe."""

import numpy as np
from engine import ANKLE_X, LS, LT, PELVIS_Y, assemble, cyc
from hanging import hanging_base


def build(ph):
    u = cyc(ph, 0.42, 0.52)
    points, hip_frac = [], []
    for f in range(len(ph)):
        pelvis = np.array([0.0, 1.05 + 0.36 * u[f], -0.01])
        dir_t, hands = hanging_base(pelvis[1], ph)
        ankles = {
            s: np.array([sgn * ANKLE_X, pelvis[1] - LT - LS + 0.03, -0.07]) for s, sgn in (("l", 1.0), ("r", -1.0))
        }
        balls = {s: ankles[s] + np.array([0.0, -0.145, 0.09]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        points.append(
            assemble(
                pelvis,
                dir_t,
                ankles,
                hands,
                balls=balls,
                knee_poles={s: np.array([0.0, 0.0, 1.0]) for s in ("l", "r")},
                elbow_poles={s: np.array([0.0, -0.3, -1.0]) for s in ("l", "r")},
            )
        )
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "chinbar"
