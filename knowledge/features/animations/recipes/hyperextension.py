"""hyperextension recipe."""

import numpy as np
from engine import ANKLE_X, PELVIS_Y, TRUNK_LEN, assemble, cyc


def build(ph):
    u = cyc(ph, 0.42, 0.52)
    points, hip_frac = [], []
    for f in range(len(ph)):
        pelvis = np.array([0.0, 0.78, 0.0])
        ext = np.array([0.0, 0.94, -0.34])
        flex = np.array([0.0, -0.42, 0.91])
        dir_t = ext + (flex - ext) * u[f]
        dir_t = dir_t / np.linalg.norm(dir_t)
        sho_c = pelvis + dir_t * TRUNK_LEN
        ankles = {s: np.array([sgn * (ANKLE_X + 0.01), 0.31, -0.69]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        hands = {s: sho_c + np.array([-sgn * 0.10, -0.13, 0.13]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        points.append(
            assemble(
                pelvis,
                dir_t,
                ankles,
                hands,
                knee_poles={s: np.array([0.0, -0.1, -1.0]) for s in ("l", "r")},
                elbow_poles={s: np.array([0.0, -1.0, 0.2]) for s in ("l", "r")},
                head_dir=dir_t + np.array([0.0, 0.2, 0.0]),
            )
        )
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "none"
