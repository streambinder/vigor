# Exercise animation recipes

This directory is the source of the exercise animations. The
animations are parametric recipes, not recorded data: each recipe
builds the movement from canonical anchor poses and solves the
mannequin joints with inverse kinematics, phase by phase.

## Generation

`generate.py` runs the pool builder, validates every built
document, and writes it minified to the output directory,
named after the exercise identifier. The pool build stages in
a temporary directory, never in the repository. Validation
rejects a wrong schema or rig, a bone set that differs from
the rig, unsorted or out-of-range key phases, non-unit
quaternions, an unknown prop, and an exercise identifier that
is not in the catalog. A failed validation fails the build.

The builder set lives under `recipes/`: `engine.py` holds the
shared construction machinery (rest-pose constants, the rigid
trunk frame, the phase solver, the document emitter),
`emit_frames.py` turns solved positions into the animation
document, one module per exercise holds that exercise's
builder, named after the exercise identifier with hyphens
replaced by underscores, and the family builders shared by
several exercises live in the generic modules (`standing.py`,
`hinge.py`, `lying.py`, `hanging.py`, `upright.py`). The
orchestration (`build.py`, `generate.py`) and the model
library (`rig.py`) live beside `recipes/`.
`build.py` discovers builders from `exercises.json`:
an exercise without a recipe module produces no document,
and the consumer falls back to that exercise's reference
media. `rig.py` holds the shared rig math (model loading,
rest pose, quaternions, IK, swing rotations) and
`emit_frames.py` turns pose frames into animation documents.

Generated documents are build artifacts. Do not commit them and do
not edit them: change the recipes and regenerate.
The knowledge image build runs the generator and the bootstrap
injects the output into the `exercise_animations` table, so the
catalog and its animations ship in the same image.

## Model and format

The mannequin is the Quaternius model, rigged with the
`vrmc-core-22` bone set. Generation reads only the node
skeleton, so the repository commits that skeleton as
`mannequin.json`, extracted from the model file; the mesh is
never used. Documents follow the
`vigor-exercise-animation/1` schema: per-bone local quaternions
relative to the rest pose over a normalized loop phase, plus a
root track with the pelvis orientation and height. Documents
carry exactly the 22 rig bones: hands are not rendered, so no
finger tracks are emitted.

## Requirements

Python 3.12 with numpy and scipy. The versions pinned in the
knowledge Dockerfile reproduce the evaluated animations exactly.
