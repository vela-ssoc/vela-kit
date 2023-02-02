package node

func ID() string {
	return instance.id
}

func Prefix() string {
	if instance.prefix == "" {
		return "share"
	}

	return instance.prefix
}
