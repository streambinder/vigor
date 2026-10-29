"""cable_face_pull recipe."""

import numpy as np
from engine import cyc
from upright import upright


def build(ph):
    w = cyc(ph, 0.42, 0.52)

    def fn(ww, sho, sgn):
        start = sho + np.array([sgn * 0.02, 0.14, 0.55])
        end = sho + np.array([sgn * 0.20, 0.24, 0.10])
        return {
            "hand": start + (end - start) * ww,
            "pole": np.array([sgn * 1.0, 0.6, -0.2]),
        }

    return upright(w, fn, "none")
