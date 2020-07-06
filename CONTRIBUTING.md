# Contributing

<!-- toc -->
-   [Getting started](#getting-started)
-   [Workflow](#workflow)
    -   [Creating Pull Requests](#creating-pull-requests)
    -   [Code Review](#code-review)
    -   [Testing](#testing)
<!-- /toc -->

# Getting started

- Fork the repository on GitHub
- Read the [Develop Octopus](./docs/octopus/develop.md) and [Develop Adaptors](./docs/adaptors/develop.md) for build instructions.

# Workflow

Asking a question, reporting a bug or sending pull requests is welcome. 

A workflow of asking a question or reporting a bug looks like:

- Create an issue with a clear direction title.
- Select the kind of the issue from the labels on the right side.
- Fill in and describe the situation in detail.
- As the discussion progresses, the kind of the issue may change or require more detailed descriptions. Long-term unresponsive and unreproducible issues will be closed, a [bot](https://github.com/probot/stale) will take care of these abandoned issues.

A workflow of sending pull request looks like:

- Create a corresponding issue as an admission ticket and associate it with this PR.
- Make commits of logical units, unit tests may be asked.
- Submit clear commit messages.
- The PR must receive approval from at least two maintainers.

## Creating pull requests

Octopus follows the standard [github pull request](https://help.github.com/articles/about-pull-requests/) process. To submit a proposed change, please develop the code/fix and add new test cases. After that, run these local verifications before submitting pull request to predict the pass or fail of continuous integration:

- Developing Adaptor:
    + Run and pass `make adaptors <adaptor directory name> test`
    + Run and pass `CROSS=true make adaptors <adaptor directory name> package only`
- Developing Octopus:
    + Run and pass `make octopus test`
    + Run and pass `CROSS=true make octopus package only`

## Code review

To make it easier for your PR to receive reviews, consider the reviewers will need you to:

- Follow [good coding guidelines](https://github.com/golang/go/wiki/CodeReviewComments).
- Write [good commit messages](https://chris.beams.io/posts/git-commit/).
- Break large changes into a logical series of smaller patches which individually make easily understandable changes, and in aggregate solve a broader issue.
- [Label](https://help.github.com/en/github/managing-your-work-on-github/applying-labels-to-issues-and-pull-requests) PR with the scope or stage label.
- [Request](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/requesting-a-pull-request-review) appropriate reviewers to review.

### Format of the commit message

The commit message of Octopus follows [Conventional Commits](https://www.conventionalcommits.org/), which provides an easy set of rules for creating an explicit commit history, it should be structured as follows:

```text
<type>[optional scope]: <description>
<BLANK LINE>
[optional body]
<BLANK LINE>
[optional footer]
```

With Conventional Commits, Octopus could generate the changelog automatically and navigate git history easily. An example of commit message:

```text
fix(dummy): correct simulation interval

adjust the interval to relate with `gear` spec.

address #7
```
More examples could refer to [here](https://www.conventionalcommits.org/en/v1.0.0-beta.4/#examples). 

#### Rules of commit message

All commit messages should follow the rules as below:

1. Separate subject from body with a blank line, and separate footer from body with another blank line.
1. Subject line should not end with a period and limit to 70 characters.
1. Use the body to describe what and why or how, and should wrap it at 80 characters.
1. All content should be lowercase except `BREAKING CHANGE`.
1. Use the imperative mood to describe all things: "change" not "changed" nor "changes".

#### Detail of message subject

Subject line must include `type` and `scope` as shown below:

- Allowed `type` values:
    + **fix**, a bug fix.
    + **feat**, a new feature.
    + **chore**, updating the code dependencies.
    + **docs**, changes to the documentation.
    + **test**, changes to add tests, refactoring tests.
    + **ci**, changes to CI and scripts.
    + **style**, changes that do not affect the meaning of the code(e.g. white-space, formatting, missing semi-colons).
    + **perf**, changes to improve the performance.
    + **refactor**, changes that neither fixes a bug nor adds a feature(e.g. renaming a variable or extracting a method).
    + **revert**, changes that revert previous commits.
- Example `scope` values:
    + **brain**/**limb**/**suctioncup**, the related content of main framework.
    + **dummy**, the related content of [dummy](./adaptors/dummy) adaptor, the same is true of other adaptors.

The `scope` can be empty, if the change is a global or difficult to assign to a single component, for example:

```text
docs: correct spelling of CHANGELOG
```

#### Detail of message body

The body should include the motivation for the change and contrasts with previous behavior, for more information about body message, please view: [https://www.freecodecamp.org/news/writing-good-commit-messages-a-practical-guide/](https://www.freecodecamp.org/news/writing-good-commit-messages-a-practical-guide/)

#### Detail of message footer

The footer should be used for:

- Referencing issues:
    Closed issues should be listed on a separate line prefixed with `address` keyword like this:
  
    ```text
    address #7
    ``` 
    
    or in the case of multiple issues:
    
    ```text
    address #8, #9, #10
    ```
- Record breaking changes:
    Breaking changes should be prefixed with `BREAKING CHANGE` keyword like this:
    
    ```text
    BREAKING CHANGE: bump Go version to 1.13
    ```
  
    or use multiple lines to mention with the description of the change, justification and migration notes:
    
    ```text
    BREAKING CHANGE:
    
    `enable-metrics` option of `limb` has dropped, therefore, monitoring metrics are enabled by default.
    
    To migrate your project, change all command, where you use `--enable-metrics`.
    ```

## Testing

There are multiple types of tests. The location of the test code varies with type, as do the specifics of the environment needed to successfully run the test:

- Unit: These confirm that a particular function behaves as intended. Unit test source code can be found adjacent to the corresponding source code within a given package.
- Integration: These tests cover interactions of package components or interactions between Octopus components and Kubernetes control plane components like API server. Integration test source code can be found in `test/integration` directory.
- End-to-end ("e2e"): These are broad tests of overall system behavior and coherence. E2E test source code can be found in `test/e2e` directory.

Continuous integration will run these tests on PRs.

### Unit testing

Usually, the unit tests easily run locally by any developer. Developed code can be PR after passing unit tests. You can run unit tests on:

- Adaptor: `make adaptors <adaptor directory name> test only`
- Octopus: `make octopus test only`

### Integration testing

Integration testing bases on [the envtest of sigs.k8s.io/controller-runtime](https://book.kubebuilder.io/reference/testing/envtest.html), it's using [Ginkgo](http://onsi.github.io/ginkgo/), a testing framework which supports [Behavior-Driven Development(BDD)](https://en.wikipedia.org/wiki/Behavior-driven_development) style. You can run integration tests on:

- Adaptor: `make adaptors <adaptor directory name> verify only`
- Octopus: `make octopus verify only`

When running integration tests, the [framework](./test/framework) will launch a local Kubernetes cluster using Docker. There are two supported local clusters inside: `kind` and `k3d`, you can use environment variable `CLUSTER_TYPE` to select, default is `k3d`. Instead of setting up a local cluster, you can also use environment variable `USE_EXISTING_CLUSTER=true` to point out an existing cluster, and then the integration tests will use the kubeconfig of the current environment to communicate with the existing cluster.

### E2E testing

TODO

### Documentation

Currently, Octopus hosted its documentation on the [cnrancher/docs-octopus](https://github.com/cnrancher/docs-octopus) repo, please submit your PR against this repo for documentation changes.
