"""barbell_standing_overhead_press recipe."""

import numpy as np
from engine import cyc
from upright import upright


def build(ph):
    w = cyc(ph, 0.42, 0.52)

    def fn(ww, sho, sgn):
        hand = sho + np.array([0.05 * sgn, -0.06 + 0.62 * ww, 0.24 - 0.20 * ww])
        return {"hand": hand, "pole": np.array([sgn * 0.8, -0.6, -0.3])}

    return upright(w, fn, "bar")
