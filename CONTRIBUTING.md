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
- As the discussion progresses, the kind of the issue may change or require more detailed descriptions. Long-term unresponsive and unreproducible issues will be closed.

A workflow of sending pull request looks like:

- Create a corresponding issue as an admission ticket and associate it with this PR.
- Make commits of logical units, unit tests may be asked.
- Submit clear commit messages.
- The PR must receive approval from at least two maintainers.

## Creating pull requests

Octopus follows the standard [github pull request](https://help.github.com/articles/about-pull-requests/) process. To submit a proposed change, please develop the code/fix and add new test cases. After that, run these local verifications before submitting pull request to predict the pass or fail of continuous integration:

- Developing Adaptor:
    + Run and pass `make adaptors <adaptor directory name> package`
    + Run and pass `make adaptors <adaptor directory name> test only`
- Developing Octopus:
    + Run and pass `make octopus package`
    + Run and pass `make octopus test only`

## Code review

To make it easier for your PR to receive reviews, consider the reviewers will need you to:

- Follow [good coding guidelines](https://github.com/golang/go/wiki/CodeReviewComments).
- Write [good commit messages](https://chris.beams.io/posts/git-commit/).
- Break large changes into a logical series of smaller patches which individually make easily understandable changes, and in aggregate solve a broader issue.
- [Label](https://help.github.com/en/github/managing-your-work-on-github/applying-labels-to-issues-and-pull-requests) PR with the scope or stage label.
- [Request](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/requesting-a-pull-request-review) appropriate reviewers to review.

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

When running integration tests, the [framework](./test/framework) will launch a local Kubernetes cluster using Docker. There are two supported local clusters inside: `kind` and `k3d`, you can use environment variable `LOCAL_CLUSTER_KIND` to select, default is `k3d`. Instead of setting up a local cluster, you can also use environment variable `USE_EXISTING_CLUSTER=true` to point out an existing cluster, and then the integration tests will use the kubeconfig of the current environment to communicate with the existing cluster.

### E2E testing

TODO

