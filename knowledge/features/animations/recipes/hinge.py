"""hang_bar_y family builder."""

import numpy as np
from engine import (
    ANKLE_H,
    ANKLE_X,
    FA,
    HIP_HALF,
    LS,
    LT,
    NPHASE,
    PELVIS_Y,
    SHO_HALF,
    TRUNK_LEN,
    UA,
    assemble,
    cyc,
)


def hang_bar_y(sho_y, sho_z, grip_half, reach=UA + FA + 0.09, floor=0.16):
    drop2 = reach**2 - (0.05 - sho_z) ** 2 - (grip_half - SHO_HALF) ** 2
    return max(floor, sho_y - np.sqrt(max(drop2, 0.01)))


def build_hinge(_name, k_bot, t_bot, dz_fix, arm_mode, prop, tempo=(0.45, 0.55), u_const=None):
    """Deadlift machinery generalised: setback fixed, pelvis derived."""
    ph = np.linspace(0, 1, NPHASE, endpoint=False)
    u = np.full(NPHASE, u_const) if u_const is not None else cyc(ph, *tempo)
    points, hip_frac = [], []
    for f in range(NPHASE):
        k = np.radians(175.0 - (175.0 - k_bot) * u[f])
        d = np.sqrt(max(LT * LT + LS * LS - 2 * LT * LS * np.cos(k), 1e-8))
        dz = dz_fix * u[f]
        y = ANKLE_H + np.sqrt(max(d * d - dz * dz - (ANKLE_X - HIP_HALF) ** 2, 1e-6))
        pelvis = np.array([0.0, y, -dz])
        t = np.radians(t_bot * u[f] ** 0.9)
        dir_t = np.array([0.0, np.cos(t), np.sin(t)])
        sho_c = pelvis + dir_t * TRUNK_LEN
        ankles = {s: np.array([sgn * ANKLE_X, ANKLE_H, 0.0]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        if arm_mode == "bar_neck":
            perp_t = np.array([0.0, -np.sin(t), np.cos(t)])
            bar = sho_c - perp_t * 0.10 - dir_t * 0.05
            hands = {s: bar + np.array([sgn * 0.34, -0.02, 0.0]) for s, sgn in (("l", 1.0), ("r", -1.0))}
            poles = {s: np.array([0.0, -1.0, -0.4]) for s in ("l", "r")}
        else:  # hang_bar
            by = hang_bar_y(sho_c[1], sho_c[2], 0.34)
            hands = {s: np.array([sgn * 0.34, by, 0.05]) for s, sgn in (("l", 1.0), ("r", -1.0))}
            poles = {s: np.array([0.0, -1.0, -0.2]) for s in ("l", "r")}
        points.append(assemble(pelvis, dir_t, ankles, hands, elbow_poles=poles))
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), prop
