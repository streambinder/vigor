"""barbell_hip_thrust recipe."""

import numpy as np
from engine import ANKLE_H, ANKLE_X, PELVIS_Y, assemble, cyc


def build(ph):
    u = cyc(ph, 0.42, 0.52)
    points, hip_frac = [], []
    bench = np.array([0.0, 0.84, -0.52])
    for f in range(len(ph)):
        pelvis = np.array([0.0, 0.36 + 0.20 * u[f], -0.03])
        dir_t = bench - pelvis
        dir_t = dir_t / np.linalg.norm(dir_t)
        ankles = {s: np.array([sgn * ANKLE_X, ANKLE_H, 0.40]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        hands = {s: pelvis + np.array([sgn * 0.25, 0.05, 0.08]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        points.append(
            assemble(
                pelvis,
                dir_t,
                ankles,
                hands,
                knee_poles={s: np.array([0.0, 0.7, 0.7]) for s in ("l", "r")},
                elbow_poles={s: np.array([0.0, -1.0, 0.0]) for s in ("l", "r")},
                head_dir=dir_t + np.array([0.0, 0.25, 0.0]),
            )
        )
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "bar"
