"""hanging_base family builder."""

import numpy as np


def hanging_base(_pelvis_y, _ph):
    dir_t = np.array([0.0, 1.0, 0.0])
    hands = {s: np.array([sgn * 0.27, 2.02, 0.0]) for s, sgn in (("l", 1.0), ("r", -1.0))}
    return dir_t, hands
