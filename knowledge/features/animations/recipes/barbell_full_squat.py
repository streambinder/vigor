"""barbell_full_squat recipe."""

import numpy as np
from engine import (
    ANKLE_H,
    FA,
    FOOT_LEN,
    HIP_HALF,
    LS,
    LT,
    NPHASE,
    PELVIS_Y,
    SHO_HALF,
    TRUNK_LEN,
    UA,
)
from rig import ik_2bone


def build(ph):
    """Barbell back squat: canonical knee/trunk anchors, IK legs, derived hip height."""
    knee_bottom = 75.0  # deg, thigh just below parallel
    trunk_bottom = 35.0  # deg of forward trunk inclination
    hipfrac_bottom = 0.50  # pelvis height at the bottom, fraction
    stance_extra = 0.03
    grip_half = 0.36
    bar_back = 0.10
    bar_drop = 0.05
    u = np.where(
        ph < 0.42,
        0.5 - 0.5 * np.cos(np.pi * ph / 0.42),
        np.where(ph < 0.52, 1.0, 0.5 + 0.5 * np.cos(np.pi * (ph - 0.52) / 0.48)),
    )
    kp = 175.0 - (175.0 - knee_bottom) * u
    tp = trunk_bottom * np.power(u, 0.9)
    ankle_x = HIP_HALF + stance_extra
    dx = ankle_x - HIP_HALF
    depth = (kp.max() - kp) / max(kp.max() - kp.min(), 1e-6)
    d_bot = np.sqrt(max(LT * LT + LS * LS - 2 * LT * LS * np.cos(np.radians(knee_bottom)), 1e-8))
    dy_bot = hipfrac_bottom * PELVIS_Y - ANKLE_H
    dz_max = np.sqrt(max(d_bot * d_bot - dy_bot * dy_bot - dx * dx, 0.01))
    yfrac_derived = np.zeros(NPHASE)
    points = []
    for f in range(NPHASE):
        k = np.radians(kp[f])
        d = np.sqrt(max(LT * LT + LS * LS - 2 * LT * LS * np.cos(k), 1e-8))
        dz = dz_max * depth[f]
        y = ANKLE_H + np.sqrt(max(d * d - dz * dz - dx * dx, 1e-6))
        yfrac_derived[f] = y / PELVIS_Y
        pelvis = np.array([0.0, y, -dz])
        t = np.radians(tp[f])
        dir_t = np.array([0.0, np.cos(t), np.sin(t)])
        perp_t = np.array([0.0, -np.sin(t), np.cos(t)])
        sho_c3 = pelvis + dir_t * TRUNK_LEN
        bar = sho_c3 - perp_t * bar_back - dir_t * bar_drop
        head = sho_c3 + np.array([0.0, np.cos(t * 0.7), np.sin(t * 0.7)]) * 0.22
        trunk_vec3 = sho_c3 - pelvis
        pts = {
            "pelvis": pelvis,
            "spine_01": pelvis + trunk_vec3 * 0.34,
            "spine_02": pelvis + trunk_vec3 * 0.67,
            "spine_03": sho_c3,
            "neck_01": sho_c3 + (head - sho_c3) * 0.35,
            "Head": head,
        }
        for side_name, sgn in (("l", 1.0), ("r", -1.0)):
            hip = pelvis + np.array([sgn * HIP_HALF, 0.0, 0.0])
            ankle = np.array([sgn * ankle_x, ANKLE_H, 0.0])
            knee = ik_2bone(hip, ankle, LT, LS, np.array([0.0, 0.15, 1.0]))
            sho = sho_c3 + np.array([sgn * SHO_HALF, 0.0, 0.0])
            hand = bar + np.array([sgn * grip_half, -0.02, 0.0])
            elbow = ik_2bone(sho, hand, UA, FA, np.array([0.0, -1.0, -0.4]))
            pts[f"clavicle_{side_name}"] = sho_c3 + (sho - sho_c3) * 0.35
            pts[f"upperarm_{side_name}"] = sho
            pts[f"lowerarm_{side_name}"] = elbow
            pts[f"hand_{side_name}"] = hand
            pts[f"thigh_{side_name}"] = hip
            pts[f"calf_{side_name}"] = knee
            pts[f"foot_{side_name}"] = ankle
            pts[f"ball_{side_name}"] = ankle + np.array([0.0, -ANKLE_H + 0.035, FOOT_LEN * 0.92])
        points.append(pts)
    return points, yfrac_derived, "bar"
