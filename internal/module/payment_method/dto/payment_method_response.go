package dto

type PaymentMethodResponse struct {
	ID                  string  `json:"id"`
	MethodType          string  `json:"method_type"`
	ProviderName        string  `json:"provider_name"`
	AccountName         *string `json:"account_name,omitempty"`
	MaskedAccountNumber *string `json:"masked_account_number,omitempty"`
	QRImageURL          *string `json:"qr_image_url,omitempty"`
	Visibility          string  `json:"visibility"`
	IsDefault           bool    `json:"is_default"`
	IsVerified          bool    `json:"is_verified"`
}
