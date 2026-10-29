"""Shared construction machinery for the exercise recipes.

Constants from the mannequin rest pose, the rigid trunk frame,
the phase solver (assemble) and the document emitter (finish).
Per-exercise builders live in one module per exercise beside
this file and import from here; family builders live in the
generic modules beside this file.
"""

import json
import sys

import numpy as np
from emit_frames import emit_frames
from rig import ik_2bone, load_glb, quat_rotate, rest_world, swing

NPHASE = 100
_glb = load_glb(sys.argv[1]) if len(sys.argv) > 1 else None
_rest, _, _ = rest_world(_glb)
RP = {n: p for n, (p, _) in _rest.items()}
LT = float(np.linalg.norm(RP["calf_l"] - RP["thigh_l"]))
LS = float(np.linalg.norm(RP["foot_l"] - RP["calf_l"]))
FOOT_LEN = float(np.linalg.norm(RP["ball_l"] - RP["foot_l"]))
HIP_HALF = abs(float(RP["thigh_l"][0] - RP["pelvis"][0]))
PELVIS_Y = float(RP["pelvis"][1])
ANKLE_H = float(RP["foot_l"][1])
TRUNK_LEN = float(np.linalg.norm(RP["spine_03"] - RP["pelvis"])) + 0.03
SHO_HALF = abs(float(RP["upperarm_l"][0] - RP["spine_03"][0]))
UA = float(np.linalg.norm(RP["lowerarm_l"] - RP["upperarm_l"]))
FA = float(np.linalg.norm(RP["hand_l"] - RP["lowerarm_l"]))
ANKLE_X = HIP_HALF + 0.03
FLAT = np.array([0.0, -ANKLE_H + 0.035, FOOT_LEN * 0.92])  # ankle->toe, planted
# Rigid trunk frame from the rig's rest geometry. assemble() used
# to place the spine/clavicle/hip points by fractional interpolation
# along the pelvis->shoulder line; those are VIRTUAL points, not the
# GLB node origins, and the mismatch is invisible upright but warps
# pitched poses (rendered shoulder 21 cm off target in the pushup).
# The frame below maps the nodes' true rest offsets rigidly.
_lat0 = RP["thigh_l"] - RP["thigh_r"]
_lat0 = _lat0 / np.linalg.norm(_lat0)
_up0 = RP["spine_03"] - RP["pelvis"]
_up0 = _up0 / np.linalg.norm(_up0)
_fwd0 = np.cross(_lat0, _up0)
_fwd0 = _fwd0 / np.linalg.norm(_fwd0)
_up0 = np.cross(_fwd0, _lat0)
_B0 = np.column_stack([_lat0, _up0, _fwd0])
OFF = {
    n: RP[n] - RP["pelvis"]
    for n in (
        "spine_01",
        "spine_02",
        "spine_03",
        "neck_01",
        "Head",
        "thigh_l",
        "thigh_r",
    )
}
CLAV_OFF = {s: RP[f"clavicle_{s}"] - RP["spine_03"] for s in ("l", "r")}
UA_OFF = {s: RP[f"upperarm_{s}"] - RP[f"clavicle_{s}"] for s in ("l", "r")}
NECK_HEAD = RP["Head"] - RP["neck_01"]
NECK_DIR0 = NECK_HEAD / np.linalg.norm(NECK_HEAD)
_VSHO = {
    s: RP["pelvis"] + _up0 * TRUNK_LEN + np.array([sgn * SHO_HALF, 0.0, 0.0]) for s, sgn in (("l", 1.0), ("r", -1.0))
}
D_HAND = {s: RP[f"upperarm_{s}"] - _VSHO[s] for s in ("l", "r")}
D_HIP = {
    s: RP[f"thigh_{s}"] - (RP["pelvis"] + np.array([sgn * HIP_HALF, 0.0, 0.0])) for s, sgn in (("l", 1.0), ("r", -1.0))
}


def cyc(ph, a=0.42, b=0.52):
    """Depth profile: 0 -> 1 in [0, a], hold to b, back to 0."""
    return np.where(
        ph < a,
        0.5 - 0.5 * np.cos(np.pi * ph / a),
        np.where(ph < b, 1.0, 0.5 + 0.5 * np.cos(np.pi * (ph - b) / (1 - b))),
    )


def rot_axis(v, axis, ang):
    axis = axis / np.linalg.norm(axis)
    return v * np.cos(ang) + np.cross(axis, v) * np.sin(ang) + axis * np.dot(axis, v) * (1 - np.cos(ang))


def trunk_R(dir_t):
    lat = np.array([1.0, 0.0, 0.0])
    up = np.asarray(dir_t, float)
    up = up / np.linalg.norm(up)
    fwd = np.cross(lat, up)
    fwd = fwd / np.linalg.norm(fwd)
    up = np.cross(fwd, lat)
    return np.column_stack([lat, up, fwd]) @ _B0.T


def girdle_rot(off, g):
    """Rotate a rest offset in the trunk frame around the lateral
    axis by g radians (girdle depression/elevation: the clavicle
    angle changes with arm elevation in a real shoulder)."""
    if not g:
        return off
    c = _B0.T @ off
    cg, sg = np.cos(g), np.sin(g)
    c = np.array([c[0], c[1] * cg - c[2] * sg, c[1] * sg + c[2] * cg])
    return _B0 @ c


def assemble(
    pelvis,
    dir_t,
    ankles,
    hands,
    knee_poles=None,
    elbow_poles=None,
    elbows=None,
    balls=None,
    head_dir=None,
    girdle=0.0,
    targets_absolute=False,
):
    """Build one phase's points dict from placed targets.

    ankles/hands: {"l": vec3, "r": vec3}. elbows: optional explicit
    elbow positions (arm-isolation work); otherwise IK with poles.
    Spine, neck, clavicle, hip and shoulder points are the rig's
    true node positions under the rigid trunk frame (see above);
    IK for arms and legs starts from those true joints.
    """
    R = trunk_R(dir_t)
    if head_dir is None:
        head_dir = dir_t + np.array([0.0, 0.45, 0.0])
        head_dir = head_dir / np.linalg.norm(head_dir)
    spine_03 = pelvis + R @ OFF["spine_03"]
    neck = pelvis + R @ OFF["neck_01"]
    head = neck + quat_rotate(swing(NECK_DIR0, head_dir), NECK_HEAD)
    pts = {
        "pelvis": pelvis,
        "spine_01": pelvis + R @ OFF["spine_01"],
        "spine_02": pelvis + R @ OFF["spine_02"],
        "spine_03": spine_03,
        "neck_01": neck,
        "Head": head,
    }
    for sname, _sgn in (("l", 1.0), ("r", -1.0)):
        hip = pelvis + R @ OFF[f"thigh_{sname}"]
        ankle = ankles[sname]
        ball = (balls or {}).get(sname, None)
        if not targets_absolute:
            ankle = ankle + R @ D_HIP[sname]
            if ball is not None:
                ball = ball + R @ D_HIP[sname]
        kpole = (knee_poles or {}).get(sname, np.array([0.0, 0.15, 1.0]))
        knee = ik_2bone(hip, ankle, LT, LS, kpole)
        pts[f"thigh_{sname}"] = hip
        pts[f"calf_{sname}"] = knee
        pts[f"foot_{sname}"] = ankle
        pts[f"ball_{sname}"] = ball if ball is not None else ankle + FLAT
        clav = spine_03 + R @ girdle_rot(CLAV_OFF[sname], girdle)
        sho = clav + R @ girdle_rot(UA_OFF[sname], girdle)
        hand = hands[sname]
        if not targets_absolute:
            hand = hand + R @ D_HAND[sname]
        if elbows and sname in elbows:
            elbow = elbows[sname]
            if not targets_absolute:
                elbow = elbow + R @ D_HAND[sname]
        else:
            epole = (elbow_poles or {}).get(sname, np.array([0.0, -1.0, -0.3]))
            elbow = ik_2bone(sho, hand, UA, FA, epole)
        pts[f"clavicle_{sname}"] = clav
        pts[f"upperarm_{sname}"] = sho
        pts[f"lowerarm_{sname}"] = elbow
        pts[f"hand_{sname}"] = hand
    return pts


def finish(points, hip_frac, out_path, name, prop, source):
    anim = emit_frames(
        points,
        sys.argv[1],
        hip_frac,
        keys=50,
        smooth_passes=1,
        source=source,
        hand_mode="plant" if name in ("push-up", "burpee", "long-plank-hold") else None,
    )
    anim["exercise"] = name
    anim["prop"] = prop
    with open(out_path, "w", encoding="utf-8") as fh:
        json.dump(anim, fh, indent=1)
    return anim
