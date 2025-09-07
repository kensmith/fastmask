not = $(if $(strip $(1)),,t)

eq = \
$(strip \
  $(if $(filter $(strip $(1)),$(strip $(2))), \
    $(if $(filter $(strip $(2)),$(strip $(1))), \
      t \
     ) \
   ) \
 )

neq = $(call not,$(call eq,$(1),$(2)))
