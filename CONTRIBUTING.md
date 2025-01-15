# Contributing

windows_exporter uses GitHub to manage reviews of pull requests.

* If you are a new contributor see: [Steps to Contribute](#steps-to-contribute)

* If you have a trivial fix or improvement, go ahead and create a pull request,
  addressing (with `@...`) a suitable maintainer of this repository (see
  [MAINTAINERS.md](MAINTAINERS.md)) in the description of the pull request.

* If you plan to do something more involved, first discuss your ideas
  as [github issue](https://github.com/prometheus-community/windows_exporter/issues).
  This will avoid unnecessary work and surely give you and us a good deal
  of inspiration. New collectors are unlikely to be accepted, since the
  `performancecounter` collector is the preferred way to collect metrics.

* Relevant coding style guidelines are the [Go Code Review
  Comments](https://code.google.com/p/go-wiki/wiki/CodeReviewComments)
  and the _Formatting and style_ section of Peter Bourgon's [Go: Best
  Practices for Production
  Environments](https://peter.bourgon.org/go-in-production/#formatting-and-style).
  gofmt and [golangci-lint](https://github.com/golangci/golangci-lint) are your friends.

* Be sure to sign off on the [DCO](https://github.com/probot/dco#how-it-works).

## Steps to Contribute

Should you wish to work on an issue, please claim it first by commenting on the GitHub issue that you want to work on it. This is to prevent duplicated efforts from contributors on the same issue.
For quickly compiling and testing your changes do:

```bash
# For building.
go build -o windows_exporter.exe ./cmd/windows_exporter/
./windows_exporter.exe

# For testing.
make test         # Make sure all the tests pass before you commit and push :)
```

To run a collection of Go linters through [`golangci-lint`](https://github.com/golangci/golangci-lint), do:
```bash
make lint
```

If it reports an issue and you think that the warning needs to be disregarded or is a false-positive, you can add a special comment `//nolint:linter1[,linter2,...]` before the offending line. Use this sparingly though, fixing the code to comply with the linter's recommendation is in general the preferred course of action. See [this section of the golangci-lint documentation](https://golangci-lint.run/usage/false-positives/#nolint-directive) for more information.

All our issues are regularly tagged so that you can also filter down the issues involving the components you want to work on. For our labeling policy refer [the wiki page](https://github.com/prometheus/prometheus/wiki/Label-Names-and-Descriptions).

## Pull Request Checklist

* Branch from the main branch and, if needed, rebase to the current main branch before submitting your pull request. If it doesn't merge cleanly with main you may be asked to rebase your changes.

* Commits should be as small as possible, while ensuring that each commit is correct independently (i.e., each commit should compile and pass tests).

* The PR title should be of the format: `subsystem: what this PR does` (for example, `cpu: Add support for thing` or `docs: fix typo`).

* If your patch is not getting reviewed or you need a specific person to review it, you can @-reply a reviewer asking for a review in the pull request or a comment.

* Add tests relevant to the fixed bug or new feature.

## Dependency management

The Prometheus project uses [Go modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more) to manage dependencies on external packages.

To add or update a new dependency, use the `go get` command:

```bash
# Pick the latest tagged release.
go get example.com/some/module/pkg@latest

# Pick a specific version.
go get example.com/some/module/pkg@vX.Y.Z
```

Tidy up the `go.mod` and `go.sum` files:

```bash
# The GO111MODULE variable can be omitted when the code isn't located in GOPATH.
GO111MODULE=on go mod tidy
```

You have to commit the changes to `go.mod` and `go.sum` before submitting the pull request.
