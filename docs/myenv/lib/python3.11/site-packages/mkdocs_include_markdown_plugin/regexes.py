"""Regular expressions used in different parts of the plugin."""

from __future__ import annotations

import re
import string


DOUBLE_QUOTED_STR_RE = r'([^"]|(?<=\\)["])+'
SINGLE_QUOTED_STR_RE = r"([^']|(?<=\\)['])+"

# In the following regular expression, the substrings "$OPENING_TAG"
# and "$CLOSING_TAG" will be replaced by the effective opening and
# closing tags in the `on_config` plugin event.
INCLUDE_TAG_RE = rf"""
    (?P<_includer_indent>[ \t\f\v\w{re.escape(string.punctuation)}]*?)$OPENING_TAG
    \s*
    include
    \s+
    (?:"(?P<double_quoted_filename>{DOUBLE_QUOTED_STR_RE})")?(?:'(?P<single_quoted_filename>{SINGLE_QUOTED_STR_RE})')?
    (?P<arguments>.*?)
    \s*
    $CLOSING_TAG
"""  # noqa: E501
