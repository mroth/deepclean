# deepclean

> be disrespectful to filesystem dirt :sparkles::wastebasket::sparkles:

Currently looks for the following:
 - `.bundle` (Ruby Bundler)
 - `node_modules` (NodeJS NPM)
 - `target` (Rust Cargo, Scala SBT)

However it's easy to overrride this list.

deepclean will take advantage of multiple cores on your machine by gathering
statistics for matched directories in parallel.

TODO: only recommend these for deletion if they are `.gitignore`'d, not tracked
in git.

## Installation

Grab a compiled binary from the [releases][1] page, or on macOS with [homebrew][2]
you can `brew install mroth/formulas/deepclean`.

[1]: https://github.com/mroth/deepclean/releases
[2]: https://brew.sh

## Usage

    Usage: deepclean [options] [dir]

    Options:
      -sort
          sort output
      -target string
          dirs to scan for (default "node_modules,.bundle,target")

Will scan the current directory or `dir` if provided.

## Errata

### Why not `find`?

It is possible to do something similar with a monster shell command.

```bash
find . \( \
      -name 'node_modules' \
      -o -name '.bundle' \
      -o -name 'target' \
    \)  -prune \
    -exec sh -c 'echo "$(find "$0" | wc -l)\t$(du -sh "$0")"' {} \;
```

On my machine that takes about ~3.5sec total. In contrast deepclean is ~670ms.
I'm on a fairly fast machine[*] and don't have a super large src dir. I imagine
that these numbers should scale similarly on very large directories or slower
disks.

[*]: 8-core Xeon, 2xSSD array in RAID-0.