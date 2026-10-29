"""dead_bug recipe."""

import numpy as np
from engine import HIP_HALF, PELVIS_Y, SHO_HALF, TRUNK_LEN, assemble


def build(ph):
    points, hip_frac = [], []
    for f in range(len(ph)):
        a = 0.5 + 0.5 * np.sin(2 * np.pi * ph[f])  # left-side weight
        pelvis = np.array([0.0, 0.16, 0.05])
        dir_t = np.array([0.0, 0.06, -1.0])
        dir_t = dir_t / np.linalg.norm(dir_t)
        sho_c = pelvis + dir_t * TRUNK_LEN
        ankles, hands, knees = {}, {}, {}
        for s, sgn in (("l", 1.0), ("r", -1.0)):
            hip = pelvis + np.array([sgn * HIP_HALF, 0.0, 0.0])
            w_leg = a if s == "l" else 1 - a
            w_arm = 1 - w_leg
            knee_n = hip + np.array([0.0, 0.40, 0.03])
            knee_e = hip + np.array([0.0, -0.06, 0.395])
            ank_n = knee_n + np.array([0.0, 0.03, 0.42])
            ank_e = knee_e + np.array([0.0, -0.02, 0.425])
            knees[s] = knee_n + (knee_e - knee_n) * w_leg
            ankles[s] = ank_n + (ank_e - ank_n) * w_leg
            sho = sho_c + np.array([sgn * SHO_HALF, 0.0, 0.0])
            hand_n = sho + np.array([0.0, 0.52, 0.02])
            hand_r = sho + np.array([0.0, -0.10, -0.47])
            hands[s] = hand_n + (hand_r - hand_n) * w_arm
        balls = {s: ankles[s] + np.array([0.0, -0.05, 0.155]) for s in ("l", "r")}
        pts = assemble(
            pelvis,
            dir_t,
            ankles,
            hands,
            balls=balls,
            knee_poles={s: np.array([0.0, 0.3, 1.0]) for s in ("l", "r")},
            elbow_poles={s: np.array([0.0, -0.5, -0.5]) for s in ("l", "r")},
            head_dir=np.array([0.0, 0.3, -1.0]),
        )
        points.append(pts)
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "none"
