package dto

type CreatePaymentMethodRequest struct {
	MethodType    string `json:"method_type" validate:"required,oneof=bank ewallet qris"`
	ProviderName  string `json:"provider_name" validate:"required,max=100"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	QRImageURL    string `json:"qr_image_url"`
	Visibility    string `json:"visibility" validate:"required,oneof=private group_members debtor_only"`
	IsDefault     bool   `json:"is_default"`
}
