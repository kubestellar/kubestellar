# Missing results in a CombinedStatus object

## Description of the Issue

A `CombinedStatus` object, which is specific to one `Binding`
(`BindingPolicy`) and one workload object, lacks an entry in `results`
for some `StatusCollector` whose name is associated with the workload
object by the `Binding`.

## Root Cause

There is no `StatusCollector` object with the name given in the `Binding`.

This could be because of a typo in the `Binding` or because something
failed to create the intended `StatusCollector`.
