"""
Writer for JavaScript objects
(not JSON syntax)

https://stackoverflow.com/questions/3975859/what-are-the-differences-between-json-and-javascript-object

In principle, JSON _could_ work.
However, we occasionally need to use identifiers, e.g. for function names;
the problem is that json.dumps() only produces strings, never literals.

"""
from typing import Any, Callable

from jsbeautifier import beautify


def dumps(obj:Any, 
          pretty:bool=True, 
          default:Callable[[Any], str]=None) -> str:
    """
    Serialize an object into a JavaScript object syntax. 

    Arguments:
    ---------
    - obj: the input object
    - pretty: indent the result
    - default: function to be used of no serialization is pssible

    Usage Note
    ----------
    When a string starts with a ^ character (caret),
    the rest of the string will be considered as an identifier.
    

    Returns
    -------
    A string containing a Javascript object
    """
    if isinstance(obj, bool):
        if obj:
            return "true"
        else:
            return "false"
    elif isinstance(obj, (int, float)):
        return obj
    elif isinstance(obj, str):
        # return strings
        if not obj:
            # empty
            r = '""'
        elif obj[0] == '^':
            # it's a literal
            r = obj[1:]
        else:
            # normal string
            r =  '"%s"' % obj
        return r
    elif isinstance(obj, dict):
        l = []
        for key, value in obj.items():
            s_value = dumps(value, False, default)
            # print("Key, value, s_value:", key, value, s_value)
            v = '%s: %s' % (key, s_value)
            l.append(v)
        #return indent('\n'.join(l), level)
        r = "{ %s }" % ', '.join(l)
        if pretty:
            r = beautify(r)
        # r = str(level) + r
        return r
    elif isinstance(obj, list):
        l = []
        for value in obj:
            s_value = dumps(value, False, default)
            l.append('%s' % s_value)
        r = "[ %s ]" % ',  '.join(l)
        if pretty:
            r = beautify(r)
        return r
    elif default:
        return default(obj)
    else:
        raise ValueError("Unrecognized type %s for Javascript" % 
                          type(obj).__name__)


if __name__ == "__main__":
    a = { "hello": "world", 
        "barbaz": "^bazbar",
        "foo": {"bar": 2},
        "bar": True}

    r = dumps(a)
    print("FIRST")
    print("- Python:", a)
    print("- Javascript:")
    print(r)

    assert '"bazbar"' not in r
    assert 'bazbar' in r

    print("SECOND")
    import yaml
    config_yaml = """
    plugins:
        - mermaid2:
            arguments:
                theme: neutral
                mermaid:
                    callback: ^myMermaidCallbackFunction

    extra_javascript:
        - https://unpkg.com/mermaid@8.5.2/dist/mermaid.min.js
        - js/extra.js
    """

    config = yaml.load(config_yaml, Loader=yaml.BaseLoader)
    print("- YAML:")
    print(config_yaml)
    print("- dumps:")
    print(dumps(config))
