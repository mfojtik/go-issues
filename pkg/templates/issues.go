package templates

const IssuesTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
 <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
		<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.4/css/bootstrap.min.css">
		<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.4/css/bootstrap-theme.min.css">
		<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.4/js/bootstrap.min.js"></script>
</head>
<body>
	<div class="container">
	{{range $user, $issues := .Issues}}
		<h4 class="text-muted">@{{$user}}</h4>
		<table class="table table-hover table-condensed">
		{{range $issues}}
			<tr>
				<td><a href="https://github.com/openshift/origin/issues/{{.Number}}" target="_blank">#{{.Number}}</a></tds>
				<td><strong>{{.Title}}</strong></td>
				<td class="text-right">
					{{range .Labels}}
					<span class="label label-default" style="background-color:#{{.Color}}">{{.Name}}</span>
					{{end}}
				</td>
			</tr>
		{{end}}
		</table>
	{{end}}
	</div>
	<!-- jQuery (necessary for Bootstrap's JavaScript plugins) -->
  <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.11.2/jquery.min.js"></script>
  <!-- Include all compiled plugins (below), or include individual files as needed -->
  <script src="js/bootstrap.min.js"></script>
</body>
</html>
`
