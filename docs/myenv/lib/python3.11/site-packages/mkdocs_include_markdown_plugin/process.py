"""Utilities for string processing."""

from __future__ import annotations

import functools
import io
import os
import re
from collections.abc import Callable, Iterator
from urllib.parse import urlparse, urlunparse


# Markdown regular expressions. Taken from the original Markdown.pl by John
# Gruber, and modified to work in Python

# Matches markdown links.
# e.g. [scikit-learn](https://github.com/scikit-learn/scikit-learn)
#
# The next Regex can raise a catastrophic backtracking, but with the current
# implementation of the plugin it is not very much likely to reach the case.
# Can be checked with dlint:
# python3 -m dlint.redos --pattern '\[(?:(?:\[[^\[\]]+\])*)?\]'
#
# In the original Markdown.pl, the nested brackets are enclosed by an atomic
# group (?>...), but atomic groups are not supported by Python in versions
# previous to Python3.11. Also, these nested brackets can be recursive in the
# Perl implementation but this doesn't seem possible in Python, the current
# implementation only reaches two levels.
MARKDOWN_LINK_REGEX = re.compile(  # noqa: DUO138
    r"""
        (                 # wrap whole match in $1
          (?<!!)          # don't match images - negative lookbehind
          \[
            (             # link text = $2
                (?:
                    [^\[\]]+  # not bracket
                    (?:
                        \[[^\[\]]+\]  # another level of nested bracket
                                      # with something inside
                        [^\[\]]*      # not bracket
                    )*
                )?        # allow for empty link text
            )
          \]
          \(             # literal paren
            [ \t]*
            <?(.*?)>?    # href = $3
            [ \t]*
            (            # $4
              (['"])     # quote char = $5
              (.*?)      # Title = $6
              \5         # matching quote
            )?           # title is optional
          \)
        )
    """,
    flags=re.VERBOSE,
)

# Matches markdown inline images.
# e.g. ![alt-text](path/to/image.png)
MARKDOWN_IMAGE_REGEX = re.compile(
    r"""
        (                # wrap whole match in $1
          !\[
            (.*?)        # alt text = $2
          \]
          \(             # literal paren
            [ \t]*
            <?(\S+?)>?   # src url = $3
            [ \t]*
            (            # $4
              (['"])     # quote char = $5
              (.*?)      # title = $6
              \5         # matching quote
              [ \t]*
            )?           # title is optional
          \)
        )
    """,
    flags=re.VERBOSE,
)

# Matches markdown link definitions.
# e.g. [scikit-learn]: https://github.com/scikit-learn/scikit-learn
MARKDOWN_LINK_DEFINITION_REGEX = re.compile(
    r"""
        ^[ ]{0,4}\[(.+)\]:   # id = $1
        [ \t]*
        \n?                # maybe *one* newline
        [ \t]*
        <?(\S+?)>?           # url = $2
        [ \t]*
        \n?                # maybe one newline
        [ \t]*
        (?:
            (?<=\s)          # lookbehind for whitespace
            ["(]
            (.+?)            # title = $3
            [")]
            [ \t]*
        )?                   # title is optional
        (?:\n+|\Z)
    """,
    flags=re.VERBOSE | re.MULTILINE,
)


def transform_p_by_p_skipping_codeblocks(
        markdown: str,
        func: Callable[[str], str],
) -> str:
    """Apply a transformation paragraph by paragraph in a Markdown text.

    Apply a transformation paragraph by paragraph in a Markdown using a
    function. Skip indented and fenced codeblock lines, where the
    transformation is never applied.
    """
    # current fenced codeblock delimiter
    _current_fcodeblock_delimiter = ''

    # inside indented codeblock
    _inside_icodeblock = False

    lines, current_paragraph = ([], '')

    def process_current_paragraph() -> None:
        lines.extend(func(current_paragraph).splitlines(keepends=True))

    for line in io.StringIO(markdown):
        if not _current_fcodeblock_delimiter and not _inside_icodeblock:
            lstripped_line = line.lstrip()
            if (
                lstripped_line.startswith('```')
                or lstripped_line.startswith('~~~')
            ):
                _current_fcodeblock_delimiter = lstripped_line[:3]
                if current_paragraph:
                    process_current_paragraph()
                    current_paragraph = ''
                lines.append(line)
            elif (
                line.replace('\t', '    ').replace('\r\n', '\n')
                == '    \n'
            ):
                _inside_icodeblock = True
                if current_paragraph:
                    process_current_paragraph()
                    current_paragraph = ''
                lines.append(line)
            else:
                current_paragraph += line
        else:
            lines.append(line)
            if _current_fcodeblock_delimiter:
                if line.lstrip().startswith(_current_fcodeblock_delimiter):
                    _current_fcodeblock_delimiter = ''
            else:
                if not line.startswith('    ') and not line.startswith('\t'):
                    _inside_icodeblock = False

    process_current_paragraph()

    return ''.join(lines)


def transform_line_by_line_skipping_codeblocks(
        markdown: str,
        func: Callable[[str], str],
) -> str:
    """Apply a transformation line by line in a Markdown text using a function.

    Skip fenced codeblock lines, where the transformation never is applied.

    Indented codeblocks are not taken into account because in the practice
    this function is never used for transformations on indented lines. See
    the PR https://github.com/mondeja/mkdocs-include-markdown-plugin/pull/95
    to recover the implementation handling indented codeblocks.
    """
    # current fenced codeblock delimiter
    _current_fcodeblock_delimiter = ''

    lines = []
    for line in io.StringIO(markdown):
        if not _current_fcodeblock_delimiter:
            lstripped_line = line.lstrip()
            if (
                lstripped_line.startswith('```')
                or lstripped_line.startswith('~~~')
            ):
                _current_fcodeblock_delimiter = lstripped_line[:3]
            else:
                line = func(line)
        elif line.lstrip().startswith(_current_fcodeblock_delimiter):
            _current_fcodeblock_delimiter = ''
        lines.append(line)

    return ''.join(lines)


def rewrite_relative_urls(
    markdown: str, source_path: str, destination_path: str,
) -> str:
    """Rewrite relative URLs in a Markdown text.

    Rewrites markdown so that relative links that were written at
    ``source_path`` will still work when inserted into a file at
    ``destination_path``.
    """
    def rewrite_url(url: str) -> str:
        scheme, netloc, path, params, query, fragment = urlparse(url)

        # absolute or mail
        if path.startswith('/') or scheme == 'mailto':
            return url

        trailing_slash = path.endswith('/')

        path = os.path.relpath(
            os.path.join(os.path.dirname(source_path), path),
            os.path.dirname(destination_path),
        )

        # ensure forward slashes are used, on Windows
        path = path.replace('\\', '/').replace('//', '/')

        if trailing_slash:
            # the above operation removes a trailing slash. Add it back if it
            # was present in the input
            path = path + '/'

        return urlunparse((scheme, netloc, path, params, query, fragment))

    def found_href(m: re.Match[str], url_group_index: int = -1) -> str:
        match_start, match_end = m.span(0)
        href = m.group(url_group_index)
        href_start, href_end = m.span(url_group_index)
        rewritten_url = rewrite_url(href)
        return (
            m.string[match_start:href_start]
            + rewritten_url
            + m.string[href_end:match_end]
        )

    found_href_url_group_index_3 = functools.partial(
        found_href,
        url_group_index=3,
    )

    def transform(paragraph: str) -> str:
        paragraph = MARKDOWN_LINK_REGEX.sub(
            found_href_url_group_index_3,
            paragraph,
        )
        paragraph = MARKDOWN_IMAGE_REGEX.sub(
            found_href_url_group_index_3,
            paragraph,
        )
        return MARKDOWN_LINK_DEFINITION_REGEX.sub(
            functools.partial(found_href, url_group_index=2),
            paragraph,
        )
    return transform_p_by_p_skipping_codeblocks(
        markdown,
        transform,
    )


def interpret_escapes(value: str) -> str:
    """Interpret Python literal escapes in a string.

    Replaces any standard escape sequences in value with their usual
    meanings as in ordinary Python string literals.
    """
    return value.encode('latin-1', 'backslashreplace').decode('unicode_escape')


def filter_inclusions(
    new_start: str | None,
    new_end: str | None,
    text_to_include: str,
) -> tuple[str, bool, bool]:
    """Filter inclusions in a text.

    Manages inclusions from files using ``start`` and ``end`` directive
    arguments.
    """
    expected_start_not_found, expected_end_not_found = (False, False)

    if new_start is not None:
        start = interpret_escapes(new_start)
        end = interpret_escapes(new_end) if new_end is not None else None

        new_text_to_include = ''

        if end is not None:
            end_found = False
            start_split = text_to_include.split(start)[1:]
            if not start_split:
                expected_start_not_found = True
            else:
                for start_text in start_split:
                    for i, end_text in enumerate(start_text.split(end)):
                        if not i % 2:
                            new_text_to_include += end_text
                            end_found = True
            if not end_found:
                expected_end_not_found = True
        else:
            if start in text_to_include:
                new_text_to_include = text_to_include.split(
                    start,
                    maxsplit=1,
                )[1]
            else:
                expected_start_not_found = True
        text_to_include = new_text_to_include

    elif new_end is not None:
        end = interpret_escapes(new_end)
        if end in text_to_include:
            text_to_include = text_to_include.split(
                end,
                maxsplit=1,
            )[0]
        else:
            expected_end_not_found = True

    return (
        text_to_include,
        expected_start_not_found,
        expected_end_not_found,
    )


def _transform_negative_offset_func_factory(
        offset: int,
) -> Callable[[str], str]:
    heading_prefix = '#' * abs(offset)
    return lambda line: line if not line.startswith('#') else (
        heading_prefix + line.lstrip('#')
        if line.startswith(heading_prefix)
        else '#' + line.lstrip('#')
    )


def _transform_positive_offset_func_factory(
        offset: int,
) -> Callable[[str], str]:
    heading_prefix = '#' * offset
    return lambda line: (
        heading_prefix + line if line.startswith('#') else line
    )


def increase_headings_offset(markdown: str, offset: int = 0) -> str:
    """Increases the headings depth of a snippet of Makdown content."""
    if not offset:
        return markdown
    return transform_line_by_line_skipping_codeblocks(
        markdown,
        _transform_positive_offset_func_factory(offset) if offset > 0
        else _transform_negative_offset_func_factory(offset),
    )


def rstrip_trailing_newlines(content: str) -> str:
    """Removes trailing newlines from a string."""
    while content.endswith('\n') or content.endswith('\r'):
        content = content.rstrip('\r\n')
    return content


def filter_paths(
        filepaths: Iterator[str],
        ignore_paths: list[str] | None = None,
) -> list[str]:
    """Filters a list of paths removing those defined in other list of paths.

    The paths to filter can be defined in the list of paths to ignore in
    several forms:

    - The same string.
    - Only the file name.
    - Only their direct directory name.
    - Their direct directory full path.

    Args:
        filepaths (list): Set of source paths to filter.
        ignore_paths (list): Paths that must not be included in the response.

    Returns:
        list: Non filtered paths ordered alphabetically.
    """
    if ignore_paths is None:
        ignore_paths = []

    response = []
    for filepath in filepaths:
        # ignore by filepath
        if filepath in ignore_paths:
            continue

        # ignore by dirpath (relative or absolute)
        if (os.sep).join(filepath.split(os.sep)[:-1]) in ignore_paths:
            continue

        # ignore if is a directory
        if not os.path.isdir(filepath):
            response.append(filepath)
    response.sort()
    return response
