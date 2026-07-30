---
name: agent-code-review
description: >
  Adversarial verification layer for reviewing agent-generated or otherwise untrusted
  implementations. Use when the author of a diff, patch, commit, or test suite is an agent,
  when a change touches tests or verification infrastructure, or when invoked as
  /agent-code-review. Adds test falsifiability checks and check-gaming detection on top of
  ordinary review; does not replace it.
---

# Agent Code Review

Treat the change as an untrusted implementation candidate. Readable code, passing tests,
high coverage, and the author's confidence are not evidence of correctness. Coverage
percentage is not evidence by itself.

Do not ask only:

> Does this code look correct?

Also ask:

> What evidence would fail if this implementation were wrong?

## Recover the contract first

Identify intended behavior, acceptance criteria, and behavior that must remain unchanged.
If the intended behavior is unclear, that is the first finding.

Do not invent a stronger contract than the task provides. A finding against a requirement
the task never stated is noise.

## Test the tests

Invoke the `test-the-tests` skill for this step: verify each test that carries weight by
actually mutating the implementation (invert a condition, remove validation, suppress an
error, skip a side effect, revert to the prior behavior) and confirming the test fails.
A suite that cannot detect a plausible fault is weak even when it passes — passing and
high coverage are not evidence by themselves.

## Check-gaming

Flag changes that:

- weaken or delete assertions;
- update golden files without explaining the behavioral change;
- skip, quarantine, or filter tests;
- lower quality thresholds;
- add test-only branches or detect CI/test environments;
- hardcode fixture-specific answers;
- rely entirely on mocks for critical behavior;
- modify verification scripts together with the implementation;
- present logs or reports without proving they belong to the reviewed commit.

Unexplained changes to tests or verification infrastructure are high risk by default.

## Inspection depth

Direct inspection of the implementation is mandatory regardless of green checks when the
change touches auth, secrets, crypto, billing, data deletion, migrations, concurrency,
release or supply-chain logic, public APIs, persistent formats, or anything hard to roll
back.

For low-risk mechanical changes with strong evidence, targeted inspection is enough. No
line-by-line narration in either case.

## Know this review's limits when you are the author

If you wrote the diff you are now reviewing, this pass is necessary but not sufficient.
Self-review — even done adversarially, even by mutating your own tests — inherits your own
design assumptions; it reliably catches localized bugs and weak tests, but is a poor
substitute for a second opinion on whether the overall approach or architecture is sound.
An independently-invoked reviewer with no memory of your rationale (a fresh subagent, or an
external tool/model such as Codex) will more reliably surface issues you can't see because
you already believe your own premises.

For anything more than a mechanical or low-risk change, escalate to an independently
invoked reviewer before merging — don't treat your own adversarial pass as the final word.
Say so explicitly in your findings when you are the diff's author and no second reviewer
has looked yet, so the person deciding whether to merge knows that gap exists.

## Output

Report only actionable findings. Each one: where the problem is visible, what fails and
under which conditions, and the smallest concrete correction or missing proof. Style
preferences are not defects.

Passing checks are claims. The review's job is to establish whether those checks are
capable of disproving the implementation.
