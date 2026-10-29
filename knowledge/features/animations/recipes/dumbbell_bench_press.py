"""dumbbell_bench_press recipe."""

from engine import cyc
from lying import lying_base


def build(ph):
    w = 1.0 - cyc(ph, 0.42, 0.52)
    return lying_base(w, 0.42, 0.36, "dumbbell")
