"""russian_twist recipe."""

import numpy as np
from engine import PELVIS_Y, TRUNK_LEN, assemble, rot_axis


def build(ph):
    points, hip_frac = [], []
    for f in range(len(ph)):
        ang = np.radians(38) * np.sin(2 * np.pi * ph[f])
        pelvis = np.array([0.0, 0.20, 0.0])
        dir_t = np.array([0.0, 0.90, -0.44])
        dir_t = dir_t / np.linalg.norm(dir_t)
        sho_c = pelvis + dir_t * TRUNK_LEN
        nrm = np.array([0.0, 0.44, 0.90])
        nrm = nrm / np.linalg.norm(nrm)
        fwd = np.array([0.0, 0.10, 1.0])
        fwd = fwd / np.linalg.norm(fwd)
        hand_m = sho_c - dir_t * 0.15 + fwd * 0.30
        hands = {}
        for s, sgn in (("l", 1.0), ("r", -1.0)):
            h = hand_m + np.array([sgn * 0.07, 0.0, 0.0])
            hands[s] = pelvis + rot_axis(h - pelvis, dir_t, ang)
        ankles = {s: np.array([sgn * 0.13, 0.10, 0.46]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        balls = {s: ankles[s] + np.array([0.0, -0.065, 0.16]) for s in ("l", "r")}
        pts = assemble(
            pelvis,
            dir_t,
            ankles,
            hands,
            balls=balls,
            knee_poles={s: np.array([0.0, 1.0, 0.2]) for s in ("l", "r")},
            elbow_poles={s: np.array([0.0, -1.0, 0.0]) for s in ("l", "r")},
        )
        pts["_twist"] = float(ang)
        points.append(pts)
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "dumbbell"
