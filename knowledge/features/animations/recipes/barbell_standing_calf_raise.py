"""barbell_standing_calf_raise recipe."""

import numpy as np
from engine import ANKLE_H, ANKLE_X, LS, LT, PELVIS_Y, TRUNK_LEN, assemble, cyc


def build(ph):
    u = cyc(ph, 0.38, 0.52)
    points, hip_frac = [], []

    for f in range(len(ph)):
        ay = ANKLE_H + 0.085 * u[f]
        pelvis = np.array([0.0, ay + LT + LS - 0.015, 0.0])
        dir_t = np.array([0.0, 1.0, 0.0])
        ankles = {s: np.array([sgn * ANKLE_X, ay, 0.0]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        balls = {s: np.array([sgn * ANKLE_X, 0.035, 0.155]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        sho_c = pelvis + dir_t * TRUNK_LEN
        bar = sho_c + np.array([0.0, -0.07, -0.10])
        hands = {s: bar + np.array([sgn * 0.34, -0.02, 0.0]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        points.append(
            assemble(
                pelvis,
                dir_t,
                ankles,
                hands,
                balls=balls,
                elbow_poles={s: np.array([0.0, -1.0, -0.4]) for s in ("l", "r")},
            )
        )
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "bar"
