package main

import (
	"fmt"
)

func (a *app) getManText(host string) []byte {
	h := fmt.Sprintf("%s://%s", a.urlScheme, host)
	man := fmt.Sprintf(`File dump service.

# Dump "foo":
echo "foo" | curl --data-binary @- %s
%s/WGBtm-RLJkE

# Get dump:
curl %s/WGBtm-RLJkE
foo

# Dump "foo" and delete it after ten minutes:
echo "foo" | curl --data-binary @- %s?deleteAfter=10m
%s/Tuo3wgzdBVX

# Dump "foo" and password protect it:
echo "foo" | curl --data-binary @- --user foo:bar %s
%s/NbbMcLcGcA9

# Get the password protected dump:
curl --user foo:bar %s/NbbMcLcGcA9
foo

# Library/CLI code:
https://github.com/osm/dumpinen

# Server code:
https://github.com/osm/dumpinen-server`, h, h, h, h, h, h, h, h)

	return []byte(man + "\r\n")
}
