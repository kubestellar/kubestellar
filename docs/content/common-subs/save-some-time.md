<!--save-some-time-start-->
!!! tip "This document is 'docs-ecutable' - you can 'run' this document, just like we do in our testing, on your local environment"
    ```
    git clone -n -b {{config.ks_branch}} {{config.repo_url}} --depth 1 {{config.site_name}}-{{page.meta.short_name}}
    cd {{config.site_name}}-{{page.meta.short_name}}
    git restore --staged Makefile Makefile.venv go.mod docs/mkdocs.yml docs/content docs/scripts/docs-ecutable.sh
    git checkout Makefile Makefile.venv go.mod docs/mkdocs.yml docs/content docs/scripts/docs-ecutable.sh
    make MANIFEST="'{{page.meta.pre_req_name}}','{{page.meta.manifest_name}}'" docs-ecutable
    ```

    ```
    # done? remove everything
    make MANIFEST="docs/content/common-subs/remove-all.md" docs-ecutable
    cd ../
    rm -rf {{config.site_name}}-{{page.meta.short_name}}
    ```
<!--save-some-time-end-->