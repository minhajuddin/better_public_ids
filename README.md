# Better Public IDs

We need better public IDs for our resources.
We need to take inspiration from https://github.com/minhajuddin/prefixed_uuids and make it better.
Any Public ID is defined a struct which describes a serialize and deserialize interface from a standard go interface.
We have a registry of Public IDs similar to prefixed_uuids where we register structs with a prefix
