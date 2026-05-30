package seeders

// SeedAll runs minimal auth/user seeders for the cleaned baseline project.
func SeedAll() error {
	if err := SeedUsers(); err != nil {
		return err
	}
	return SeedSystemAdmins()
}
