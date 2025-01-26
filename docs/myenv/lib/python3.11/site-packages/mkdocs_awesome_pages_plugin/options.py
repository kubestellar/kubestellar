class Options:
    def __init__(self, *, filename: str, collapse_single_pages: bool, strict: bool):
        self.filename = filename
        self.collapse_single_pages = collapse_single_pages
        self.strict = strict
