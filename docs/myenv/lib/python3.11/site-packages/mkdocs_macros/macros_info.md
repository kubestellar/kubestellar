{#
Template for the macro_info() command
(C) Laurent Franceschetti 2019
#}


# Macros Plugin Environment

## General List
All available variables and filters within the macros plugin:
{{ context() | pretty }}

## Config Information
Standard MkDocs configuration information. Do not try to modify.

e.g. {{ "`{{ config.docs_dir }}`" }}

See also the [MkDocs documentation on the config object](https://www.MkDocs.org/user-guide/custom-themes/#config).

{{ context(config)| pretty }}

## Macros
These macros have been defined programmatically for this environment
(module or pluglets). 
{{ context(macros)| pretty }}

## Git Information
Information available on the last commit and the git repository containing the
documentation project:

e.g. {{ "`{{ git.message }}`" }}

{{ context(git)| pretty }}

## Page Attributes
Provided by MkDocs. These attributes change for every page
(the attributes shown are for this page).

e.g. {{ "`{{ page.title }}`" }}

See also the [MkDocs documentation on the page object](https://www.MkDocs.org/user-guide/custom-themes/#page).


{{ context(page)| pretty }}

To have all titles of all pages, use:

```
{% raw %}
{% for page in navigation.pages %}
- {{ page.title }}
{% endfor %}
{% endraw %}
```

## Plugin Filters
These filters are provided as a standard by the macros plugin.
{{ context(filters)| pretty }}

## Builtin Jinja2 Filters
These filters are provided by Jinja2 as a standard.

See also the [Jinja2 documentation on builtin filters](https://jinja.palletsprojects.com/en/2.11.x/templates/#builtin-filters)).

{{ context(filters_builtin) | pretty }}
