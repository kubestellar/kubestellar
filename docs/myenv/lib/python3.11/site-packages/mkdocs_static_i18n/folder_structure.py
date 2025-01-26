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


class I18nFolderFiles(Files):
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
        Since I18nFolderFile find their own language versions, we need to avoid adding
        them multiple times when a localized version of a file is considered.

        The first I18nFolderFile is sufficient to cover all their possible localized versions.
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
        return True if self.get_file_from_path(path) else False

    def get_file_from_path(self, path):
        """Return a File instance with File.src_path equal to path."""
        expected_src_path = Path(path)
        root_folder = expected_src_path.parts[0]
        expected_src_paths = [
            expected_src_path,
            expected_src_path.relative_to(root_folder),
            Path(self.locale) / Path(expected_src_path),
        ]
        for src_path in filter(lambda s: Path(s) in expected_src_paths, self.src_paths):
            return self.src_paths.get(os.path.normpath(src_path))

    def get_localized_page_from_url(self, url, language):
        """Return the I18nFolderFile instance from our files that match the given url and language"""
        if language:
            url = f"{language}/{url}"
        url = url.rstrip(".") or "."
        for file in self:
            if not file.is_documentation_page():
                continue
            if file.url == url:
                return file


class I18nFolderFile(File):
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
        self.abs_src_path = file_from.abs_src_path
        self.docs_dir = docs_dir
        self.page = file_from.page
        self.site_dir = site_dir
        self.src_path = file_from.src_path

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
        self.root_folder = Path(file_from.src_path).parts[0]

        # the name
        self.name = Path(self.initial_src_path).name
        self.dest_name = self.name

        if self.root_folder not in self.all_languages:
            # non localized root folder, file should be copied as-is
            # in the destination language path
            self.locale = self.dest_language
            self.dest_path = Path(self.locale) / Path(file_from.dest_path)
            self.abs_dest_path = Path(self.site_dir) / Path(self.dest_path)
        elif language == "":
            # default version file
            self.locale = self.default_language
            self.dest_path = Path(self.initial_dest_path).relative_to(self.locale)
            self.abs_dest_path = Path(self.site_dir) / Path(self.dest_path)
        else:
            # in localized folder file
            self.locale = self.dest_language
            self.dest_path = file_from.dest_path
            self.abs_dest_path = file_from.abs_dest_path

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
            f"I18nFolderFile(src_path='{self.src_path}', abs_src_path='{self.abs_src_path}',"
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
                url = "."
            else:
                url = dirname + "/"
        if self.dest_language:
            if url == ".":
                url += "/"
            # else:
            #     url = "/" + url
        return urlquote(url)

    def url_relative_to(self, other):
        """Return url for file relative to other i18n file."""
        return utils.get_relative_url(
            self.url,
            other.url
            if (isinstance(other, File) or isinstance(other, I18nFolderFile))
            else other,
        )


def on_files(self, files, config):
    """"""
    main_files = I18nFolderFiles([])
    main_files.default_locale = self.default_language
    main_files.locale = self.default_language
    for language in self.all_languages:
        self.i18n_files[language] = I18nFolderFiles([])
        self.i18n_files[language].default_locale = self.default_language
        self.i18n_files[language].locale = language

    for fileobj in files:

        file_locale = Path(fileobj.src_path).parts[0]

        if file_locale not in self.all_languages:
            if config["docs_dir"] in fileobj.abs_src_path:
                i18n_ffile = I18nFolderFile(
                    fileobj,
                    "",
                    all_languages=self.all_languages,
                    default_language=self.default_language,
                    docs_dir=config["docs_dir"],
                    site_dir=config["site_dir"],
                    use_directory_urls=config.get("use_directory_urls"),
                )
                main_files.append(i18n_ffile)
                for language in self.all_languages:
                    i18n_ffile = I18nFolderFile(
                        fileobj,
                        language,
                        all_languages=self.all_languages,
                        default_language=self.default_language,
                        docs_dir=config["docs_dir"],
                        site_dir=config["site_dir"],
                        use_directory_urls=config.get("use_directory_urls"),
                    )
                    self.i18n_files[language].append(i18n_ffile)
            else:
                # file is bundled by theme
                continue
        else:

            i18n_ffile = I18nFolderFile(
                fileobj,
                file_locale,
                all_languages=self.all_languages,
                default_language=self.default_language,
                docs_dir=config["docs_dir"],
                site_dir=config["site_dir"],
                use_directory_urls=config.get("use_directory_urls"),
            )
            self.i18n_files[file_locale].append(i18n_ffile)

            if file_locale == self.default_language:
                i18n_ffile = I18nFolderFile(
                    fileobj,
                    "",
                    all_languages=self.all_languages,
                    default_language=self.default_language,
                    docs_dir=config["docs_dir"],
                    site_dir=config["site_dir"],
                    use_directory_urls=config.get("use_directory_urls"),
                )
                main_files.append(i18n_ffile)

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

    return main_files


def on_nav(self, nav, config, files):
    """ """
    # translate default nav, see #113
    if self._maybe_translate_titles(self.default_language, nav):
        log.info(f"Translated default navigation to {self.default_language}")

    # check if the navigation is manually configured, see #145
    manual_nav = config.get("nav") is not None

    for language, lang_config in self.config["languages"].items():
        # skip nav generation for languages that we do not build
        if lang_config["build"] is False:
            continue
        if self.i18n_configs[language]["nav"]:
            self._fix_config_navigation(language, self.i18n_files[language])

        self.i18n_navs[language] = get_navigation(
            self.i18n_files[language], self.i18n_configs[language]
        )
        if manual_nav is False:
            if self.i18n_navs[language].items[0].children is None:
                # the structure is weird, say it but do not crash hard, see #152
                log.warning(
                    f"The structure of folder '{config['docs_dir']}/{language}' "
                    "does not look right, expect navigation or url inconsistencies"
                )
            else:
                # the expected folder structure starts with a [Section(title='LANG')]
                # so we render our navigation using it as a root
                self.i18n_navs[language].items = (
                    self.i18n_navs[language].items[0].children
                )
            for item in self.i18n_navs[language]:
                if config["use_directory_urls"] is True:
                    expected_url = f"{language}/"
                else:
                    expected_url = f"{language}/index.html"
                if item.is_page and item.url.strip() == expected_url:
                    self.i18n_navs[language].homepage = item
                    break
            else:
                raise Exception(f"could not find homepage Page(url='{expected_url}')")

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

        if language == self.default_language:
            for section in nav.items:
                if section.title == self.default_language.capitalize():
                    nav.items = section.children
                    break
            else:
                if manual_nav is False:
                    raise Exception(
                        f"could not find default version Section(title='{self.default_language.capitalize()}')"
                    )
            for item in nav:
                if config["use_directory_urls"] is True:
                    expected_url = ""
                else:
                    expected_url = "index.html"
                if item.is_page and item.url == expected_url:
                    nav.homepage = item
                    break
            else:
                raise Exception(
                    f"could not find default homepage Page(url='{expected_url}')"
                )

    return nav
