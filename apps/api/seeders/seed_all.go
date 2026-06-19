package seeders

// SeedAll runs the baseline seeders needed for local development.
func SeedAll() error {
	if err := SeedMonetization(); err != nil {
		return err
	}
	if err := SeedUsers(); err != nil {
		return err
	}
	if err := SeedSystemAdmins(); err != nil {
		return err
	}
	if err := SeedTransactions(); err != nil {
		return err
	}
	if err := SeedRFQs(); err != nil {
		return err
	}
	return SeedContent()
}
