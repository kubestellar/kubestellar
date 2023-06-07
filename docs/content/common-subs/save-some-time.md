<!--save-some-time-start-->
!!! tip "This document is 'docs-ecutable' - you can 'run' this document, just like we do in our testing, on your local environment"
    ```
    git clone -n -b {{config.ks_branch}} {{config.repo_url}} --depth 1 {{config.site_name}}-{{page.meta.short_name}}
    cd {{config.site_name}}-{{page.meta.short_name}}
    git restore --staged Makefile Makefile.venv go.mod docs
    git checkout Makefile Makefile.venv go.mod docs
    make MANIFEST="'{{page.meta.pre_req_name}}','{{page.meta.manifest_name}}'" run-doc-shells
    ```

    ```
    # done? remove everything  
    make MANIFEST="content/common-subs/remove-all.md" run-doc-shells
    cd ../
    rm -rf {{config.site_name}}-{{page.meta.short_name}}
    ```
<!--save-some-time-end-->