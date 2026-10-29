# Generate Exercise Animations at Knowledge Build Time

## Status

Proposed. The repository owner ratifies this decision.

## Context and Problem Statement

Thirty-six exercise animations exist as evaluated parametric
recipes. The animations must reach the product without
inflating the repository and without drifting from the
exercise catalog they illustrate. The generated documents are
large as text (about 124 KB each formatted) and are rewritten
as a whole whenever the generator changes, so committing them
makes the history grow by the full set on every regeneration.
The catalog is 1,489 exercises: committed documents would
reach about 189 MB in the tree at full coverage. The knowledge
image already ships self-consistent: its build bakes the
catalog into a database dump from committed source files.

## Considered Options

1. Commit the generated documents next to the catalog data.
   The history grows by the full document set on every
   pool-wide regeneration, and the tree reaches about
   189 MB at full coverage.
2. Store the documents outside the repository (object
   storage or a content delivery service) and fetch them at
   runtime. The catalog and its animations version apart, a
   second deployable surface appears, and rollback no longer
   restores a consistent pair.
3. Commit the recipes, the mapping, and the model under
   `knowledge/recipes/`. The knowledge image build runs the
   generator, validates the documents, and the bootstrap
   bakes them into the database next to the catalog.

## Decision Outcome

Chosen option: option 3.

The knowledge build gains a pinned Python generation stage
(Python 3.12.12, numpy 1.26.4, scipy 1.17.1). Generation is
deterministic: the vendored generator reproduces the
evaluated animations exactly. Validation inside the generator
fails the image build on any malformed document, and the
bootstrap fails on a document that does not parse, so an
inconsistent catalog-animation pair cannot ship.

Justification: option 1 trades a small source tree for an
unbounded history. Option 2 breaks the one property the
knowledge image already guarantees, a self-consistent baked
catalog, and adds an operational surface. Option 3 keeps the
repository at source size, keeps the pair consistent by
construction, and reuses the build and bootstrap machinery
the catalog already uses.

## Consequences

- Positive: the repository carries kilobytes of recipes
  instead of megabytes of generated documents, at any
  coverage.
- Positive: catalog and animations version, ship, and roll
  back as one image.
- Positive: a broken recipe fails the image build, loudly,
  before publication.
- Negative: the knowledge build gains a Python stage and
  about twenty seconds of generation at the current
  coverage.
- Neutral: the application does not read the animations yet.
  Backend exposure and the player are a later cycle; this
  decision fixes where the data is born and how it ships.
