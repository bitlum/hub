package graphql

import (
	"bytes"
	"fmt"
	"text/template"
)

type graphiQLParams struct {
	QueryPath        string
	SubsPath         string
	PublicHost       string
	ListenPort       string
	SecureListenPort string
	FAQ              string
}

func renderGraphiQL(p graphiQLParams) ([]byte, error) {
	page, err := template.New("graphiQLPage").Parse(
		graphiQlTemplate)
	if err != nil {
		return nil, fmt.Errorf("unable to parse graphiQL"+
			" template: %v", err)
	}

	var buf bytes.Buffer

	err = page.Execute(&buf, p)
	if err != nil {
		return nil, fmt.Errorf("unable to execute graphiQL"+
			"template: %v", err)
	}

	return buf.Bytes(), nil
}

const graphiQlTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8" />
  <title>GraphiQL</title>
  <meta name="robots" content="noindex" />
  <style>
    html, body {
      height: 100%;
      margin: 0;
      overflow: hidden;
      width: 100%;
    }
  </style>
  <link href="//unpkg.com/graphiql@0.11.11/graphiql.css" rel="stylesheet" />
  <script src="//unpkg.com/react@15.6.1/dist/react.min.js"></script>
  <script src="//unpkg.com/react-dom@15.6.1/dist/react-dom.min.js"></script>
  <script src="//unpkg.com/graphiql@0.11.11/graphiql.min.js"></script>
  <script src="//cdn.jsdelivr.net/fetch/2.0.1/fetch.min.js"></script>
</head>
<body>
  <script>
    var wsProto = "ws"
	var wsPort = "{{ .ListenPort }}"
    if (location.protocol === 'https:') {
       wsProto = "wss"
       wsPort = "{{ .SecureListenPort }}"
    }

    if (wsPort !== "") {
       wsPort = ":"+wsPort
    }

    // Defines a GraphQL fetcher using the fetch API.
    function graphQLHttpFetcher(graphQLParams) {
      return fetch('{{ .QueryPath }}', {
        method: 'post',
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(graphQLParams),
        credentials: 'same-origin',
      }).then(function (response) {
        return response.text();
      }).then(function (responseBody) {
        try {
          return JSON.parse(responseBody);
        } catch (error) {
          return responseBody;
        }
      });}

    // Render <GraphiQL /> into the body.
    ReactDOM.render(
      React.createElement(GraphiQL, {
        fetcher: graphQLHttpFetcher,
        query: "{{ .FAQ }}",
      }),
      document.body
    );
  </script>
</body>
</html>`
