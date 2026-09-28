package entity

import "time"

type PaymentMethod struct {
	AccountName   *string
	DeletedAt     *time.Time
	QRImageURL    *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ID            string
	UserID        string
	MethodType    string
	ProviderName  string
	Visibility    string
	Status        string
	AccountNumber []byte
	IsDefault     bool
	IsVerified    bool
}
