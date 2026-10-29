"""dumbbell_biceps_curl recipe."""

from engine import cyc
from upright import elbow_flex, upright


def build(ph):
    u = cyc(ph, 0.40, 0.50)
    return upright(u, elbow_flex, "dumbbell")
