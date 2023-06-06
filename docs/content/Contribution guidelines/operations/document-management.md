##

### Global Variables
There are many global variables defined in the [docs/mkdocs.yml]({{ config.repo_raw_url }}/{{ config.ks_branch }}/docs/mkdocs.yml).  The following are some very common variables you are encouraged to use in our documentation.  Use of these variables/macros allows our documentation to have github branch context and take advantage of our evolution without breaking

    - site_name: {{ config.site_name }}
    - repo_url: {{ config.repo_url }}
    - site_url: {{ config.site_url }}
    - repo_default_file_path: {{ config.repo_default_path }}
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
&nbsp;&nbsp;&nbsp;&nbsp;- A more extensive and detailed list is located at [mkdocs information](all-macros.md) <br />
&nbsp;&nbsp;&nbsp;&nbsp;- We also check for broken links as part of our PR pipeline.  For more information check out our [Broken Links Crawler]({{ config.repo_url }}/actions/workflows/broken-links-crawler.yml)<br />

### Including external markdown
We make extensive use of 'include-markdown' to help us keep our documentation modular and up-to-date.  To use 'include-markdown' you must add a block in your document that refers to a block in your external document content:

In your original markdown document, add a block that refers to the external markdown you want to include:
![Include Markdown](./include-markdown-example.png)

In the document you want to include, add the start and end tags you configured in the include-markdown block in your original document:
![Included Markdown](./included-markdown-example.png)

for more information on the 'include-markdown' plugin for mkdocs look [here](https://github.com/mondeja/mkdocs-include-markdown-plugin)

### Serving up documents locally
You can view and modify our documentation in your local development environment.  Simply checkout one of our branches.

```shell
git clone git@github.com:{{ config.repo_short_name }}.git
cd {{ config.repo_default_path }}/docs
git checkout {{ config.ks_branch }}
```

You can view and modify our documentation in the branch you have checked out by using `mkdocs serve` from [mkdocs](https://www.mkdocs.org):

```shell
pip install -r requirements.txt
mkdocs serve
```
Then open a browser to [`http://localhost:8000/`](http://localhost:8000/)

Another way to view (not modify - this method reflects what has been deployed to the `gh-pages` branch of our repo) all branches/versions of our documentation locally using 'mike' [mike for mkdocs](https://github.com/jimporter/mike):

```shell
git clone git@github.com:{{ config.repo_short_name }}.git
cd {{ config.repo_default_path }}
git checkout {{ config.ks_branch }}
make serve-docs
```
Then open a browser to [`http://localhost:8000/`](http://localhost:8000/)

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
- [https://kubestellar.io/code](https://kubestellar.io/code) - our current GH repo (wherever that is)
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

- For a codeblock that can be 'tested' as part of our CI, use the <b><i>`shell`</i></b> block:
<br/><b>code:</b>
````
```shell
mkdocs serve
```
````
<b>output:</b>


- For a codeblock that should be 'tested', BUT <b>not</b> shown, use the <b><i>`.bash`</i></b> with the plain codeblock, and the <b><i>'.hide-me'</i></b> style (great for hiding a sleep command that user does not need to run, but CI does):
<br/><b>code:</b>
````
``` {.bash .hide-me}
sleep 10
```
````
<b>output:</b>


- For a codeblock that should <u>not</u> be 'tested' as part of our CI, use the <b><i>`.bash`</i></b> with the plain codeblock, and <b>without</b> the <b><i>'.hide-me'</b></i> style:
<br/><b>code:</b>
````
``` { .bash }
mkdocs server
```
````
<b>output:</b>
``` { .bash }
mkdocs server
```

- For a codeblock that should not be 'tested' and not include a 'copy' icon (great for output-only instances), use the <b><i>`.bash`</i></b> codeblock <b>without</b> the <b><i>'.no-copy'</b></i> style:
<br/><b>code:</b>
````
``` { .bash .no-copy }
I0412 15:15:57.867837   94634 shared_informer.go:282] Waiting for caches to sync for placement-translator
I0412 15:15:57.969533   94634 shared_informer.go:289] Caches are synced for placement-translator
I0412 15:15:57.970003   94634 shared_informer.go:282] Waiting for caches to sync for what-resolver
```
````
<b>output:</b>
``` { .bash .no-copy }
I0412 15:15:57.867837   94634 shared_informer.go:282] Waiting for caches to sync for placement-translator
I0412 15:15:57.969533   94634 shared_informer.go:289] Caches are synced for placement-translator
I0412 15:15:57.970003   94634 shared_informer.go:282] Waiting for caches to sync for what-resolver
```

- For language-specific highlighting (yaml, etc.), use the <b><i>yaml</i></b> codeblock
<br/><b>code:</b>
````
```yaml

```
````
<b>output:</b>
```yaml
nav:
  - Home: index.md
  - QuickStart: Getting-Started/quickstart.md
  - Contributing: 
      - Guidelines: Contribution guidelines/CONTRIBUTING.md
```

- For a codeblock that has a title, use the <b><i>'title'</i></b> parameter in conjunction with the plain codeblock:
<br/><b>code:</b>
````
``` title="testing.sh"
#!/bin/sh
echo hello KubeStellar
```
````
<b>output:</b>
``` title="testing.sh"
#!/bin/sh
echo hello KubeStellar
```

(other variations are possible, PR an update to the [kubestellar.css]({{ config.repo_url }}/blob/{{ config.ks_branch }}/docs/overrides/stylesheets/kubestellar.css) file and, once approved, use the style on the plain codeblock in your documentation.)

### Testing/Running Docs
How do we ensure that our documented examples work?  Simple, we 'execute' our documentation in our CI.  We built automation called 'run-doc-shells' which can be invoked to test any markdown (.md) file in our repository. You could use it in your project as well - afterall it is opensource.

#### The way it works:
- create your .md file as you normally would
- add codeblocks that can be tested, tested but hidden, or not tested at all:
    - use <b><i>'shell'</i></b> to indicate code you want to be tested
    - use <b><i>'.bash'</i></b> with the plain codeblock, and the <b><i>'.hide-md'</i></b> style for code you want to be tested, but hidden from the reader (some like this, but its not cool if you want others to run your instructions without hiccups)
    - use plain codeblock (```) if you want to show sample output that is not to be tested
- you can use 'include-markdown' blocks, and they will also be executed (or not), depending on the codeblock style you use in the included markdown files.

#### The GH Workflow:
- One example of the GH Workflow is located in our {{ config.repo_short_name }} at [{{ config.repo_url }}/blob/{{ config.ks_branch }}/.github/workflows/run-doc-shells-qs.yml]({{ config.repo_url }}/blob/{{ config.ks_branch }}/.github/workflows/run-doc-shells-qs.yml)

#### The secret sauce:
- The code that makes all this possible is at [{{ config.repo_url }}/blob/{{ config.ks_branch }}/docs/scripts/run-doc-shells.sh]({{ config.repo_url }}/blob/{{ config.ks_branch }}/docs/scripts/run-doc-shells.sh)
    - This code parses the .md file you give it to pull out all the 'shell' and '.bash .hide-me' blocks
    - The code is smart enough to traverse the include-markdown blocks and include the 'shell' and '.bash .hide-me' blocks in them
    - It then creates a file called 'generate_script.sh' which is then run at the end of the run-doc-shells execution.

All of this is invoke in a target in our [makefile]({{ config.repo_url }}/blob/{{config.ks_branch}}/Makefile)
``` {.bash .no-copy}
.PHONY: run-doc-shells
run-doc-shells: venv
	. $(VENV)/activate; \
	MANIFEST=$(MANIFEST) docs/scripts/run-doc-shells.sh
```

You give the path from that follows the '{{config.repo_url}}/docs' path, and name of the .md file you want to 'execute'/'test' as the value for the <b><i>MANIFEST</i></b> variable:

``` title="How to 'make' our run-doc-shells target"
make MANIFEST="'content/Getting-Started/quickstart.md'" run-doc-shells
```

<b>note:</b> there are single and double-quotes used here to avoid issues with 'spaces' used in files names or directories.  Use the single and double-quotes as specified in the quickstart example here.

### Important files in our gh-pages branch
#### index.html and home.html
In the 'gh-pages' branch there are two(2) important files that redirect the github docs url to our {{ config.site_name }} doc site hosted with [GoDaddy.com](https://godaddy.com).

[{{config.repo_url}}/blob/gh-pages/home.html]({{config.repo_url}}/blob/gh-pages/home.html)
[{{config.repo_url}}/blob/gh-pages/index.html]({{config.repo_url}}/blob/gh-pages/index.html)

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
```title="CNAME"
docs.kubestellar.io
```

#### versions.json
The versions.json file contains the version and alias information required by 'mike' to properly serve our doc site.  This file is maintained by the 'mike' environment and should not be edited by hand.

```json
[{"version": "release-0.2", "title": "release-0.2", "aliases": ["stable"]}, {"version": "{{config.ks_branch}}", "title": "{{config.ks_branch}}", "aliases": ["latest", "unstable"]}]
```