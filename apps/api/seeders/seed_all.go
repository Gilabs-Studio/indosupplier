package seeders

// SeedAll runs the baseline seeders needed for local development.
func SeedAll() error {
	if err := SeedUsers(); err != nil {
		return err
	}
	return SeedSystemAdmins()
}
