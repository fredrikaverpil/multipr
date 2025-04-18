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
      rg --hidden 'interval: daily' -g dependabot.yml
changes:
  - name: Update dependabot schedule
    shell: bash # optional
    cmd: sed -i 's/daily/weekly/g' .github/dependabot.yml
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
```

> [!NOTE]
>
> - For search syntax, consult the
>   [GitHub CLI `gh search code` docs](https://cli.github.com/manual/gh_search_code).
> - For repository search syntax, consult the
>   > [GitHub CLI `gh search repos` docs](https://cli.github.com/manual/gh_search_repos).
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
> Certain features such as `-draft` is not supported in private GitHub
> repositories when on a free personal plan. Read more
> [here](https://docs.github.com/en/get-started/learning-about-github/githubs-plans).

## How `multipr` works

1. A user-defined GitHub `gh search code` query is the base for cloning down git
   repositories to local disk.
1. Git repositories are cloned down into a `$(pwd)/jobs` folder.
1. A user-defined local identification phase (using e.g. `rg`) decides which of
   the cloned down repositories are eligible for modification (exit code 0 means
   eligible).
1. For each eligible repository:
   - Fetch all, reset hard and checkout the default branch.
   - Check out a new user-defined branch.
   - Perform code changes via user-defined shell commands.
   - Create user-defined git commit.
   - Create (or edit existing) pull request via `gh pr [create|edit]`.

## Development and contribution

```sh
# compile and run
go run ./cmd/multipr -job job.yml
```

Contributions are very welcome!

## Similar projects

- [mani](https://github.com/alajmo/mani)
- [remote-find-replace](https://github.com/einride/remote-find-replace)
