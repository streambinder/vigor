"""squat_geo family builder."""

import numpy as np
from engine import (
    ANKLE_H,
    ANKLE_X,
    HIP_HALF,
    LS,
    LT,
    NPHASE,
    PELVIS_Y,
    SHO_HALF,
    TRUNK_LEN,
    assemble,
    cyc,
)


def squat_geo(k_bot, hipfrac_bot):
    """Pelvis path for a squat-family rep from the knee anchor."""
    d_bot = np.sqrt(max(LT * LT + LS * LS - 2 * LT * LS * np.cos(np.radians(k_bot)), 1e-8))
    dy_bot = hipfrac_bot * PELVIS_Y - ANKLE_H
    dz_max = np.sqrt(max(d_bot * d_bot - dy_bot * dy_bot - (ANKLE_X - HIP_HALF) ** 2, 0.01))
    return dz_max


def squat_pelvis(u, k_top, k_bot, dz_max):
    k = np.radians(k_top - (k_top - k_bot) * u)
    d = np.sqrt(max(LT * LT + LS * LS - 2 * LT * LS * np.cos(k), 1e-8))
    dz = dz_max * u
    y = ANKLE_H + np.sqrt(max(d * d - dz * dz - (ANKLE_X - HIP_HALF) ** 2, 1e-6))
    return np.array([0.0, y, -dz])


def build_standing(_name, k_bot, t_bot, hipfrac_bot, arm_mode, prop, tempo=(0.42, 0.52)):
    ph = np.linspace(0, 1, NPHASE, endpoint=False)
    u = cyc(ph, *tempo)
    dz_max = squat_geo(k_bot, hipfrac_bot)
    points, hip_frac = [], []
    for f in range(NPHASE):
        pelvis = squat_pelvis(u[f], 175.0, k_bot, dz_max)
        t = np.radians(t_bot * u[f] ** 0.9)
        dir_t = np.array([0.0, np.cos(t), np.sin(t)])
        sho_c = pelvis + dir_t * TRUNK_LEN
        ankles = {s: np.array([sgn * ANKLE_X, ANKLE_H, 0.0]) for s, sgn in (("l", 1.0), ("r", -1.0))}
        elbows = None
        if arm_mode == "genie":
            # crossed arms (the reference front-squat variant): each
            # hand rests on the opposite shoulder, elbows point down
            # and out; the bar sits on the delts, not in the hands
            hands = {s: sho_c + np.array([-sgn * 0.10, -0.05, 0.17]) for s, sgn in (("l", 1.0), ("r", -1.0))}
            poles = {s: np.array([0.0, -1.0, 0.2]) for s in ("l", "r")}
            elbows = {
                s: sho_c + np.array([sgn * (SHO_HALF + 0.05), -0.30, 0.13]) for s, sgn in (("l", 1.0), ("r", -1.0))
            }
        elif arm_mode == "goblet":
            hands = {s: sho_c + np.array([sgn * 0.05, -0.06, 0.26]) for s, sgn in (("l", 1.0), ("r", -1.0))}
            poles = {s: np.array([0.0, -1.0, 0.2]) for s in ("l", "r")}
        else:
            raise ValueError(f"unknown arm_mode: {arm_mode}")
        points.append(assemble(pelvis, dir_t, ankles, hands, elbow_poles=poles, elbows=elbows))
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), prop
