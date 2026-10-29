"""barbell-good-morning recipe."""

from hinge import build_hinge


def build(_ph):
    return build_hinge("goodmorning", 165.0, 62.0, 0.33, "bar_neck", "bar")
