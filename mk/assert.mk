assert = \
$(strip \
  $(if $(strip $(1)), \
    $(comment assertion holds), \
    $(error $(strip $(2))) \
   ) \
 )

nostrip-assert = \
$(strip \
  $(if $(1), \
    $(comment assertion holds), \
    $(error $(strip $(2))) \
   ) \
 )
