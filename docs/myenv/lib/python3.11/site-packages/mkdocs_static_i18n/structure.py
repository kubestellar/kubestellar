import logging
from re import compile

from mkdocs.config.base import ValidationError
from mkdocs.config.config_options import Type

RE_LOCALE = compile(r"(^[a-z]{2}_[A-Z]{2}$)|(^[a-z]{2}$)")

log = logging.getLogger("mkdocs.plugins." + __name__)


class Locale(Type):
    """
    Locale Config Option

    Validate the locale config option against a given Python type.
    """

    def _validate_locale(self, value):
        # we allow the special case of the default language
        if value == "default":
            return
        if not RE_LOCALE.match(value):
            raise ValidationError(
                "Language code values must be either ISO-639-1 lower case "
                "or represented with they territory/region/county codes, "
                f"received '{value}' expected forms examples: 'en' or 'en_US'."
            )

    def _get_lang_dict_value(self, lang_key, lang_value):
        """
        Normalize the 'languages' option dict with backward compatibility.

        We support legacy:

            languages:
              en: English
              fr: Français

        And the new way:

            languages:
              en:
                name: English
                build: true
              fr:
                name: Français
                build: true
        """
        allowed_keys = set(["name", "link", "build", "site_name", "fixed_link"])
        lang_config = {
            "build": True,
            "link": f"./{lang_key}/" if lang_key != "default" else "./",
            "fixed_link": None,
            "name": lang_key,
            "site_name": None,
        }
        if isinstance(lang_value, str):
            lang_config["name"] = lang_value
        elif isinstance(lang_value, dict):
            unsupported_keys = set(lang_value.keys()).difference(allowed_keys)
            if unsupported_keys:
                log.warning(
                    f"'plugins.i18n.languages.{lang_key}' unsupported options: {','.join(unsupported_keys)}"
                )
            for key in lang_config:
                if key in lang_value:
                    lang_config[key] = lang_value[key]
        return lang_config

    def run_validation(self, value):
        value = super().run_validation(value)
        # default_language
        if isinstance(value, str):
            self._validate_locale(value)
        # languages
        if isinstance(value, dict):
            languages = {}
            for lang_key, lang_value in value.items():
                self._validate_locale(lang_key)
                languages[lang_key] = self._get_lang_dict_value(lang_key, lang_value)
            value = languages
        return value
