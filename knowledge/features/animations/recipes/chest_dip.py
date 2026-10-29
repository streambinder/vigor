"""chest_dip recipe."""

import numpy as np
from engine import PELVIS_Y, assemble, cyc


def build(ph):
    u = cyc(ph, 0.42, 0.52)
    points, hip_frac = [], []
    for f in range(len(ph)):
        pelvis = np.array([0.0, 1.06 - 0.22 * u[f], 0.02])
        lean = np.radians(4.0 + 18.0 * u[f])
        dir_t = np.array([0.0, np.cos(lean), np.sin(lean)])
        hands = {s: np.array([sgn * 0.26, 0.93, 0.0]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        ankles = {s: np.array([sgn * 0.11, pelvis[1] - 0.64, 0.27]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        balls = {s: ankles[s] + np.array([0.0, -0.13, -0.10]) for s in ("l", "r")}
        points.append(
            assemble(
                pelvis,
                dir_t,
                ankles,
                hands,
                balls=balls,
                knee_poles={s: np.array([0.0, -0.2, 1.0]) for s in ("l", "r")},
                elbow_poles={s: np.array([0.0, 0.2, -1.0]) for s in ("l", "r")},
            )
        )
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "dipbars"
