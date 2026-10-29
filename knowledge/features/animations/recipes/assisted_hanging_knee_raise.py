"""assisted_hanging_knee_raise recipe."""

import numpy as np
from engine import HIP_HALF, LS, LT, PELVIS_Y, assemble, cyc
from hanging import hanging_base


def build(ph):
    u = cyc(ph, 0.42, 0.52)
    points, hip_frac = [], []
    for f in range(len(ph)):
        pelvis = np.array([0.0, 1.02, 0.0])
        dir_t, hands = hanging_base(pelvis[1], ph)
        ankles = {}
        for s, sgn in (("l", 1.0), ("r", -1.0)):
            hip = pelvis + np.array([sgn * HIP_HALF, 0.0, 0.0])
            a = np.radians(95) * u[f]
            knee = hip + LT * np.array([0.0, -np.cos(a), np.sin(a)])
            a2 = a * 0.25
            ankles[s] = knee + LS * np.array([0.0, -np.cos(a2), np.sin(a2)])
        balls = {s: ankles[s] + np.array([0.0, -0.145, 0.09]) for s in ("l", "r")}
        pts = assemble(
            pelvis,
            dir_t,
            ankles,
            hands,
            balls=balls,
            knee_poles={s: np.array([0.0, 0.0, 1.0]) for s in ("l", "r")},
            elbow_poles={s: np.array([0.0, -1.0, -0.2]) for s in ("l", "r")},
        )
        points.append(pts)
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "chinbar"
