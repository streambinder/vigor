"""lever_leg_extension recipe."""

import numpy as np
from engine import HIP_HALF, LS, PELVIS_Y, assemble, cyc


def build(ph):
    e = 1.0 - cyc(ph, 0.42, 0.52)  # 1 = extended
    points, hip_frac = [], []
    for f in range(len(ph)):
        pelvis = np.array([0.0, 0.52, -0.30])
        dir_t = np.array([0.0, 1.0, -0.06])
        dir_t = dir_t / np.linalg.norm(dir_t)
        ankles, balls = {}, {}
        for s, sgn in (("l", 1.0), ("r", -1.0)):
            knee = np.array([sgn * HIP_HALF, 0.50, 0.10])
            ang = (1 - e[f]) * np.radians(88)
            ankle = knee + LS * np.array([0.0, -np.cos(ang), np.sin(ang)])
            ankles[s] = ankle
            fd = np.array([0.0, -0.15 - 0.75 * e[f], 0.98 - 0.10 * e[f]])
            balls[s] = ankle + fd / np.linalg.norm(fd) * 0.173
        hands = {s: np.array([sgn * 0.30, 0.44, -0.22]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        # knees are placed by the segment direction, override after IK:
        pts = assemble(
            pelvis,
            dir_t,
            ankles,
            hands,
            balls=balls,
            knee_poles={s: np.array([0.0, 0.2, 1.0]) for s in ("l", "r")},
            elbow_poles={s: np.array([0.0, -1.0, 0.0]) for s in ("l", "r")},
        )
        for s, sgn in (("l", 1.0), ("r", -1.0)):
            pts[f"calf_{s}"] = np.array([sgn * HIP_HALF, 0.50, 0.10])
        points.append(pts)
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "none"
