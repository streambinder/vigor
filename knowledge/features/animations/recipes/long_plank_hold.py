"""long_plank_hold recipe."""

import numpy as np
from engine import PELVIS_Y, assemble


def build(ph):
    points, hip_frac = [], []
    for f in range(len(ph)):
        sway = 0.008 * np.sin(2 * np.pi * ph[f])
        pelvis = np.array([0.0, 0.304 + sway, -0.12])
        dir_t = np.array([0.0, 0.222, 0.975])
        dir_t = dir_t / np.linalg.norm(dir_t)
        ankles = {s: np.array([sgn * 0.10, 0.12, -0.93]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        balls = {s: ankles[s] + np.array([0.0, -0.095, -0.135]) for s in ("l", "r")}
        elbows = {s: np.array([sgn * 0.20, 0.055, 0.30]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        hands = {s: np.array([sgn * 0.21, 0.03, 0.56]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        points.append(
            assemble(
                pelvis,
                dir_t,
                ankles,
                hands,
                balls=balls,
                elbows=elbows,
                knee_poles={s: np.array([0.0, -1.0, 0.0]) for s in ("l", "r")},
            )
        )
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "none"
