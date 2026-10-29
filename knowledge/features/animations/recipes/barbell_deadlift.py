"""barbell_deadlift recipe."""

import numpy as np
from engine import (
    ANKLE_H,
    CLAV_OFF,
    FA,
    FOOT_LEN,
    HIP_HALF,
    LS,
    LT,
    NECK_DIR0,
    NECK_HEAD,
    NPHASE,
    OFF,
    PELVIS_Y,
    UA,
    UA_OFF,
    girdle_rot,
    trunk_R,
)
from rig import ik_2bone, quat_rotate, swing


def build(ph):
    """Barbell deadlift: canonical hinge anchors, solved setup, bar from the grip."""
    trunk_setup = 48.0  # deg of trunk inclination at the setup
    grip_half = 0.34
    plate_r = 0.225  # bar centre height with plates on the floor
    grip_reach = 0.09  # fingers wrap beyond the hand bone root
    arm_reach = UA + FA + grip_reach
    u = np.where(
        ph < 0.45,
        0.5 - 0.5 * np.cos(np.pi * ph / 0.45),
        np.where(ph < 0.55, 1.0, 0.5 + 0.5 * np.cos(np.pi * (ph - 0.55) / 0.45)),
    )
    tp = trunk_setup * np.power(u, 0.9)
    ankle_x = HIP_HALF + 0.03
    dx = ankle_x - HIP_HALF
    t_set = np.radians(trunk_setup)
    dir_set = np.array([0.0, np.cos(t_set), np.sin(t_set)])
    r_set = trunk_R(dir_set)
    off_thigh = OFF["thigh_l"]
    # girdle protraction: a hanging arm pulls the shoulder girdle
    # forward and down; without it the rig's shoulder sits ~5 cm too
    # high and back for the arms to reach a floor bar at all
    sho_off = OFF["spine_03"] + girdle_rot(CLAV_OFF["l"], 0.30) + girdle_rot(UA_OFF["l"], 0.30)
    # canonical anchors, fixed: pelvis 0.60, setback 0.34, bar over
    # mid-foot. The knee derives from the leg closure. An exact-reach
    # solve was tried and rejected: it drags the setup to a deep
    # squat, while a real setup carries a few cm of the reach in
    # scapular protraction (modelled above) and a fully straight arm.
    py_sol, dz_max, bar_z_sol = 0.60, 0.34, 0.02
    hip_set = np.array([0.0, py_sol, -dz_max]) + r_set @ off_thigh
    d_set = float(np.linalg.norm(hip_set - np.array([ankle_x, ANKLE_H, 0.0])))
    knee_setup = float(np.degrees(np.arccos(np.clip((LT * LT + LS * LS - d_set**2) / (2 * LT * LS), -1.0, 1.0))))
    kp = 175.0 - (175.0 - knee_setup) * u
    depth = (kp.max() - kp) / max(kp.max() - kp.min(), 1e-6)
    yfrac_derived = np.zeros(NPHASE)
    points = []
    for f in range(NPHASE):
        k = np.radians(kp[f])
        d = np.sqrt(max(LT * LT + LS * LS - 2 * LT * LS * np.cos(k), 1e-8))
        t = np.radians(tp[f])
        dir_t = np.array([0.0, np.cos(t), np.sin(t)])
        r = trunk_R(dir_t)
        dz = dz_max * depth[f]
        hip0 = r @ off_thigh
        y = ANKLE_H - hip0[1] + np.sqrt(max(d * d - (dz - hip0[2]) ** 2 - dx * dx, 1e-6))
        yfrac_derived[f] = y / PELVIS_Y
        pelvis = np.array([0.0, y, -dz])
        spine03 = pelvis + r @ OFF["spine_03"]
        neck = pelvis + r @ OFF["neck_01"]
        head_dir = np.array([0.0, np.cos(t * 0.7), np.sin(t * 0.7)])
        head = neck + quat_rotate(swing(NECK_DIR0, head_dir), r @ NECK_HEAD)
        sho_true = pelvis + r @ sho_off
        drop2 = arm_reach**2 - (bar_z_sol - sho_true[2]) ** 2 - (grip_half - abs(sho_off[0])) ** 2
        bar_y = max(plate_r, sho_true[1] - np.sqrt(max(drop2, 0.01)))
        pts = {"pelvis": pelvis, "Head": head, "neck_01": neck}
        for name in ("spine_01", "spine_02", "spine_03"):
            pts[name] = pelvis + r @ OFF[name]
        for side_name, sgn in (("l", 1.0), ("r", -1.0)):
            hip = pelvis + r @ OFF[f"thigh_{side_name}"]
            ankle = np.array([sgn * ankle_x, ANKLE_H, 0.0])
            knee = ik_2bone(hip, ankle, LT, LS, np.array([0.0, 0.15, 1.0]))
            clav = spine03 + r @ CLAV_OFF[side_name]
            sho = clav + r @ UA_OFF[side_name]
            hand = np.array([sgn * grip_half, bar_y, bar_z_sol])
            elbow = ik_2bone(sho, hand, UA, FA, np.array([0.0, -1.0, -0.2]))
            pts[f"clavicle_{side_name}"] = clav
            pts[f"upperarm_{side_name}"] = sho
            pts[f"lowerarm_{side_name}"] = elbow
            pts[f"hand_{side_name}"] = hand
            pts[f"thigh_{side_name}"] = hip
            pts[f"calf_{side_name}"] = knee
            pts[f"foot_{side_name}"] = ankle
            pts[f"ball_{side_name}"] = ankle + np.array([0.0, -ANKLE_H + 0.035, FOOT_LEN * 0.92])
        points.append(pts)
    return points, yfrac_derived, "bar"
