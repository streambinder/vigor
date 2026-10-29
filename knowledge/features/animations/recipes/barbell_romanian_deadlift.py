"""barbell-romanian-deadlift recipe."""

from hinge import build_hinge


def build(_ph):
    return build_hinge("rdl", 168.0, 63.0, 0.3, "hang_bar", "bar")
