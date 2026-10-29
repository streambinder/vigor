"""barbell_lying_triceps_extension_skull_crusher recipe."""

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
    u = cyc(ph, 0.42, 0.52)  # flexion toward the forehead
    points, hip_frac = [], []
    for f in range(len(ph)):
        pelvis = np.array([0.0, 0.52, -0.25])
        dir_t = np.array([0.0, 0.03, -1.0])
        dir_t = dir_t / np.linalg.norm(dir_t)
        sho_c = pelvis + dir_t * TRUNK_LEN
        ankles = {s: np.array([sgn * (ANKLE_X + 0.02), ANKLE_H, 0.33]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        elbows, hands = {}, {}
        for s, sgn in (("l", 1.0), ("r", -1.0)):
            sho = sho_c + np.array([sgn * SHO_HALF, 0.0, 0.0])
            elbows[s] = sho + np.array([0.0, UA, 0.03])
            ang = u[f] * 1.85
            hands[s] = elbows[s] + FA * np.array([0.0, np.cos(ang), -np.sin(ang) - 0.02])
        points.append(
            assemble(
                pelvis,
                dir_t,
                ankles,
                hands,
                elbows=elbows,
                knee_poles={s: np.array([0.0, 1.0, 0.3]) for s in ("l", "r")},
                head_dir=np.array([0.0, 0.15, -1.0]),
            )
        )
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "bar"
