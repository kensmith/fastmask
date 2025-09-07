include $(include-dir)/dir.mk
all-includes := $(wildcard $(mydir)/*.mk)
all-excludes := all.mk dir.mk # this file and its dependencies
$(foreach .exclude,$(all-excludes), \
  $(eval all-includes := $(call filter-out, %$(.exclude),$(all-includes))) \
 )
$(foreach mk,$(all-includes), \
  $(eval include $(mk)) \
 )
