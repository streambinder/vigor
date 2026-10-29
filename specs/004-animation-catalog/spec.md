# Feature Specification: Animation Catalog Generation

**Feature Branch**: `feat/animation-catalog`

**Created**: 2026-10-27

**Status**: Draft

**Input**: Owner direction: the evaluated exercise animations
exist as generated JSON documents outside the repository. The
owner asked whether committing those documents scales to the
catalog size, then directed the form this cycle takes: the
recipes live in `knowledge/`, the knowledge image build runs
the generation, and the image stays self-consistent, as it
already is for the catalog data.

## Clarifications

### Investigation 2026-10-27 (hypothesis challenge)

The premise was checked against the generator output and the
knowledge build before any design.

- Committing the generated documents as they come out of the
  generator does not scale. One document averages 124 KB of
  formatted text. At full catalog coverage (1,489 exercises)
  the tree would carry about 189 MB, and every pool-wide
  regeneration rewrites every document, so the repository
  history grows by that amount per regeneration round.
- The information content is small. The curves are smooth and
  densely sampled: the pool compresses to about 6 KB per
  animation, and minified emission is about 56 KB per document,
  2.0 MB for the current 36.
- The knowledge image already ships self-consistent: the build
  bakes the catalog into a database dump from the committed
  feature files. Generating the animations in the same build
  follows the existing pattern and adds no deployable unit, no
  external storage, and no version pair that can drift.
- The recipes are analytic. The two standalone generators load
  reference measurements and never read them; the recipes build
  from canonical anchors alone. Generation needs the mannequin
  skeleton and numpy with scipy, nothing else: no media files, no
  machine learning runtime.
- The generator output is deterministic. The vendored
  generator, after lint normalization, reproduces the evaluated
  pool byte for byte under canonical comparison, for all 36
  animations.
- Coverage today is 36 evaluated recipes. Each recipe module
  is named after its catalog exercise, so no mapping file
  exists: the builder set is the animation set, and an
  exercise without a recipe module produces no document and
  falls back to its reference media at consumption time.
  Two recipes are nearest-pattern approximations: the leg
  press recipe covers the sled press record whose name
  carries a side-view suffix, and the hanging knee raise
  recipe covers the assisted variant, because the catalog
  holds no unassisted record. Both stay flagged for
  editorial review.
- Delivery to the application is out of scope. The renderer
  decision is still open, so no player ships in this cycle.
  This cycle makes the animations exist, versioned with the
  catalog, inside the knowledge database. Backend exposure
  and the player follow in a later cycle.

## User Scenarios and Testing _(mandatory)_

### User Story 1 - One image carries catalog and animations (Priority: P1)

The operator deploys a knowledge image. The image was built
from one commit: the catalog files and the animation recipes
come from that commit, and the database inside the image holds
both. A rollback to an earlier image restores a consistent
pair, never a catalog from one version with animations from
another.

**Why this priority**: consistency of the pair is the reason
the owner chose generation at build time over committed
documents and over external storage.

**Independent Test**: build the knowledge image from the
branch, start a database from it, and count the animation rows
per exercise in the baked data.

**Acceptance Scenarios**:

1. **Given** the branch recipes, **When** the
   knowledge image build runs, **Then** the build generates
   and validates 36 animation documents and the baked database
   holds one animation row for each exercise with a recipe.
2. **Given** a recipe change that produces an invalid
   document, **When** the image build runs, **Then** the
   build fails before any image is published.
3. **Given** a checkout without the generated documents,
   **When** the bootstrap runs, **Then** the catalog injects
   as before and the animation step skips without an error.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: The recipes and the mannequin skeleton are
  committed under `knowledge/features/animations/`.
  Generated animation documents are never committed.
- **FR-002**: The knowledge image build runs the generator.
  The generator validates every document: schema and rig
  identifiers, the bone set of the rig, sorted key phases
  inside the loop range, unit quaternions, the prop
  vocabulary, and membership of the exercise identifier in
  the catalog. Any violation fails the build.
- **FR-003**: The bootstrap injects the generated documents
  into an `exercise_animations` table, one row per exercise,
  keyed by the exercise identifier. A missing animations
  directory skips the step. A malformed document fails the
  bootstrap.
- **FR-004**: Generation is deterministic under the pinned
  toolchain (Python 3.12.12, numpy 1.26.4, scipy 1.17.1).
  The same recipes produce the same documents on every build.
- **FR-005**: The change is additive. Existing catalog tables,
  records, and consumers keep their current behavior. No
  component reads the new table in this cycle.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: The knowledge tests cover the ingestion paths
  (directory present, directory missing, malformed document,
  store error) and the knowledge coverage gate holds.
- **SC-002**: A generator run produces 36 documents that all
  pass validation, and their canonical hashes equal the
  evaluated pool hashes for all 36 animations.
- **SC-003**: The knowledge image builds in CI with the
  generation stage, and the full Go test suite, super-linter,
  gofumpt, and golangci-lint pass on the branch.
- **SC-004**: After the rollout, the production knowledge
  database holds 36 animation rows keyed by catalog exercise
  identifiers, observed on the deployed image.

## Assumptions

- The renderer and the application player are a later cycle.
  This cycle ends at the knowledge database.
- The two approximate recipes (leg press, hanging
  knee raise) are accepted for the pipeline rollout and stay
  open for editorial correction.
- The mannequin skeleton, extracted from the model file, is
  committed as a build input; the model binary is not.
  Its licence terms were reviewed during the animation
  research as permitting embedding in a product, not
  standalone redistribution; a final check against the
  current terms precedes any public distribution decision.
