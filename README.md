# deepclean

> be disrespectful to filesystem dirt :sparkles::wastebasket::sparkles:

Currently looks for the following:
 - .bundle (Ruby)
 - node_modules (NPM)
 - target (Rust)

TODO: only recommend these for deletion if they are `.gitignore`'d, not tracked in git.

deepclean will take advantage of multiple cores on your machine.


### Why not `find`?

It is possible to do something similar with a monster shell command.

```shell
find . \( \
      -name 'node_modules' \
      -o -name '.bundle' \
      -o -name 'target' \
    \)  -prune \
    -exec sh -c 'echo "$(find "$0" | wc -l)\t$(du -sh "$0")"' {} \;
```

On my machine that takes about 3.2 seconds total.  deepclean is ~670ms.
