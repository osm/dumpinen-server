package main

const man = `File dump service.

# Dump "foo":
echo "foo" | curl --data-binary @- https://dumpinen.com
WGBtm-RLJkE

# Get dump:
curl https://dumpinen.com/WGBtm-RLJkE
foo

# Dump "foo" and delete it after ten minutes:
echo "foo" | curl --data-binary @- https://dumpinen.com?deleteAfter=10m
Tuo3wgzdBVX

# Dump "foo" and password protect it:
echo "foo" | curl --data-binary @- --user foo:bar https://dumpinen.com
NbbMcLcGcA9

# Get the password protected dump:
curl --user foo:bar https://dumpinen.com/NbbMcLcGcA9
foo

# Library/CLI code:
https://github.com/osm/dumpinen

# Server code:
https://github.com/osm/dumpinen-server`
