package main

type UI struct {
	IsMain    bool
	IsText    bool
	IsFile    bool
	IsAbout   bool
	IsError   bool
	ErrorText string
	Host      string
}

const uiHTML = `<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>
			{{if .IsMain}}dumpinen{{end}}
			{{if .IsText}}dumpinen - text{{end}}
			{{if .IsFile}}dumpinen - file{{end}}
			{{if .IsAbout}}dumpinen - about{{end}}
			{{if .IsError}}dumpinen - error{{end}}
		</title>
		<style type="text/css">
			a {
				color: #000000;
			}
			a:hover {
				text-decoration: none;
			}
			body {
				font-family: monospace;
			}
			label {
				margin-right: 10px;
				margin-bottom: 10px;
			}
			#navigation {
				margin-top: 0;
				padding-top: 0;
				margin-bottom: 30px;
			}
			.row {
				margin-top: 20px;
				margin-bottom: 20px;
			}
			.rowNarrow {
				margin: 0;
				padding: 0;
			}
			.active {
				font-weight: bold;
			}
			textarea {
				-webkit-box-sizing: border-box;
				-moz-box-sizing: border-box;
				box-sizing: border-box;
				width: 100%;
				height: 50vh;
			}
		</style>
	</head>
	<body>
		<div id="main">
			<div id="navigation">
				<nav>
					{{if .IsMain }}
					<a class="active" href="/">manual</a> |
					{{else}}
					<a href="/">manual</a> |
					{{end}}
					{{if .IsText }}
					<a class="active" href="/text">text</a> |
					{{else}}
					<a href="/text">text</a> |
					{{end}}
					{{if .IsFile }}
					<a class="active" hreF="/file">file</a> |
					{{else}}
					<a hreF="/file">file</a> |
					{{end}}
					{{if .IsAbout}}
					<a class="active" hreF="/about">about</a>
					{{else}}
					<a hreF="/about">about</a>
					{{end}}
				</nav>
			</div>
			<div id="content">
				{{if .IsError}}
				<div class="row">
					<div class="rowNarrow">
						<h2>Error</h2>
						<p>{{.ErrorText}}</p>
					</div>
				</div>
				{{end}}
				{{if .IsMain}}
				<div class="row">
					<div class="rowNarrow">
						# Dump "foo":
					</div>
					<div class="rowNarrow">
						echo "foo" | curl --data-binary @- {{.Host}}
					</div>
					<div class="rowNarrow">
						{{.Host}}/WGBtm-RLJkE
					</div>
				</div>
				<div class="row">
					<div class="rowNarrow">
						# Get dump:
					</div>
					<div class="rowNarrow">
						curl {{.Host}}/WGBtm-RLJkE
					</div>
					<div class="rowNarrow">
						foo
					</div>
				</div>
				<div class="row">
					<div class="rowNarrow">
						# Dump "foo" and delete it after ten minutes:
					</div>
					<div class="rowNarrow">
						echo "foo" | curl --data-binary @- {{.Host}}?deleteAfter=10m
					</div>
					<div class="rowNarrow">
						{{.Host}}/Tuo3wgzdBVX
					</div>
				</div>
				<div class="row">
					<div class="rowNarrow">
						# Dump "foo" and password protect it:
					</div>
					<div class="rowNarrow">
						echo "foo" | curl --data-binary @- --user foo:bar {{.Host}}
					</div>
					<div class="rowNarrow">
						{{.Host}}/NbbMcLcGcA9
					</div>
				</div>
				<div class="row">
					<div class="rowNarrow">
						# Get the password protected dump:
					</div>
					<div class="rowNarrow">
						curl --user foo:bar {{.Host}}/NbbMcLcGcA9
					</div>
					<div class="rowNarrow">
						foo
					</div>
				</div>
				{{end}}
				{{if .IsAbout}}
				<div class="row">
					<div class="rowNarrow">
						Dumpinen is a free text and file dumping service.
					</div>
				</div>
				<div class="row">
					<div class="rowNarrow">
						You are not allowed to store illegal content on dumpinen.
					</div>
				</div>
				<div class="row">
					<div class="rowNarrow">
						Source code:
					</div>
					<div class="rowNarrow">
						<a href="https://github.com/osm/dumpinen">
							https://github.com/osm/dumpinen
						</a>
					</div>
					<div class="rowNarrow">
						<a href="https://github.com/osm/dumpinen-server">
							https://github.com/osm/dumpinen-server
						</a>
					</div>
				</div>
				{{end}}
				{{if or .IsText .IsFile}}
				{{if .IsText }}
				<form action="/dump" method="post">
				{{end}}
				{{if .IsFile }}
				<form action="/dump" enctype="multipart/form-data" method="post">
				{{end}}
					{{if .IsText}}
					<div class="row">
						<div class="rowNarrow">
							<p>Enter the text to dump.</p>
						</div>
						<div class="rowNarrow">
							<textarea autofocus required name="text"></textarea>
						</div>
					</div>
					{{end}}
					{{if .IsFile}}
					<div class="row">
						<div class="rowNarrow">
							<p>Select the file to dump.</p>
						</div>
						<div class="rowNarrow">
							<input autofocus required type="file" name="file">
						</div>
					</div>
					{{end}}
					<div class="row">
						<div class="rowNarrow">
							<p>Dump lifetime.</p>
						</div>
						<div class="rowNarrow">
							<select name="deleteAfter">
								<option value="">Infinite</option>
								<option value="10m">Ten minutes</option>
								<option value="1h">One hour</option>
								<option value="24h">24 hours</option>
								<option value="168h">One week</option>
							</select>
						</div>
					</div>
					<div class="row">
						<div class="rowNarrow">
							<p>Fill in the username and password to protect your dump.</p>
						</div>
						<div class="rowNarrow">
							<label>Username:</label><input type="text" name="username">
						</div>
						<div class="rowNarrow">
							<label>Password:</label><input type="password" name="password">
						</div>
					</div>
					<div class="row">
						<button>Dump</dump>
					</div>
				</form>
				{{end}}
			</div>
		</div>
	</body>
</html>`
