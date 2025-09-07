mydir = $(strip \
  $(patsubst %/,%, \
    $(dir $(lastword $(MAKEFILE_LIST))) \
   ) \
)
