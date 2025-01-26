"""Module where the `on_page_markdown` plugin event is processed."""

from __future__ import annotations

import glob
import html
import logging
import os
import re
import textwrap
from collections.abc import MutableMapping
from typing import TYPE_CHECKING, Any
from urllib.parse import urlparse
import urllib.request

from mkdocs_include_markdown_plugin import process
from mkdocs_include_markdown_plugin.config import (
    CONFIG_DEFAULTS,
    DEFAULT_CLOSING_TAG,
    DEFAULT_COMMENTS,
    DEFAULT_OPENING_TAG,
    create_include_tag,
)
from mkdocs_include_markdown_plugin.files_watcher import FilesWatcher
from mkdocs_include_markdown_plugin.regexes import (
    DOUBLE_QUOTED_STR_RE,
    SINGLE_QUOTED_STR_RE,
)


if TYPE_CHECKING:  # remove this for mypyc compiling
    import sys

    if sys.version_info >= (3, 8):
        from typing import TypedDict
    else:
        from typing_extensions import TypedDict

    from mkdocs.structure.pages import Page

    class DirectiveBoolArgument(TypedDict):  # noqa: D101
        value: bool
        regex: re.Pattern[str]


logger = logging.getLogger('mkdocs.plugins.mkdocs_include_markdown_plugin')

TRUE_FALSE_STR_BOOL = {
    'true': True,
    'false': False,
}

TRUE_FALSE_BOOL_STR = {
    True: 'true',
    False: 'false',
}

def is_url(string):
    try:
        result = urlparse(string)
        return all([result.scheme, result.netloc])
    except ValueError:
        return False

def bool_arg(arg: str) -> re.Pattern[str]:
    """Return a compiled regexp to match a boolean argument."""
    return re.compile(rf'{arg}=(\w+)')


def str_arg(arg: str) -> re.Pattern[str]:
    """Return a compiled regexp to match a string argument."""
    return re.compile(
        rf'{arg}=(?:"({DOUBLE_QUOTED_STR_RE})")?'
        rf"(?:'({SINGLE_QUOTED_STR_RE})')?",
    )


ARGUMENT_REGEXES = {
    'start': str_arg('start'),
    'end': str_arg('end'),
    'exclude': str_arg('exclude'),
    'encoding': str_arg('encoding'),

    # bool
    'comments': bool_arg('comments'),
    'preserve-includer-indent': bool_arg('preserve-includer-indent'),
    'dedent': bool_arg('dedent'),
    'trailing-newlines': bool_arg('trailing-newlines'),
    'rewrite-relative-urls': bool_arg('rewrite-relative-urls'),

    # int
    'heading-offset': re.compile(r'heading-offset=(-?\d+)'),
}


def parse_filename_argument(
        match: re.Match[str],
) -> tuple[str | None, str | None]:
    """Return the filename argument matched by ``match``."""
    raw_filename = match.group('double_quoted_filename')
    if raw_filename is None:
        raw_filename = match.group('single_quoted_filename')
        if raw_filename is None:
            filename = None
        else:
            filename = raw_filename.replace("\\'", "'")
    else:
        filename = raw_filename.replace('\\"', '"')
    return filename, raw_filename


def parse_string_argument(match: re.Match[str]) -> str | None:
    """Return the string argument matched by ``match``."""
    value = match.group(1)
    if value is None:
        value = match.group(3)
        if value is not None:
            value = value.replace("\\'", "'")
    else:
        value = value.replace('\\"', '"')
    return value


def lineno_from_content_start(content: str, start: int) -> int:
    """Return the line number of the first line of ``start`` in ``content``."""
    return content[:start].count('\n') + 1


def read_file(file_path: str, encoding: str) -> str:
    """Read a file and return its content."""
    with open(file_path, encoding=encoding) as f:
        return f.read()

def read_http(target_url: str, encoding: str) -> str:
    """Read an http location and return its content."""
    req = urllib.request.Request(target_url)
    with urllib.request.urlopen(req) as response:
        return response.read().decode('UTF-8')
    
def get_file_content(
    markdown: str,
    page_src_path: str,
    docs_dir: str,
    include_tag_regex: re.Pattern[str],
    include_markdown_tag_regex: re.Pattern[str],
    default_encoding: str,
    default_preserve_includer_indent: bool,
    default_dedent: bool,
    default_trailing_newlines: bool,
    default_comments: bool = DEFAULT_COMMENTS,
    cumulative_heading_offset: int = 0,
    files_watcher: FilesWatcher | None = None,
) -> str:
    """Return the content of the file to include."""
    def found_include_tag(match: re.Match[str]) -> str:
        directive_match_start = match.start()

        _includer_indent = match.group('_includer_indent')

        filename, raw_filename = parse_filename_argument(match)
        if filename is None:
            lineno = lineno_from_content_start(
                markdown,
                directive_match_start,
            )
            logger.error(
                "Found no path passed including with 'include'"
                f' directive at {os.path.relpath(page_src_path, docs_dir)}'
                f':{lineno}',
            )
            return ''

        arguments_string = match.group('arguments')

        if os.path.isabs(filename):
            file_path_glob = filename
        elif is_url(filename):
            file_path_glob = filename
        else:
            file_path_glob = os.path.join(
                os.path.abspath(os.path.dirname(page_src_path)),
                filename,
            )

        exclude_match = ARGUMENT_REGEXES['exclude'].search(arguments_string)
        if exclude_match is None:
            ignore_paths: list[str] = []
        else:
            exclude_string = parse_string_argument(exclude_match)
            if exclude_string is None:
                lineno = lineno_from_content_start(
                    markdown,
                    directive_match_start,
                )
                logger.error(
                    "Invalid empty 'exclude' argument in 'include' directive"
                    f' at {os.path.relpath(page_src_path, docs_dir)}:{lineno}',
                )
                ignore_paths = []
            else:

                if os.path.isabs(exclude_string):
                    exclude_globstr = exclude_string
                else:
                    exclude_globstr = os.path.realpath(
                        os.path.join(
                            os.path.abspath(os.path.dirname(page_src_path)),
                            exclude_string,
                        ),
                    )
                ignore_paths = glob.glob(exclude_globstr)

        file_paths_to_include = process.filter_paths(
            glob.iglob(file_path_glob, recursive=True),
            ignore_paths=ignore_paths,
        )
        if is_url(filename):
            file_paths_to_include=[file_path_glob]
            logger.info('url found: ' + file_path_glob)


        if not file_paths_to_include:
            lineno = lineno_from_content_start(
                markdown,
                directive_match_start,
            )
            logger.error(
                f"No files found including '{raw_filename}'"
                f' at {os.path.relpath(page_src_path, docs_dir)}'
                f':{lineno}',
            )
            return ''
        elif files_watcher is not None:
            if not is_url(file_path_glob):
                files_watcher.included_files.extend(file_paths_to_include)
            else:
                readable_files_to_include = ', '.join(file_paths_to_include)
                plural_suffix = 's' if len(file_paths_to_include) > 1 else ''
                lineno = lineno_from_content_start(
                    markdown,
                    directive_match_start,
                )
                logger.warning(
                    f"Not adding a watcher for {file_path_glob} of 'include-markdown'"
                    f' directive at {os.path.relpath(page_src_path, docs_dir)}'
                    f':{lineno} not detected in the file{plural_suffix}'
                    f' {readable_files_to_include}',
                )

        bool_options: dict[str, DirectiveBoolArgument] = {
            'preserve-includer-indent': {
                'value': default_preserve_includer_indent,
                'regex': ARGUMENT_REGEXES['preserve-includer-indent'],
            },
            'dedent': {
                'value': default_dedent,
                'regex': ARGUMENT_REGEXES['dedent'],
            },
            'trailing-newlines': {
                'value': default_trailing_newlines,
                'regex': ARGUMENT_REGEXES['trailing-newlines'],
            },
        }

        for arg_name, arg in bool_options.items():
            bool_arg_match = arg['regex'].search(arguments_string)
            if bool_arg_match is None:
                continue
            try:
                bool_options[arg_name]['value'] = TRUE_FALSE_STR_BOOL[
                    bool_arg_match.group(
                        1,
                    ) or TRUE_FALSE_BOOL_STR[arg['value']]
                ]
            except KeyError:
                lineno = lineno_from_content_start(
                    markdown,
                    directive_match_start,
                )
                logger.error(
                    f"Invalid value for '{arg_name}' argument of 'include'"
                    f' directive at {os.path.relpath(page_src_path, docs_dir)}'
                    f':{lineno}. Possible values are true or false.',
                )
                return ''

        start_match = ARGUMENT_REGEXES['start'].search(arguments_string)
        if start_match:
            start = parse_string_argument(start_match)
            if start is None:
                lineno = lineno_from_content_start(
                    markdown,
                    directive_match_start,
                )
                logger.error(
                    "Invalid empty 'start' argument in 'include' directive at "
                    f'{os.path.relpath(page_src_path, docs_dir)}:{lineno}',
                )
        else:
            start = None

        end_match = ARGUMENT_REGEXES['end'].search(arguments_string)
        if end_match:
            end = parse_string_argument(end_match)
            if end is None:
                lineno = lineno_from_content_start(
                    markdown,
                    directive_match_start,
                )
                logger.error(
                    "Invalid empty 'end' argument in 'include' directive at "
                    f'{os.path.relpath(page_src_path, docs_dir)}:{lineno}',
                )
        else:
            end = None

        encoding_match = ARGUMENT_REGEXES['encoding'].search(arguments_string)
        if encoding_match:
            encoding = parse_string_argument(encoding_match)
            if encoding is None:
                lineno = lineno_from_content_start(
                    markdown,
                    directive_match_start,
                )
                logger.error(
                    "Invalid empty 'encoding' argument in 'include'"
                    ' directive at '
                    f'{os.path.relpath(page_src_path, docs_dir)}:{lineno}',
                )
                encoding = 'utf-8'
        else:
            encoding = default_encoding

        text_to_include = ''
        expected_but_any_found = [start is not None, end is not None]
        for file_path in file_paths_to_include:
            if is_url(filename):
                new_text_to_include = read_http(file_path, encoding)
            else:
                new_text_to_include = read_file(file_path, encoding)

            if start is not None or end is not None:
                new_text_to_include, *expected_not_found = (
                    process.filter_inclusions(
                        start,
                        end,
                        new_text_to_include,
                    )
                )
                for i in range(2):
                    if expected_but_any_found[i] and not expected_not_found[i]:
                        expected_but_any_found[i] = False

            # nested includes
            new_text_to_include = get_file_content(
                new_text_to_include,
                file_path,
                docs_dir,
                include_tag_regex,
                include_markdown_tag_regex,
                default_encoding,
                default_preserve_includer_indent,
                default_dedent,
                default_trailing_newlines,
                files_watcher=files_watcher,
            )

            # trailing newlines right stripping
            if not bool_options['trailing-newlines']['value']:
                new_text_to_include = process.rstrip_trailing_newlines(
                    new_text_to_include,
                )

            if bool_options['dedent']['value']:
                new_text_to_include = textwrap.dedent(new_text_to_include)

            # includer indentation preservation
            if bool_options['preserve-includer-indent']['value']:
                new_text_to_include = ''.join(
                    _includer_indent + line
                    for line in (
                        new_text_to_include.splitlines(keepends=True)
                        or ['']
                    )
                )
            else:
                new_text_to_include = _includer_indent + new_text_to_include

            text_to_include += new_text_to_include

        # warn if expected start or ends haven't been found in included content        
        for i, argname in enumerate(['start', 'end']):
            if expected_but_any_found[i]:
                value = locals()[argname]
                readable_files_to_include = ', '.join([
                    os.path.relpath(fpath, docs_dir)
                    for fpath in file_paths_to_include
                ])
                plural_suffix = 's' if len(file_paths_to_include) > 1 else ''
                lineno = lineno_from_content_start(
                    markdown,
                    directive_match_start,
                )
                logger.warning(
                    f"Delimiter {argname} '{value}' of 'include'"
                    f' directive at {os.path.relpath(page_src_path, docs_dir)}'
                    f':{lineno} not detected in the file{plural_suffix}'
                    f' {readable_files_to_include}',
                )

        return text_to_include

    def found_include_markdown_tag(match: re.Match[str]) -> str:
        directive_match_start = match.start()

        _includer_indent = match.group('_includer_indent')
        _empty_includer_indent = ' ' * len(_includer_indent)

        filename, raw_filename = parse_filename_argument(match)
        if filename is None:
            lineno = lineno_from_content_start(
                markdown,
                directive_match_start,
            )
            logger.error(
                "Found no path passed including with 'include-markdown'"
                f' directive at {os.path.relpath(page_src_path, docs_dir)}'
                f':{lineno}',
            )
            return ''
        
        arguments_string = match.group('arguments')

        if os.path.isabs(filename):
            file_path_glob = filename
        elif is_url(filename):
            file_path_glob = filename
        else:
            file_path_glob = os.path.join(
                os.path.abspath(os.path.dirname(page_src_path)),
                filename,
            )

        exclude_match = ARGUMENT_REGEXES['exclude'].search(arguments_string)
        if exclude_match is None:
            ignore_paths: list[str] = []
        else:
            exclude_string = parse_string_argument(exclude_match)
            if exclude_string is None:
                lineno = lineno_from_content_start(
                    markdown,
                    directive_match_start,
                )
                logger.error(
                    "Invalid empty 'exclude' argument in 'include-markdown'"
                    f' directive at {os.path.relpath(page_src_path, docs_dir)}'
                    f':{lineno}',
                )
                ignore_paths = []
            else:
                if os.path.isabs(exclude_string):
                    exclude_globstr = exclude_string
                else:
                    exclude_globstr = os.path.realpath(
                        os.path.join(
                            os.path.abspath(os.path.dirname(page_src_path)),
                            exclude_string,
                        ),
                    )
                ignore_paths = glob.glob(exclude_globstr)

        file_paths_to_include = process.filter_paths(
            glob.iglob(file_path_glob, recursive=True),
            ignore_paths=ignore_paths,
        )

        if is_url(filename):
            file_paths_to_include=[file_path_glob]
            logger.info('url found: ' + file_path_glob)

        if not file_paths_to_include:
            lineno = lineno_from_content_start(
                markdown,
                directive_match_start,
            )
            logger.error(
                f"No files found including '{raw_filename}' at"
                f' {os.path.relpath(page_src_path, docs_dir)}'
                f':{lineno}',
            )
            return ''
        elif files_watcher is not None:
            if not is_url(file_path_glob):
                files_watcher.included_files.extend(file_paths_to_include)
            else:
                readable_files_to_include = ', '.join(file_paths_to_include)
                plural_suffix = 's' if len(file_paths_to_include) > 1 else ''
                lineno = lineno_from_content_start(
                    markdown,
                    directive_match_start,
                )
                logger.warning(
                    f"Not adding a watcher for {file_path_glob} of 'include-markdown'"
                    f' directive at {os.path.relpath(page_src_path, docs_dir)}'
                    f':{lineno} not detected in the file{plural_suffix}'
                    f' {readable_files_to_include}',
                )

        bool_options: dict[str, DirectiveBoolArgument] = {
            'rewrite-relative-urls': {
                'value': True,
                'regex': ARGUMENT_REGEXES['rewrite-relative-urls'],
            },
            'comments': {
                'value': default_comments,
                'regex': ARGUMENT_REGEXES['comments'],
            },
            'preserve-includer-indent': {
                'value': default_preserve_includer_indent,
                'regex': ARGUMENT_REGEXES['preserve-includer-indent'],
            },
            'dedent': {
                'value': default_dedent,
                'regex': ARGUMENT_REGEXES['dedent'],
            },
            'trailing-newlines': {
                'value': default_trailing_newlines,
                'regex': ARGUMENT_REGEXES['trailing-newlines'],
            },
        }

        for arg_name, arg in bool_options.items():
            bool_arg_match = arg['regex'].search(arguments_string)
            if bool_arg_match is None:
                continue
            try:
                bool_options[arg_name]['value'] = TRUE_FALSE_STR_BOOL[
                    bool_arg_match.group(
                        1,
                    ) or TRUE_FALSE_BOOL_STR[arg['value']]
                ]
            except KeyError:
                lineno = lineno_from_content_start(
                    markdown,
                    directive_match_start,
                )
                logger.error(
                    f"Invalid value for '{arg_name}' argument of"
                    " 'include-markdown' directive at"
                    f' {os.path.relpath(page_src_path, docs_dir)}'
                    f':{lineno}. Possible values are true or false.',
                )
                return ''

        # start and end arguments
        start_match = ARGUMENT_REGEXES['start'].search(arguments_string)
        if start_match:
            start = parse_string_argument(start_match)
            if start is None:
                lineno = lineno_from_content_start(
                    markdown,
                    directive_match_start,
                )
                logger.error(
                    "Invalid empty 'start' argument in 'include-markdown'"
                    f' directive at {os.path.relpath(page_src_path, docs_dir)}'
                    f':{lineno}',
                )
        else:
            start = None

        end_match = ARGUMENT_REGEXES['end'].search(arguments_string)
        if end_match:
            end = parse_string_argument(end_match)
            if end is None:
                lineno = lineno_from_content_start(
                    markdown,
                    directive_match_start,
                )
                logger.error(
                    "Invalid empty 'end' argument in 'include-markdown'"
                    f' directive at {os.path.relpath(page_src_path, docs_dir)}'
                    f':{lineno}',
                )
        else:
            end = None

        encoding_match = ARGUMENT_REGEXES['encoding'].search(arguments_string)
        if encoding_match:
            encoding = parse_string_argument(encoding_match)
            if encoding is None:
                lineno = lineno_from_content_start(
                    markdown,
                    directive_match_start,
                )
                logger.error(
                    "Invalid empty 'encoding' argument in 'include-markdown'"
                    ' directive at '
                    f'{os.path.relpath(page_src_path, docs_dir)}:{lineno}',
                )
                encoding = 'utf-8'
        else:
            encoding = default_encoding

        # heading offset
        offset = 0
        offset_match = ARGUMENT_REGEXES['heading-offset'].search(
            arguments_string,
        )
        if offset_match:
            offset += int(offset_match.group(1))

        separator = '\n' if bool_options['trailing-newlines']['value'] else ''
        if not start and not end:
            start_end_part = ''
        else:
            start_end_part = f"'{html.escape(start)}' " if start else "'' "
            start_end_part += f"'{html.escape(end)}' " if end else "'' "

        # if any start or end strings are found in the included content
        # but the arguments are specified, we must raise a warning
        #
        # `True` means that no start/end strings have been found in content
        # but they have been specified, so the warning(s) must be raised
        expected_but_any_found = [start is not None, end is not None]

        text_to_include = ''
        for file_path in file_paths_to_include:
            if is_url(filename):
                new_text_to_include = read_http(file_path, encoding)
            else:
                new_text_to_include = read_file(file_path, encoding)

            if start is not None or end is not None:
                new_text_to_include, *expected_not_found = (
                    process.filter_inclusions(
                        start,
                        end,
                        new_text_to_include,
                    )
                )
                for i in range(2):
                    if expected_but_any_found[i] and not expected_not_found[i]:
                        expected_but_any_found[i] = False

            # nested includes
            new_text_to_include = get_file_content(
                new_text_to_include,
                file_path,
                docs_dir,
                include_tag_regex,
                include_markdown_tag_regex,
                default_encoding,
                default_preserve_includer_indent,
                default_dedent,
                default_trailing_newlines,
                default_comments=default_comments,
                files_watcher=files_watcher,
            )

            # trailing newlines right stripping
            if not bool_options['trailing-newlines']['value']:
                new_text_to_include = process.rstrip_trailing_newlines(
                    new_text_to_include,
                )

            # relative URLs rewriting
            if bool_options['rewrite-relative-urls']['value']:
                new_text_to_include = process.rewrite_relative_urls(
                    new_text_to_include,
                    source_path=file_path,
                    destination_path=page_src_path,
                )

            # comments
            if bool_options['comments']['value']:
                new_text_to_include = (
                    f'{_includer_indent}'
                    f'<!-- BEGIN INCLUDE {html.escape(filename)}'
                    f' {start_end_part}-->{separator}{new_text_to_include}'
                    f'{separator}<!-- END INCLUDE -->'
                )
            else:
                new_text_to_include = (
                    f'{_includer_indent}{new_text_to_include}'
                )

            # dedent
            if bool_options['dedent']['value']:
                new_text_to_include = textwrap.dedent(new_text_to_include)

            # includer indentation preservation
            if bool_options['preserve-includer-indent']['value']:
                new_text_to_include = ''.join(
                    (_empty_includer_indent if i > 0 else '') + line
                    for i, line in enumerate(
                        new_text_to_include.splitlines(keepends=True)
                        or [''],
                    )
                )

            if offset_match:
                new_text_to_include = process.increase_headings_offset(
                    new_text_to_include,
                    offset=offset + cumulative_heading_offset,
                )

            text_to_include += new_text_to_include

        # warn if expected start or ends haven't been found in included content
        for i, argname in enumerate(['start', 'end']):
            if expected_but_any_found[i]:
                value = locals()[argname]
                readable_files_to_include = ', '.join([
                    os.path.relpath(fpath, docs_dir)
                    for fpath in file_paths_to_include
                ])
                plural_suffix = 's' if len(file_paths_to_include) > 1 else ''
                lineno = lineno_from_content_start(
                    markdown,
                    directive_match_start,
                )
                logger.warning(
                    f"Delimiter {argname} '{value}' of 'include-markdown'"
                    f' directive at {os.path.relpath(page_src_path, docs_dir)}'
                    f':{lineno} not detected in the file{plural_suffix}'
                    f' {readable_files_to_include}',
                )

        return text_to_include

    markdown = include_tag_regex.sub(
        found_include_tag,
        markdown,
    )
    markdown = include_markdown_tag_regex.sub(
        found_include_markdown_tag,
        markdown,
    )
    return markdown


def on_page_markdown(
    markdown: str,
    page: Page,
    docs_dir: str,
    config: MutableMapping[str, Any] | None = None,
    files_watcher: FilesWatcher | None = None,
) -> str:
    """Process markdown content of a page."""
    if config is None:
        config = {}

    return get_file_content(
        markdown,
        page.file.abs_src_path,
        docs_dir,
        config.get(
            '_include_tag',
            create_include_tag(
                config.get('opening_tag', DEFAULT_OPENING_TAG),
                config.get('closing_tag', DEFAULT_CLOSING_TAG),
            ),
        ),
        config.get(
            '_include_markdown_tag',
            create_include_tag(
                config.get('opening_tag', DEFAULT_OPENING_TAG),
                config.get('closing_tag', DEFAULT_CLOSING_TAG),
                tag='include-markdown',
            ),
        ),
        config.get('encoding', CONFIG_DEFAULTS['encoding']),
        config.get(
            'preserve_includer_indent',
            CONFIG_DEFAULTS['preserve_includer_indent'],
        ),
        config.get('dedent', CONFIG_DEFAULTS['dedent']),
        config.get('trailing_newlines', CONFIG_DEFAULTS['trailing_newlines']),
        default_comments=config.get('comments', CONFIG_DEFAULTS['comments']),
        files_watcher=files_watcher,
    )
