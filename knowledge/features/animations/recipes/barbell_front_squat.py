"""barbell-front-squat recipe."""

from standing import build_standing


def build(_ph):
    return build_standing("frontsquat", 78.0, 26.0, 0.5, "genie", "bar")
