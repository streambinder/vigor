"""stand_ankles family builder."""

import numpy as np
from engine import ANKLE_H, ANKLE_X, FA, PELVIS_Y, SHO_HALF, TRUNK_LEN, UA, assemble


def stand_ankles():
    return {s: np.array([sgn * ANKLE_X, ANKLE_H, 0.0]) for s, sgn in (("l", 1.0), ("r", -1.0))}


def elbow_flex(u, sho, sgn, grip_x=None):
    elbow = sho + np.array([0.0, -UA, 0.01])
    ang = 0.08 + u * 2.15
    hand = elbow + FA * np.array([0.0, -np.cos(ang), np.sin(ang)])
    if grip_x is not None:
        hand[0] = sgn * grip_x
    return {"hand": hand, "elbow": elbow}


def upright(ph_u, arm_fn, prop, pelvis_y=None, t_deg=0.0):
    """Standing builder: pelvis fixed, arms driven by arm_fn(u, side)."""
    points, hip_frac = [], []
    for f in range(len(ph_u)):
        pelvis = np.array([0.0, pelvis_y or PELVIS_Y * 0.99, 0.0])
        t = np.radians(t_deg)
        dir_t = np.array([0.0, np.cos(t), np.sin(t)])
        sho_c = pelvis + dir_t * TRUNK_LEN
        hands, elbows, poles = {}, {}, {}
        for s, sgn in (("l", 1.0), ("r", -1.0)):
            sho = sho_c + np.array([sgn * SHO_HALF, 0.0, 0.0])
            out = arm_fn(ph_u[f], sho, sgn)
            hands[s] = out["hand"]
            if "elbow" in out:
                elbows[s] = out["elbow"]
            else:
                poles[s] = out.get("pole", np.array([0.0, -1.0, -0.3]))
        points.append(
            assemble(
                pelvis,
                dir_t,
                stand_ankles(),
                hands,
                elbow_poles=poles or None,
                elbows=elbows or None,
            )
        )
        hip_frac.append(pelvis[1] / PELVIS_Y)
    return points, np.array(hip_frac), prop
