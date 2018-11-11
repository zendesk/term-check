package github

// ClientOption an option function to customize Client
type ClientOption func(*Client)

// WithPrivateKeyPath sets client's privateKeyPath
func WithPrivateKeyPath(privateKeyPath string) ClientOption {
	return func(c *Client) {
		c.privateKeyPath = privateKeyPath
	}
}

// WithAppID sets client's GitHub app id
func WithAppID(appID int) ClientOption {
	return func(c *Client) {
		c.appID = appID
	}
}
