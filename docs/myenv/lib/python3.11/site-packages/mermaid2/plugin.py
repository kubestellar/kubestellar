"""
Main plugin module for mermaid2
"""

import os


from mkdocs.plugins import BasePlugin
from mkdocs.config.config_options import Type as PluginType
from bs4 import BeautifulSoup

from . import pyjs
from .util import info, libname, url_exists


# ------------------------
# Constants and utilities
# ------------------------
# the default (recommended) mermaid lib:
JAVASCRIPT_VERSION = '10.4.0'
JAVASCRIPT_PRE_10 = "https://unpkg.com/mermaid@%s/dist/mermaid.min.js"
# New format (ESM):
JAVASCRIPT = "https://unpkg.com/mermaid@%s/dist/mermaid.esm.min.mjs"




# Two conditions for activating custom fences:
SUPERFENCES_EXTENSION = 'pymdownx.superfences'
CUSTOM_FENCE_FN = 'fence_mermaid_custom' 



# ------------------------
# Plugin
# ------------------------
class MarkdownMermaidPlugin(BasePlugin):
    """
    Plugin for interpreting Mermaid code
    """
    config_scheme = (

        ('version', PluginType(str, default=JAVASCRIPT_VERSION)),
        ('javascript', PluginType(str, default=None)),
        ('arguments', PluginType(dict, default={})),
        # ('custom_loader', PluginType(bool, default=False))
    )


    # ------------------------
    # Properties
    # Do not call them before on_config was run!
    # ------------------------
    @property
    def full_config(self):
        """
        The full plugin's configuration object,
        which also includes the contents of the yaml config file.
        """
        return self._full_config  
    
    @property
    def mermaid_args(self):
        """
        The arguments for mermaid, found in the config file.
        """
        return self._mermaid_args
    
    @property
    def mermaid_version(self) -> str:
        """
        The version of mermaid
        This information comes from the YAML file parameter,
        or, if empty, from JAVASCRIPT_VERSION.
        """
        version = self.config['version'] or JAVASCRIPT_VERSION
        assert version, "No correct version of mermaid is provided!"
        return version
    
    @property
    def mermaid_major_version(self) -> int:
        """
        Major version of mermaid (e.g. 8. 9, 10) as int
        """
        major = self.mermaid_version.split('.')[0]
        try:
            return int(major)
        except: 
            ValueError("Mermaid version provided has incorrect format")


    @property
    def extra_javascript(self) -> str:
        """
        Provides the mermaid.js library defined in mkdocs.yml 
        under extra_javascript.
        
        To be recognized, the library must have 'mermaid' in the filename.

        WARNING:
            Using extra_javascript for that purpose was the original way,
            but is now DEPRECATED; it bypasses the new and better mechanisms
            for selecting the javascript library.
            It will insert the mermaid library in all pages, regardless
            of whether a mermaid diagram is present or not.
        """
        if not hasattr(self, '_extra_javascript'):
            # As of mkdocs 1.5, extra_javascript is a list of objects; 
            # no longer a string. Call to str was used.
            # Patched in 1.5.1, with __fspath___ method, 
            # see https://github.com/mkdocs/mkdocs/issues/3310
            # But we keep it, to guarantee it's a string. 
            extra_javascript = map(str, self.full_config.get('extra_javascript', []))
            for lib in extra_javascript:
                # check that 'mermaid' is in the filename, minus the extension.
                basename = os.path.basename(lib)
                basename, ext = os.path.splitext(basename)
                if  'mermaid' in basename.lower():
                    self._extra_javascript = lib
                    return lib
            self._extra_javascript = None
        return self._extra_javascript


    @property
    def javascript(self) -> str:
        """
        Provides the url/pathanme of mermaid library according to version
        (distinction on the default, between version < 10 and after)
        """
        if not hasattr(self, '_javascript'):
            # check if a mermaid javascript parameter exists:
            javascript = self.config['javascript']
            if not javascript:
                if self.mermaid_major_version < 10:
                    javascript = JAVASCRIPT_PRE_10 % self.mermaid_version
                else:
                    # newer versions
                    javascript = JAVASCRIPT % self.mermaid_version
                # make checks
            if not url_exists(javascript, 
                              local_base_dir=self.full_config['docs_dir']):
                raise FileNotFoundError("Cannot find Mermaid library: %s" %
                                        javascript)
            self._javascript = javascript
        return self._javascript
    

    @property
    def activate_custom_loader(self) -> bool:
        """
        Predicate: activate the custom loader for superfences?
        The rule is to activate:
            1. superfences extension is activated
            2. it specifies 'fence_mermaid_custom' as
               as format function (instead of fence_mermaid)
        """
        try:
            return self._activate_custom_loader
        except AttributeError:
            # first call:
            # superfences_installed = ('pymdownx.superfences' in 
            #             self.full_config['markdown_extensions'])
            # custom_loader = self.config['custom_loader']
            # self._activate_custom_loader = (superfences_installed and 
            #                                 custom_loader)
            # return self._activate_custom_loader
            self._activate_custom_loader = False
            superfences_installed = (SUPERFENCES_EXTENSION in 
                         self.full_config['markdown_extensions'])
            if superfences_installed:
                # get the config extension configs
                mdx_configs = self.full_config['mdx_configs']
                # get the superfences config, if exists:
                superfence_config = mdx_configs.get(SUPERFENCES_EXTENSION)
                if superfence_config:
                    info("Found superfences config: %s" % superfence_config)
                    custom_fences = superfence_config.get('custom_fences', [])
                    for fence in custom_fences:
                        format_fn = fence.get('format')
                        if format_fn.__name__ == CUSTOM_FENCE_FN:
                            self._activate_custom_loader = True
                            info("Found '%s' function: " 
                                 "activate custom loader for superfences" 
                                 % CUSTOM_FENCE_FN)
                            break
                    
            return self._activate_custom_loader

    # ------------------------
    # Event handlers
    # ------------------------
    def on_config(self, config):
        """
        The initial configuration
        store the configuration in properties
        """
        # the full config info for the plugin is there
        # we copy it into our own variable, to keep it accessible
        self._full_config = config
        # Storing the arguments to be passed to the Javascript library;
        # they are found under `mermaid2:arguments` in the config file:
        self._mermaid_args = self.config['arguments']
        # Here we used the standard self.config property
        # (this can get confusing...)
        assert isinstance(self.mermaid_args, dict)
        info("Initialization arguments:", self.mermaid_args)
        # info on the javascript library:
        if self.extra_javascript:
            info("Explicit mermaid javascript library from extra_javascript:\n  ", 
                 self.extra_javascript)
            info("WARNING: Using extra_javascript is now DEPRECATED; "
                 "use mermaid:javascript instead!")
        elif self.config['javascript']:
            info("Using specified javascript library: %s" %
                 self.config['javascript'])
        else:
            info("Using javascript library (%s):\n  "% 
                  self.config['version'],
                  self.javascript)
            
    def on_post_page(self, output_content, config, page, **kwargs):
        """
        Actions for each page:
        generate the HTML code for all code items marked as 'mermaid'
        """
        if "mermaid" not in output_content:
            # Skip unecessary HTML parsing
            return output_content
        soup = BeautifulSoup(output_content, 'html.parser')
        page_name = page.title
        # first, determine if the page has diagrams:
        if self.activate_custom_loader:
            # the custom loader has its specific marking
            # <pre class = 'mermaid'><code> ... </code></pre>
            info("Custom loader activated")
            mermaids = len(soup.select("pre.mermaid code"))
        else:
            # standard mermaid can accept two types of marking:
            # <pre><code class = 'mermaid'> ... </code></pre>
            # but since we want only <div> for best compatibility,
            # it needs to be replaced
            # NOTE: Python-Markdown changed its representation of code blocks
            # https://python-markdown.github.io/change_log/release-3.3/
            pre_code_tags = (soup.select("pre code.mermaid") or 
                            soup.select("pre code.language-mermaid"))
            no_found = len(pre_code_tags)
            if no_found:
                info("Page '%s': found %s diagrams "
                     "(with <pre><code='[language-]mermaid'>), converting to <div>..." % 
                        (page_name, len(pre_code_tags)))
                for tag in pre_code_tags:
                    content = tag.text
                    new_tag = soup.new_tag("div", attrs={"class": "mermaid"})
                    new_tag.append(content)
                    # replace the parent:
                    tag.parent.replaceWith(new_tag)
            # Count the diagrams <div class = 'mermaid'> ... </div>
            mermaids = len(soup.select("div.mermaid"))
        # if yes, add the javascript snippets:
        if mermaids:
            info("Page '%s': found %s diagrams, adding scripts" % 
                    (page_name, mermaids))
            # insertion of the <script> tag, with the initialization arguments
            new_tag = soup.new_tag("script")
            js_code = [] # the code lines
            if not self.extra_javascript:
                javascript = self.javascript.strip()
                if not javascript.startswith("http"):
                    # it is necessary to adapt the link
                    javascript = os.path.relpath(javascript,
                                            os.path.dirname(page.url))
                # if no extra library mentioned,
                # add the <SCRIPT> tag needed for mermaid
                if javascript.endswith('.mjs'):
                    # <script type="module">
                    # import mermaid from ...
                    new_tag['type'] = "module"
                    js_code.append('import mermaid from "%s";' 
                                   % javascript)
                else:
                    # <script src="...">
                    # generally for self.mermaid_major_version < 10:
                    new_tag['src'] = javascript
                    # it's necessary to close and reopen the tag:
                    soup.body.append(new_tag)
                    new_tag = soup.new_tag("script")

            # (self.mermaid_args), as found in the config file.
            if self.activate_custom_loader:
                # if the superfences extension is present, use the specific loader
                self.mermaid_args['startOnLoad'] = False
                js_args =  pyjs.dumps(self.mermaid_args) 
                js_code.append("window.mermaidConfig = {default: %s};" %
                               js_args)
            else:
                # normal case
                js_args =  pyjs.dumps(self.mermaid_args) 
                js_code.append("mermaid.initialize(%s);" % js_args)
            # merge the code lines into a string:
            new_tag.string = "\n".join(js_code)
            soup.body.append(new_tag)
        return str(soup)
