"""dumbbell_shrug recipe."""

import numpy as np
from engine import FA, PELVIS_Y, SHO_HALF, UA, assemble, cyc
from upright import stand_ankles


def build(ph):
    u = cyc(ph, 0.35, 0.55)
    points, hip_frac = [], []
    for f in range(len(ph)):
        pelvis = np.array([0.0, PELVIS_Y * 0.99, 0.0])
        lift = u[f] * 0.055
        dir_t = np.array([0.0, 1.0, 0.0])
        pts = assemble(
            pelvis,
            dir_t,
            stand_ankles(),
            {
                s: np.array(
                    [
                        sgn * (SHO_HALF + 0.05),
                        PELVIS_Y * 0.99 + lift - UA - FA + 0.02,
                        0.02,
                    ]
                )
                for s, sgn in (("l", 1.0), ("r", -1.0))
            },
            elbow_poles={s: np.array([0.0, -1.0, -0.2]) for s in ("l", "r")},
        )
        # shoulder lift is expressed by raising the whole arm line;
        # the spine bones stay, acceptable at 5 cm on a shrug
        points.append(pts)
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), "dumbbell"
