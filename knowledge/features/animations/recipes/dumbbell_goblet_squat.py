"""dumbbell-goblet-squat recipe."""

from standing import build_standing


def build(_ph):
    return build_standing("goblet", 85.0, 30.0, 0.52, "goblet", "dumbbell")
