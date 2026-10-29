"""push_up recipe."""

import numpy as np
from engine import PELVIS_Y, assemble, cyc


def build(ph):
    d = cyc(ph, 0.42, 0.52)
    points, hip_frac = [], []
    for f in range(len(ph)):
        pts, hf = pushup_pose(d[f])
        points.append(pts)
        hip_frac.append(hf)
    return points, np.array(hip_frac), "none"


def pushup_pose(d):
    # Anchors solved in page space (solve_pushup.py, bounded least
    # squares against page_fk): the rendered pose, not the
    # construction, hits the targets -- collinear body line at both
    # extremes, elbow 178 at the top and 85 at the bottom folded
    # toward the feet, hands planted. The old hand-tuned anchors
    # were unreachable once assemble() adopted the rig's true
    # shoulder (hand target 0.64 m from a 0.547 m arm).
    pelvis = np.array([0.0, 0.3833 - 0.0809 * d, -0.07])
    pitch = 0.3012 - 0.3022 * d
    dir_t = np.array([0.0, np.sin(pitch), np.cos(pitch)])
    ankles = {s: np.array([sgn * 0.10, 0.12, -0.8935 - 0.0562 * d]) for s, sgn in (("l", 1.0), ("r", -1.0))}
    balls = {s: ankles[s] + np.array([0.0, -0.095, -0.135]) for s in ("l", "r")}
    hands = {s: np.array([sgn * 0.25, 0.03, 0.3037 + 0.0298 * d]) for s, sgn in (("l", 1.0), ("r", -1.0))}
    return (
        assemble(
            pelvis,
            dir_t,
            ankles,
            hands,
            balls=balls,
            knee_poles={s: np.array([0.0, -1.0, 0.0]) for s in ("l", "r")},
            elbow_poles={s: np.array([sgn * 0.3, 0.44, 0.0]) for s, sgn in (("l", 1.0), ("r", -1.0))},
            head_dir=np.array([0.0, -0.15, 1.0]),
            targets_absolute=True,
        ),
        pelvis[1] / PELVIS_Y,
    )
