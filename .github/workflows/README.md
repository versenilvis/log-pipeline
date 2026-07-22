# CI/CD Workflows

This folder contains the GitHub Actions workflows for this project.

## `ci.yml` - Main CI pipeline

Runs on every push to `main` and every pull request.

```
lint -> test -> build-and-push (only on push to main)
```

| Job              | What it does                                                                           | Details                                                                                  |
| ---------------- | -------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `lint`           | Runs `golangci-lint` against the whole codebase                                        | Checks code style and static issues, runs first because it is fastest                    |
| `test`           | Spins up Postgres + Redis as service containers, runs migrations, then `go test ./...` | Real integration tests with service containers in CI -> ensures production behavior      |
| `build-and-push` | Builds Docker images for all 7 services using a matrix strategy, then pushes to GHCR   | Runs only on push to `main` -> pushes built images to `ghcr.io` ready for K8s deployment |

**Note:** `build-and-push` only runs on `push` to `main`, guarded by:
```yaml
if: github.ref == 'refs/heads/main' && github.event_name == 'push'
```
Pull requests only go through `lint` + `test` -> no image build or push for unmerged code.

## `spelling.yml` - Typo checker

Runs [`crate-ci/typos`](https://github.com/crate-ci/typos) on every push/PR to catch spelling mistakes in code, comments, and docs.

```yaml
name: Spelling
permissions:
  contents: read
on:
  push:
    branches:
      - main
  pull_request:
env:
  CLICOLOR: 1
jobs:
  spelling:
    name: Spell Check with Typos
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          persist-credentials: false
      - name: Spell Check Repo
        uses: crate-ci/typos@v1.29.4
```

Runs lightweight and fast -> scans spelling across the entire repository (code, comments, docs).

### False positives - `_typos.toml`

Some words in this project (identifiers, service names, domain-specific terms) look like typos to the checker but are intentional. These are whitelisted in `_typos.toml` at the repo root:

```toml
[default.extend-words]
hel = "hel"
wrk = "wrk"
Wrk = "Wrk"
Brane = "Brane"
HGE = "HGE"
udid = "udid"
OT = "OT"
crypted = "crypted"
uncomplete = "uncomplete"
edn = "edn"
```

If `typos` flags a new word as a false positive, add it to this file using `word = "word"` -> no code edits needed, just whitelist it here.

## Local testing before pushing

Use [`act`](https://github.com/nektos/act) to dry-run workflows locally without pushing:

```bash
act -j lint
act -j test
```

`build-and-push` does not need local testing since it only runs on `main` and requires a real `GITHUB_TOKEN` -> verifying `lint` and `test` locally is sufficient.