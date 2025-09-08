# build targets
* make watch-build
* make watch-fastmask

# install
Fastmask requires go 1.25 or later because it uses the experimental JSON v2
library. (This rewrite is because I wanted to learn how that library works. It
strangely lacks MarshalIndent and the workaround it ... odd.)

GOEXPERIMENT=jsonv2 go install github.com/kensmith/fastmask/fastmask@latest
