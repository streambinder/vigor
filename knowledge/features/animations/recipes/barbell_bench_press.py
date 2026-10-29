"""barbell_bench_press recipe."""

from engine import cyc
from lying import lying_base


def build(ph):
    w = 1.0 - cyc(ph, 0.42, 0.52)  # 1 = arms extended (top)
    return lying_base(w, 0.34, 0.38, "bar")
