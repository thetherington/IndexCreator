## IndexCreator import

Subcommand used to import inSITE 'import files (tar.gz)

### Synopsis

This subcommand is used to import a inSITE index import tar.gz or a directory containing import files
	
Example Usage:
  ./IndexCreator import log-syslog-informational-2023.03.15.tar.gz
  ./IndexCreator import log-syslog-informational-directory

```
IndexCreator import [flags]
```

### Options

```
  -a, --app string   inSITE Elasticsearch Maintenance Program (default "mnt-1")
  -h, --help         help for import
```

### SEE ALSO

* [IndexCreator](IndexCreator.md)	 - Auto inSITE Index Creator and Importer tool

###### Auto generated by spf13/cobra on 21-Mar-2023
