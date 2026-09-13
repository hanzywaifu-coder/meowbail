package protocol

// Config returns the client configuration (read-only access for pkg domains).
// NOTE: returns the field directly — never call c.Config() here (recursion).
func (c *Client) Config() *Config {
	if c == nil {
		return nil
	}
	return c.config
}
