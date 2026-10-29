"""dumbbell_hammer_curl recipe."""

from engine import SHO_HALF, cyc
from upright import elbow_flex, upright


def build(ph):
    u = cyc(ph, 0.40, 0.50)
    return upright(
        u,
        lambda uu, sho, sgn: elbow_flex(uu, sho, sgn, grip_x=SHO_HALF - 0.01),
        "dumbbell",
    )
