package node

func (nd *node) valid() error {
	if err := socket(&resolve); err != nil {
		return err
	}

	return nil
}

func (nd *node) Name() string {
	return nd.id
}
