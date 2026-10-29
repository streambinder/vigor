"""Build animations for every catalog exercise that has a recipe module.

Discovery, not a registry: for each exercise in exercises.json the
module name is the exercise id with hyphens replaced by underscores.
A missing module means the exercise has no recipe, so no animation
is built for it and the consumer falls back to the exercise's
reference media. An import error inside an existing recipe module
is a real failure and propagates, never a silent skip. Recipe
modules and the shared machinery (engine, generics,
emit_frames) live in the recipes/ subdirectory; the model
library (rig) lives beside this file and stays importable
through the script directory.

Usage: build.py <mannequin.json> <outdir> [exercise-id ...]
"""

import importlib
import json
import sys
from pathlib import Path

import numpy as np

sys.path.insert(0, str(Path(__file__).resolve().parent / "recipes"))

CATALOG = Path(__file__).resolve().parent.parent / "exercises.json"


def main():
    # engine lives in recipes/, so it is imported lazily here,
    # after the path insertion above (pylint allows lazy imports)
    import engine as E

    # sys.argv[1] is the mannequin path, consumed by engine at import time
    outdir = sys.argv[2]
    wanted = set(sys.argv[3:])
    catalog = json.loads(CATALOG.read_text(encoding="utf-8"))
    out = Path(outdir)
    out.mkdir(parents=True, exist_ok=True)
    ph = np.linspace(0, 1, E.NPHASE, endpoint=False)
    built, skipped = [], []
    for row in catalog:
        ex_id = row["id"]
        if wanted and ex_id not in wanted:
            continue
        mod_name = ex_id.replace("-", "_")
        if not mod_name.isidentifier():
            skipped.append(ex_id)
            continue
        try:
            module = importlib.import_module(mod_name)
        except ModuleNotFoundError as err:
            if err.name != mod_name:
                raise
            skipped.append(ex_id)
            continue
        points, hip_frac, prop = module.build(ph)
        anim = E.finish(points, hip_frac, str(out / f"{ex_id}.json"), ex_id, prop, f"recipe pool v1: {ex_id}")
        anim["displayName"] = row["name"]
        (out / f"{ex_id}.json").write_text(json.dumps(anim, indent=1), encoding="utf-8")
        built.append(ex_id)
    print(f"built {len(built)} animations, {len(skipped)} exercises without a recipe")
    return 0


if __name__ == "__main__":
    sys.exit(main())
