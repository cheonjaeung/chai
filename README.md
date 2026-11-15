# Chai

**Chai** ([tɕʰa.i]) is an interactive diff tool.

In the AI era, writing code is faster than before, but checking it has become harder to keep up with.
Many developers still prefer using Git from the command line, but they often turn to a GUI or a rich TUI to view diffs.
Chai aims to make viewing diffs fast and convenient without leaving the terminal.

> [!IMPORTANT]
> This project is under development and not ready for production use.
> External contributions are not accepted at this time.

## Usage

### CLI

You can use Chai as a CLI.
Pass two files to Chai.

```sh
chai FILE1 FILE2
```

### Git Integration

Chai can be used as a pager and diff filter for Git.
Add this to your `~/.gitconfig`.

```gitconfig
[core]
    pager = chai

[interactive]
    diffFilter = chai --git-diff-filter
```

## About the Name

The name **Chai** comes from the Korean word **"차이"**, meaning **"difference"**.
It is pronounced **[tɕʰa.i]**, similar to **"cha-ee"**.
