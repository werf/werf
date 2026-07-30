---
name: test-the-tests
description: Verify a test actually falsifies the behavior it claims to cover, via real mutation. Use right after writing or changing a test, when auditing an existing suite's real coverage, or as a step inside a code review.
---

# Test the Tests

Passing and high coverage are not evidence, whoever wrote the test — including you, just
now. A test proves it runs; it does not prove it discriminates correct from incorrect
behavior. Ask:

> What fault would this test fail to catch?

## The mutation loop

For each test that carries real weight (guards a fix, a regression, an invariant — not a
trivial getter):

1. Pick the smallest plausible fault: invert a condition, change `<` to `<=`, remove a
   validation, suppress an error, skip a side effect, hardcode a return value, revert to
   the prior (buggy) behavior.
2. Apply it to the implementation, run the test, confirm it fails for the right reason —
   not an unrelated crash.
3. Revert immediately, confirm the tree is clean (`git status`/`git diff`), confirm the
   suite passes again. NEVER leave mutated code in the repository between steps.

If running the mutation isn't practical, name the smallest mutation that should be tried
instead of skipping the exercise — that name is itself the finding.

## Common ways a test looks strong but isn't

- **The assertion holds under the bug too.** A chain assertion like
  `index(a) < index(b) < index(c)` can pass under both the fix and the regression it's
  meant to catch if the scenario doesn't force them to disagree. Reshape the scenario (add
  a case whose correct answer differs from what the bug would produce) instead of trusting
  a single run.
- **It re-asserts a second recording of the same event** instead of a property true by
  construction — comparing two independently-recorded orderings is racier and weaker than
  asserting a monotonic counter or a guaranteed dependency order.
- **It exercises the unit in isolation**, never the real wiring a regression would actually
  break. Prefer driving the real entry point end-to-end when the risk is in the wiring.
- **It asserts implementation trivia** (mock called N times, internal helper ran) instead
  of externally observable behavior.

## Output

For each test verified this way: the fault tried, whether it failed as expected, and if
not, the smallest change that would make it discriminate. For tests not mutated, name the
smallest fault that should be tried next — don't assert confidence without it.
