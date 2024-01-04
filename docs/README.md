<!--readme-for-documentation-start-->
## Overview

Our documentation is powered by [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/) with some 
additional plugins and tools:

- [awesome-pages plugin](https://github.com/lukasgeiter/mkdocs-awesome-pages-plugin)
- [macros plugin](https://mkdocs-macros-plugin.readthedocs.io/en/latest/)
- [mike](https://github.com/jimporter/mike) for multiple version support

We have support in place for multiple languages (i18n), although we currently only have documentation in English. If 
you're interested in contributing translations, please let us know!

### Serving up documents locally
You can view and modify our documentation in your local development environment.  Simply checkout one of our branches.

```shell
git clone git@github.com:{{ config.repo_short_name }}.git
cd {{ config.repo_default_file_path }}/docs
git checkout {{ config.ks_branch }}
```

You can view and modify our documentation in the branch you have checked out by using `mkdocs serve` from [mkdocs](https://www.mkdocs.org).  We have a Python requirements file in `requirements.txt`, and a Makefile target that builds a Python virtual environment and installs the requirements there.  You can either install those requirements into your global Python environment or use the Makefile target.  To install those requirements into your global Python environment, do the following usual thing.

```shell
pip install -r requirements.txt
```

Alternatively, use the following commands to use the Makefile target to construct an adequate virtual environment and enter it.

```shell
( cd ..; make venv )
. venv/bin/activate
```

Then, using your chosen environment with the requirements installed, build and serve the documents with the following command.

```shell
mkdocs serve
```
Then open a browser to [`http://localhost:8000/`](http://localhost:8000/)

Another way to view (not modify - this method reflects what has been deployed to the `gh-pages` branch of our repo) all branches/versions of our documentation locally using 'mike' [mike for mkdocs](https://github.com/jimporter/mike):

```shell
git clone git@github.com:{{ config.repo_short_name }}.git
cd {{ config.repo_default_file_path }}
git checkout {{ config.ks_branch }}
cd docs
mike set-default {{ config.ks_branch }}
cd ..
make serve-docs
```
Then open a browser to [`http://localhost:8000/`](http://localhost:8000/)

## File structure

All documentation-related items live in `docs` (with the small exception of various `make` targets and some helper 
scripts in `hack`).

The structure of `docs` is as follows:

| Path                        | Description                                                                       |
|-----------------------------|-----------------------------------------------------------------------------------|
| config/$language/mkdocs.yml | Language-specific `mkdocs` configuration.                                         |
| content/$language           | Language-specific website content.                                                |
| generated/branch            | All generated content for all languages for the current version.                  |
| generated/branch/$language  | Generated content for a single language. Never added to git.                      |
| generated/branch/index.html | Minimal index for the current version that redirects to the default language (en) |
| overrides                   | Global (not language-specific) content.                                           |
| Dockerfile                  | Builds the kubestellar-docs image containing mkdocs + associated tooling.                 |
| mkdocs.yml                  | Minimal `mkdocs` configuration for `mike` for multi-version support.              |
| requirements.txt            | List of Python modules used to build the site.                                    |

### Global Variables
There are many global variables defined in the <a href="{{ config.repo_raw_url }}/{{ config.ks_branch }}/docs/mkdocs.yml">docs/mkdocs.yml</a>.  The following are some very common variables you are encouraged to use in our documentation.  Use of these variables/macros allows our documentation to have github branch context and take advantage of our evolution without breaking

    - site_name: {{ config.site_name }}
    - repo_url: {{ config.repo_url }}
    - site_url: {{ config.site_url }}
    - repo_default_file_path: {{ config.repo_default_file_path }}
    - repo_short_name: {{ config.repo_short_name }}
    - docs_url: {{ config.docs_url }}
    - repo_raw_url: {{ config.repo_raw_url }}
    - edit_uri: {{ config.edit_uri }}
    - ks_branch: {{ config.ks_branch }}
    - ks_tag: {{ config.ks_tag }}

to use a variables/macro in your documentation reference like this:

\{\{ config.<var_name\> \}\}

and in context that can look something like this:

bash <(curl -s \{\{ config.repo_raw_url \}\}/\{\{ config.ks_branch \}\}/bootstrap/bootstrap-kubestellar.sh) --kubestellar-version \{\{ config.ks_tag \}\}


<b>note:</b><br /> 
&nbsp;&nbsp;&nbsp;&nbsp;- A more extensive and detailed list is located at [mkdocs information](./content/Contribution%20guidelines/operations/all-macros.md) <br />
&nbsp;&nbsp;&nbsp;&nbsp;- We also check for broken links as part of our PR pipeline.  For more information check out our <a href="{{ config.repo_url }}/actions/workflows/broken-links-crawler.yml">Broken Links Crawler</a><br />

### Including external markdown
We make extensive use of 'include-markdown' to help us keep our documentation modular and up-to-date.  To use 'include-markdown' you must add a block in your document that refers to a block in your external document content:

In your original markdown document, add a block that refers to the external markdown you want to include:
![Include Markdown](./content/Contribution%20guidelines/operations/include-markdown-example.png)

In the document you want to include, add the start and end tags you configured in the include-markdown block in your original document:
![Included Markdown](./content/Contribution%20guidelines/operations/included-markdown-example.png)

for more information on the 'include-markdown' plugin for mkdocs look [here](https://github.com/mondeja/mkdocs-include-markdown-plugin)

### Supported aliases for our documentation
We currently support 3 aliases for our documentation:

    - from the release major.minor branch:
        - [{{ config.docs_url }}/stable]({{ config.docs_url }}/stable)
    - from the main branch:
        - [{{ config.docs_url }}/unstable]({{ config.docs_url }}/unstable)
        - [{{ config.docs_url }}/latest]({{ config.docs_url }}/latest)

### Shortcut URLs
We have a few shortcut urls that come in handy when referring others to our project:

<b>note:</b> You need to join our mailing list first to get access to some of the links that follow ([{{ config.docs_url }}/joinus](https://kubestellar.io/joinus))

- [https://kubestellar.io/agenda](https://kubestellar.io/agenda) - our community meeting agenda google doc
- [https://kubestellar.io/blog](https://kubestellar.io/blog) - our medium reading list
- [https://kubestellar.io/code](https://kubestellar.io/code) - our current GitHub repo (wherever that is)
- [https://kubestellar.io/community](https://kubestellar.io/community) - our stable docs community page
- [https://kubestellar.io/drive](https://kubestellar.io/drive) - our google drive
- [https://kubestellar.io/joinus](https://kubestellar.io/joinus) - our dev mailing list where you join and get our invites
- [https://kubestellar.io/join_us](https://kubestellar.io/join_us) - also, our dev mailing list
- [https://kubestellar.io/linkedin](https://kubestellar.io/linkedin) - our linkedin filter (soon, our page)
- [https://kubestellar.io/tv](https://kubestellar.io/tv) - our youtube channel
- [https://kubestellar.io/youtube](https://kubestellar.io/tv) - also, our youtube channel
- [https://kubestellar.io/infomercial](https://kubestellar.io/infomercial) - our infomercial that premieres on June 12th at 9am

and.. the very important…
- [https://kubestellar.io/quickstart](https://kubestellar.io/quickstart) - our 'stable' quickstart

### Codeblocks
mkdocs has some very helpful ways to include blocks of code in a style that makes it clear to our readers that console interaction is necessary in the documentation.  There are options to include a plain codeblock (```), shell (shell), console (console - no used in our documentation), language or format-specific (yaml, etc.), and others.  For more detailed information, checkout the [mkdocs information on codeblocks](https://squidfunk.github.io/mkdocs-material/reference/code-blocks/).

Here are some examples of how we use codeblocks:

- For a codeblock that can be 'tested' (and seen by the reader) as part of our CI, use the <b><i>`shell`</i></b> block:
<br/><b>codeblock:</b>
````
```shell
mkdocs serve
```
````
<b>as seen by reader:</b>
```shell
mkdocs serve
```
<br/>

- For a codeblock that should be 'tested', BUT <b>not</b> seen by the reader, use the <b><i>`.bash`</i></b> with the plain codeblock, and the <b><i>'.hide-me'</i></b> style (great for hiding a sleep command that user does not need to run, but CI does):
<br/><b>codeblock:</b>
````
``` {.bash .hide-me}
sleep 10
```
````
<b>as seen by reader:</b>
```
```
<br/>

- For a codeblock that should <u>not</u> be 'tested' as part of our CI, use the <b><i>`.bash`</i></b> with the plain codeblock, and <b>without</b> the <b><i>'.hide-me'</b></i> style:
<br/><b>codeblock:</b>
````
``` {.bash}
mkdocs server
```
````
<b>as seen by reader:</b>
``` {.bash}
mkdocs server
```
<br/>

- For a codeblock that should not be 'tested', be seen by the reader, and not include a 'copy' icon (great for output-only instances), use the <b><i>`.bash`</i></b> codeblock <b>without</b> the <b><i>'.no-copy'</b></i> style:
<br/><b>codeblock:</b>
```` {.bash .no-copy}
``` {.bash .no-copy}
I0412 15:15:57.867837   94634 shared_informer.go:282] Waiting for caches to sync for placement-translator
I0412 15:15:57.969533   94634 shared_informer.go:289] Caches are synced for placement-translator
I0412 15:15:57.970003   94634 shared_informer.go:282] Waiting for caches to sync for what-resolver
```
````
<b>as seen by reader:</b>
``` {.bash .no-copy}
I0412 15:15:57.867837   94634 shared_informer.go:282] Waiting for caches to sync for placement-translator
I0412 15:15:57.969533   94634 shared_informer.go:289] Caches are synced for placement-translator
I0412 15:15:57.970003   94634 shared_informer.go:282] Waiting for caches to sync for what-resolver
```
<br/>

- For language-specific highlighting (yaml, etc.), use the <b><i>yaml</i></b> codeblock
<br/><b>codeblock:</b>
````
```yaml
nav:
  - Home: index.md
  - QuickStart: Getting-Started/quickstart.md
  - Contributing: 
      - Guidelines: Contribution guidelines/CONTRIBUTING.md
```
````
<b>as seen by reader:</b>
```yaml
nav:
  - Home: index.md
  - QuickStart: Getting-Started/quickstart.md
  - Contributing: 
      - Guidelines: Contribution guidelines/CONTRIBUTING.md
```
<br/>

- For a codeblock that has a title, and will not be tested, use the <b><i>'title'</i></b> parameter in conjunction with the plain codeblock (greater for showing or prescribing contents of files):
<br/><b>codeblock:</b>
````
``` title="testing.sh"
#!/bin/sh
echo hello KubeStellar
```
````
<b>as seen by reader:</b>
``` title="testing.sh"
#!/bin/sh
echo hello KubeStellar
```
<br/>

(other variations are possible, PR an update to the <a href="{{ config.repo_url }}/blob/{{ config.ks_branch }}/docs/overrides/stylesheets/kubestellar.css">kubestellar.css</a> file and, once approved, use the style on the plain codeblock in your documentation.)

### Testing/Running Docs
How do we ensure that our documented examples work?  Simple, we 'execute' our documentation in our CI.  We built automation called 'docs-ecutable' which can be invoked to test any markdown (.md) file in our repository. You could use it in your project as well - afterall it is opensource.

#### The way it works:
- create your .md file as you normally would
- add codeblocks that can be tested, tested but hidden, or not tested at all:
    - use <b><i>'shell'</i></b> to indicate code you want to be tested
    - use <b><i>'.bash'</i></b> with the plain codeblock, and the <b><i>'.hide-md'</i></b> style for code you want to be tested, but hidden from the reader (some like this, but its not cool if you want others to run your instructions without hiccups)
    - use plain codeblock (```) if you want to show sample output that is not to be tested
- you can use 'include-markdown' blocks, and they will also be executed (or not), depending on the codeblock style you use in the included markdown files.

#### The GitHub Workflow:
- One example of the GitHub Workflow is located in our {{ config.repo_short_name }} at <a href="{{ config.repo_url }}/blob/{{ config.ks_branch }}/.github/workflows/docs-ecutable-where-resolver.yml">{{ config.repo_url }}/blob/{{ config.ks_branch }}/.github/workflows/docs-ecutable-where-resolver.yml</a>

#### The secret sauce:
- The code that makes all this possible is at <a href="{{ config.repo_url }}/blob/{{ config.ks_branch }}/docs/scripts/docs-ecutable.sh">{{ config.repo_url }}/blob/{{ config.ks_branch }}/docs/scripts/docs-ecutable.sh</a>
    - This code parses the .md file you give it to pull out all the 'shell' and '.bash .hide-me' blocks
    - The code is smart enough to traverse the include-markdown blocks and include the 'shell' and '.bash .hide-me' blocks in them
    - It then creates a file called 'generate_script.sh' which is then run at the end of the docs-ecutable execution.

All of this is invoke in a target in our <a href="{{ config.repo_url }}/blob/{{config.ks_branch}}/Makefile">Makefile</a>
``` {.bash .no-copy}
.PHONY: docs-ecutable
docs-ecutable: 
	MANIFEST=$(MANIFEST) docs/scripts/docs-ecutable.sh
```

You give the path from that follows the '{{config.repo_url}}/docs' path, and name of the .md file you want to 'execute'/'test' as the value for the <b><i>MANIFEST</i></b> variable:

``` title="How to 'make' our docs-ecutable target"
make MANIFEST="'docs/content/Getting-Started/quickstart.md'" docs-ecutable
```

<b>note:</b> there are single and double-quotes used here to avoid issues with 'spaces' used in files names or directories.  Use the single and double-quotes as specified in the quickstart example here.

### Important files in our gh-pages branch
#### index.html and home.html
In the 'gh-pages' branch there are two(2) important files that redirect the github docs url to our {{ config.site_name }} doc site hosted with [GoDaddy.com](https://godaddy.com).

<a href="{{config.repo_url}}/blob/gh-pages/home.html">{{config.repo_url}}/blob/gh-pages/home.html</a>
<a href="{{config.repo_url}}/blob/gh-pages/index.html">{{config.repo_url}}/blob/gh-pages/index.html</a>

both files have content similar to:
```html title="index.html and home.html"
<!DOCTYPE html>
<html>
<head>
<title>KubeStellar</title>
<meta http-equiv="content-type" content="text/html; charset=utf-8" >
<meta http-equiv="refresh" content="0; URL=https://docs.kubestellar.io/stable" />
</head>
```

Do not remove these files!
#### CNAME
The CNAME file has to be in the gh-pages root to allow github to recognize the url tls cert served by our hosting provider.  Do not remove this file!

the CNAME file must have the following content in it:
``` title="CNAME"
docs.kubestellar.io
```

#### versions.json
The versions.json file contains the version and alias information required by 'mike' to properly serve our doc site.  This file is maintained by the 'mike' environment and should not be edited by hand.

```json
[{"version": "release-0.2", "title": "release-0.2", "aliases": ["stable"]}, {"version": "{{config.ks_branch}}", "title": "{{config.ks_branch}}", "aliases": ["latest", "unstable"]}]
```

### In case of emergency
If you find yourself in a jam and the pages are not showing up at kubestellar.io or docs.kubestellar.io, check the following
1) Is the index.html, home.html, CNAME, and versions.json file in the gh-pages branch alongside the folders for the compiled documents?  If not, then recreate those files as indicated above (except for versions.json which is programmatically created by 'mike').
2) Is GitHub settings for 'Pages' for the domain pointing at the https://docs.kubestellar.io url?  If not, paste it in and check off 'enforce https'.  This can happen if the CNAME file goes missing from the gh-pages branch.

### How to recreate the gh-pages branch
To recreate the gh-pages branch, do the following:
- checkout the gh-pages branch to your local system
```shell
git clone -b gh-pages {{ config.repo_url }} KubeStellar
cd KubeStellar
```
- delete all files in the branch and push it to GitHub
```shell
rm -rf *
git add; git commit -m "removing all gh-pages files"; git push -u origin gh-pages
```
-- switch to the 'main' branch
```shell
git checkout main
git pull
```
- switch to /docs and run 'mike deploy' for the main branch for alias 'unstable' and 'latest'
```shell
cd docs
mike deploy --push --rebase --update-aliases main unstable
mike deploy --push --rebase --update-aliases main latest
```
- switch to the 'release' branch and 'mike deploy' for the release branch for alias 'stable' (your release name will vary)
```shell
git checkout release-0.2
git pull
mike deploy --push --rebase --update-aliases release-0.2 stable
```
- switch back to the gh-pages branch and recreate the home.html, index.html, and CNAME files as needed (make sure you back out of the docs path first before switching to gh-pages because that path does not exist in that branch)
```shell
cd ..
git checkout gh-pages
git pull
vi index.html
vi home.html
vi CNAME
```
- push the new files into gh-pages
```shell
git add .;git commit -m "add index, home, and CNAME files";git push -u origin gh-pages
```
- go into the GitHub UI and go to the settings for the project and click on 'Pages' to add https://docs.kubestellar.io as the domain and check the box to enforce https.

- if the above did not work, then you might have an issue with the GoDaddy domain (expired, files missing, etc.)

## Publishing Workflow

All documentation building and publishing is done using GitHub Actions in
[docs-gen-and-push.yaml](../.github/workflows/docs-gen-and-push.yaml). The overall sequence is:
<!--readme-for-documentation-end-->
