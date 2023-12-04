The executables that are scripts are maintained in the following
subdirectories, categorized according to how they are used when the
core components are deployed as Kubernetes workload and without regard
to what contributors running the central controllers as bare processes
do.

- **inner**: scripts used from within the core image and not by users outside the core image.
- **outer**: scripts used by users outside the core image and not inside the core image.
- **overlap**: scripts used both inside the core image and by users outside the core image.
