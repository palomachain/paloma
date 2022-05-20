# Contributing

Thank you for considering making contributions to Paloma. We welcome all messenger birds! 

Contributing to this repo can mean many things such as participating in
discussions or proposing new features, improvements or bug fixes. To ensure a
smooth and timely workflow for all contributors, the general procedure for
contributing has been established:

1. If you would like to contribute, first do your best to check if discussions
   already exist as either a Github  [Issue](https://github.com/palomachain/paloma/issues) or
   [PR](https://github.com/palomachain/paloma/pulls). Be sure to also check out
   our public [Telegram Channel](https://t.me/palomachain). Existing discussions will help you
   gain context on the current status of the proposed contribution or topic. If
   one does not exist, feel free to start one.
2. If you would like to create a [Github Issue](https://github.com/palomachain/paloma/issues),
   either [open](https://github.com/palomachain/paloma/issues/new/choose) or
   [find](https://github.com/palomachain/paloma/issues) an issue you'd like to
   help with. If the issue already exists, attempt to participate in thoughtful
   discussion on that issue.
3. If you would like to contribute:
   1. If the issue is a proposal, ensure that the proposal has been discussed
      and accepted.
   2. Ensure that nobody else has already begun working on this issue. If they
      have, make sure to contact them to potentially collaborate.
   3. If nobody has been assigned for the issue and you would like to work on it,
      make a comment on the issue to inform the community of your intentions to
      begin work.
   4. Follow standard GitHub best practices, i.e. fork the repo, branch from the
      HEAD of `main`, make commits, and submit a PR to `main`
      - For core developers working within the repo, to ensure a clear ownership
        of branches, branches must be named with the convention `{moniker}/{issue#}-branch-name`.
   5. Be sure to submit the PR in `Draft` mode. Submit your PR early, even if
      it's incomplete as this indicates to the community you're working on
      something and allows them to provide comments early in the development
      process.
   6. When the code is complete it can be marked `Ready for Review` and follow
      the PR readiness checklist.

## Coding Guidelines

We follow the Cosmos SDK [Coding Guidelines](https://github.com/cosmos/cosmos-sdk/blob/main/CODING_GUIDELINES.md). Specifically:

- API & Design SHOULD be proposed and reviewed before the main implementaion.
- Minimize code duplication
- Define [Acceptance tests](https://github.com/cosmos/cosmos-sdk/blob/main/CODING_GUIDELINES.md#acceptance-tests) or while implementing new features.
  - Prefer use of acceptance test framework, like [gocuke](https://github.com/regen-network/gocuke/)
  - For unit tests or integration tests use [go mock](https://github.com/golang/mock) for creating mocks. Generate mock interface implementations using `go generate`.

## Architecture Decision Records (ADR) or EPICS

When proposing an architecture decision for Paloma, please start by
opening an [Issue](https://github.com/palomachain/paloma/issues/new/choose) or a
[Discussion](https://github.com/palomachain/paloma/discussions/new) with a summary
of the proposal. 

Once the proposal has been discussed and there is rough alignment on a high-level
approach to the design, the [ADR creation process](https://github.com/umee-network/umee/blob/main/docs/architecture/PROCESS.md) can begin. We are following this process to ensure all involved parties
are in agreement before any party begins coding the proposed implementation.

If you would like to see examples of how these are written, please refer to the
current [ADRs](https://github.com/umee-network/umee/tree/main/docs/architecture).

## Branching Model

The Paloma repos adheres to the [trunk based development branching](https://trunkbaseddevelopment.com/)
model and utilizes [semantic versioning](https://semver.org/).

### PR Targeting

Ensure that you base and target your PR against the `main` branch.

All feature additions should be targeted against `main`. Bug fixes for an
outstanding release candidate should be targeted against the release candidate
branch.

### PR & Merge Procedure

- Ensure the PR branch is rebased on `main`.
- Ensure you provided unit tests and integration tests.
- Run `make test-unit test-e2e` to ensure that all tests pass.
- Merge the PR!

## Release procedure

We follow semver (from v3.0.0):

- major version update (eg 2.x.x -> 3.0.0) has API breaking changes or signals major feature update
- minor version update (eg 2.1.x -> 2.2.0) has no API nor state machine breaking changes. It can provide new functionality or bug fixes.
- patch version update (eg 2.1.0 -> 2.1.1) has no API nor state machine breaking changes nor new features. It only contains hot fixes.

### Major Release Procedure

All major changes related to major version update are first released for testnet. We use `-betaX` (eg `2.0.0-beta1`, `2.0.0-beta2` ...) releases for testnet. Once the code is stabilized we create a release candidate (eg `2.0.0-rc1`). If no issues are found the latest release candidate become the major release.