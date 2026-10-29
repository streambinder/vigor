"""burpee recipe."""

import numpy as np
from engine import (
    ANKLE_H,
    ANKLE_X,
    D_HAND,
    D_HIP,
    FA,
    FLAT,
    PELVIS_Y,
    SHO_HALF,
    TRUNK_LEN,
    UA,
    assemble,
    trunk_R,
)


def build(ph):
    """Keyframed on TARGETS, solved per frame.

    The previous version interpolated the solved joint positions of
    the knots; mid-lerp the knee/foot points swung through the floor
    (ankle 30 cm under ground in the plank transition, caught by the
    autonomous evaluator). Here every frame lerps the assemble()
    inputs (pelvis, trunk direction, ankle/hand targets, ball
    offsets, poles) and solves IK fresh, like every other recipe.
    """
    sides = (("l", 1.0), ("r", -1.0))
    d_kpole = np.array([0.0, 0.15, 1.0])
    d_epole = np.array([0.0, -1.0, -0.3])
    up = np.array([0.0, 1.0, 0.0])

    def K(pelvis, dir_t, ankles, ball_off, hands, kpole, epole):
        return {
            "pelvis": np.array(pelvis, float),
            "dir_t": np.array(dir_t, float),
            "ankles": {s: np.array([sgn * ankles[0], ankles[1], ankles[2]]) for s, sgn in sides},
            "ball_off": np.array(ball_off, float),
            "hands": {s: np.array([sgn * hands[0], hands[1], hands[2]]) for s, sgn in sides},
            "kpole": np.array(kpole, float),
            "epole": np.array(epole, float),
            "epole_side": None,
        }

    def K_ep(pelvis, dir_t, ankles, ball_off, hands, kpole, epole_l):
        k = K(pelvis, dir_t, ankles, ball_off, hands, kpole, epole_l)
        k["epole_side"] = True
        return k

    stand_ank = (ANKLE_X, ANKLE_H, 0.0)
    hang_y = PELVIS_Y * 0.985 + TRUNK_LEN - UA - FA + 0.02
    knots = [
        (
            0.00,
            K(
                (0.0, PELVIS_Y * 0.985, 0.0),
                up,
                stand_ank,
                FLAT,
                (SHO_HALF + 0.04, hang_y, 0.02),
                d_kpole,
                d_epole,
            ),
        ),
        (
            0.16,
            K(
                (0.0, 0.46, -0.28),
                (0.0, np.cos(np.radians(40)), np.sin(np.radians(40))),
                stand_ank,
                FLAT,
                (0.24, 0.12, 0.40),
                d_kpole,
                d_epole,
            ),
        ),
        (
            0.30,
            K(
                (0.0, 0.44, -0.30),
                (0.0, np.cos(np.radians(55)), np.sin(np.radians(55))),
                stand_ank,
                FLAT,
                (0.24, 0.03, 0.38),
                d_kpole,
                d_epole,
            ),
        ),
        # airborne kick-back: feet already back and off the floor,
        # pelvis still high -- without this knot the lerp folds the
        # legs under the body and the knees touch down mid-transition
        (
            0.40,
            K(
                (0.0, 0.45, -0.16),
                (0.0, np.cos(np.radians(78)), np.sin(np.radians(78))),
                (ANKLE_X, 0.24, -0.84),
                (0.0, -0.13, -0.10),
                (0.24, 0.03, 0.38),
                d_kpole,
                d_epole,
            ),
        ),
        (0.44, K_ep(*_virt_knot(0.0), (0.0, -1.0, 0.0), (0.3, 0.44, 0.0))),
        (0.54, K_ep(*_virt_knot(1.0), (0.0, -1.0, 0.0), (0.3, 0.44, 0.0))),
        (
            0.61,
            K(
                (0.0, 0.45, -0.16),
                (0.0, np.cos(np.radians(78)), np.sin(np.radians(78))),
                (ANKLE_X, 0.24, -0.84),
                (0.0, -0.13, -0.10),
                (0.24, 0.03, 0.38),
                d_kpole,
                d_epole,
            ),
        ),
        (
            0.68,
            K(
                (0.0, 0.44, -0.30),
                (0.0, np.cos(np.radians(55)), np.sin(np.radians(55))),
                stand_ank,
                FLAT,
                (0.24, 0.03, 0.38),
                d_kpole,
                d_epole,
            ),
        ),
        (
            0.78,
            K(
                (0.0, 1.16, 0.0),
                up,
                (ANKLE_X, 0.50, -0.05),
                (0.0, -0.145, 0.09),
                (0.30, 2.02, 0.0),
                d_kpole,
                (0.0, -1.0, 0.2),
            ),
        ),
        (
            0.88,
            K(
                (0.0, 1.16, 0.0),
                up,
                (ANKLE_X, 0.50, -0.05),
                (0.0, -0.145, 0.09),
                (0.30, 2.02, 0.0),
                d_kpole,
                (0.0, -1.0, 0.2),
            ),
        ),
        (
            1.00,
            K(
                (0.0, PELVIS_Y * 0.985, 0.0),
                up,
                stand_ank,
                FLAT,
                (SHO_HALF + 0.04, hang_y, 0.02),
                d_kpole,
                d_epole,
            ),
        ),
    ]
    points, hip_frac = [], []
    for f in range(len(ph)):
        p = ph[f]
        for i in range(len(knots) - 1):
            p0, k0 = knots[i]
            p1, k1 = knots[i + 1]
            if p0 <= p <= p1:
                s = 0.0 if p1 == p0 else (p - p0) / (p1 - p0)
                s = s * s * (3 - 2 * s)

                def mix(a, b, s=s):
                    return a + (b - a) * s

                pelvis = mix(k0["pelvis"], k1["pelvis"])
                dir_t = mix(k0["dir_t"], k1["dir_t"])
                dir_t = dir_t / np.linalg.norm(dir_t)
                ankles = {ss: mix(k0["ankles"][ss], k1["ankles"][ss]) for ss in ("l", "r")}
                lift = 0.0
                if i in (2, 3, 5, 6):
                    # kick-back / return: the feet travel through the air;
                    # a floor-level ankle path drags the toes under ground
                    s_raw = 0.0 if p1 == p0 else (p - p0) / (p1 - p0)
                    lift = np.sin(np.pi * s_raw) * 0.08
                    ankles = {ss: ankles[ss] + np.array([0.0, lift, 0.0]) for ss in ("l", "r")}
                ball_off = mix(k0["ball_off"], k1["ball_off"])
                balls = {ss: ankles[ss] + ball_off + np.array([0.0, lift * 1.6, 0.0]) for ss in ("l", "r")}
                hands = {ss: mix(k0["hands"][ss], k1["hands"][ss]) for ss in ("l", "r")}
                kpole = mix(k0["kpole"], k1["kpole"])
                ep0, ep1 = k0["epole"], k1["epole"]
                elbow_poles = {}
                for ss, sgn in sides:
                    e0 = np.array([ep0[0] * sgn, ep0[1], ep0[2]]) if k0["epole_side"] else ep0
                    e1 = np.array([ep1[0] * sgn, ep1[1], ep1[2]]) if k1["epole_side"] else ep1
                    elbow_poles[ss] = mix(e0, e1)
                points.append(
                    assemble(
                        pelvis,
                        dir_t,
                        ankles,
                        hands,
                        balls=balls,
                        knee_poles={ss: kpole for ss in ("l", "r")},
                        elbow_poles=elbow_poles,
                    )
                )
                hip_frac.append(pelvis[1] / PELVIS_Y)
                break
    return points, np.array(hip_frac), "none"


def pushup_targets(d):
    """pushup_pose's inputs as burpee knot values (x, y, z forms)."""
    pelvis = np.array([0.0, 0.3833 - 0.0809 * d, -0.07])
    pitch = 0.3012 - 0.3022 * d
    dir_t = np.array([0.0, np.sin(pitch), np.cos(pitch)])
    ank = (0.10, 0.12, -0.8935 - 0.0562 * d)
    hands = (0.25, 0.03, 0.3037 + 0.0298 * d)
    return pelvis, dir_t, ank, (0.0, -0.095, -0.135), hands


def _virt_knot(d):
    """pushup_targets converted to assemble()'s virtual-target mode
    (the page-solved anchors minus the true-joint offsets, which
    assemble adds back)."""

    pelvis, dir_t, ank, ball_off, hands = pushup_targets(d)
    R = trunk_R(dir_t)
    dh, dp = R @ D_HAND["l"], R @ D_HIP["l"]
    ank_v = (ank[0], ank[1] - dp[1], ank[2] - dp[2])
    hands_v = (hands[0], hands[1] - dh[1], hands[2] - dh[2])
    return pelvis, dir_t, ank_v, ball_off, hands_v
