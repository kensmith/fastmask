install-go-tool = \
$(strip \
  $(eval .tool := $1) \
  $(eval .url := $($(.tool)-url)) \
  $(call assert,$(.url),missing value for $(.tool)-url) \
  $(eval .target := $(GOBIN)/$(.tool)) \
  $(eval $(.target):;go install $(.url)) \
  $(eval tools:|$(.target)) \
  $(eval $(.tool) := $(.target)) \
 )
