"""Direct animation emitter with full bone frames.

The house converter (landmarks_to_json.convert) orients every bone
by a shortest-arc SWING from its rest direction to the measured
direction. A swing carries no twist: the bone's rotation around its
own axis is whatever the shortest arc happens to produce. On screen
that reads as feet rolling over (toes flipped) and forearms
contorting at the elbow — the two defects Davide flagged on the
recipe squat. Positions alone cannot fix it; orientations must be
emitted as full frames.

This emitter takes constructed joint positions (model space, the
rig_positions layout) and builds each bone's world orientation as
swing + an anatomically determined twist:

- thigh / calf: twist aligns the bone's lateral axis with the knee
  hinge axis, cross(thigh_dir, shank_dir); falls back to the pelvis
  lateral axis when the leg is nearly straight.
- upperarm / forearm / hand: same, with the elbow hinge axis.
- foot: twist keeps the foot's lateral axis horizontal (sole down,
  no roll); the toe bone shares the foot frame.
- spine / neck / clavicles: twist keeps the lateral axis aligned
  with the pelvis lateral axis (the trunk does not twist).
- pelvis and head: as in the converter (full basis / neck offset).

The q_rel chaining, root track and keyframe selection mirror the
converter exactly, so the output format is unchanged.

Usage as a library: emit_frames(frames_pos, glb_path, hip_frac).
"""

import numpy as np
from rig import (
    BONE_MAP,
    SEGMENTS,
    basis_quat,
    load_glb,
    node_local_trs,
    quat_conj,
    quat_from_matrix,
    quat_mul,
    quat_rotate,
    rest_world,
    slerp,
    swing,
)


def _axis_angle(axis, angle):
    axis = axis / np.linalg.norm(axis)
    s = np.sin(angle / 2.0)
    return np.array([axis[0] * s, axis[1] * s, axis[2] * s, np.cos(angle / 2.0)])


def _twist(wq_swing, rest_q, s0_world, s_tgt, d):
    """Rotate wq_swing around the bone direction d so the bone's
    secondary axis (s0 at rest, in world terms) lands on s_tgt."""
    s0_local = quat_rotate(quat_conj(rest_q), s0_world)
    s_now = quat_rotate(wq_swing, s0_local)
    a = s_now - d * np.dot(s_now, d)
    b = s_tgt - d * np.dot(s_tgt, d)
    if np.linalg.norm(a) < 0.05 or np.linalg.norm(b) < 0.05:
        return wq_swing
    a /= np.linalg.norm(a)
    b /= np.linalg.norm(b)
    ang = np.arctan2(np.dot(np.cross(a, b), d), np.dot(a, b))
    return quat_mul(_axis_angle(d, ang), wq_swing)


def emit_frames(
    frames_pos,
    glb_path,
    hip_frac,
    keys=50,
    smooth_passes=1,
    source="parametric recipe: full-frame emission",
    duration=2.4,
    hand_mode=None,
):
    glb = load_glb(glb_path)
    rest, parent, _ = rest_world(glb)
    nodes = glb["nodes"]
    idx_of = {n.get("name"): i for i, n in enumerate(nodes)}
    inv_name = {i: n.get("name") for i, n in enumerate(nodes)}
    rest_pos = {name: wp for name, (wp, _) in rest.items()}
    rest_world_q = {name: wq for name, (_, wq) in rest.items()}
    rest_local = {}
    for node_name in BONE_MAP.values():
        _, r, _ = node_local_trs(nodes[idx_of[node_name]])
        rest_local[node_name] = r
    rest_dir = {}
    for bone, child in SEGMENTS.items():
        d = rest_pos[child] - rest_pos[bone]
        rest_dir[bone] = d / np.linalg.norm(d)
    rest_pelvis_basis = basis_quat(
        rest_pos["thigh_r"] - rest_pos["thigh_l"],
        (rest_pos["upperarm_l"] + rest_pos["upperarm_r"]) / 2 - rest_pos["pelvis"],
    )
    lateral0 = rest_pos["thigh_l"] - rest_pos["thigh_r"]
    lateral0 = lateral0 / np.linalg.norm(lateral0)
    up0 = np.array([0.0, 1.0, 0.0])
    # rest hand bases (knuckle axis, palm axis, finger direction)
    # for the planted-hand mode
    _HAND_B0 = {}
    for _s in ("l", "r"):
        _f = rest_pos[f"middle_02_{_s}"] - rest_pos[f"hand_{_s}"]
        _f = _f / np.linalg.norm(_f)
        _k = rest_pos[f"index_01_{_s}"] - rest_pos[f"pinky_01_{_s}"]
        _k = _k / np.linalg.norm(_k)
        _y = np.cross(_f, _k)
        _y = _y / np.linalg.norm(_y)
        _k = np.cross(_y, _f)
        _HAND_B0[_s] = np.column_stack([_k, _y, _f])

    frames_q = {name: [] for name in BONE_MAP.values()}
    root_q = []
    last_hinge = {}
    for _f, pos in enumerate(frames_pos):
        wq = {}
        hip_axis = pos["thigh_r"] - pos["thigh_l"]
        lateral = hip_axis / np.linalg.norm(hip_axis)
        lateral = -lateral  # rest lateral0 points thigh_l - thigh_r
        sho_c = (pos["upperarm_l"] + pos["upperarm_r"]) / 2
        torso = sho_c - pos["pelvis"]
        wq["pelvis"] = quat_mul(
            quat_mul(basis_quat(hip_axis, torso), quat_conj(rest_pelvis_basis)),
            rest_world_q["pelvis"],
        )

        def hinge(key, d1, d2, fallback, lateral):
            h = np.cross(d1, d2)
            if np.linalg.norm(h) > 0.15:
                h = h / np.linalg.norm(h)
                if np.dot(h, lateral) < 0:
                    h = -h
                last_hinge[key] = h
                return h
            return last_hinge.get(key, fallback)

        dirs = {}
        for bone, child in SEGMENTS.items():
            d = pos[child] - pos[bone]
            dirs[bone] = d / np.linalg.norm(d) if np.linalg.norm(d) > 1e-9 else rest_dir[bone]

        def propagate(parent_bone, child_bone, h, dirs, wq):
            """Child frame = parent frame hinged by the joint angle
            around the hinge axis. Pure hinge: no roll can enter,
            and the child direction is exact by construction."""
            d1, d2 = dirs[parent_bone], dirs[child_bone]
            if np.linalg.norm(np.cross(d1, d2)) < 0.02:
                return quat_mul(
                    wq[parent_bone],
                    quat_mul(quat_conj(rest_world_q[parent_bone]), rest_world_q[child_bone]),
                )
            phi = np.arctan2(np.dot(np.cross(d1, d2), h), np.dot(d1, d2))
            return quat_mul(_axis_angle(h, phi), wq[parent_bone])

        for bone in SEGMENTS:
            base = quat_mul(swing(rest_dir[bone], dirs[bone]), rest_world_q[bone])
            if bone.startswith("thigh"):
                side = bone[-1]
                h = hinge(
                    f"leg{side}",
                    dirs[f"thigh_{side}"],
                    dirs[f"calf_{side}"],
                    lateral,
                    lateral,
                )
                wq[bone] = _twist(base, rest_world_q[bone], lateral0, h, dirs[bone])
            elif bone.startswith("calf"):
                side = bone[-1]
                h = hinge(
                    f"leg{side}",
                    dirs[f"thigh_{side}"],
                    dirs[f"calf_{side}"],
                    lateral,
                    lateral,
                )
                wq[bone] = propagate(f"thigh_{side}", bone, h, dirs, wq)
            elif bone.startswith("upperarm"):
                side = bone[-1]
                h = hinge(
                    f"arm{side}",
                    dirs[f"upperarm_{side}"],
                    dirs[f"lowerarm_{side}"],
                    lateral,
                    lateral,
                )
                wq[bone] = _twist(base, rest_world_q[bone], lateral0, h, dirs[bone])
            elif bone.startswith("lowerarm"):
                side = bone[-1]
                h = hinge(
                    f"arm{side}",
                    dirs[f"upperarm_{side}"],
                    dirs[f"lowerarm_{side}"],
                    lateral,
                    lateral,
                )
                wq[bone] = propagate(f"upperarm_{side}", bone, h, dirs, wq)
            elif bone.startswith("foot"):
                tgt = np.cross(up0, dirs[bone])
                if np.linalg.norm(tgt) < 0.1:
                    tgt = last_hinge.get(f"foot{bone[-1]}", lateral)
                else:
                    tgt = tgt / np.linalg.norm(tgt)
                    last_hinge[f"foot{bone[-1]}"] = tgt
                wq[bone] = _twist(base, rest_world_q[bone], np.array([1.0, 0.0, 0.0]), tgt, dirs[bone])
            else:  # spine, neck, clavicles: keep the trunk untwisted
                wq[bone] = _twist(base, rest_world_q[bone], lateral0, lateral, dirs[bone])
        twist = pos.get("_twist", 0.0)
        if twist:
            axis = torso / np.linalg.norm(torso)
            for bone, frac in (
                ("spine_01", 0.35),
                ("spine_02", 0.65),
                ("spine_03", 1.0),
                ("neck_01", 1.1),
            ):
                wq[bone] = quat_mul(_axis_angle(axis, twist * frac), wq[bone])
        for side in ("l", "r"):
            # hands are not segment bones: rigid with the forearm,
            # via the rest offset (same trick the converter uses
            # for the head on the neck)
            wq[f"hand_{side}"] = quat_mul(
                wq[f"lowerarm_{side}"],
                quat_mul(
                    quat_conj(rest_world_q[f"lowerarm_{side}"]),
                    rest_world_q[f"hand_{side}"],
                ),
            )
            if hand_mode == "plant" and pos[f"hand_{side}"][1] < pos[f"upperarm_{side}"][1] - 0.05:
                # planted hand: palm on the floor, fingers pointing
                # along the body's forward -- a wrist bend the rigid
                # rest offset cannot express
                f_t = np.array([torso[0], 0.0, torso[2]])
                if np.linalg.norm(f_t) < 0.1:
                    f_t = np.array([0.0, 0.0, 1.0])
                f_t = f_t / np.linalg.norm(f_t)
                x_t = -lateral if side == "l" else lateral
                y_t = np.cross(f_t, x_t)
                y_t = y_t / np.linalg.norm(y_t)
                x_t = np.cross(y_t, f_t)
                Bt = np.column_stack([x_t, y_t, f_t])
                R = Bt @ _HAND_B0[side].T
                wq[f"hand_{side}"] = quat_mul(quat_from_matrix(R), rest_world_q[f"hand_{side}"])
        wq["ball_l"] = wq["foot_l"]
        wq["ball_r"] = wq["foot_r"]
        wq["Head"] = quat_mul(
            wq["neck_01"],
            quat_mul(quat_conj(rest_world_q["neck_01"]), rest_world_q["Head"]),
        )
        root_delta = quat_mul(wq["pelvis"], quat_conj(rest_world_q["pelvis"]))
        for node_name in BONE_MAP.values():
            i = idx_of[node_name]
            par = parent.get(i)
            par_name = inv_name.get(par) if par is not None else None
            if par_name is None:
                par_wq = np.array([0.0, 0.0, 0.0, 1.0])
            elif par_name in wq:
                par_wq = wq[par_name]
            else:
                par_wq = quat_mul(root_delta, rest_world_q[par_name])
            q_local = quat_mul(quat_conj(par_wq), wq[node_name])
            q_rel = quat_mul(quat_conj(rest_local[node_name]), q_local)
            frames_q[node_name].append(q_rel)
        root_q.append(root_delta)

    n = len(frames_pos)
    sel = np.linspace(0, n - 1, keys).round().astype(int)

    def smooth(qs):
        out = list(qs)
        for _ in range(smooth_passes):
            out = [slerp(out[i - 1], out[i], 0.25) if i > 0 else out[i] for i in range(len(out))]
            out = [slerp(out[i], out[i + 1], 0.25) if i < len(out) - 1 else out[i] for i in range(len(out))]
        return out

    inv_map = {v: k for k, v in BONE_MAP.items()}
    bones_out = {}
    for node_name, qs in frames_q.items():
        picked = smooth([qs[i] for i in sel])
        picked[-1] = picked[0]
        bones_out[inv_map[node_name]] = {
            "keys": [
                {"p": round(i / (keys - 1), 4), "q": [round(float(x), 4) for x in q]} for i, q in enumerate(picked)
            ]
        }
    root_picked = smooth([root_q[i] for i in sel])
    root_picked[-1] = root_picked[0]
    h_picked = [float(hip_frac[i]) for i in sel]
    h_picked[-1] = h_picked[0]
    return {
        "schema": "vigor-exercise-animation/1",
        "rig": "vrmc-core-22",
        "source": source,
        "durationHintSec": duration,
        "loop": True,
        "root": {
            "keys": [
                {
                    "p": round(i / (keys - 1), 4),
                    "hipHeight": round(float(np.clip(h_picked[i], 0.05, 1.6)), 4),
                    "q": [round(float(x), 4) for x in root_picked[i]],
                }
                for i in range(keys)
            ]
        },
        "bones": bones_out,
    }
