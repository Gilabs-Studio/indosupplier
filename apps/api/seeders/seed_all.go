package seeders

// SeedAll runs the baseline seeders needed for local development.
func SeedAll() error {
	if err := SeedUsers(); err != nil {
		return err
	}
	if err := SeedSystemAdmins(); err != nil {
		return err
	}
	if err := SeedTransactions(); err != nil {
		return err
	}
	return SeedRFQs()
}
