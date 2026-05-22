package mapper

import (
	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
)

func ToCertificationResponse(cert *models.EmployeeCertification) dto.EmployeeCertificationResponse {
	resp := dto.EmployeeCertificationResponse{
		ID:                cert.ID,
		EmployeeID:        cert.EmployeeID,
		CertificateName:   cert.CertificateName,
		IssuedBy:          cert.IssuedBy,
		IssueDate:         cert.IssueDate.Format("2006-01-02"),
		CertificateFile:   cert.CertificateFile,
		CertificateNumber: cert.CertificateNumber,
		Description:       cert.Description,
		IsExpired:         cert.IsExpired(),
		DaysUntilExpiry:   cert.DaysUntilExpiry(),
		CreatedAt:         &cert.CreatedAt,
		UpdatedAt:         &cert.UpdatedAt,
	}

	if cert.ExpiryDate != nil {
		expiryStr := cert.ExpiryDate.Format("2006-01-02")
		resp.ExpiryDate = &expiryStr
	}

	return resp
}

func ToCertificationResponseList(certs []*models.EmployeeCertification) []dto.EmployeeCertificationResponse {
	responses := make([]dto.EmployeeCertificationResponse, len(certs))
	for i, cert := range certs {
		responses[i] = ToCertificationResponse(cert)
	}
	return responses
}

func ToCertificationBriefResponse(cert *models.EmployeeCertification) *dto.EmployeeCertificationBriefResponse {
	if cert == nil {
		return nil
	}

	resp := &dto.EmployeeCertificationBriefResponse{
		ID:                cert.ID,
		CertificateName:   cert.CertificateName,
		IssuedBy:          cert.IssuedBy,
		IssueDate:         cert.IssueDate.Format("2006-01-02"),
		CertificateNumber: cert.CertificateNumber,
		CertificateFile:   cert.CertificateFile,
		IsExpired:         cert.IsExpired(),
		DaysUntilExpiry:   cert.DaysUntilExpiry(),
	}

	if cert.ExpiryDate != nil {
		expiryStr := cert.ExpiryDate.Format("2006-01-02")
		resp.ExpiryDate = &expiryStr
	}

	return resp
}
