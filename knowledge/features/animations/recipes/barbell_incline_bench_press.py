"""barbell_incline_bench_press recipe."""

import numpy as np
from engine import ANKLE_H, ANKLE_X, PELVIS_Y, TRUNK_LEN, assemble, cyc


def build(ph):
    w = 1.0 - cyc(ph, 0.42, 0.52)
    points, hip_frac = [], []
    for f in range(len(ph)):
        pelvis = np.array([0.0, 0.40, 0.02])
        dir_t = np.array([0.0, 0.64, -0.77])
        dir_t = dir_t / np.linalg.norm(dir_t)
        sho_c = pelvis + dir_t * TRUNK_LEN
        ankles = {s: np.array([sgn * (ANKLE_X + 0.02), ANKLE_H, 0.52]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        hands = {
            s: sho_c + np.array([sgn * 0.15, 0.06 + 0.47 * w[f], 0.18 - 0.12 * w[f]])
            for s, sgn in (("l", 1.0), ("r", -1.0))
        }
        points.append(
            assemble(
                pelvis,
                dir_t,
                ankles,
                hands,
                knee_poles={s: np.array([0.0, 1.0, 0.3]) for s in ("l", "r")},
                elbow_poles={s: np.array([sgn * 0.5, -0.8, -0.2]) for s, sgn in (("l", 1.0), ("r", -1.0))},
                head_dir=dir_t + np.array([0.0, 0.3, 0.0]),
            )
        )
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "bar"
