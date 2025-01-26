# --------------------------------------------
# Main part of the plugin
# Defines the VariablesPlugin class
#
# Laurent Franceschetti (c) 2018
# MIT License
# --------------------------------------------

import os
import traceback
from copy import copy
import importlib


import yaml
from jinja2 import (Environment, FileSystemLoader, TemplateSyntaxError,
                    Undefined, DebugUndefined, StrictUndefined)
from mkdocs.plugins import BasePlugin
from mkdocs.config import config_options
from mkdocs.config.config_options import Type as PluginType


from .util import (install_package, parse_package, trace, debug,
        update, SuperDict, import_local_module, format_chatter, LOG)
from .context import define_env

# ------------------------------------------
# Initialization
# ------------------------------------------

# The subsets of the YAML file that will be used for the variables:
YAML_VARIABLES = 'extra'

# The default name of the Python module:
DEFAULT_MODULE_NAME = 'main'  # main.py

# Possible behavior in case of ignored variables or macros (first is default)


class LaxUndefined(Undefined):
    "Pass anything wrong as blank"

    def _fail_with_undefined_error(self, *args, **kwargs):
        return ''


UNDEFINED_BEHAVIOR = {'keep': DebugUndefined,
                      'silent': Undefined,
                      'strict': StrictUndefined,
                      # lax will even pass unknown objects:
                      'lax': LaxUndefined}

# By default undefined jinja2 variables AND macros will be left as-is
# see https://stackoverflow.com/a/53134416
DEFAULT_UNDEFINED_BEHAVIOR = 'keep'

# Return codes in case of error
ERROR_MACRO = 100


# ------------------------------------------
# Plugin
# ------------------------------------------


class MacrosPlugin(BasePlugin):
    """
    Inject config 'extra' variables into the markdown
    plus macros / variables defined in external module.

    The python code is located in 'main.py' or in a 'main' package
    in the root directory of the website
    (unless you want to redefine that name in the 'python_module' value
    in the mkdocs.yml file)
    """

    # what is under the 'macros' namespace (will go into the config property):
    J2_STRING = PluginType(str, default='')
    config_scheme = (
        # main python module:
        ('module_name',  PluginType(str,
                                    default=DEFAULT_MODULE_NAME)),
        ('modules', PluginType(list,
                               default=[])),
        # include directory for templates ({% include ....%}):
        ('include_dir',  J2_STRING),
        # list of additional yaml files:
        ('include_yaml', PluginType(list, default=[])),
        # for altering the j2 markers, in case of need:
        ('j2_block_start_string',    J2_STRING),
        ('j2_block_end_string',      J2_STRING),
        ('j2_variable_start_string', J2_STRING),
        ('j2_variable_end_string',   J2_STRING),
        # for behavior of unknown macro (e.g. other plugin):
        ('on_undefined',  PluginType(str, default=DEFAULT_UNDEFINED_BEHAVIOR)),
        # for CD/CI set that parameter to true
        ('on_error_fail', PluginType(bool, default=False)),
        ('verbose', PluginType(bool, default=False))
    )

    def start_chatting(self, prefix: str, color: str = 'yellow'):
        "Generate a chatter function (trace for macros)"
        def chatter(*args):
            """
            Defines a tracer for the Verbose mode, to be used in macros.
            If `verbose: true` in the YAML config file (under macros plugin), 
            it will start "chattering"  
            (talking a lot and in a friendly way,
            about mostly unimportant things).
            Otherwise, it will remain silent.

            If you change the `verbose` while the local server is activated,
            (`mkdocs server`) this should be instantly reflected.

            Usage:
            -----
            chatter = env.make_chatter('MY_MODULE_NAME')
            chatter("This is a dull debug message.")

            Will result in:

            INFO    -  [macros - Simple module] - This is a dull info message.
            """
            if self.config['verbose']:
                LOG.info(format_chatter(*args, prefix=prefix, color=color))

        return chatter

    # ------------------------------------------------
    # These properties are available in the env object
    # ------------------------------------------------

    @property
    def conf(self):
        """
        Dictionary containing of the whole config file (by default: mkdocs.yml)

        This property may be useful if the code in the module needs to access
        general configuration information.

        NOTE: this property is called 'conf', because there is already
              a 'config' property in a BasePlugin object,
              which is the data connected to the macros plugin
              (in the yaml file)
        """
        try:
            return self._conf
        except AttributeError:
            raise AttributeError("Conf property of macros plugin "
                                 "was called before it was initialized!")

    @property
    def variables(self):
        "The cumulative list of variables, initialized by on_config()"
        try:
            return self._variables
        except AttributeError:
            raise AttributeError("Property called before on_config()")

    @property
    def macros(self):
        "The cumulative list of macros, initialized by on_config()"
        try:
            return self._macros
        except AttributeError:
            raise AttributeError("Property called before on_config()")

    @property
    def filters(self):
        "The list of filters defined in the module, initialized by on_config()"
        try:
            return self._filters
        except AttributeError:
            self._filters = {}
            return self._filters

    @property
    def project_dir(self):
        "The directory of project"
        # we calculate it from the configuration file
        CONFIG_FILE = self.conf['config_file_path']
        return os.path.dirname(os.path.abspath(CONFIG_FILE))

    def macro(self, v, name=''):
        """
        Registers a variable as a macro in the template,
        i.e. in the variables dictionary:

            env.macro(myfunc)

        Optionally, you can assign a different name:

            env.macro(myfunc, 'funcname')


        You can also use it as a decorator:

        @env.macro
        def foo(a):
            return a ** 2

        More info:
        https://stackoverflow.com/questions/6036082/call-a-python-function-from-jinja2
        """

        name = name or v.__name__
        self.macros[name] = v
        return v

    def filter(self, v, name=''):
        """
        Register a filter in the template,
        i.e. in the filters dictionary:

            env.filter(myfunc)

        Optionally, you can assign a different name:

            env.filter(myfunc, 'filtername')


        You can also use it as a decorator:

        @env.filter
        def reverse(x):
            "Reverse a string (and uppercase)"
            return x.upper().[::-1]

        See: https://jinja.palletsprojects.com/en/2.10.x/api/#custom-filters
        """

        name = name or v.__name__
        self.filters[name] = v
        return v

    @property
    def page(self):
        """
        The page information
        """
        try:
            return self._page
        except AttributeError:
            raise AttributeError("Too early: page information is not available"
                                 "at this stage!")

    @property
    def raw_markdown(self):
        """
        The page information
        """
        try:
            return self._raw_markdown
        except AttributeError:
            raise AttributeError("Too early: raw markdown is not available"
                                 "at this stage!")

    @raw_markdown.setter
    def raw_markdown(self, value):
        """
        Used to set the raw markdown
        """
        if not isinstance(value, str):
            raise ValueError("Value provided to attribute raw_markdown "
                             "should be a string")
        # check whether attribute is accessible:
        self.raw_markdown
        self._raw_markdown = value

    # ----------------------------------
    # Function lists, for later events
    # ----------------------------------

    @property
    def pre_macro_functions(self):
        """
        List of pre-macro functions contained in modules.
        These are deferred to the on_page_markdown() event.
        """
        try:
            return self._pre_macro_functions
        except AttributeError:
            raise AttributeError("You called the pre_macro_functions property "
                                 "too early. Does not exist yet !")

    @property
    def post_macro_functions(self):
        """
        List of post-macro functions contained in modules.
        These are deferred to the on_page_markdown() event.
        """
        try:
            return self._post_macro_functions
        except AttributeError:
            raise AttributeError("You called the post_macro_functions property "
                                 "too early. Does not exist yet !")

    @property
    def post_build_functions(self):
        """
        List of post build functions contained in modules.
        These are deferred to the on_post_build() event.
        """
        try:
            return self._post_build_functions
        except AttributeError:
            raise AttributeError("You called post_build_functions property "
                                 "too early. Does not exist yet !")

    # ----------------------------------
    # load elements
    # ----------------------------------

    def _load_yaml(self):
        "Load the the external yaml files"
        for el in self.config['include_yaml']:
            # el is either a filename or {key: filename} single-entry dict
            try:
                [[key, filename]] = el.items()
            except AttributeError:
                key = None
                filename = el
            # Paths are be relative to the project root.
            filename = os.path.join(self.project_dir, filename)
            if os.path.isfile(filename):
                with open(filename) as f:
                    # load the yaml file
                    # NOTE: for the SafeLoader argument, see: https://github.com/yaml/pyyaml/wiki/PyYAML-yaml.load(input)-Deprecation
                    content = yaml.load(f, Loader=yaml.SafeLoader)
                    trace("Loading yaml file:", filename)
                if key is not None:
                    content = {key: content}
                update(self.variables, content)
            else:
                trace("WARNING: YAML configuration file was not found!",
                      filename)

    def _load_module(self, module, module_name):
        """
        Load a single module

        Add variables and functions to the config dictionary,
        via the python module
        (located in the same directory as the Yaml config file).

        This function enriches the variables dictionary

        The python module must contain the following hook:

        define_env(env):
            "Declare environment for jinja2 templates for markdown"

            env.variables['a'] = 5

            @env.macro
            def bar(x):
                ...

            @env.macro
            def baz(x):
                ...

            @env.filter
            def foobar(x):
                ...

        """
        if not module:
            return
        trace("Found external Python module '%s' in:" % module_name,
              self.project_dir)
        # execute the hook for the macros
        function_found = False
        if hasattr(module, 'define_env'):
            module.define_env(self)
            function_found = True
        if hasattr(module, 'declare_variables'):
            # this is for compatibility (DEPRECATED)
            module.declare_variables(self.variables, self.macro)
            trace("You are using declare_variables() in the python "
                  "module '%s'. Prefer the define_env() function "
                  "(see documentation)!" % module_name)
            function_found = True
        if not function_found:
            raise NameError("No valid function found in module '%s'" %
                            module_name)
        # DECLARE additional event functions
        # NOTE: each of these functions requires self (the environment).

        def add_function(funcname: str, funclist: list):
            "Add an optional function to the module"
            if hasattr(module, funcname):
                func = getattr(module, funcname)
                funclist.append(func)
        add_function('on_pre_page_macros',  self.pre_macro_functions)
        add_function('on_post_page_macros', self.post_macro_functions)
        add_function('on_post_build',       self.post_build_functions)

    def _load_modules(self):
        "Load all modules"
        self._pre_macro_functions = []
        self._post_macro_functions = []
        self._post_build_functions = []

        # pluglets installed modules (as in pip list)
        modules = self.config['modules']
        if modules:
            trace("Preinstalled modules: ", ','.join(modules))
        for m in modules:
            # split the name of package in source (pypi) and module name
            source_name, module_name = parse_package(m)
            try:
                module = importlib.import_module(module_name)
            except ModuleNotFoundError:
                try:
                    # if absent, install (from pypi)
                    trace("Module '%s' not found, installing (source: '%s')" %
                          (module_name, source_name))
                    install_package(source_name)
                    # install package raises NameError
                    module = importlib.import_module(module_name)
                except (NameError, ModuleNotFoundError):
                    raise ModuleNotFoundError("Could not import installed "
                                              "module '%s' (missing?)" %
                                              module_name,
                                              name=module_name)
            self._load_module(module, module_name)
        # local module (file or dir)
        local_module_name = self.config['module_name']
        debug("Project dir '%s'" % self.project_dir)
        module = import_local_module(self.project_dir, local_module_name)
        if module:
            trace("Found local Python module '%s' in:" % local_module_name,
                  self.project_dir)
            self._load_module(module, local_module_name)

        else:
            if local_module_name == DEFAULT_MODULE_NAME:
                # do not do anything if there is no main module
                # trace("No module")
                pass
            else:
                raise ImportError("Macro plugin could not find custom '%s' "
                                  "module in '%s'." %
                                  (local_module_name, self.project_dir))

    def render(self, markdown):
        """
        Render a page through jinja2: it executes the macros

        Returns
        -------
        A pure markdown/HTML page.

        Notes
        -----
        - Must called by '_on_page_markdown()'
        - If the YAML header of the page contains `ignore_macros: true`
          then NO rendering will be done, and the markdown will be returned
          as is.
        """
        # copy the page variables and update with the meta data
        # in the YAML header:
        page_variables = copy(self.variables)
        meta_variables = self.variables['page'].meta
        # it must be possible to completely
        if meta_variables:
            # a trick to force of a page NOT to be interpreted,
            if meta_variables.get('ignore_macros') == True:
                return markdown
            # trace("Metavariables for '%s':" % self.variables['page'].title,
            #                 meta_variables)
            page_variables.update(meta_variables)

        # expand the template
        on_error_fail = self.config['on_error_fail']
        try:
            md_template = self.env.from_string(markdown)
            # Execute the jinja2 template and return
            return md_template.render(**page_variables)
        except TemplateSyntaxError as e:
            line = markdown.splitlines()[e.lineno-1]
            output = ["# _Macro Syntax Error_",
                      "_Line %s in Markdown file:_ **%s** " %
                      (e.lineno, e.message),
                      "```python",
                      line,
                      "```"]
            error = "\n".join(output)
            trace("ERROR", error)
            if on_error_fail:
                exit(ERROR_MACRO)
            else:
                # default case
                return error
        except Exception as e:
            output = ["# _Macro Rendering Error_",
                      "",
                      "**%s**: %s" % (type(e).__name__, e),
                      "", "",
                      "```",
                      traceback.format_exc(),
                      "```"]
            error = "\n".join(output)
            trace("ERROR", error)
            if on_error_fail:
                exit(ERROR_MACRO)
            else:
                # default case
                return error

    # ----------------------------------
    # Standard Hooks for a mkdocs plugin
    # ----------------------------------

    def on_config(self, config):
        """
        Called once (initialization)
        From the configuration file, builds a Jinj2 environment
        with variables, functions and filters.
        """
        # WARNING: this is not the config argument:
        trace("Macros arguments:", self.config)
        # define the variables and macros as dictionaries
        # (for update function to work):
        self._variables = SuperDict()
        self._macros = SuperDict()

        # load the extra variables
        extra = dict(config.get(YAML_VARIABLES))
        # make a copy for documentation:
        self.variables['extra'] = extra
        # actual variables (top level will be loaded later)

        # export the whole data passed as argument, in case of need:
        self._conf = config
        # add a copy to the template variables
        # that copy may be manipulated
        self.variables['config'] = copy(config)
        assert self.variables['config'] is not config

        # load other yaml files
        self._load_yaml()

        # load the standard plugin context
        define_env(self)

        # at this point load the actual variables from extra (YAML file)
        self.variables.update(extra)

        # add variables, functions and filters from the Python module:
        # by design, this MUST be the last step, so that programmers have
        # full control on what happened in the configuration files
        self._load_modules()
        # Provide information:
        debug("Variables:", list(self.variables.keys()))
        if len(extra):
            trace("Extra variables (config file):", list(extra.keys()))
            debug("Content of extra variables (config file):", extra)
        if self.filters:
            trace("Extra filters (module):", list(self.filters.keys()))

        # -------------------
        # Create the jinja2 environment:
        # -------------------
        DOCS_DIR = config.get('docs_dir')
        debug("Docs directory:", DOCS_DIR)
        # define the include directory:
        # NOTE: using DOCS_DIR as default is not ideal,
        # because those files get rendered as well, which is incorrect
        # since they are partials; but we do not want to break existing installs
        include_dir = self.config['include_dir'] or DOCS_DIR
        if not os.path.isdir(include_dir):
            raise FileNotFoundError("MACROS ERROR: Include directory '%s' "
                                    "does not exist!" %
                                    include_dir)
        if self.config['include_dir']:
            trace("Includes directory:", include_dir)
        else:
            debug("Includes directory:", include_dir)
        # get the behavior in case of unknown variable (default: keep)
        on_undefined = self.config['on_undefined']
        if on_undefined not in UNDEFINED_BEHAVIOR:
            raise ValueError("Illegal value for undefined macro parameter '%s'" % on_undefined)
        undefined = UNDEFINED_BEHAVIOR[on_undefined]
        debug("Undefined behavior:", undefined)
        env_config = {
            'loader': FileSystemLoader(include_dir),
            'undefined': undefined
        }
        # read the config variables for jinja2:
        for key, value in self.config.items():
            # take definitions in config_scheme where key starts with 'j2_'
            # (if value is not empty)
            # and forward them to jinja2
            # this is used for the markers
            if key.startswith('j2_') and value:
                variable_name = key.split('_', 1)[1]  # remove prefix
                trace("Found j2 variable '%s': '%s'" %
                      (variable_name, value))
                env_config[variable_name] = value

        # finally build the environment:
        self.env = Environment(**env_config)

        # -------------------
        # Process macros
        # -------------------
        # reference all macros
        self.variables['macros'] = copy(self.macros)
        # add the macros to the environment's global (not to the template!)
        self.env.globals.update(self.macros)

        # -------------------
        # Process filters
        # -------------------
        # reference all filters, for doc [these are copies, so no black magic]
        # NOTE: self.variables is reflected in the list of variables
        #       in the jinja2 environment (same object)
        self.variables['filters'] = copy(self.filters)
        self.variables['filters_builtin'] = copy(self.env.filters)
        # update environment with the custom filters:
        self.env.filters.update(self.filters)

    def on_nav(self, nav, config, files):
        """
        Called after the site navigation is created.
        Capture the nav and files objects so they can be used by
        templates.
        """
        # nav has useful properties like 'pages' and 'items'
        # see: https://github.com/mkdocs/mkdocs/blob/master/mkdocs/structure/nav.py
        self.variables['navigation'] = nav
        # files has collection of files discovered in docs_dir
        # see: https://github.com/mkdocs/mkdocs/blob/master/mkdocs/structure/files.py
        # NOTE: useful for writing macros that check for the existence of files; e.g., a macro to mark a link as disabled, if its target doesn't exist
        self.variables['files'] = files

    def on_serve(self, server, config, **kwargs):
        """
        Called when the serve command is used during development.
        This is to add files or directories to the list of "watched" 
        files for auto-reloading.
        """
        # define directories to add, keep non nulls
        additional = [self.config['include_dir']  # markdown includes
                      ]
        additional = [el for el in additional if el]
        if additional:
            trace("We will also watch:", additional)
        # necessary because of a bug in mkdocs:
        # more information in:
        # https://github.com/mkdocs/mkdocs/issues/1952))
        try:
            builder = list(server.watcher._tasks.values())[0]["func"]
        except AttributeError:
            # change in mkdocs 1.2, see: https://www.mkdocs.org/about/release-notes/#backward-incompatible-changes-in-12
            # this parameter is now optional
            builder = None
        # go ahead and watch
        for el in additional:
            if el:
                server.watch(el, builder)

    def on_page_markdown(self, markdown, page, config,
                         site_navigation=None, **kwargs):
        """
        Pre-rendering for each page of the website.
        It uses the jinja2 directives, together with
        variables, macros and filters, to create pure markdown code.
        """
        # the site_navigation argument has been made optional
        # (deleted in post-1.0 mkdocs, but maintained here
        # for backward compatibility)
        # We REALLY want the same object
        self._page = page
        if not self.variables:
            return markdown
        else:
            # Update the page info in the document
            # page is an object with a number of properties (title, url, ...)
            # see: https://github.com/mkdocs/mkdocs/blob/master/mkdocs/structure/pages.py
            self.variables["page"] = copy(page)
            # execute the pre-macro functions in the various modules
            for func in self.pre_macro_functions:
                func(self)
            # render the macros
            self._raw_markdown = self.render(markdown)
            # execute the post-macro functions in the various modules
            for func in self.post_macro_functions:
                func(self)
            return self.raw_markdown

    def on_post_build(self, config: config_options.Config):
        """
        Hook for post build actions, typically adding
        raw files to the setup.
        """
        # execute the functions in the various modules
        for func in self.post_build_functions:
            func(self)
