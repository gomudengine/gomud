# Contributing to GoMud

Want to add code to GoMud? First off, thank you! Contributions from you or people like you gives GoMud the opportunity to be the absolute best it can.

There aren't a ton of rules, but here are some good guidelines to follow when suggesting changes, additions, or reviewing code.

## Quicklinks

- [Contributing to GoMud](#contributing-to-gomud)
  - [Quicklinks](#quicklinks)
    - [Pull Requests](#pull-requests)
    - [Pull Request Reviews](#pull-request-reviews)
  - [Getting Help](#getting-help)

### Pull Requests

If you have planned changes that will have a wide-spread impact, break current functionality meaningfully, or change core ways the code operates, make sure it is discussed fully and agreed upon as an actionable item in an issue discussion.

In general, PRs should:

- Attempt to limit the size and scope of the code change. Grouping a number of small changes is fine if complexity is also small. 
- Include unit tests for any packages you are adding or changing.
  - Unit tests should use [testify](https://github.com/stretchr/testify). Do not auto generate/commit mocks with the mockery tool.
- Include reasonable code documentation as well as update relevant markdown documentation.
- A PR template will automatically pre-populate your Pull Request. Fill in the fields requested, as they are required.
  - Additional sections can be added to a PR, such as screenshots, or other named sections of your choosing. Include these after the required sections.

### Pull Request Reviews

Some general guidelines when reviewing PR's:

1. Do not review PR's marked as DRAFT's.
2. Reviewers should be clear about the reason for requesting a change if it is contrary to the current implementation.
3. Try to be clear about whether a comment on code is intended as a suggestion vs. a change request. In fact, prefixing with `SUGGESTION:` or `CONSIDER:` makes this very clear that it is not a blocking issue.
4. Be clear, direct, even terse - but be respectful. Disrespectful or confrontational reviews are a quick way to end up excluded.

## Getting Help

Unsure about something? Create a discussion thread about it, or join the [Discord Server](https://discord.gg/cjukKvQWyy) for live discussion about features, code, and other questions you may have.
