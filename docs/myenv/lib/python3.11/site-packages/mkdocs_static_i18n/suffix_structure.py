import logging
import os
from pathlib import Path
from re import compile
from urllib.parse import quote as urlquote

from mkdocs import utils
from mkdocs.structure.files import File, Files
from mkdocs.structure.nav import get_navigation

RE_LOCALE = compile(r"(^[a-z]{2}_[A-Z]{2}$)|(^[a-z]{2}$)")

log = logging.getLogger("mkdocs.plugins." + __name__)


class I18nFiles(Files):
    """
    This class extends MkDocs' Files class to support links and assets that
    have a translated locale suffix.

    Since MkDocs relies on the file.src_path of pages and assets we have to
    derive the file.src_path and check for a possible .<locale>.<suffix> file
    to use instead of the link / asset referenced in the markdown source.
    """

    locale = None
    translated = False

    def append(self, file):
        """
        Since i18nFile find their own language versions, we need to avoid adding
        them multiple times when a localized version of a file is considered.

        The first i18nFile is sufficient to cover all their possible localized versions.
        """
        for inside_file in self:
            if inside_file.dest_path == file.dest_path:
                return
        super().append(file)

    def __contains__(self, path):
        """
        Return a bool stipulating whether or not we found a translated version
        of the given path or the path itself.

        Since our plugin automatically localize links, this is useful for the
        mkdocs.structure.pages / path_to_url() method to point to the localized
        version of the file, if present.
        """
        expected_src_path = Path(path)
        expected_src_paths = [
            expected_src_path.with_suffix(f".{self.locale}{expected_src_path.suffix}"),
            expected_src_path.with_suffix(
                f".{self.default_locale}{expected_src_path.suffix}"
            ),
            expected_src_path,
        ]
        return any(filter(lambda s: Path(s) in expected_src_paths, self.src_paths))

    def get_file_from_path(self, path):
        """Return a File instance with File.src_path equal to path."""
        expected_src_path = Path(path)
        expected_src_paths = [
            expected_src_path.with_suffix(f".{self.locale}{expected_src_path.suffix}"),
            expected_src_path.with_suffix(
                f".{self.default_locale}{expected_src_path.suffix}"
            ),
            expected_src_path,
        ]
        for src_path in filter(lambda s: Path(s) in expected_src_paths, self.src_paths):
            return self.src_paths.get(os.path.normpath(src_path))

    def get_localized_page_from_url(self, url, language):
        """Return the I18nFile instance from our files that match the given url and language"""
        if language:
            url = f"{language}/{url}"
        url = url.rstrip(".") or "."
        # TODO: when we bump to mkdocs > 1.4 we can (finally) normalize
        # the url convention again
        if url.endswith("/./"):
            url = url[:-2]
        for file in self:
            if not file.is_documentation_page():
                continue
            if file.url == url:
                return file


class I18nFile(File):
    """
    This is a i18n aware version of a mkdocs.structure.files.File
    """

    def __init__(
        self,
        file_from,
        language,
        all_languages=None,
        default_language=None,
        docs_dir=None,
        site_dir=None,
        use_directory_urls=None,
    ) -> None:
        # preserved from mkdocs.structure.files.File
        # since they are not calculated
        self.page = file_from.page
        self.docs_dir = docs_dir
        self.site_dir = site_dir

        # i18n addons
        self.all_languages = all_languages
        self.alternates = {lang: None for lang in self.all_languages}
        self.default_language = default_language
        self.dest_language = language
        self.initial_abs_dest_path = file_from.abs_dest_path
        self.initial_abs_src_path = file_from.abs_src_path
        self.initial_dest_path = file_from.dest_path
        self.initial_src_path = file_from.src_path
        self.locale_suffix = None

        # the name without any suffix
        self.name = self._get_name()

        # find src_path
        expected_paths = [
            (
                language,
                Path(docs_dir)
                / Path(f"{self.non_i18n_src_path}.{language}{self.suffix}"),
            ),
            (
                default_language,
                Path(docs_dir)
                / Path(f"{self.non_i18n_src_path}.{default_language}{self.suffix}"),
            ),
            (None, Path(docs_dir) / Path(f"{self.non_i18n_src_path}{self.suffix}")),
        ]
        for locale_suffix, expected_path in expected_paths:
            if Path(expected_path).exists():

                self.src_path = expected_path.relative_to(self.docs_dir)
                self.abs_src_path = Path(self.docs_dir) / Path(self.src_path)
                #
                self.locale_suffix = locale_suffix
                if self.locale_suffix:
                    self.dest_name = Path(self.name).with_suffix(
                        Path(self.name).suffix + self.suffix
                    )
                else:
                    self.dest_name = Path(expected_path).name
                #
                self.dest_path = self._get_dest_path(use_directory_urls)
                self.abs_dest_path = (
                    Path(self.site_dir)
                    / Path(self.dest_language)
                    / Path(self.dest_path)
                )
                break
        else:
            self.src_path = file_from.src_path
            self.abs_src_path = file_from.abs_src_path
            #
            self.dest_path = file_from.dest_path
            self.abs_dest_path = file_from.abs_dest_path
            #
            self.dest_name = self.name

        # set url
        self.url = self._get_url(use_directory_urls)

        # set ourself as our own alternate
        self.alternates[self.dest_language or self.default_language] = self

        # mkdocs expects strings for those
        self.abs_dest_path = str(self.abs_dest_path)
        self.abs_src_path = str(self.abs_src_path)
        self.dest_path = str(self.dest_path)
        self.src_path = str(self.src_path)

    def __repr__(self):
        return (
            f"I18nFile(src_path='{self.src_path}', abs_src_path='{self.abs_src_path}',"
            f" dest_path='{self.dest_path}', abs_dest_path='{self.abs_dest_path}',"
            f" name='{self.name}', locale_suffix='{self.locale_suffix}',"
            f" dest_language='{self.dest_language}', dest_name='{self.dest_name}',"
            f" url='{self.url}')"
        )

    @property
    def non_i18n_src_path(self):
        """
        Return the path of the given page without any suffix.
        """
        if self._is_localized() is None:
            non_i18n_src_path = Path(self.initial_src_path).with_suffix("")
        else:
            non_i18n_src_path = (
                Path(self.initial_src_path).with_suffix("").with_suffix("")
            )
        return non_i18n_src_path

    def _is_localized(self):
        """
        Returns the locale detected in the file's suffixes <name>.<locale>.<suffix>.
        """
        for language in self.all_languages:
            initial_file_suffixes = Path(self.initial_src_path).suffixes
            expected_suffixes = [f".{language}", Path(self.initial_src_path).suffix]
            if len(initial_file_suffixes) >= len(expected_suffixes):
                if (
                    # fmt: off
                    initial_file_suffixes[-len(expected_suffixes):]
                    == expected_suffixes
                ):
                    return language
        return None

    @property
    def suffix(self):
        return Path(self.initial_src_path).suffix

    def _get_name(self):
        """Return the name of the file without it's extension."""
        return (
            "index"
            if self.non_i18n_src_path.name in ("index", "README")
            else self.non_i18n_src_path.name
        )

    def _get_dest_path(self, use_directory_urls):
        """Return destination path based on source path."""
        parent, _ = os.path.split(self.src_path)
        if self.is_documentation_page():
            if use_directory_urls is False or self.name == "index":
                # index.md or README.md => index.html
                # foo.md => foo.html
                return os.path.join(parent, self.name + ".html")
            else:
                # foo.md => foo/index.html
                return os.path.join(parent, self.name, "index.html")
        else:
            return os.path.join(parent, self.dest_name)

    def _get_url(self, use_directory_urls):
        """Return url based in destination path."""
        url = str(self.dest_path).replace(os.path.sep, "/")
        dirname, filename = os.path.split(url)
        if use_directory_urls and filename == "index.html":
            if dirname == "":
                url = "./"
            else:
                url = dirname + "/"
        if self.dest_language:
            if url in [".", "./"]:
                url = self.dest_language + "/"
            else:
                url = self.dest_language + "/" + url
        return urlquote(url)

    def url_relative_to(self, other):
        """Return url for file relative to other i18n file."""
        return utils.get_relative_url(
            self.url,
            other.url
            if (isinstance(other, File) or isinstance(other, I18nFile))
            else other,
        )


def on_files(self, files, config):
    """"""
    main_files = I18nFiles([])
    main_files.default_locale = self.default_language
    main_files.locale = self.default_language
    for language in self.all_languages:
        self.i18n_files[language] = I18nFiles([])
        self.i18n_files[language].default_locale = self.default_language
        self.i18n_files[language].locale = language

    for fileobj in files:

        main_i18n_file = I18nFile(
            fileobj,
            "",
            all_languages=self.all_languages,
            default_language=self.default_language,
            docs_dir=config["docs_dir"],
            site_dir=config["site_dir"],
            use_directory_urls=config.get("use_directory_urls"),
        )
        if (
            self.default_language_options is not None
            and self.default_language_options["build"] is True
        ):
            main_files.append(main_i18n_file)

        for language in self.all_languages:
            i18n_file = I18nFile(
                fileobj,
                language,
                all_languages=self.all_languages,
                default_language=self.default_language,
                docs_dir=config["docs_dir"],
                site_dir=config["site_dir"],
                use_directory_urls=config.get("use_directory_urls"),
            )
            # this 'append' method is reimplemented in I18nFiles to avoid duplicates
            self.i18n_files[language].append(i18n_file)
            if (
                main_i18n_file.is_documentation_page()
                and language != self.default_language
                and main_i18n_file.src_path == i18n_file.src_path
            ):
                log.debug(
                    f"file {main_i18n_file.src_path} is missing translation in '{language}'"
                )

    # these comments are here to help me debug later if needed
    # print([{p.src_path: p.url} for p in main_files.documentation_pages()])
    # print([{p.src_path: p.url} for p in self.i18n_files["en"].documentation_pages()])
    # print([{p.src_path: p.url} for p in self.i18n_files["fr"].documentation_pages()])
    # print([{p.src_path: p.url} for p in main_files.static_pages()])
    # print([{p.src_path: p.url} for p in self.i18n_files["en"].static_pages()])
    # print([{p.src_path: p.url} for p in self.i18n_files["fr"].static_pages()])

    # populate pages alternates
    # main default version
    for page in main_files.documentation_pages():
        for language in self.all_languages:
            # do not list languages not being built as alternates
            if self.config["languages"].get(language, {}).get("build", False) is False:
                continue
            alternate = self.i18n_files[language].get_localized_page_from_url(
                page.url, language
            )
            if alternate:
                page.alternates[language] = alternate
            else:
                log.warning(
                    f"could not find '{language}' alternate for the default version of page '{page.src_path}'"
                )
    # localized versions
    # for files in self.i18n_files.values():
    #     for page in files.documentation_pages():
    #         url = page.url
    #         if url.startswith(f"{files.locale}/"):
    #             url = url.replace(f"{files.locale}/", "", 1) or "."
    #         for language in self.all_languages:
    #             alternate = self.i18n_files[language].get_localized_page_from_url(
    #                 url, language
    #             )
    #             if not alternate:
    #                 page.alternates[
    #                     language
    #                 ] = main_files.get_localized_page_from_url(url, "")
    #             if alternate:
    #                 page.alternates[language] = alternate
    #             else:
    #                 log.warning(
    #                     f"could not find '{language}' alternate for the '{files.locale}' version of page '{page.src_path}'"
    #                 )

    return main_files


def on_nav(self, nav, config, files):
    """ """
    # translate default nav, see #113
    if self._maybe_translate_titles(self.default_language, nav):
        log.info(f"Translated default navigation to {self.default_language}")

    for language, lang_config in self.config["languages"].items():
        # skip nav generation for languages that we do not build
        if lang_config["build"] is False:
            continue
        if self.i18n_configs[language]["nav"]:
            self._fix_config_navigation(language, self.i18n_files[language])

        self.i18n_navs[language] = get_navigation(
            self.i18n_files[language], self.i18n_configs[language]
        )

        # If awesome-pages is used, we want to use it to structurate our
        # localized navigations as well
        if "awesome-pages" in config["plugins"]:
            self.i18n_navs[language] = config["plugins"]["awesome-pages"].on_nav(
                self.i18n_navs[language],
                config=self.i18n_configs[language],
                files=self.i18n_files[language],
            )

        if self.config["nav_translations"].get(language, {}):
            if self._maybe_translate_titles(language, self.i18n_navs[language]):
                log.info(f"Translated navigation to {language}")

        # detect and set nav homepage
        for page in self.i18n_files[language].documentation_pages():
            if page.url in (f"{language}/", f"{language}/index.html"):
                self.i18n_navs[language].homepage = page
                break
        else:
            log.warning(f"could not find homepage for locale '{language}'")

    return nav
