# multipr

Create (and update) pull requests en masse.

> [!WARNING]
>
> This is alpha/beta quality software and breaking changes might occur at any
> time!

## Requirements

- Go
- Bash
- sed
- Git
- [GitHub CLI](https://cli.github.com/) (and authenticated)

## Quickstart

```sh
# install
go install github.com/fredrikaverpil/multipr/cmd/multipr@latest
```

```yml
# job.yml

search:
  github:
    method: code
    query: --owner myorg --filename CODEOWNERS "my team"
identify:
  - name: Find daily interval
    cmd: |
      rg --hidden 'interval: daily' -g dependabot.yml
changes:
  - name: Update dependabot schedule
    cmd: sed -i 's/daily/weekly/g' .github/dependabot.yml
pr:
  github:
    branch: multipr/dependabot-interval
    title: "ci(dependabot): update interval"
    body: |
      ## Update dependabot interval
      ### Why?
      It's too noisy with a daily interval.
      ### What?
      - Update schedule to weekly interval.
      ### Notes
      - This PR was generated with [multipr](https://github.com/fredrikaverpil/multipr)
```

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

## Help, troubleshooting

- For search syntax, consult the
  [GitHub CLI `search code` docs](https://cli.github.com/manual/gh_search_code).

## Development and contribution

```sh
# compile and run
go run ./cmd/multipr -job job.yml
```

Contributions are very welcome!

### To-do

- [ ] Tests.
- [ ] CI stuff like linting, formatting, vulnerability checks etc.
- [ ] Potentially detect default branch without requiring `sed`.
