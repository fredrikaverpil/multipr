# multipr

Create (and update) pull requests en masse.

> [!WARNING]
>
> - This is alpha/beta quality software and breaking changes might occur at any
>   time!
> - I've developed and used `multipr` on macOS only, so your mileage may vary on
>   other platforms. Open an issue if you encounter any problems! 😅✌️

## Requirements

`multipr` will run sub-commands in the shell. It will internally call the
following, which are therefore expected to exist locally:

- `bash` (can be configured with CLI argument `-shell`)
- `git`
- `gh` (authenticated [GitHub CLI](https://cli.github.com/))

## Quickstart

Download the `multipr` binary from the
[releases](https://github.com/fredrikaverpil/multipr/releases) or install it
with Go:

```sh
# install
go install github.com/fredrikaverpil/multipr/cmd/multipr@latest
```

Create a `job.yml` file:

```yml
# job.yml

search:
  github:
    method: code # methods available: code | repos
    query: --owner myorg --filename CODEOWNERS "my team"
identify:
  - name: Find daily interval
    shell: bash # optional
    cmd: |
      rg --hidden 'interval: daily' .github --glob 'dependabot.yml'
changes:
  - name: Update dependabot schedule
    shell: bash # optional
    cmd: |
      sed -i 's/daily/weekly/g' .github/dependabot.yml
pr:
  github:
    branch: multipr/dependabot-interval
    title: "ci(dependabot): update interval"
    body: |
      ## Update dependabot interval
      ### Why?
      It's too noisy with a daily interval
      ### What?
      - Update schedule to weekly interval
      ### Notes
      - This PR was generated with [multipr](https://github.com/fredrikaverpil/multipr)
      - Job yaml:
        {yaml}
```

> [!NOTE]
>
> - For search syntax, consult the
>   [GitHub CLI `gh search` docs](https://cli.github.com/manual/gh_search).
>   Search methods supported are `code` and `repos`.
> - The `shell` command field is optional, can be set to some other shell on a
>   per-command basis and defaults to `bash` (or whatever you specify with CLI
>   argument `-shell`). Will execute like `<shell> -c <cmd>`.
> - On macOS, the default `sed` implementation is BSD-based and incompatible
>   with GNU `sed` syntax used in the examples. You can install GNU `sed` with
>   `brew install gnu-sed` and use it as `gsed` in your commands.

Run `multipr`:

```sh
# carefully perform manual review of each step
multipr -job job.yml -review

# dry-run without manual review
multipr -job job.yml

# publish draft PRs
multipr -job job.yml -publish -draft

# update PRs, making them ready for review
multipr -job job.yml -publish
```

## Usage

```text
Usage of multipr:
  -clean
        Remove cloned repositories before run
  -debug
        Print all commands and their output
  -draft
        Make PRs into drafts
  -help
        Show help
  -job string
        Path to the YAML job file (required)
  -manual-commit
        User manages git commits in shell commands
  -publish
        Publish PRs
  -review
        Manual review of each major step
  -shell string
        Shell to use for executing commands (default "bash")
  -show-diffs
        Show each git diff (default true)
  -skip-search
        Skip search for repositories
  -workers int
        Number of workers to use for concurrency (default: 2x CPU cores)
```

> [!NOTE]
>
> Certain features are not supported in private GitHub repositories when on a
> free personal plan. Read more
> [here](https://docs.github.com/en/get-started/learning-about-github/githubs-plans).

## How `multipr` works

1. A user-defined GitHub `gh search` query is the base for cloning down git
   repositories to local disk. Git repositories are cloned down into a
   `$(pwd)/jobs` folder.
1. A user-defined local identification phase (using e.g. `find` or `rg`) decides
   which of the cloned down repositories are fully eligible for modification
   (exit code 0 means eligible). This phase exists because it may not always be
   possible to achieve this via `gh search`.
1. For each eligible repository:
   - Fetch all, reset hard and checkout the default branch.
   - Check out a new user-defined branch.
   - Perform code changes via user-defined shell commands.
   - Create user-defined git commit.
   - Create (or edit existing) pull request via `gh pr [create|edit]`.

## Commands for identifying and replacing file contents

> [!NOTE]
>
> - All examples expect GNU `sed` (`brew install gnu-sed` for macOS). If using
>   macOS BSD `sed`, you must pass an empty string to `sed`, like: `sed -i ''`
> - Arguments like `-print0` and `-0` caters for null-delimiting filenames to
>   avoid issues where file names may contain spaces, tabs or even newlines.
> - Paths to directories can generally be substituted with `.` to search from
>   each cloned down repo's root.
>
> Use regex (basic regex by default; GNU sed supports `-E` for extended regex):
>
> - GNU (Linux): `sed -E -i 's/foo[0-9]+/bar/g' file`
> - BSD/macOS: `sed -E -i '' 's/foo[0-9]+/bar/g' file`

```sh
# Spawns one `sed` command per file found by `find`
find ./path/to/dir -type f -name '*.ext' -print0 | xargs -0 sed -i 's/SEARCH/REPLACE/g'

# Spawns one `find` command, which executes only one big `sed` command
find ./path/to/dir -type f -name "*.ext" -exec sed -i 's/SEARCH/REPLACE/g' {} +
```

```sh
# Spawns one `sed` command per file found by ripgrep
rg --files-with-matches --hidden -0 'PATTERN' ./path/to/dir --glob '*.ext' | xargs -0 sed -i 's/SEARCH/REPLACE/g'
```

> [!TIP]
>
> You can execute a script on your machine which performs the search-replace
> (replace the `sed` command with e.g. `python3 script.py`). This is especially
> nice when you need to apply more logic than just replacing a string.

## Development and contribution

```sh
# compile and run
go run ./cmd/multipr -job job.yml
```

Contributions are very welcome! ❤️

## Similar projects

- [mani](https://github.com/alajmo/mani)
- [remote-find-replace](https://github.com/einride/remote-find-replace)
