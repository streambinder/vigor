"""cable_pushdown recipe."""

import numpy as np
from engine import FA, UA, cyc
from upright import upright


def build(ph):
    e = cyc(ph, 0.40, 0.50)  # extension

    def fn(ee, sho, _sgn):
        elbow = sho + np.array([0.0, -UA, 0.02])
        a = np.radians(82 - 76 * ee)
        hand = elbow + FA * np.array([0.0, -np.sin(a), np.cos(a)])
        return {"hand": hand, "elbow": elbow}

    points, hf, _ = upright(e, fn, "none")
    return points, hf, "none"
