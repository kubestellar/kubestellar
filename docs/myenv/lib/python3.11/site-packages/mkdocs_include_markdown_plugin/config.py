"""Plugin configuration."""

from __future__ import annotations

import re

from mkdocs.config.config_options import Type as MkType

from mkdocs_include_markdown_plugin.regexes import INCLUDE_TAG_RE


DEFAULT_COMMENTS = True
DEFAULT_OPENING_TAG = '{%'
DEFAULT_CLOSING_TAG = '%}'

CONFIG_DEFAULTS = {
    'opening_tag': DEFAULT_OPENING_TAG,
    'closing_tag': DEFAULT_CLOSING_TAG,
    'encoding': 'utf-8',
    'preserve_includer_indent': True,
    'dedent': False,
    'trailing_newlines': True,
    'comments': DEFAULT_COMMENTS,
}

CONFIG_SCHEME = (
    (
        'opening_tag',
        MkType(str, default=DEFAULT_OPENING_TAG),
    ),
    (
        'closing_tag',
        MkType(str, default=DEFAULT_CLOSING_TAG),
    ),
    (
        'encoding',
        MkType(str, default=CONFIG_DEFAULTS['encoding']),
    ),
    (
        'preserve_includer_indent',
        MkType(bool, default=CONFIG_DEFAULTS['preserve_includer_indent']),
    ),
    (
        'dedent',
        MkType(bool, default=CONFIG_DEFAULTS['dedent']),
    ),
    (
        'trailing_newlines',
        MkType(bool, default=CONFIG_DEFAULTS['trailing_newlines']),
    ),
    (
        'comments',
        MkType(bool, default=DEFAULT_COMMENTS),
    ),
)


def create_include_tag(
    opening_tag: str, closing_tag: str, tag: str = 'include',
) -> re.Pattern[str]:
    """Create a regex pattern to match an inclusion tag directive.

    Replaces the substrings '$OPENING_TAG' and '$CLOSING_TAG' from
    INCLUDE_TAG_REGEX by the effective tag.
    """
    return re.compile(
        INCLUDE_TAG_RE.replace(' include', f' {tag}').replace(
            '$OPENING_TAG', re.escape(opening_tag),
        ).replace('$CLOSING_TAG', re.escape(closing_tag)),
        flags=re.VERBOSE | re.DOTALL,
    )
