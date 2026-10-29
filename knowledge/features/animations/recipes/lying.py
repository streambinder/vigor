"""lying_base family builder."""

import numpy as np
from engine import ANKLE_H, ANKLE_X, PELVIS_Y, SHO_HALF, TRUNK_LEN, assemble


def lying_base(press_w, grip, _lift, prop, pelvis_y=0.52, pelvis_z=-0.25, _hand_z=-0.20):
    points, hip_frac = [], []
    for f in range(len(press_w)):
        pelvis = np.array([0.0, pelvis_y, pelvis_z])
        dir_t = np.array([0.0, 0.03, -1.0])
        dir_t = dir_t / np.linalg.norm(dir_t)
        ankles = {s: np.array([sgn * (ANKLE_X + 0.02), ANKLE_H, 0.33]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        # hand path is shoulder-relative and arcs: at the bottom the
        # load sits just over the sternum (elbow ~85 deg), at lockout
        # it is at full reach over the shoulders. The old absolute
        # path moved only vertically at a fixed z in front of the
        # shoulders, which capped the elbow ROM at ~40 deg (caught
        # by the autonomous evaluator against the take's ~92 deg).
        sho_c = pelvis + dir_t * TRUNK_LEN
        w = press_w[f]
        off_bot = np.array([0.0, 0.30, 0.16])
        off_top = np.array([0.0, 0.56, 0.02])
        off = off_bot + (off_top - off_bot) * w
        hands = {s: sho_c + np.array([sgn * (grip - SHO_HALF), 0.0, 0.0]) + off for s, sgn in (("l", 1.0), ("r", -1.0))}
        points.append(
            assemble(
                pelvis,
                dir_t,
                ankles,
                hands,
                knee_poles={s: np.array([0.0, 1.0, 0.3]) for s in ("l", "r")},
                elbow_poles={s: np.array([sgn * 0.4, -1.0, 0.0]) for s, sgn in (("l", 1.0), ("r", -1.0))},
                head_dir=np.array([0.0, 0.15, -1.0]),
            )
        )
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), prop
