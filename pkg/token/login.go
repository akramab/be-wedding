package token

import "net/http"

/*
IndexPage renders the html content for the index page.
*/
const IndexPage = `
<html>
	<head>
		<title>OAuth-2 Test</title>
	</head>
	<body>
		<h2>OAuth-2 Test</h2>
		<p>
			Login with the following,
		</p>
		<ul>
			<form action="/auth/google" method="post">
   				<button type="submit" name="" value="" class="btn-link">Login</button>
			</form>
		</ul>
	</body>
</html>
`

func HandleMain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(IndexPage))
}
