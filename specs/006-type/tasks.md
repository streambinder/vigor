# Cycle 006 tasks

- [ ] T000 Cycle 005 convergence bookkeeping: observe SC-004
      (36 rows in the production knowledge database) and check
      cycle 005 T008 in its tasks file.
- [x] T001 Knowledge: `type` and `flags` fields on every
      catalog exercise. Formal lists `types.json` and
      `flags.json` with canonical IDs and aliases.
      Rule classification plus a curated plyometric list.
      The `mode` field is checked first. Bootstrap seeds
      the vocabularies and rejects missing or unknown
      values. Type and Flags lines in the embedded
      exercise text.
- [x] T002 Backend request: `Type` and `Flags` on
      `PostTrainingRequest` (JSON fields `type` and
      `flags`). Thread them through `GenerateTraining`
      at the same rank as `methodology`. Unknown value: 400. Absent Type: `mixed`. Absent Flags: no
      flag filter.
- [x] T003 Backend derivation and retrieval: free-text
      derivation of Type and Flags in
      `DeriveFreeTextParams`, with the formal allowed
      values in the derivation prompt. An explicit choice
      stays authoritative. Type and Flag filter in
      `retrieveBySimilarity` for the work pool only.
      Type and Flags shown for each candidate in the
      selection prompt, with the rule as reinforcement.
- [x] T004 Backend availability: endpoint returning the
      available Types with counts for an equipment list.
      Validation 400 listing available Types. Empty
      filtered pool rejected with 400.
- [x] T005 App: Type selector in the generation modal
      next to the equipment mode, showing only available
      Types, default Mixed, chosen value sent in the
      `type` field. No Flags selector this cycle.
- [x] T006 Tests: retrieval admits only the requested
      Type and Flags. `mixed` and absent Type behave as
      today. Derivation and explicit-choice precedence.
      Bootstrap rejection. Availability endpoint returns
      only servable Types. App request carries the chosen
      Type. Integration harness: one trajectory per Type
      value, plus default, plus one with the
      `plyometric` Flag.
- [ ] T007 Device verification by Davide: one generated
      workout per Type value. Work exercises match the
      requested modality. CI green job by job, including
      the app jobs.

## Convergence notes

- Implementation ships in two commits: docs, then code.
  Every file appears in exactly one commit.
- T000 stays open: the production knowledge database is
  not readable from the implementation machine, so the
  36-row observation and the cycle 005 T008 check remain
  for Davide.
- T007 stays open: device verification belongs to Davide.
