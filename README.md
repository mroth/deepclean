# deepclean

> be disrespectful to filesystem dirt :sparkles::wastebasket::sparkles:

I often have a bunch of junk dependency files sitting around in my source folder
that I don't actually need. Periodically I want to clean them up on inactive
projects, as I recently discovered when I wanted to transfer my src dir to a new
computer, and it was taking forever due to over _half a million_ junk files in
`node_modules` directories.

<img width="600" src="docs/demo.svg">

Currently looks for the following:
 - `.bundle` (Ruby Bundler)
 - `node_modules` (NodeJS NPM)
 - `target` (Rust Cargo, Scala SBT)

However it's easy to overrride this list.

Deepclean is very fast -- it will take advantage of multiple cores on your
machine by gathering statistics for matched directories in parallel.

Nothing is actually deleted at the moment due to paranoia, just surfaced in the
UI so the user can decide on their own how to handle.

_TODO: only recommend these for deletion if they are `.gitignore`'d, not tracked
in git._

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