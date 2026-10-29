"""Generate the exercise animation catalog from the parametric recipes.

Runs the pool builder, validates every built document, and writes
it minified to the output directory under the exercise identifier.
Only exercises with a recipe module produce a document; the rest
fall back to their reference media at consumption time. The pool
build stages in a temporary directory, never in the repository.

Usage: generate.py [output-dir] [catalog-json]

Defaults: output-dir is ./animations-out, catalog-json is
../exercises.json. The process exits nonzero when any
animation fails validation, so a broken recipe fails the build.
"""

import json
import math
import subprocess
import sys
import tempfile
from pathlib import Path

from rig import BONE_MAP

HERE = Path(__file__).resolve().parent
PROPS = {"bar", "dumbbell", "kb", "chinbar", "latbar", "dipbars", "none"}


def run(*args):
    subprocess.run([sys.executable, *args], cwd=HERE, check=True)


def check_keys(keys, label, errors, hip=False):
    phases = [k["p"] for k in keys]
    if phases != sorted(phases) or phases[0] < 0.0 or phases[-1] > 1.0:
        errors.append(f"{label}: key phases not sorted inside 0..1")
    for k in keys:
        q = k["q"]
        if len(q) != 4 or not 0.99 <= math.sqrt(sum(c * c for c in q)) <= 1.01:
            errors.append(f"{label}: quaternion not unit length")
            return
        if hip and not 0.05 <= k["hipHeight"] <= 1.6:
            errors.append(f"{label}: hipHeight outside the emitter clip range")
            return


def validate(anim, exercise_id, catalog_ids, bone_names):
    errors = []
    if anim.get("schema") != "vigor-exercise-animation/1":
        errors.append(f"{exercise_id}: unexpected schema {anim.get('schema')!r}")
    if anim.get("rig") != "vrmc-core-22":
        errors.append(f"{exercise_id}: unexpected rig {anim.get('rig')!r}")
    if anim.get("loop") is not True:
        errors.append(f"{exercise_id}: animation does not loop")
    if anim.get("exercise") != exercise_id:
        errors.append(f"{exercise_id}: exercise field is {anim.get('exercise')!r}")
    if anim.get("prop") not in PROPS:
        errors.append(f"{exercise_id}: unexpected prop {anim.get('prop')!r}")
    if exercise_id not in catalog_ids:
        errors.append(f"{exercise_id}: {exercise_id} is not a catalog exercise")
    bones = anim.get("bones", {})
    if set(bones) != bone_names:
        errors.append(f"{exercise_id}: bone set differs from the rig bone map")
    for name, track in bones.items():
        check_keys(track["keys"], f"{exercise_id}/{name}", errors)
    check_keys(anim["root"]["keys"], f"{exercise_id}/root", errors, hip=True)
    return errors


def main():
    outdir = Path(sys.argv[1]) if len(sys.argv) > 1 else HERE / "animations-out"
    catalog = Path(sys.argv[2]) if len(sys.argv) > 2 else HERE.parent / "exercises.json"
    bone_names = set(BONE_MAP)
    catalog_ids = {row["id"] for row in json.loads(catalog.read_text(encoding="utf-8"))}
    outdir.mkdir(parents=True, exist_ok=True)
    errors = []
    with tempfile.TemporaryDirectory() as tmp:
        pool = Path(tmp) / "pool"
        run("build.py", "mannequin.json", str(pool))
        for path in sorted(pool.glob("*.json")):
            exercise_id = path.stem
            anim = json.loads(path.read_text(encoding="utf-8"))
            errors.extend(validate(anim, exercise_id, catalog_ids, bone_names))
            payload = json.dumps(anim, separators=(",", ":"))
            (outdir / f"{exercise_id}.json").write_text(payload, encoding="utf-8")
    if errors:
        for error in errors:
            print(f"INVALID {error}")
        return 1
    print(f"generated {len(list(outdir.glob('*.json')))} animations in {outdir}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
