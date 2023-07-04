<!--save-some-time-start-->
!!! tip "This document is 'docs-ecutable' - you can 'run' this document, just like we do in our testing, on your local environment"
    ```
    git clone -b {{config.ks_branch}} {{config.repo_url}}
    cd {{config.repo_default_file_path}}
    make MANIFEST="'{{page.meta.pre_req_name}}','{{page.meta.manifest_name}}'" docs-ecutable
    ```

    ```
    # done? remove everything
    make MANIFEST="docs/content/common-subs/remove-all.md" docs-ecutable
    cd ..
    rm -rf {{config.repo_default_file_path}}
    ```
<!--save-some-time-end-->
