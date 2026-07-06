package dto

type VerificationBusinessInfo struct {
	LegalName       string `json:"legalName"`
	NIBNumber       string `json:"nibNumber"`
	EstablishedDate string `json:"establishedDate"`
	Industry        string `json:"industry"`
	Phone           string `json:"phone"`
	Description     string `json:"description"`
	Address         string `json:"address"`
	NIBFileURL      string `json:"nibFileUrl"`
	NPWPNumber      string `json:"npwpNumber"`
	NPWPFileURL     string `json:"npwpFileUrl"`
	AktaFileURL     string `json:"aktaFileUrl"`
}

type VerificationStakeholderInfo struct {
	DirectorName string `json:"directorName"`
	DirectorNIK  string `json:"directorNik"`
}

type VerificationBankAccountInfo struct {
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
}

type VerificationResponse struct {
	BusinessInfo    VerificationBusinessInfo    `json:"businessInfo"`
	StakeholderInfo VerificationStakeholderInfo `json:"stakeholderInfo"`
	BankAccountInfo VerificationBankAccountInfo `json:"bankAccountInfo"`
	Status          string                      `json:"status"`
}

type UpdateVerificationRequest struct {
	BusinessInfo    *VerificationBusinessInfo    `json:"businessInfo"`
	StakeholderInfo *VerificationStakeholderInfo `json:"stakeholderInfo"`
	BankAccountInfo *VerificationBankAccountInfo `json:"bankAccountInfo"`
}
