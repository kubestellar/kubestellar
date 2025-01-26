import os
from typing import Optional


class cd:
    """Context manager for changing the current working directory"""

    def __init__(self, new_path):
        self.new_path = os.path.expanduser(new_path)

    def __enter__(self):
        self.saved_path = os.getcwd()
        os.chdir(self.new_path)

    def __exit__(self, exception_type, value, traceback):
        os.chdir(self.saved_path)


def dirname(path: Optional[str]) -> Optional[str]:
    """Returns the directory component of a pathname and None if the argument is None"""
    if path is not None:
        return os.path.dirname(path)


def basename(path: Optional[str]) -> Optional[str]:
    """Returns the final component of a pathname and None if the argument is None"""
    if path is not None:
        return os.path.basename(path)


def normpath(path: Optional[str]) -> Optional[str]:
    """Normalizes the path, returns None if the argument is None"""
    if path is not None:
        return os.path.normpath(path)


def join_paths(path1: Optional[str], path2: Optional[str]) -> Optional[str]:
    """Joins two paths if neither of them is None"""
    if path1 is not None and path2 is not None:
        return os.path.join(path1, path2)
