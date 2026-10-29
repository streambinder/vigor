"""Shared rig math: GLB skeleton loading, quaternion helpers, pose frames.

The Quaternius rig (mannequin and superhero share it) uses Unreal-style
bone names. The parametric JSON format stores, per bone, a quaternion that
is the bone's local rotation RELATIVE to its rest-pose local rotation:
    local = rest_local * q_rel
The root track stores the pelvis orientation the same way, plus hipHeight
as a fraction of the standing hip height.
"""

import json
import struct

import numpy as np

# vrmc-core-22 name -> Quaternius node name
BONE_MAP = {
    "hips": "pelvis",
    "spine": "spine_01",
    "chest": "spine_02",
    "upperChest": "spine_03",
    "neck": "neck_01",
    "head": "Head",
    "leftShoulder": "clavicle_l",
    "rightShoulder": "clavicle_r",
    "leftUpperArm": "upperarm_l",
    "leftLowerArm": "lowerarm_l",
    "leftHand": "hand_l",
    "rightUpperArm": "upperarm_r",
    "rightLowerArm": "lowerarm_r",
    "rightHand": "hand_r",
    "leftUpperLeg": "thigh_l",
    "leftLowerLeg": "calf_l",
    "leftFoot": "foot_l",
    "leftToes": "ball_l",
    "rightUpperLeg": "thigh_r",
    "rightLowerLeg": "calf_r",
    "rightFoot": "foot_r",
    "rightToes": "ball_r",
}

# rig joint -> (child joint, reference joint) for frame construction.
# The frame's first axis follows the bone direction (joint -> child); the
# second axis is the component of (child -> reference) orthogonal to it.
FRAME_DEFS = {
    "pelvis": ("spine_01", "spine_03"),
    "spine_01": ("spine_02", "spine_03"),
    "spine_02": ("spine_03", "neck_01"),
    "spine_03": ("neck_01", "Head"),
    "neck_01": ("Head", None),
    "Head": (None, None),
    "clavicle_l": ("upperarm_l", "lowerarm_l"),
    "upperarm_l": ("lowerarm_l", "hand_l"),
    "lowerarm_l": ("hand_l", None),
    "hand_l": (None, None),
    "clavicle_r": ("upperarm_r", "lowerarm_r"),
    "upperarm_r": ("lowerarm_r", "hand_r"),
    "lowerarm_r": ("hand_r", None),
    "hand_r": (None, None),
    "thigh_l": ("calf_l", "foot_l"),
    "calf_l": ("foot_l", "ball_l"),
    "foot_l": ("ball_l", None),
    "ball_l": (None, None),
    "thigh_r": ("calf_r", "foot_r"),
    "calf_r": ("foot_r", "ball_r"),
    "foot_r": ("ball_r", None),
    "ball_r": (None, None),
}


def quat_mul(a, b):
    """Hamilton product a*b for (x, y, z, w) quaternions."""
    ax, ay, az, aw = a
    bx, by, bz, bw = b
    return np.array(
        [
            aw * bx + ax * bw + ay * bz - az * by,
            aw * by - ax * bz + ay * bw + az * bx,
            aw * bz + ax * by - ay * bx + az * bw,
            aw * bw - ax * bx - ay * by - az * bz,
        ]
    )


def quat_conj(q):
    return np.array([-q[0], -q[1], -q[2], q[3]])


def quat_rotate(q, v):
    return quat_mul(quat_mul(q, np.array([v[0], v[1], v[2], 0.0])), quat_conj(q))[:3]


def quat_from_matrix(m):
    t = np.trace(m)
    if t > 0:
        s = np.sqrt(t + 1.0) * 2
        q = np.array(
            [
                (m[2, 1] - m[1, 2]) / s,
                (m[0, 2] - m[2, 0]) / s,
                (m[1, 0] - m[0, 1]) / s,
                0.25 * s,
            ]
        )
    elif m[0, 0] > m[1, 1] and m[0, 0] > m[2, 2]:
        s = np.sqrt(1.0 + m[0, 0] - m[1, 1] - m[2, 2]) * 2
        q = np.array(
            [
                0.25 * s,
                (m[0, 1] + m[1, 0]) / s,
                (m[0, 2] + m[2, 0]) / s,
                (m[2, 1] - m[1, 2]) / s,
            ]
        )
    elif m[1, 1] > m[2, 2]:
        s = np.sqrt(1.0 + m[1, 1] - m[0, 0] - m[2, 2]) * 2
        q = np.array(
            [
                (m[0, 1] + m[1, 0]) / s,
                0.25 * s,
                (m[1, 2] + m[2, 1]) / s,
                (m[0, 2] - m[2, 0]) / s,
            ]
        )
    else:
        s = np.sqrt(1.0 + m[2, 2] - m[0, 0] - m[1, 1]) * 2
        q = np.array(
            [
                (m[0, 2] + m[2, 0]) / s,
                (m[1, 2] + m[2, 1]) / s,
                0.25 * s,
                (m[1, 0] - m[0, 1]) / s,
            ]
        )
    return q / np.linalg.norm(q)


def slerp(a, b, t):
    d = float(np.dot(a, b))
    if d < 0:
        b, d = -b, -d
    if d > 0.9995:
        q = a + t * (b - a)
        return q / np.linalg.norm(q)
    th = np.arccos(np.clip(d, -1.0, 1.0))
    s = np.sin(th)
    return (np.sin((1 - t) * th) * a + np.sin(t * th) * b) / s


def load_glb(path):
    # Accepts a GLB (reads its JSON chunk) or the extracted node JSON
    # the repository commits as mannequin.json. Only the node skeleton
    # is ever read: generation never touches the mesh.
    with open(path, "rb") as fh:
        data = fh.read()
    if data[:4] == b"glTF":
        jl, _ = struct.unpack("<I4s", data[12:20])
        return json.loads(data[20 : 20 + jl])
    return json.loads(data)


def node_local_trs(node):
    t = np.array(node.get("translation", [0.0, 0.0, 0.0]))
    r = np.array(node.get("rotation", [0.0, 0.0, 0.0, 1.0]), dtype=float)
    s = np.array(node.get("scale", [1.0, 1.0, 1.0]))
    return t, r / np.linalg.norm(r), s


def rest_world(glb):
    """FK over rest locals. Returns {name: (world_pos, world_quat)}."""
    nodes = glb["nodes"]
    name_of = {i: n.get("name", f"node{i}") for i, n in enumerate(nodes)}
    parent = {}
    for i, n in enumerate(nodes):
        for c in n.get("children", []):
            parent[c] = i
    out = {}

    def compute(i):
        if name_of[i] in out:
            return out[name_of[i]]
        t, r, _ = node_local_trs(nodes[i])
        if i in parent:
            pp, pq = compute(parent[i])
            wp = pp + quat_rotate(pq, t)
            wq = quat_mul(pq, r)
        else:
            wp, wq = t, r
        out[name_of[i]] = (wp, wq)
        return out[name_of[i]]

    for i in range(len(nodes)):
        compute(i)
    return out, parent, name_of


def ik_2bone(root, target, l1, l2, pole):
    """Joint position for a 2-bone chain root->mid->target."""
    d = np.linalg.norm(target - root)
    d = min(d, l1 + l2 - 1e-6)
    a = (l1 * l1 - l2 * l2 + d * d) / (2 * d)
    h = np.sqrt(max(l1 * l1 - a * a, 0.0))
    axis = (target - root) / np.linalg.norm(target - root)
    perp = pole - axis * np.dot(pole, axis)
    perp /= max(np.linalg.norm(perp), 1e-9)
    return root + axis * a + perp * h


SEGMENTS = {  # bone joint -> child joint (positions in rig space)
    "spine_01": "spine_02",
    "spine_02": "spine_03",
    "spine_03": "neck_01",
    "neck_01": "Head",
    "clavicle_l": "upperarm_l",
    "upperarm_l": "lowerarm_l",
    "lowerarm_l": "hand_l",
    "clavicle_r": "upperarm_r",
    "upperarm_r": "lowerarm_r",
    "lowerarm_r": "hand_r",
    "thigh_l": "calf_l",
    "calf_l": "foot_l",
    "foot_l": "ball_l",
    "thigh_r": "calf_r",
    "calf_r": "foot_r",
    "foot_r": "ball_r",
}


def swing(a, b):
    """Shortest-arc quaternion rotating unit vector a to unit vector b."""
    a = a / np.linalg.norm(a)
    b = b / np.linalg.norm(b)
    d = float(np.dot(a, b))
    if d > 0.99999:
        return np.array([0.0, 0.0, 0.0, 1.0])
    if d < -0.99999:
        axis = np.cross(a, [1.0, 0.0, 0.0])
        if np.linalg.norm(axis) < 1e-6:
            axis = np.cross(a, [0.0, 0.0, 1.0])
        axis /= np.linalg.norm(axis)
        return np.array([axis[0], axis[1], axis[2], 0.0])
    axis = np.cross(a, b)
    q = np.array([axis[0], axis[1], axis[2], 1.0 + d])
    return q / np.linalg.norm(q)


def basis_quat(right, up_seed):
    """Quaternion with columns (right, up, forward) orthonormalised."""
    x = right / np.linalg.norm(right)
    y = up_seed - x * np.dot(up_seed, x)
    y /= np.linalg.norm(y)
    z = np.cross(x, y)
    return quat_from_matrix(np.column_stack([x, y, z]))
