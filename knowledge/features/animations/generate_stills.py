"""Generate one filled-silhouette SVG still per exercise animation document.

Pipeline (all programmatic, no hand-drawn assets, no screenshots):
  animation doc + Quaternius GLB -> skin the mesh to the peak frame
  (linear blend skinning with the doc's bone quaternions + root track,
  PLAYER convention) -> perspective camera (player frameCamera) ->
  rasterize figure + prop triangles into one mask -> binary closing ->
  marching-squares trace -> closed-loop Douglas-Peucker simplify ->
  Catmull-Rom Bezier smoothing (tension 0.5) -> one black SVG path,
  transparent background.

Render truth is app/assets/player/player.html:
  - pose: pelvis world quat = root_q * rest_world_quat, position =
    (0, hipHeight * rest_pelvis_y, 0); bones: world = parent_world *
    (rest_local * rel_q); parentless non-pelvis nodes keep rest world;
    nodes with no bone mapping or bone "hips" use identity rel.
  - camera: bounds over ALL keyframes' 22 vrmc bone world positions +
    origin (the player samples 12 phases; all frames is an accepted
    superset), center = box center, d = max(size) * 1.8 + 0.7,
    cam_pos = center + (d * 0.34, 0.1, d * 0.94), lookAt center,
    perspective fov 38 degrees, aspect 1.
  - props: player geometry + placement at the peak frame (makeBar,
    makeDB, makeKB, makePlainBar, makeDipBars, updateProps).
  - test-model cuts (render-only): hand triangles (40 finger joints)
    and toe-front triangles (ball_l/r, ball_leaf_l/r) are dropped
    with the player's stripHands rule: a triangle is dropped iff ALL
    THREE of its vertices are cut-weighted (any of its 4 joints in
    the cut set with weight > 0).

Peak frame rule (unchanged from v1): the keyframe with maximum total
joint rotation from rest.

Dependencies: numpy + stdlib only. The animations-builder stage
installs numpy and scipy, so this needs NO Dockerfile change for
dependencies (scipy is available but unused: binary closing is
numpy shifts, rasterization is numpy barycentric).

Usage: generate_stills.py [animations-dir] [output-dir] [glb]
Defaults: ./animations-out, ./stills-out, ./quaternius-mannequin.glb.
The third argument MUST be the GLB (mesh + skins); a skeleton-only
JSON is rejected. The process exits nonzero when any document fails
validation, so a broken animation fails the build.
"""

import itertools
import json
import math
import re
import struct
import sys
from pathlib import Path

import numpy as np
from rig import BONE_MAP, load_glb, node_local_trs, quat_mul, rest_world

HERE = Path(__file__).resolve().parent
VIEW = 200.0
RASTER = 500
SCALE = RASTER / VIEW
FOV = math.radians(38.0)
TAN_HALF = math.tan(FOV / 2.0)
ID_PATTERN = re.compile(r"^[a-z0-9-]+$")
FINGER_PREFIXES = ("index", "middle", "pinky", "ring", "thumb")
TOE_NAMES = {"ball_l", "ball_r", "ball_leaf_l", "ball_leaf_r"}
KNOWN_PROPS = {"bar", "dumbbell", "kb", "chinbar", "latbar", "dipbars", "none"}


def angle_of(q):
    """Rotation angle in radians of a unit quaternion."""
    return 2.0 * math.acos(min(1.0, abs(float(q[3]))))


def peak_frame(doc):
    """Index of the keyframe with maximum total joint rotation."""
    bones = doc["bones"]
    count = len(doc["root"]["keys"])
    best, best_total = 0, -1.0
    for i in range(count):
        total = sum(angle_of(bones[name]["keys"][i]["q"]) for name in bones if name != "hips")
        if total > best_total:
            best, best_total = i, total
    return best


# ---------------------------------------------------------------- GLB reading
def read_glb(path):
    """Parse a GLB container -> (gltf json, binary chunk)."""
    data = Path(path).read_bytes()
    if data[:4] != b"glTF":
        raise ValueError(f"{path}: not a GLB file (needs mesh + skins, not mannequin.json)")
    jl, _ = struct.unpack("<I4s", data[12:20])
    gltf = json.loads(data[20 : 20 + jl])
    blen, _ = struct.unpack("<I4s", data[20 + jl : 28 + jl])
    bin_data = data[28 + jl : 28 + jl + blen]
    return gltf, bin_data


def read_accessor(gltf, bin_data, index):
    acc = gltf["accessors"][index]
    bv = gltf["bufferViews"][acc["bufferView"]]
    comp = {5120: "i1", 5121: "u1", 5122: "i2", 5123: "u2", 5125: "u4", 5126: "f4"}[acc["componentType"]]
    count = acc["count"]
    ncomp = {"SCALAR": 1, "VEC2": 2, "VEC3": 3, "VEC4": 4, "MAT4": 16}[acc["type"]]
    start = bv.get("byteOffset", 0) + acc.get("byteOffset", 0)
    stride = bv.get("byteStride")
    item = np.dtype(comp).itemsize * ncomp
    if stride and stride != item:
        raw = np.frombuffer(bin_data, dtype=comp, count=count * stride // np.dtype(comp).itemsize, offset=start)
        raw = raw.reshape(count, stride // np.dtype(comp).itemsize)[:, :ncomp]
    else:
        raw = np.frombuffer(bin_data, dtype=comp, count=count * ncomp, offset=start)
        raw = raw.reshape(count, ncomp)
    if acc["type"] == "MAT4":
        # Column-major storage: transpose each row-major reshape.
        return raw.reshape(count, 4, 4).transpose(0, 2, 1).astype(float)
    return raw.astype(float)


def load_mesh(gltf, bin_data):
    """Mesh primitives + skin joints + inverse bind matrices."""
    prims = []
    for mesh in gltf["meshes"]:
        for pr in mesh["primitives"]:
            at = pr["attributes"]
            prims.append(
                {
                    "pos": read_accessor(gltf, bin_data, at["POSITION"])[:, :3],
                    "joints": read_accessor(gltf, bin_data, at["JOINTS_0"]).astype(int),
                    "weights": read_accessor(gltf, bin_data, at["WEIGHTS_0"]),
                    "idx": read_accessor(gltf, bin_data, pr["indices"]).astype(int).reshape(-1, 3),
                }
            )
    skin_data = gltf["skins"][0]
    ibm = read_accessor(gltf, bin_data, skin_data["inverseBindMatrices"])
    return prims, skin_data["joints"], ibm


# ---------------------------------------------------------------- pose
def quat_to_matrix(q):
    x, y, z, w = q
    return np.array(
        [
            [1 - 2 * (y * y + z * z), 2 * (x * y - z * w), 2 * (x * z + y * w)],
            [2 * (x * y + z * w), 1 - 2 * (x * x + z * z), 2 * (y * z - x * w)],
            [2 * (x * z - y * w), 2 * (y * z + x * w), 1 - 2 * (x * x + y * y)],
        ]
    )


def compose(pos, quat):
    m = np.eye(4)
    m[:3, :3] = quat_to_matrix(quat)
    m[:3, 3] = pos
    return m


def prepare_rig(glb):
    """Doc-independent rig data (computed once)."""
    rest_map, parent, name_of = rest_world(glb)
    nodes = glb["nodes"]
    rest = {name_of[i]: node_local_trs(node) for i, node in enumerate(nodes)}
    inv_map = {node: bone for bone, node in BONE_MAP.items()}
    name_to_idx = {n: i for i, n in name_of.items()}
    rest_pelvis_pos, rest_pelvis_quat = rest_map["pelvis"]
    return {
        "rest_map": rest_map,
        "parent": parent,
        "name_of": name_of,
        "rest": rest,
        "inv_map": inv_map,
        "name_to_idx": name_to_idx,
        "rest_pelvis_pos": rest_pelvis_pos,
        "rest_pelvis_quat": rest_pelvis_quat,
        "bone_nodes": [name_to_idx[node] for node in BONE_MAP.values()],
    }


def pose_matrices(rig, doc, frame):
    """World matrix per node index at one keyframe, PLAYER convention."""
    rest_map = rig["rest_map"]
    parent = rig["parent"]
    name_of = rig["name_of"]
    rest = rig["rest"]
    inv_map = rig["inv_map"]
    root_key = doc["root"]["keys"][frame]
    hip_height = float(root_key["hipHeight"])
    rest_pelvis_pos = rig["rest_pelvis_pos"]
    rest_pelvis_quat = rig["rest_pelvis_quat"]
    mats = {}

    def compute(i):
        name = name_of[i]
        if i in mats:
            return mats[i]
        trans, rest_rot, _ = rest[name]
        if name == "pelvis":
            pos = np.array([0.0, hip_height * float(rest_pelvis_pos[1]), 0.0])
            quat = quat_mul(np.array(root_key["q"], dtype=float), rest_pelvis_quat)
            m = compose(pos, quat)
        elif i not in parent:
            pos, quat = rest_map[name]
            m = compose(pos, quat)
        else:
            pm = compute(parent[i])
            bone = inv_map.get(name)
            rel = (
                np.array([0.0, 0.0, 0.0, 1.0])
                if bone is None or bone == "hips"
                else np.array(doc["bones"][bone]["keys"][frame]["q"], dtype=float)
            )
            m = pm @ compose(trans, quat_mul(rest_rot, rel))
        mats[i] = m
        return m

    for i in range(len(rest)):
        compute(i)
    return mats


def is_finger_joint(name):
    for p in FINGER_PREFIXES:
        for s in ("01", "02", "03"):
            if name in (f"{p}_{s}_l", f"{p}_{s}_r"):
                return True
        if name in (f"{p}_04_leaf_l", f"{p}_04_leaf_r"):
            return True
    return False


def strip_kept(prim, finger_joints, toe_joints):
    """Triangle indices kept after the test-model cuts (hands + toes)."""
    keep = []
    for tri in prim["idx"]:

        def vert_cut(v, cut):
            return any(prim["joints"][v, k] in cut and prim["weights"][v, k] > 0 for k in range(4))

        all_finger = all(vert_cut(v, finger_joints) for v in tri)
        all_toe = all(vert_cut(v, toe_joints) for v in tri)
        if not (all_finger or all_toe):
            keep.append(tri)
    return np.array(keep)


def skin(prim, joint_nodes, ibm, mats):
    """Linear blend skinning of one primitive at a posed frame."""
    w = prim["weights"]
    w = w / np.maximum(w.sum(axis=1, keepdims=True), 1e-12)
    out = np.zeros_like(prim["pos"])
    for k in range(4):
        for j in set(prim["joints"][:, k].tolist()):
            sel = prim["joints"][:, k] == j
            if not sel.any():
                continue
            m = mats[joint_nodes[j]] @ ibm[j]
            out[sel] += w[sel, k : k + 1] * (prim["pos"][sel] @ m[:3, :3].T + m[:3, 3])
    return out


# ---------------------------------------------------------------- camera
def compute_camera(rig, doc):
    """Player frameCamera: bounds over all keyframes + origin."""
    pts = [np.zeros(3)]
    for f in range(len(doc["root"]["keys"])):
        mats = pose_matrices(rig, doc, f)
        for ni in rig["bone_nodes"]:
            pts.append(mats[ni][:3, 3])
    pts = np.array(pts)
    mn, mx = pts.min(axis=0), pts.max(axis=0)
    center, size = (mn + mx) / 2.0, mx - mn
    d = float(size.max()) * 1.8 + 0.7
    cam_pos = center + np.array([d * 0.34, 0.1, d * 0.94])
    z_axis = cam_pos - center
    z_axis /= np.linalg.norm(z_axis)
    x_axis = np.cross(np.array([0.0, 1.0, 0.0]), z_axis)
    x_axis /= np.linalg.norm(x_axis)
    y_axis = np.cross(z_axis, x_axis)
    return cam_pos, x_axis, y_axis, z_axis


def project(world, cam):
    """World points -> (viewBox screen coords, view depth)."""
    cam_pos, x_axis, y_axis, z_axis = cam
    rel = world - cam_pos
    vx, vy = rel @ x_axis, rel @ y_axis
    depth = -(rel @ z_axis)
    x_ndc = vx / (depth * TAN_HALF)
    y_ndc = vy / (depth * TAN_HALF)
    return np.column_stack([(x_ndc + 1.0) / 2.0 * VIEW, (1.0 - y_ndc) / 2.0 * VIEW]), depth


# ---------------------------------------------------------------- prop meshes
def _basis(axis):
    a = axis / (np.linalg.norm(axis) or 1.0)
    ref = np.array([0.0, 1.0, 0.0]) if abs(a[1]) < 0.9 else np.array([1.0, 0.0, 0.0])
    y = np.cross(ref, a)
    y /= np.linalg.norm(y)
    return a, y, np.cross(a, y)


def cylinder_tris(p0, p1, radius, nseg=16):
    """Prism between two points (caps included), as (3, 3) triangles."""
    p0, p1 = np.array(p0, dtype=float), np.array(p1, dtype=float)
    axis = p1 - p0
    length = np.linalg.norm(axis)
    if length < 1e-12:
        return []
    _, y, z = _basis(axis)
    ring0, ring1 = [], []
    for i in range(nseg):
        th = 2.0 * math.pi * i / nseg
        off = radius * (math.cos(th) * y + math.sin(th) * z)
        ring0.append(p0 + off)
        ring1.append(p1 + off)
    tris = []
    for i in range(nseg):
        j = (i + 1) % nseg
        tris.append(np.array([ring0[i], ring1[i], ring1[j]]))
        tris.append(np.array([ring0[i], ring1[j], ring0[j]]))
        tris.append(np.array([p0, ring0[j], ring0[i]]))
        tris.append(np.array([p1, ring1[i], ring1[j]]))
    return tris


def sphere_tris(center, radius, nlat=10, nlon=16):
    """UV sphere mesh (pole vertices + rings), as (3, 3) triangles."""
    center = np.array(center, dtype=float)
    rings = []
    for k in range(1, nlat):
        phi = math.pi * k / nlat
        ring = []
        for i in range(nlon):
            th = 2.0 * math.pi * i / nlon
            ring.append(
                center + radius * np.array([math.sin(phi) * math.cos(th), math.cos(phi), math.sin(phi) * math.sin(th)])
            )
        rings.append(ring)
    north = center + radius * np.array([0.0, 1.0, 0.0])
    south = center + radius * np.array([0.0, -1.0, 0.0])
    tris = []
    for i in range(nlon):
        j = (i + 1) % nlon
        tris.append(np.array([north, rings[0][j], rings[0][i]]))
        tris.append(np.array([south, rings[-1][i], rings[-1][j]]))
    for r in range(len(rings) - 1):
        for i in range(nlon):
            j = (i + 1) % nlon
            tris.append(np.array([rings[r][i], rings[r + 1][i], rings[r + 1][j]]))
            tris.append(np.array([rings[r][i], rings[r + 1][j], rings[r][j]]))
    return tris


def prop_tris(prop, jp):
    """Player prop geometry at the peak frame, in world coords.

    jp: {node name: world position}. Mirrors updateProps + makeBar /
    makeDB / makeKB / makePlainBar / makeDipBars. The kettlebell
    handle torus is approximated by tube segments along its arc.
    """
    hl, hr = jp["hand_l"], jp["hand_r"]
    mid = (hl + hr) / 2.0
    axis = hl - hr
    if np.linalg.norm(axis) < 1e-9:
        axis = np.array([1.0, 0.0, 0.0])
    axis = axis / np.linalg.norm(axis)
    tris = []
    if prop == "bar":
        tris += cylinder_tris(mid - axis * 1.1, mid + axis * 1.1, 0.016, 16)
        for s in (-1, 1):
            c = mid + axis * (s * 0.86)
            tris += cylinder_tris(c - axis * 0.0225, c + axis * 0.0225, 0.225, 24)
    elif prop == "dumbbell":
        for hand, fore in ((hl, jp["lowerarm_l"]), (hr, jp["lowerarm_r"])):
            ax = hand - fore
            if np.linalg.norm(ax) < 1e-9:
                ax = np.array([1.0, 0.0, 0.0])
            ax = ax / np.linalg.norm(ax)
            tris += cylinder_tris(hand - ax * 0.17, hand + ax * 0.17, 0.014, 8)
            for s in (-1, 1):
                c = hand + ax * (s * 0.12)
                tris += cylinder_tris(c - ax * 0.025, c + ax * 0.025, 0.075, 16)
    elif prop == "kb":
        tris += sphere_tris(mid + np.array([0.0, -0.16, 0.0]), 0.105)
        arc_center = mid + np.array([0.0, -0.055, 0.0])
        arc = [arc_center + 0.075 * np.array([math.cos(th), math.sin(th), 0.0]) for th in np.linspace(0.0, math.pi, 13)]
        for a0, a1 in itertools.pairwise(arc):
            tris += cylinder_tris(a0, a1, 0.016, 8)
    elif prop == "chinbar":
        tris += cylinder_tris(mid - axis * 0.625, mid + axis * 0.625, 0.016, 10)
    elif prop == "latbar":
        tris += cylinder_tris(mid - axis * 0.775, mid + axis * 0.775, 0.016, 10)
    elif prop == "dipbars":
        for s in (-1, 1):
            c = np.array([s * 0.26, hl[1] - 0.01, hl[2] - 0.04])
            tris += cylinder_tris(c - np.array([0.0, 0.0, 0.26]), c + np.array([0.0, 0.0, 0.26]), 0.018, 10)
    elif prop == "none":
        pass
    else:
        raise ValueError(f"unknown prop {prop!r}")
    return tris


# ---------------------------------------------------------------- raster + trace
def rasterize(tris_screen):
    """Screen triangles (500-scale coords) -> bool mask, numpy only."""
    mask = np.zeros((RASTER, RASTER), dtype=bool)
    for tri in tris_screen:
        xs, ys = tri[:, 0], tri[:, 1]
        x0, x1 = max(0, int(np.floor(xs.min()))), min(RASTER, int(np.ceil(xs.max())) + 1)
        y0, y1 = max(0, int(np.floor(ys.min()))), min(RASTER, int(np.ceil(ys.max())) + 1)
        if x1 <= x0 or y1 <= y0:
            continue
        area2 = (ys[1] - ys[2]) * (xs[0] - xs[2]) + (xs[2] - xs[1]) * (ys[0] - ys[2])
        if area2 == 0:
            continue
        gx, gy = np.meshgrid(np.arange(x0, x1) + 0.5, np.arange(y0, y1) + 0.5)
        w0 = ((ys[1] - ys[2]) * (gx - xs[2]) + (xs[2] - xs[1]) * (gy - ys[2])) / area2
        w1 = ((ys[2] - ys[0]) * (gx - xs[2]) + (xs[0] - xs[2]) * (gy - ys[2])) / area2
        w2 = 1.0 - w0 - w1
        inside = (w0 >= -1e-9) & (w1 >= -1e-9) & (w2 >= -1e-9)
        if inside.any():
            mask[y0:y1, x0:x1] |= inside
    return mask


def binary_closing(mask):
    """3x3 binary closing (dilate then erode), 1 iteration, numpy shifts."""
    padded = np.pad(mask, 1, mode="constant", constant_values=False)
    dil = np.zeros_like(mask)
    for dy in range(3):
        for dx in range(3):
            dil |= padded[dy : dy + mask.shape[0], dx : dx + mask.shape[1]]
    padded = np.pad(dil, 1, mode="constant", constant_values=False)
    ero = np.ones_like(mask)
    for dy in range(3):
        for dx in range(3):
            ero &= padded[dy : dy + mask.shape[0], dx : dx + mask.shape[1]]
    return ero


def trace_mask(mask):
    """Binary mask -> list of loops (each a list of (x, y) pixel coords)."""
    inside = mask > 127 if mask.dtype != bool else mask
    tl = inside[:-1, :-1]
    tr = inside[:-1, 1:]
    br = inside[1:, 1:]
    bl = inside[1:, :-1]
    case = (tl.astype(int) << 3) | (tr.astype(int) << 2) | (br.astype(int) << 1) | bl.astype(int)
    table = {
        1: [("l", "b")],
        2: [("b", "r")],
        3: [("l", "r")],
        4: [("t", "r")],
        5: [("t", "r"), ("l", "b")],
        6: [("t", "b")],
        7: [("t", "l")],
        8: [("t", "l")],
        9: [("t", "b")],
        10: [("t", "l"), ("b", "r")],
        11: [("t", "r")],
        12: [("l", "r")],
        13: [("b", "r")],
        14: [("l", "b")],
    }

    def edge_key(edge, i, j):
        if edge == "t":
            return ("h", i, j)
        if edge == "b":
            return ("h", i + 1, j)
        if edge == "l":
            return ("v", i, j)
        return ("v", i, j + 1)

    def edge_pt(key):
        kind, i, j = key
        return (j + 0.5, i) if kind == "h" else (j, i + 0.5)

    segments = []
    touch = {}
    ys, xs = np.nonzero((case != 0) & (case != 15))
    for i, j in zip(ys.tolist(), xs.tolist()):
        for ea, eb in table[int(case[i, j])]:
            ka, kb = edge_key(ea, i, j), edge_key(eb, i, j)
            si = len(segments)
            segments.append((ka, kb))
            touch.setdefault(ka, []).append(si)
            touch.setdefault(kb, []).append(si)
    used = [False] * len(segments)
    loops = []
    for s0 in range(len(segments)):
        if used[s0]:
            continue
        loop = [edge_pt(segments[s0][0])]
        cur, cur_key = s0, segments[s0][1]
        used[s0] = True
        while cur_key != segments[s0][0]:
            nxt = [s for s in touch[cur_key] if not used[s]]
            if not nxt:
                break
            cur = nxt[0]
            used[cur] = True
            a, b = segments[cur]
            nxt_key = b if a == cur_key else a
            loop.append(edge_pt(nxt_key))
            cur_key = nxt_key
        loops.append(loop)
    return loops


def _dp_open(pts, tol):
    """Douglas-Peucker kept indices for an open polyline."""
    keep = {0, len(pts) - 1}
    stack = [(0, len(pts) - 1)]
    while stack:
        a, b = stack.pop()
        if b - a < 2:
            continue
        ab = pts[b] - pts[a]
        denom = np.linalg.norm(ab)
        if denom == 0:
            continue
        d = np.abs(np.cross(ab, pts[a + 1 : b] - pts[a])) / denom
        if len(d) == 0:
            continue
        k = int(np.argmax(d)) + a + 1
        if d.max() > tol:
            keep.add(k)
            stack.append((a, k))
            stack.append((k, b))
    return keep


def simplify(loop, tol):
    """Douglas-Peucker on a closed loop (split at the farthest point:
    a naive open pass sees the repeated start point as a zero-length
    chord and keeps nothing)."""
    pts = np.array(loop, dtype=float)
    if len(pts) > 1 and np.allclose(pts[0], pts[-1]):
        pts = pts[:-1]
    if len(pts) <= 2:
        return [tuple(p) for p in pts]
    k = int(np.argmax(np.linalg.norm(pts - pts[0], axis=1)))
    keep = set(_dp_open(pts[: k + 1], tol))
    seg2 = np.vstack([pts[k:], pts[0:1]])
    for i in _dp_open(seg2, tol):
        keep.add(0 if i == len(seg2) - 1 else k + i)
    return [tuple(pts[i]) for i in sorted(keep)]


def smooth_closed(pts, tension=0.5):
    """Closed polygon -> cubic Bezier path (Catmull-Rom, tension knob).

    The curve passes through every simplified point (no shrinkage).
    Tension scales the tangent vectors: 1.0 = full rounding, 0.0 =
    straight lines, 0.5 = half rounding.
    """
    n = len(pts)
    if n < 3:
        return "M " + " L ".join(f"{x:.2f} {y:.2f}" for x, y in pts) + " Z"
    p = [np.array(q, dtype=float) for q in pts]
    d = f"M {p[0][0]:.2f} {p[0][1]:.2f}"
    for i in range(n):
        p0, p1, p2, p3 = p[(i - 1) % n], p[i], p[(i + 1) % n], p[(i + 2) % n]
        c1 = p1 + tension * (p2 - p0) / 6.0
        c2 = p2 - tension * (p3 - p1) / 6.0
        d += f" C {c1[0]:.2f} {c1[1]:.2f}, {c2[0]:.2f} {c2[1]:.2f}, {p2[0]:.2f} {p2[1]:.2f}"
    return d + " Z"


# ---------------------------------------------------------------- build
def validate_doc(doc, exercise_id):
    """Fail the build on a document the stills cannot pose."""
    if doc.get("schema") != "vigor-exercise-animation/1":
        raise ValueError(f"{exercise_id}: unexpected schema {doc.get('schema')!r}")
    if doc.get("rig") != "vrmc-core-22":
        raise ValueError(f"{exercise_id}: unexpected rig {doc.get('rig')!r}")
    missing = [name for name in BONE_MAP if name not in doc.get("bones", {})]
    if missing:
        raise ValueError(f"{exercise_id}: missing bone tracks {missing}")
    count = len(doc["root"]["keys"])
    for name, track in doc["bones"].items():
        if len(track["keys"]) != count:
            raise ValueError(f"{exercise_id}: bone {name} has mismatched key count")
    if doc.get("prop") not in KNOWN_PROPS:
        raise ValueError(f"{exercise_id}: unknown prop {doc.get('prop')!r}")


def render_still(doc, rig, prims, kept, joint_nodes, ibm):
    """One still SVG + stats for a document. Returns (svg, points, fill)."""
    frame = peak_frame(doc)
    cam = compute_camera(rig, doc)
    mats = pose_matrices(rig, doc, frame)
    name_of = rig["name_of"]
    joint_pos = {name_of[i]: mats[i][:3, 3] for i in range(len(name_of))}
    tris_screen = []
    for prim, kept_tris in zip(prims, kept):
        sverts = skin(prim, joint_nodes, ibm, mats)
        scr, dep = project(sverts, cam)
        for tri in kept_tris:
            if (dep[tri] <= 0.01).any():
                continue
            tris_screen.append(scr[tri] * SCALE)
    for tri in prop_tris(doc["prop"], joint_pos):
        w = np.array(tri)
        scr, dep = project(w, cam)
        if (dep <= 0.01).any():
            continue
        tris_screen.append(scr * SCALE)
    mask = rasterize(tris_screen)
    fill = float(mask.mean()) * 100.0
    closed = binary_closing(mask)
    parts = []
    total_pts = 0
    for loop in trace_mask(closed):
        if len(loop) < 12:
            continue
        simp = simplify(loop, 1.0)
        pts = [(x / SCALE, y / SCALE) for x, y in simp]
        total_pts += len(pts)
        parts.append(smooth_closed(pts))
    if not parts:
        raise ValueError(f"{doc.get('exercise')}: empty silhouette")
    path_data = " ".join(parts)
    svg = (
        f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {VIEW:.0f} {VIEW:.0f}">\n'
        f"  <title>{doc['displayName']}</title>\n"
        f'  <path d="{path_data}" fill="#000" fill-rule="evenodd" />\n</svg>\n'
    )
    return svg, frame, total_pts, fill


def main():
    """Build one SVG still per animation document."""
    anim_dir = Path(sys.argv[1]) if len(sys.argv) > 1 else HERE / "animations-out"
    out_dir = Path(sys.argv[2]) if len(sys.argv) > 2 else HERE / "stills-out"
    glb_path = Path(sys.argv[3]) if len(sys.argv) > 3 else HERE / "quaternius-mannequin.glb"
    gltf, bin_data = read_glb(glb_path)
    glb_json = load_glb(str(glb_path))
    prims, joint_nodes, ibm = load_mesh(gltf, bin_data)
    rig = prepare_rig(glb_json)
    finger_joints = {
        j for j, node in enumerate(joint_nodes) if is_finger_joint(glb_json["nodes"][node].get("name", ""))
    }
    toe_joints = {j for j, node in enumerate(joint_nodes) if glb_json["nodes"][node].get("name", "") in TOE_NAMES}
    print(f"finger joints: {len(finger_joints)}, toe joints: {len(toe_joints)}")
    kept = [strip_kept(p, finger_joints, toe_joints) for p in prims]
    for n, (prim, kept_tris) in enumerate(zip(prims, kept)):
        print(f"prim {n}: {len(prim['idx'])} tris -> {len(kept_tris)} after strip")
    # Rest-skin self-check: skinning at rest must reproduce bind positions.
    rest_mats = {i: compose(*rig["rest_map"][rig["name_of"][i]]) for i in range(len(glb_json["nodes"]))}
    for n, prim in enumerate(prims):
        err = float(np.abs(skin(prim, joint_nodes, ibm, rest_mats) - prim["pos"]).max())
        print(f"prim {n}: rest-skin max error {err:.2e}")
        if err > 1e-3:
            raise ValueError(f"prim {n}: rest-skin check failed ({err:.2e})")
    docs = sorted(anim_dir.glob("*.json"))
    if not docs:
        raise ValueError(f"no animation documents in {anim_dir}")
    out_dir.mkdir(parents=True, exist_ok=True)
    for doc_path in docs:
        doc = json.loads(doc_path.read_text())
        exercise_id = doc.get("exercise") or doc_path.stem
        if not ID_PATTERN.match(exercise_id):
            raise ValueError(f"unexpected exercise id {exercise_id!r}")
        validate_doc(doc, exercise_id)
        svg, frame, points, fill = render_still(doc, rig, prims, kept, joint_nodes, ibm)
        (out_dir / f"{exercise_id}.svg").write_text(svg)
        print(f"{exercise_id}: frame {frame}, {points} path points, fill {fill:.1f}%")
    print(f"wrote {len(docs)} stills to {out_dir}")


if __name__ == "__main__":
    main()
