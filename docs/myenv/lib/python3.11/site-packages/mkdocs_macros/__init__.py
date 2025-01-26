# -------------------
# These can be imported in macro code
# -------------------

# from .plugin import MacrosPlugin
# for fixing URLS in macros
from .context import fix_url, is_relative as is_relative_url
from .util import SuperDict