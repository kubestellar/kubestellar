# Go template execute

This command simply parses a go template and some JSON and/or YAML
data and executes that template with that data.

This command requires exactly one positional argument, the pathname of
a file holding the input template.

This command accepts the following flags that supply data.  The data
starts out as an empty JSON object and is augmented by data read
according to these flags.

- `--data-json $spec` - parse a JSON object
- `--data-yaml $spec` - parse a YAML object
- `--data-field-json fieldName=$spec` - parses JSON data to add as the
  value of the given field. Can be given more than once.
- `--data-field-yaml fieldName=$spec` - parses YAML data to add as the
  value of the given field. Can be given more than once.

A `$spec` can be one of the following.
- `-`, meaning to read stdin.  Nothing good will come of using this
  twice in one command.
- `@` + pathname of a file whose contents are to be read.
- a literal to be read.

Following is an example invocation.

```shell
go-template-execute my-tmpl.txt --data-json '{"fieldx": 42}' --data-field-yaml Values=@valsfile.yaml
```

This will read data from the YAML object stored the file named
`valsfile.yaml` and add it as the value of the field named "Values".
Then it will parse the given JSON literal, adding to the top-level
object.  The template will be parsed from the file `my-tmpl.txt`.

