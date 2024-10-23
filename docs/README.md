<!--readme-for-documentation-start-->
## Kubestellar Website Build Overview 

### Websites

We have two web sites, as follows.

- `https://kubestellar.io`. This is hosted by GoDaddy and administered by [Andy Anderson](mailto://andy@clubanderson.com). It contains a few redirects. The most important is that `https://kubestellar.io/` redirects to `https://docs.kubestellar.io/`.
- `https://docs.kubestellar.io`. This is a GitHub pages website based on the `github.com/kubestellar/kubestellar/` repository.

**A contributor may have their own copy of the website**, at `https://${repo_owner}.github.io/${fork_name}`, if they have set up the fork properly to render the webpages. See the section below on **Serving up documents globally from a fork of the repo via GitHub**.

### GitHub pages

Our documentation is powered by [mike](https://github.com/jimporter/mike) and [MkDocs](https://www.mkdocs.org/). MkDocs is powered by [Python-Markdown](https://pypi.org/project/Markdown/). These are immensely configurable and extensible. You can see our MkDocs configuration in `docs/mkdocs.yml`. Following are some of the choices we have made.

- The MkDocs theme is [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/).
- MkDocs plugin [awesome-pages](https://github.com/lukasgeiter/mkdocs-awesome-pages-plugin) for greater control over how navigation links are shown.
- MkDocs plugin [macros](https://mkdocs-macros-plugin.readthedocs.io/en/latest/).
- [Our own slightly improved vintage](https://github.com/clubanderson/mkdocs-include-markdown-plugin) of the `include-markdown` MkDocs plugin, allowing the source to be factored into re-used files.
- Python-Markdown extension [SuperFences](https://facelessuser.github.io/pymdown-extensions/extensions/superfences/), supporting fenced code blocks that play nice with other markdown features.
- Python-Markdown extension [Highlight](https://facelessuser.github.io/pymdown-extensions/extensions/highlight/), for syntax highlighting of fenced code.
- [Pygments](https://pypi.org/project/Pygments/) for even fancier code highlighting.
- MkDocs plugin [mkdocs-static-i18n](https://github.com/ultrabug/mkdocs-static-i18n/tree/0.53#readme) to support multiple languages. We currently only have documentation in English. If you're interested in contributing translations, please let us know!

-----

## Rendering and Previewing modifications to the website
You may preview possible changes to the website by either rendering them globally from a fork on GitHub, or by downloading and rendering the documents locally.


### Serving up documents globally from a fork of the repository via GitHub
You can also take advantage of the "Generate and Push Docs" Action via the github web interface to create an online, shareable rendering of the website.
This is particularly useful for documentation PRs, as it allows you to share a preview of your proposed changes directly via a URL to a working website.
To take advantage of this action, you must ensure that you have forked the repository properly, so your fork includes the gh-pages branch required for the Action to run properly:

#### Creating a fork that can use the Generate and Push Docs Action

 1. Log into your GitHub account via webbrowser
 2. Navigate to github.com/kubestellar/kubestellar
 3. Select the **Forks** dropdown and click on the plus sign to create a new fork ![image](https://github.com/user-attachments/assets/c0b56897-c2c4-479f-ba16-c1edd4690946)
 4. In the resulting dialog select your account as the owner, pick a repository name for the fork, and _be sure to uncheck the_ **"copy the** main **branch only"** _box_ <br />
![image](https://github.com/user-attachments/assets/c5909ddd-3bf6-44c2-9102-c07c7e1d6a05)

#### If you already created a fork but only included the main branch
You can remedy the problem by propagating the _gh-pages_ branch into your fork using _git_ commands
 
#### Generating a website rendered from a branch of your fork
1. Work on the documents in a branch of your fork of the repository, and commit the changes
2. (If you have been working on a local copy of the files, push the changes to the fork, then log into the GitHub webpage for your fork)
3. Switch to the Actions tab in the top menu bar of the repository page
4. Select Generate and Push Docs from the list of Actions on the left
5. Click on the Run Workflow button on the right ![image](https://github.com/user-attachments/assets/5d3d23be-6c8f-454e-bf2c-0c58c8894957)
6. Select the branch you wish to render and click on the second Run Workflow Button ![image](https://github.com/user-attachments/assets/427b827d-555c-4d36-b9c8-485eda002428)
7. If that workflow completes successfully, it will automatically call the **Pages build and deployment** workflow.
8. You can observe the progress of the workflows on the Actions page; a green checkmark circle indicates successful completion.<br />![image](https://github.com/user-attachments/assets/b9ce40f8-b744-4b3c-bc20-a4814243e85e)
9. After a minute or so, you should be able to preview your new version of the website at `https://${repo_owner}.github.io/${fork_name}/${branch_name}`

#### Automatically generate webpages
If you create a branch of your fork that begins with **doc-** (e.g. _doc-myversion_) the workflow will trigger automatically when you commit changes to the branch.

#### Switching between versions
Each branch of your fork will render as its own version. You can use the release dropdown inside the rendered pages to quickly switch between versions.

Note: the **main** branch will render as `https://${repo_owner}.github.io/${fork_name}/main`, **NOT** as "unreleased-development" which is a special alias on the main kubestellar.io website.

#### Removing outdated (draft branch) versions after rendering
You can use `mike` to remove versions, or replace gh-pages with a copy of the shared version. 
More details on these techniques will be added here soon.

-----

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

-----

## Supported aliases for our documentation

`mike` has a concept of aliases. We currently maintain only one alias.

- `latest` ([{{config.docs_url}}/latest](https://docs.kubestellar.io/latest)), for the latest regular release.

The publishing workflow updates these aliases. The latest regular release is determined by picking the first version listed by `mike list` that matches the regexp `release-[0-9.]*`.

## Publishing from the branch named "main"

The branch named "main" also gets published as a "version" on the
website, but with a different name. This is not done by `mike`
aliasing, because that only _adds_ a version. The branch named "main"
is published as the version named "unreleased-development".

## Shortcut URLs

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
- [https://kubestellar.io/quickstart](https://kubestellar.io/quickstart) - our 'stable' Getting Started recipe

## Jinja templating

Our documentation stack includes [Jinja](https://jinja.palletsprojects.com/en/3.1.x/). The Jinja constructs --- \{\# comment \#\}, \{\{ expression \}\}, and {&#37; statement &#37;} --- can appear in the markdown sources.

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
    - ks_latest_regular_release: {{ config.ks_latest_regular_release }}
    - ks_latest_release: {{ config.ks_latest_release }}

to use a variables/macro in your documentation reference like this:

\{\{ config.<var_name\> \}\}

and in context that can look something like this:

bash <(curl -s \{\{ config.repo_raw_url \}\}/\{\{ config.ks_branch \}\}/bootstrap/bootstrap-kubestellar.sh) --kubestellar-version \{\{ config.ks_tag \}\}


<b>note:</b><br /> 
&nbsp;&nbsp;&nbsp;&nbsp;- We also check for broken links as part of our PR pipeline.  For more information check out our <a href="{{ config.repo_url }}/actions/workflows/broken-links-crawler.yml">Broken Links Crawler</a><br />

### Navigation (website menu)

The navigation for the documentation is _also_ configured in  <a href="{{ config.repo_raw_url }}/{{ config.ks_branch }}/docs/mkdocs.yml">docs/mkdocs.yml</a>.
The section which begins with **nav:** lays out the navigation structure and which markdown files correspond to each topic. 


### Page variables

A markdown source file can contribute additional variables by defining them in `name: value` lines at the start of the file, set off by lines of triple dashes. For example, suppose a markdown file begins with the following.

```markdown
---
short_name: example1
manifest_name: 'docs/content/Coding Milestones/PoC2023q1/example1.md'
---
```

These variables can be referenced as \{\{ page.meta.short_name \}\} and \{\{ page.meta.manifest_name \}\}.

### Including external markdown
We make extensive use of 'include-markdown' to help us keep our documentation modular and up-to-date.  To use 'include-markdown' you must add a block in your document that refers to a block in your external document content:

In your original markdown document, add a block that refers to the external markdown you want to include:
![Include Markdown](./content/Contribution%20guidelines/operations/include-markdown-example.png)

In the document you want to include, add the start and end tags you configured in the include-markdown block in your original document:
![Included Markdown](./content/Contribution%20guidelines/operations/included-markdown-example.png)

for more information on the 'include-markdown' plugin for mkdocs look [here](https://github.com/mondeja/mkdocs-include-markdown-plugin)

### Codeblocks
mkdocs has some very helpful ways to include blocks of code in a style that makes it clear to our readers that console interaction is necessary in the documentation.  There are options to include a plain codeblock (```), shell (shell), console (console - no used in our documentation), language or format-specific (yaml, etc.), and others.  For more detailed information, checkout the [mkdocs information on codeblocks](https://squidfunk.github.io/mkdocs-material/reference/code-blocks/).

**NOTE**: the docs-ecutable technology does _not_ apply Jinja, at any stage; Jinja source inside executed code blocks will not be expanded by Jinja but rather seen directly by `bash`.

Here are some examples of how we use codeblocks.

#### Seen and executed

For a codeblock that can be 'tested' (and seen by the reader) as part of our CI, use the <b><i>`shell`</i></b> block:
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

#### Executed but not seen

(Think hard before hiding stuff from your reader.)

For a codeblock that should be 'tested', BUT <b>not</b> seen by the reader, use the <b><i>`.bash`</i></b> with the plain codeblock, and the <b><i>'.hide-me'</i></b> style (great for hiding a sleep command that user does not need to run, but CI does):
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

#### Seen but not executed

(To avoid confusing readers of the HTML, this should be used only for _output_ seen in a shell session.)

For a codeblock that should <u>not</u> be 'tested' as part of our CI, use the <b><i>`.bash`</i></b> with the plain codeblock, and <b>without</b> the <b><i>'.hide-me'</b></i> style:
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

#### Seen but not executed and no copy button

For a codeblock that should not be 'tested', be seen by the reader, and not include a 'copy' icon (great for output-only instances), use the <b><i>`.bash`</i></b> codeblock <b>without</b> the <b><i>'.no-copy'</b></i> style:
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

#### Other language-specific highlighting

For other language-specific highlighting (yaml, etc.), use the <b><i>yaml</i></b> codeblock
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

#### Codeblock with a title

For a codeblock that has a title, and will not be tested, use the <b><i>'title'</i></b> parameter in conjunction with the plain codeblock (greater for showing or prescribing contents of files):
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
- An example workflow using the newer technology is located in our {{ config.repo_short_name }} repo at <a href="{{ config.repo_url }}/blob/{{ config.ks_branch }}/.github/workflows/docs-ecutable-example1.yml">{{ config.repo_url }}/blob/{{ config.ks_branch }}/.github/workflows/docs-ecutable-example1.yml</a>

#### The original secret sauce:
- The original code that made all this possible is at <a href="{{ config.repo_url }}/blob/{{ config.ks_branch }}/docs/scripts/docs-ecutable.sh">{{ config.repo_url }}/blob/{{ config.ks_branch }}/docs/scripts/docs-ecutable.sh</a>
    - This code parses the .md file you give it to pull out all the 'shell' and '.bash .hide-me' blocks
    - The code is smart enough to traverse the include-markdown blocks and include the 'shell' and '.bash .hide-me' blocks in them
    - The Jinja constructs are not expanded by this code.
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

#### The new and improved secret sauce:
- The newer code for executing bash snippets in documentation is at <a href="{{ config.repo_url }}/blob/{{ config.ks_branch }}/docs/scripts/execute-html.sh">{{ config.repo_url }}/blob/{{ config.ks_branch }}/docs/scripts/execute-html.sh</a>
    - This code parses the HTML generated by MkDocs to extract all the fenced code blocks tagged for the "shell" language.
    - This HTML scraping is relatively easy because it does not have to work on general HTML but only the HTML generated by _our_ stack from _our_ sources. The use of the option setting `pygments_lang_class: true` for the Python-Markdown extension `pymdownx.highlight` plays a critical role, getting the source language into the generated HTML.
    - Because it reads the generated HTML, invisible code blocks are not extracted.
    - Because it reads the generated HTML, the Jinja constructs have their usual effects.
    - This script is given the name of the HTML file to read and the current working directory to establish at the start of the extracted bash.
    - It then creates a file called 'generated_script.sh' which is then run.

All of this is invoked in a target in our <a href="{{ config.repo_url }}/blob/{{config.ks_branch}}/Makefile">Makefile</a>
``` {.bash .no-copy}
.PHONY: execute-html
execute-html: venv
	. $(VENV)/activate; \
	cd docs; \
	mkdocs build; \
	scripts/execute-html.sh "$$PWD/.." "generated/$(MANIFEST)/index.html"
```

The `make` target requires the variable `MANIFEST` to be set to the directory that contains the generated `index.html` file, relative to '{{config.repo_url}}/docs/generated'. This is the name of the markdown source file, relative to '{{config.repo_url}}/docs/content' and with the `.md` extension dropped.

``` title="How to 'make' a docs-ecutable target"
make MANIFEST="Coding Milestones/PoC2023q1/example1" execute-html
```

<b>note:</b> this target has no special needs for quoting --- which is not to deny the quoting that your shell needs.

### Important files in our gh-pages branch
#### index.html and home.html
These appear in the branch named `gh-pages` and redirect from the root to the version named `latest`. The one named `index.html` is managed by `mike set-default`. The other should be kept consistent.

- <a href="{{config.repo_url}}/blob/gh-pages/home.html">{{config.repo_url}}/blob/gh-pages/home.html</a>
- <a href="{{config.repo_url}}/blob/gh-pages/index.html">{{config.repo_url}}/blob/gh-pages/index.html</a>

both files have content similar to:
```html title="index.html and home.html"
<!DOCTYPE html>
<html>
<head>
<title>KubeStellar</title>
<meta http-equiv="content-type" content="text/html; charset=utf-8" >
<meta http-equiv="refresh" content="0; URL=https://docs.kubestellar.io/latest" />
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
[{"version": "release-0.22.0", "title": "release-0.22.0", "aliases": ["latest"]}, {"version": "release-0.22.0-rc3", "title": "release-0.22.0-rc3", "aliases": []}, {"version": "release-0.21.2", "title": "release-0.21.2", "aliases": []}, {"version": "release-0.21.2-rc1", "title": "release-0.21.2-rc1", "aliases": []}, {"version": "release-0.21.1", "title": "release-0.21.1", "aliases": []}, {"version": "release-0.21.0", "title": "release-0.21.0", "aliases": []}, {"version": "release-0.14", "title": "release-0.14", "aliases": []}]
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
- switch to the 'release' branch, switch to /docs and run 'mike deploy' for the release branch. Add the alias 'latest' for the latest release 
```shell
git checkout release-0.22
git pull
mike deploy --push --rebase --update-aliases release-0.22 latest
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

### How to delete a rendering of a branch

Use `mike delete $branch_name`, either acting locally on your checked out `gh-pages` branch (after pull and before git commit and push) or acting more directly on the remote repo using `--remote` and `--push`. See [the mike delete command doc](https://github.com/jimporter/mike?tab=readme-ov-file#deleting-docs).

## Publishing Workflow

All documentation building and publishing is done using GitHub Actions in
`.github/workflows/docs-gen-and-push.yml`. This workflow is triggered either manually or by a push to a branch named `main` or `release-<something>` or `doc-<something>`. This workflow will actually do something _ONLY_ if either (a) it is acting on the shared GitHub repository at `github.com/kubestellar/kubestellar` and on behalf of the repository owner or (b) it is acting on a contributor's fork of that repo and on behalf of that same contributor. The published site appears at `https://pages.github.io/kubestellar/${branch}` in case (a) and at `https://${repo_owner}.github.io/${fork_name}/${branch}` in case (b). This workflow will build and publish a website _version_ whose name is the same as the name of the branch that it is working on. This workflow will also update the relevant `mike` alias, if necessary.

<!--readme-for-documentation-end-->
