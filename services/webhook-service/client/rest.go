package client

// RESTClient is a search service client that connects via REST.
type RESTClient struct {
	baseURL string
}

// New creates a new search service client.
func New(baseURL string) *RESTClient {
	return &RESTClient{
		baseURL: baseURL,
	}
}
