"""dumbbell_lateral_raise recipe."""

import numpy as np
from engine import FA, UA, cyc, rot_axis
from upright import upright


def build(ph):
    u = cyc(ph, 0.40, 0.50)

    def fn(uu, sho, sgn):
        beta = np.radians(4 + uu * 76)
        d0 = np.array([sgn * 0.10, -1.0, 0.02])
        d = rot_axis(d0 / np.linalg.norm(d0), np.array([0.0, 0.0, 1.0]), sgn * beta)
        elbow = sho + d * UA
        hand = elbow + rot_axis(d, np.array([0.0, 0.0, 1.0]), sgn * np.radians(6)) * FA
        return {"hand": hand, "elbow": elbow}

    return upright(u, fn, "dumbbell")
